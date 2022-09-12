/*
This file contains task runner.
Task runner composes status monitor and stream downloader   concrete task config.
The config can be load from a config file.
*/
package recording

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/keuin/slbr/bilibili"
	"github.com/keuin/slbr/common"
	"io"
	"os"
	"path"
	"time"
)

// TaskResult represents an execution result of a task.
type TaskResult struct {
	Task  *TaskConfig
	Error error
}

const kReadChunkSize = 128 * 1024
const kSpecialExtName = "partial"

// runTaskWithAutoRestart
// start a monitor&download task.
// The task will be restarted infinitely until the context is closed,
// which means it will survive when the live is ended. (It always waits for the next start)
// During the process, its status may change.
// Note: this method is blocking.
func (t *RunningTask) runTaskWithAutoRestart() error {
	for {
		t.status = StRunning
		err := tryRunTask(t)
		if errors.Is(err, bilibili.ErrRoomIsClosed) {
			t.status = StRestarting
			t.logger.Info("Restarting task...")
			continue
		} else if err != nil && !errors.Is(err, context.Canceled) {
			t.logger.Error("Task stopped with an error: %v", err)
			return fmt.Errorf("task stopped: %v", err)
		} else {
			t.logger.Info("Task stopped: %v", t.String())
			return nil
		}
	}
}

// tryRunTask does the actual work. It will return when in the following cases:
//   - the task context is cancelled
//   - the task is restarting (e.g. because of the end of live)
//   - some unrecoverable error happens (e.g. a protocol error caused by a bilibili protocol update)
func tryRunTask(t *RunningTask) error {
	netTypes := t.Transport.AllowedNetworkTypes
	t.logger.Info("Network types: %v", netTypes)
	bi := bilibili.NewBilibiliWithNetType(netTypes, t.logger)
	t.logger.Info("Start task: room %v", t.RoomId)

	t.logger.Info("Getting notification server info...")
	authKey, dmUrl, err := getDanmakuServer(&t.TaskConfig, bi)
	if err != nil {
		return err
	}
	t.logger.Info("Success.")

	// run live status watcher asynchronously
	t.logger.Info("Starting watcher...")
	chWatcherEvent := make(chan WatcherEvent)
	chWatcherDown := make(chan struct{})

	// start and recover watcher asynchronously
	// the watcher may also be stopped by the downloader goroutine
	watcherCtx, stopWatcher := context.WithCancel(t.ctx)
	defer stopWatcher()
	go watcherRecoverableLoop(
		watcherCtx,
		dmUrl,
		authKey,
		t,
		bi,
		chWatcherEvent,
		chWatcherDown,
	)

	// The stream download goroutine may fail due to wrong watcher state.
	// But this is likely temporarily, so we should restart the downloader
	// until the state turns to closed.

	recorderCtx, stopRecorder := context.WithCancel(t.ctx)
	defer stopRecorder()
	for {
		select {
		case <-t.ctx.Done():
			t.logger.Info("Task is stopped.")
			return nil
		case <-chWatcherDown:
			// watcher is down and unrecoverable, stop this task
			return fmt.Errorf("task (room %v) stopped: watcher is down and unrecoverable", t.RoomId)
		case ev := <-chWatcherEvent:
			switch ev {
			case WatcherLiveStart:
				cancelled := false
				var err2 error
				// restart recorder if interrupted by I/O errors
				for !cancelled {
					cancelled, err2 = record(recorderCtx, bi, t)
					if errors.Is(err2, io.ErrUnexpectedEOF) {
						t.logger.Warning("Reading is interrupted because of an unexpected EOF. Retrying...")
						cancelled = false
					}
				}
				t.logger.Error("Error when copying live stream: %v", err2)
				if err2 == nil || errors.Is(err2, bilibili.ErrRoomIsClosed) || errors.Is(err2, io.EOF) {
					t.logger.Info("Live is ended. Stop recording.")
					return bilibili.ErrRoomIsClosed
				}
				t.logger.Error("Cannot recover from unexpected error: %v", err2)
				t.logger.Info("Task is cancelled. Stop recording.")
			case WatcherLiveStop:
				// once the live is ended, the watcher will no longer receive live start event
				// we have to restart the watcher
				return bilibili.ErrRoomIsClosed
			}
		}
	}
}

