package recording

/*
In this file we implement task lifecycle management.
Concrete task works are done in the `runner.go` file.
*/

import (
	"context"
	"fmt"
	"github.com/keuin/slbr/common/retry"
	"github.com/keuin/slbr/logging"
	"time"
)

type TaskStatus int

const (
	StNotStarted TaskStatus = iota
	StRunning
	StRestarting
	StStopped
)

var (
	ErrTaskIsAlreadyStarted = fmt.Errorf("task is already started")
	ErrTaskIsStopped        = fmt.Errorf("restarting a stopped task is not allowed")
)

// RunningTask is an augmented TaskConfig struct
// that contains volatile runtime information.
type RunningTask struct {
	TaskConfig
	// ctx: the biggest context this task uses. It may create children contexts.
	ctx context.Context
	// status: running status
	status TaskStatus
	// hookStarted: called asynchronously when the task is started. This won't be called when restarting.
	hookStarted func()
	// hookStopped: called asynchronously when the task is stopped. This won't be called when restarting.
	hookStopped func()
	// logger: where to print logs
	logger logging.Logger
}

func NewRunningTask(
	config TaskConfig,
	ctx context.Context,
	hookStarted func(),
	hookStopped func(),
	logger logging.Logger,
) RunningTask {
	return RunningTask{
		TaskConfig:  config,
		ctx:         ctx,
		status:      StNotStarted,
		hookStarted: hookStarted,
		hookStopped: hookStopped,
		logger:      logger,
	}
}

func (t *RunningTask) StartTask() error {
	st := t.status
	switch st {
	case StNotStarted:
		// TODO real start
		go func() {
			defer func() { t.status = StStopped }()
			t.hookStarted()
			defer t.hookStopped()
			// do the task
			t.runTaskWithAutoRestart()
		}()
		return nil
	case StRunning:
		return ErrTaskIsAlreadyStarted
	case StRestarting:
		return ErrTaskIsAlreadyStarted
	case StStopped:
		// we don't allow starting a stopped task
		// because some state needs to be reset
		// just create a new task and run
		return ErrTaskIsStopped
	}
	panic(fmt.Errorf("invalid task status: %v", st))
}

func AutoRetryWithTask[T any](
	t *RunningTask,
	supplier func() (T, error),
) (T, error) {
	return retry.AutoRetry[T](
		t.ctx,
		supplier,
		t.Transport.MaxRetryTimes,
		time.Duration(t.Transport.RetryIntervalSeconds)*time.Second,
		&t.logger,
	)
}

func AutoRetryWithConfig[T any](
	ctx context.Context,
	logger logging.Logger,
	t *TaskConfig,
	supplier func() (T, error),
) (T, error) {
	return retry.AutoRetry[T](
		ctx,
		supplier,
		t.Transport.MaxRetryTimes,
		time.Duration(t.Transport.RetryIntervalSeconds)*time.Second,
		&logger,
	)
}
