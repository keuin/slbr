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
	"github.com/keuin/slbr/logging"
	"io"
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

const kSpecialExtName = "partial"

var errLiveEnded = common.NewRecoverableTaskError("live is ended", nil)

// runTaskWithAutoRestart
// start a monitor&download task.
// The task will be restarted infinitely until the context is closed,
// which means it will survive when the live is ended. (It always waits for the next start)
// During the process, its status may change.
// Note: this method is blocking.
func (t *RunningTask) runTaskWithAutoRestart() {
	t.status = StRunning
loop:
	for {
		switch err := tryRunTask(t); err.(type) {
		case nil:
			t.logger.Info("Task stopped: %v", t.String())
		case *common.RecoverableTaskError:
			if err != errLiveEnded {
				t.logger.Error("Temporary error: %v", err)
			}
			t.status = StRestarting
		default:
			if !errors.Is(err, context.Canceled) {
				t.logger.Error("Cannot recover from error: %v", err)
			}
			break loop
		}
	}
	t.logger.Info("Task stopped: %v", t.String())
}

// tryRunTask does the actual work. It will return when in the following cases:
// RecoverableError (end of live, IO error)
// UnrecoverableError (protocol error)
// context.Cancelled (the task is stopping)
func tryRunTask(t *RunningTask) error {
	netTypes := t.Transport.AllowedNetworkTypes
	t.logger.Info("Network types: %v", netTypes)
	bi := bilibili.NewBilibiliWithNetType(netTypes, t.logger)
	t.logger.Info("Start task: room %v", t.RoomId)

	t.logger.Info("Getting notification server info...")

	type dmServerInfo struct {
		AuthKey string
		DmUrl   string
	}

	dmInfo, err := AutoRetryWithTask(
		t, func() (info dmServerInfo, err error) {
			info.AuthKey, info.DmUrl, err = getDanmakuServer(&t.TaskConfig, bi)
			return
		},
	)
	if err != nil {
		return common.NewRecoverableTaskError("cannot get notification server info", err)
	}

	t.logger.Info("Success.")

	// wait for watcher goroutine
	wg := sync.WaitGroup{}
	defer wg.Wait()

	liveStatusChecker := func() (bool, error) {
		resp, err := bi.GetRoomPlayInfo(t.RoomId)
		if err != nil {
			return false, err
		}
		if resp.Code != 0 {
			return false, fmt.Errorf("bilibili API error: %v", resp.Message)
		}
		return resp.Data.LiveStatus.IsStreaming(), nil
	}

	// run live status watcher asynchronously
	t.logger.Info("Starting watcher...")

	wg.Add(1)
	chWatcherError := make(chan error)
	ctxWatcher, stopWatcher := context.WithCancel(t.ctx)
	defer stopWatcher()
	go func() {
		var err error
		defer wg.Done()
		run := true
	loop:
		for run {
			err = watch(
				ctxWatcher,
				t.TaskConfig,
				dmInfo.DmUrl,
				dmInfo.AuthKey,
				liveStatusChecker,
				t.logger,
			)
			switch err.(type) {
			case nil:
				// live is started, stop watcher loop and start the recorder
				break loop
			case *common.RecoverableTaskError:
				// if the watcher fails and recoverable, just try to recover
				// because the recorder has not started yet
				run = true
				t.logger.Error("Error occurred in live status watcher: %v", err)
				break
			case *common.UnrecoverableTaskError:
				// the watcher cannot recover, so the task should be stopped
				run = false
				t.logger.Error("Error occurred in live status watcher: %v", err)
			default:
				run = false
				// the task is being cancelled
				if errors.Is(err, context.Canceled) {
					break loop
				}
				// unknown error type, this should not happen
				t.logger.Error("Unexpected type of error in watcher: %v", err)
			}
			if run {
				t.logger.Info("Restarting watcher...")
			} else {
				t.logger.Error("Cannot restart watcher to recover from that error.")
			}
		}
		chWatcherError <- err
	}()

	// wait for live start signal or the watcher stops abnormally
	switch errWatcher := <-chWatcherError; errWatcher.(type) {
	case nil:
		// live is started, start recording
		// (now the watcher should have stopped)
		return func() error {
			var err error
			run := true
			for run {
				err = record(t.ctx, bi, &t.TaskConfig, t.logger)
				switch err.(type) {
				case nil:
					// live is ended
					t.logger.Info("The live is ended. Restarting current task...")
					return errLiveEnded
				case *common.RecoverableTaskError:
					// here we don't know if the live is ended, so we have to do a check
					t.logger.Warning("Recording is interrupted. Checking live status...")
					isLiving, err2 := AutoRetryWithTask(t, liveStatusChecker)
					if err2 != nil {
						return common.NewRecoverableTaskError(
							"when handling an error, another error occurred",
							fmt.Errorf("first: %v, second: %w", err, err2),
						)
					}
					if isLiving {
						t.logger.Info("This is a temporary error. Restarting recording...")
					} else {
						t.logger.Info("The live is ended. Restarting current task...")
						return errLiveEnded
					}
					run = isLiving
					break
				default:
					run = false
					if errors.Is(err, context.Canceled) {
						t.logger.Info("Recorder is stopped.")
					} else if errors.Is(err, io.EOF) {
						t.logger.Info("The live seems to be closed normally.")
					} else if errors.Is(err, io.ErrUnexpectedEOF) {
						t.logger.Warning("Reading is interrupted because of an unexpected EOF.")
					} else {
						t.logger.Error("Error when copying live stream: %v", err)
					}
					t.logger.Info("Stop recording.")
				}
			}
			return err
		}()
	case *common.UnrecoverableTaskError:
		// watcher is stopped and cannot restart
		return common.NewUnrecoverableTaskError("failed to watch live status", errWatcher)
	default:
		// watcher is cancelled, stop running the task
		if errors.Is(errWatcher, context.Canceled) {
			return errWatcher
		}
		// unexpected error, this is a programming error
		return common.NewUnrecoverableTaskError("unexpected error type", errWatcher)
	}
}

