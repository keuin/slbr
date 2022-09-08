/*
This file contains task runner.
Task runner composes status monitor and stream downloader   concrete task config.
The config can be load from a config file.
*/
package recording

import (
	"bilibili-livestream-archiver/bilibili"
	"bilibili-livestream-archiver/common"
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"sync"
	"time"
)

// TaskResult represents an execution result of a task.
type TaskResult struct {
	Task  *TaskConfig
	Error error
}

// RunTask start a monitor&download task and
// put its execution result into a channel.
func RunTask(ctx context.Context, wg *sync.WaitGroup, task *TaskConfig) {
	defer wg.Done()
	err := doTask(ctx, task)
	logger := log.Default()
	if err != nil && !errors.Is(err, context.Canceled) {
		logger.Printf("A task stopped with an error (room %v): %v", task.RoomId, err)
	} else {
		logger.Printf("Task stopped (room %v): %v", task.RoomId, task.String())
	}
}

// doTask do the actual work, but returns synchronously.
func doTask(ctx context.Context, task *TaskConfig) error {
	logger := log.Default()
	bi := bilibili.NewBilibili()
	logger.Printf("Start task: room %v", task.RoomId)

	authKey, url, err := getStreamingServer(task, logger, bi)
	if err != nil {
		return err
	}

	// run live status watcher asynchronously
	logger.Println("Starting watcher...")
	chWatcherEvent := make(chan WatcherEvent)
	chWatcherDown := make(chan struct{})

	// start and recover watcher asynchronously
	// the watcher may also be stopped by the downloader goroutine
	watcherCtx, stopWatcher := context.WithCancel(ctx)
	defer stopWatcher()
	go watcherRecoverableLoop(watcherCtx, url, authKey, task, bi, chWatcherEvent, chWatcherDown)

	// The stream download goroutine may fail due to wrong watcher state.
	// But this is likely temporarily, so we should restart the downloader
	// until the state turns to closed.

	// We store the last modified live status
	// in case there is a false-positive duplicate.
	lastStatusIsLiving := false
	recorderCtx, stopRecorder := context.WithCancel(ctx)
	defer stopRecorder()
	for {
		select {
		case <-ctx.Done():
			logger.Printf("Task (room %v) is stopped.", task.RoomId)
			return nil
		case <-chWatcherDown:
			// watcher is down and unrecoverable, stop this task
			return fmt.Errorf("task (room %v) stopped: watcher is down and unrecoverable", task.RoomId)
		case ev := <-chWatcherEvent:
			switch ev {
			case WatcherLiveStart:
				if lastStatusIsLiving {
					logger.Println("Duplicate adjacent WatcherLiveStart event. Ignoring.")
					continue
				}
				go func() {
					cancelled := false
					var err2 error
					// restart recorder if interrupted by I/O errors
					for !cancelled {
						cancelled, err2 = record(recorderCtx, bi, task)
						if err2 == bilibili.ErrRoomIsClosed {
							sec := task.Watch.LiveInterruptedRestartSleepSeconds
							if sec == 0 {
								// default: 3s
								// TODO move this to default config value (not easily supported by viper)
								time.Sleep(3 * time.Second)
							}
							if sec > 0 {
								logger.Printf("Sleep for %vs before restart recording.", sec)
								time.Sleep(time.Duration(sec) * time.Second)
							}
						}
					}
					logger.Printf("Task is cancelled. Stop recording. (room %v)", task.RoomId)
				}()
				lastStatusIsLiving = true
			case WatcherLiveStop:
				lastStatusIsLiving = false
			}
		}
	}
}

