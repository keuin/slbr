package common

import (
	"context"
	"log"
	"time"
)

// AutoRetry retries the supplier automatically, with given time limit and interval.
// If maximum retry time limit is reached and the supplier still fails,
// the last error will be returned.
// If logger is not nil, retry information will be printed to it.
func AutoRetry[T any](
	ctx context.Context,
	supplier func() (T, error),
	maxRetryTimes int,
	retryInterval time.Duration,
	logger *log.Logger) (T, error) {
	var err error
	for i := 0; i < maxRetryTimes; i++ {
		ret, err := supplier()
		if err != nil {
			if logger != nil {
				logger.Printf("Try %v/%v (sleep %vs): %v\n",
					i, maxRetryTimes, retryInterval, err)
			}
			timer := time.NewTimer(retryInterval)
			select {
			case <-timer.C:
				// time to have the next try
				continue
			case <-ctx.Done():
				// context is cancelled
				var zero T
				return zero, ctx.Err()
			}
		}
		// success
		return ret, nil
	}
	if logger != nil {
		logger.Printf("Max retry times reached, but it still fails. Last error: %v", err)
	}
	var zero T
	return zero, err
}