// record. When cancelled, the caller should clean up immediately and stop the task.
// Errors:
// RecoverableError
// UnrecoverableError
// context.Cancelled
// nil (live is ended normally)
func record(
	ctx context.Context,
	bi bilibili.Bilibili,
	task *TaskConfig,
	logger logging.Logger,
) error {
	logger.Info("Getting room profile...")

	profile, err := AutoRetryWithConfig(
		ctx,
		logger,
		task,
		func() (bilibili.RoomProfileResponse, error) {
			return bi.GetRoomProfile(task.RoomId)
		},
	)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		return common.NewRecoverableTaskError("failed to get living room information", err)
	}

	logger.Info("Getting stream url...")
	urlInfo, err := AutoRetryWithConfig(
		ctx,
		logger,
		task,
		func() (bilibili.RoomUrlInfoResponse, error) {
			return bi.GetStreamingInfo(task.RoomId)
		},
	)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		return common.NewRecoverableTaskError("failed to get live info", err)
	}
	if len(urlInfo.Data.URLs) == 0 {
		j, err2 := json.Marshal(urlInfo)
		if err2 != nil {
			j = []byte("(not available)")
		}
		logger.Error("No stream was provided. Response: %v", string(j))
		return common.NewUnrecoverableTaskError("invalid live info", fmt.Errorf("no stream provided"))
	}
	streamSource := urlInfo.Data.URLs[0]

	var extName string

	// the real extension name (without renaming)
	originalExtName := common.Errorable[string](common.GetFileExtensionFromUrl(streamSource.URL)).OrElse("flv")

	if task.Download.UseSpecialExtNameBeforeFinishing {
		extName = kSpecialExtName
	} else {
		extName = originalExtName
	}

	baseName := GenerateFileName(profile.Data.Title, time.Now())
	fileName := common.CombineFileName(baseName, extName)
	saveDir := task.Download.SaveDirectory
	filePath := path.Join(saveDir, fileName)

	var file *os.File

	// TODO refactor, move file close logic to CopyLiveStream
	// rename the extension name to originalExtName when finish writing
	defer func() {
		if file == nil {
			// the file is not created
			return
		}
		if extName == originalExtName {
			return
		}
		from := filePath
		to := path.Join(saveDir, common.CombineFileName(baseName, originalExtName))
		err := os.Rename(from, to)
		if err != nil {
			logger.Error("Cannot rename %v to %v: %v", from, to, err)
			return
		}
		logger.Info("Rename file \"%s\" to \"%s\".", from, to)
	}()
	defer func() { _ = file.Close() }()

	writeBufferSize := task.Download.DiskWriteBufferBytes
	logger.Info("Write buffer size: %v byte", writeBufferSize)
	err = bi.CopyLiveStream(ctx, task.RoomId, streamSource, func() (f *os.File, e error) {
		f, e = os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if e != nil {
			file = f
		}
		logger.Info("Recording live stream to file \"%v\"...", filePath)
		return
	}, writeBufferSize)
	if common.IsErrorOfType(err, &common.UnrecoverableTaskError{}) {
		logger.Error("Cannot record: %v", err)
		return err
	} else if errors.Is(err, context.Canceled) || err == nil {
		return err
	}
	logger.Error("Error when copying live stream: %v", err)
	return common.NewRecoverableTaskError("stream copy was unexpectedly interrupted", err)
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