// record. When cancelled, the caller should clean up immediately and stop the task.
func record(
	ctx context.Context,
	bi bilibili.Bilibili,
	task *RunningTask,
) (cancelled bool, err error) {
	task.logger.Info("Getting room profile...")

	profile, err := common.AutoRetry(
		ctx,
		func() (bilibili.RoomProfileResponse, error) {
			return bi.GetRoomProfile(task.RoomId)
		},
		task.Transport.MaxRetryTimes,
		time.Duration(task.Transport.RetryIntervalSeconds)*time.Second,
		&task.logger,
	)
	if errors.Is(err, context.Canceled) {
		cancelled = true
		return
	}
	if err != nil {
		// still error, abort
		task.logger.Error("Cannot get room information: %v. Stopping current task.", err)
		cancelled = true
		return
	}

	task.logger.Info("Getting stream url...")
	urlInfo, err := common.AutoRetry(
		ctx,
		func() (bilibili.RoomUrlInfoResponse, error) {
			return bi.GetStreamingInfo(task.RoomId)
		},
		task.Transport.MaxRetryTimes,
		time.Duration(task.Transport.RetryIntervalSeconds)*time.Second,
		&task.logger,
	)
	if errors.Is(err, context.Canceled) {
		cancelled = true
		return
	}
	if err != nil {
		task.logger.Error("Cannot get streaming info: %v", err)
		cancelled = true
		return
	}
	if len(urlInfo.Data.URLs) == 0 {
		j, err2 := json.Marshal(urlInfo)
		if err2 != nil {
			j = []byte("(not available)")
		}
		task.logger.Error("No stream returned from API. Response: %v", string(j))
		cancelled = true
		return
	}
	streamSource := urlInfo.Data.URLs[0]

	var extName string

	// the real extension name (without renaming)
	originalExtName := common.Errorable[string](common.GetFileExtensionFromUrl(streamSource.URL)).OrElse("flv")

	if task.TaskConfig.Download.UseSpecialExtNameBeforeFinishing {
		extName = kSpecialExtName
	} else {
		extName = originalExtName
	}

	baseName := GenerateFileName(profile.Data.Title, time.Now())
	fileName := common.CombineFileName(baseName, extName)
	saveDir := task.Download.SaveDirectory
	filePath := path.Join(saveDir, fileName)

	// rename the extension name to originalExtName when finish writing
	defer func() {
		if extName == originalExtName {
			return
		}
		from := filePath
		to := path.Join(saveDir, common.CombineFileName(baseName, originalExtName))
		err := os.Rename(from, to)
		if err != nil {
			task.logger.Error("Cannot rename %v to %v: %v", from, to, err)
			return
		}
		task.logger.Info("Rename file \"%s\" to \"%s\".", from, to)
	}()

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		task.logger.Error("Cannot open file for writing: %v", err)
		cancelled = true
		return
	}
	defer func() { _ = file.Close() }()

	writeBufferSize := task.Download.DiskWriteBufferBytes
	if writeBufferSize < kReadChunkSize {
		writeBufferSize = kReadChunkSize
	}
	if mod := writeBufferSize % kReadChunkSize; mod != 0 {
		writeBufferSize += kReadChunkSize - mod
	}
	writeBuffer := make([]byte, writeBufferSize)
	task.logger.Info("Write buffer size: %v byte", writeBufferSize)
	task.logger.Info("Recording live stream to file \"%v\"...", filePath)
	err = bi.CopyLiveStream(ctx, task.RoomId, streamSource, file, writeBuffer, kReadChunkSize)
	cancelled = err == nil || errors.Is(err, context.Canceled)
	if !cancelled {
		// real error happens
		task.logger.Error("Error when copying live stream: %v", err)
	}
	return
}

// watcherRecoverableLoop run watcher forever until the context is cancelled.
func watcherRecoverableLoop(
	ctx context.Context,
	url string,
	authKey string,
	task *RunningTask,
	bi bilibili.Bilibili,
	chWatcherEvent chan<- WatcherEvent,
	chWatcherDown chan<- struct{},
) {
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
			task.logger,
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
			task.logger.Fatal("Watcher stopped due to an unrecoverable error: %v", err)
			// shutdown the whole task
			chWatcherDown <- struct{}{}
			return
		case ErrTransport:
			task.logger.Error("Watcher stopped due to an I/O error: %v", err)
			waitSeconds := task.Transport.RetryIntervalSeconds
			task.logger.Warning(
				"Sleep for %v second(s) before restarting watcher.\n",
				waitSeconds,
			)
			time.Sleep(time.Duration(waitSeconds) * time.Second)
			task.logger.Info("Retrying...")
		}
	}
}

func getDanmakuServer(
	task *TaskConfig,
	bi bilibili.Bilibili,
) (string, string, error) {
	dmInfo, err := bi.GetDanmakuServerInfo(task.RoomId)
	if err != nil {
		return "", "", fmt.Errorf("failed to read stream server info: %w", err)
	}
	if len(dmInfo.Data.HostList) == 0 {
		return "", "", fmt.Errorf("no available stream server")
	}

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