// record. When cancelled, the caller should clean up immediately and stop the task.
func record(
	ctx context.Context,
	bi bilibili.Bilibili,
	task *TaskConfig,
) (cancelled bool, err error) {
	logger := log.Default()
	logger.Printf("INFO: Getting room profile...")

	profile, err := common.AutoRetry(
		ctx,
		func() (bilibili.RoomProfileResponse, error) {
			return bi.GetRoomProfile(task.RoomId)
		},
		task.Transport.MaxRetryTimes,
		time.Duration(task.Transport.RetryIntervalSeconds)*time.Second,
		logger,
	)
	if errors.Is(err, context.Canceled) {
		cancelled = true
		return
	}
	if err != nil {
		// still error, abort
		logger.Printf("ERROR: Cannot get room information: %v. Stopping current task.", err)
		cancelled = true
		return
	}

	urlInfo, err := common.AutoRetry(
		ctx,
		func() (bilibili.RoomUrlInfoResponse, error) {
			return bi.GetStreamingInfo(task.RoomId)
		},
		task.Transport.MaxRetryTimes,
		time.Duration(task.Transport.RetryIntervalSeconds)*time.Second,
		logger,
	)
	if errors.Is(err, context.Canceled) {
		cancelled = true
		return
	}
	if err != nil {
		logger.Printf("ERROR: Cannot get streaming info: %v", err)
		cancelled = true
		return
	}
	if len(urlInfo.Data.URLs) == 0 {
		j, err2 := json.Marshal(urlInfo)
		if err2 != nil {
			j = []byte("(not available)")
		}
		logger.Printf("ERROR: No stream returned from API. Response: %v", string(j))
		cancelled = true
		return
	}
	streamSource := urlInfo.Data.URLs[0]

	fileName := fmt.Sprintf(
		"%s.%s",
		GenerateFileName(profile.Data.Title, time.Now()),
		common.Errorable[string](common.GetFileExtensionFromUrl(streamSource.URL)).OrElse("flv"),
	)
	filePath := path.Join(task.Download.SaveDirectory, fileName)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		logger.Printf("ERROR: Cannot open file for writing: %v", err)
		cancelled = true
		return
	}
	defer func() { _ = file.Close() }()

	// buffered writer
	fWriter := bufio.NewWriterSize(file, task.Download.DiskWriteBufferBytes)
	defer func() {
		err := fWriter.Flush()
		if err != nil {
			logger := log.Default()
			logger.Printf("Failed to flush buffered file write data: %v", err)
		}
	}()
	logger.Printf("Write buffer size: %v byte", fWriter.Size())

	logger.Printf("Recording live stream to file \"%v\"...", filePath)
	err = bi.CopyLiveStream(ctx, task.RoomId, streamSource, fWriter)
	cancelled = err == nil || errors.Is(err, context.Canceled)
	if !cancelled {
		// real error happens
		logger.Printf("Error when copying live stream: %v", err)
	}
	return
}

// watcherRecoverableLoop run watcher forever until the context is cancelled.
func watcherRecoverableLoop(
	ctx context.Context,
	url string,
	authKey string,
	task *TaskConfig,
	bi bilibili.Bilibili,
	chWatcherEvent chan WatcherEvent,
	chWatcherDown chan<- struct{},
) {
	logger := log.Default()

	for {
		err, errReason := watch(
			ctx,
			url,
			authKey,
			task.RoomId,
			func() (bool, error) {
				resp, err := bi.GetRoomPlayInfo(task.RoomId)
				if err != nil {
					return false, err
				}
				if resp.Code != 0 {
					return false, fmt.Errorf("bilibili API error: %v", resp.Message)
				}
				return resp.Data.LiveStatus.IsStreaming(), nil
			},
			chWatcherEvent,
		)

		// the context is cancelled, stop watching
		if errors.Is(err, context.Canceled) {
			return
		}

		switch errReason {
		case ErrSuccess:
			// stop normally, the context is closed
			return
		case ErrProtocol:
			logger.Printf("FATAL: Watcher stopped due to an unrecoverable error: %v", err)
			// shutdown the whole task
			chWatcherDown <- struct{}{}
			return
		case ErrTransport:
			logger.Printf("ERROR: Watcher stopped due to an I/O error: %v", err)
			waitSeconds := task.Transport.RetryIntervalSeconds
			logger.Printf(
				"WARNING: Sleep for %v second(s) before restarting watcher.\n",
				waitSeconds,
			)
			time.Sleep(time.Duration(waitSeconds) * time.Second)
			logger.Printf("Retrying...")
		}
	}
}

func getStreamingServer(
	task *TaskConfig,
	logger *log.Logger,
	bi bilibili.Bilibili,
) (string, string, error) {
	logger.Println("Getting stream server info...")
	dmInfo, err := bi.GetDanmakuServerInfo(task.RoomId)
	if err != nil {
		return "", "", fmt.Errorf("failed to read stream server info: %w", err)
	}
	if len(dmInfo.Data.HostList) == 0 {
		return "", "", fmt.Errorf("no available stream server")
	}
	logger.Println("Success.")

	// get authkey and ws url
	authKey := dmInfo.Data.Token
	host := dmInfo.Data.HostList[0]
	url := fmt.Sprintf("wss://%s:%d/sub", host.Host, host.WssPort)
	return authKey, url, nil
}

func GenerateFileName(roomName string, t time.Time) string {
	ts := fmt.Sprintf(
		"%d-%02d-%02d-%02d-%02d-%02d",
		t.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
		t.Minute(),
		t.Second(),
	)
	return fmt.Sprintf("%s_%s", roomName, ts)
}
