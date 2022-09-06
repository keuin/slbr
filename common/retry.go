package common

import (
	"log"
	"time"
)

// AutoRetry retries the supplier automatically, with given time limit and interval.
// If maximum retry time limit is reached and the supplier still fails,
// the last error will be returned.
// If logger is not nil, retry information will be printed to it.
func AutoRetry[T any](
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
			time.Sleep(retryInterval)
			continue
		}
		// success
		return ret, nil
	}
	if logger != nil {
		logger.Printf("Max retry times reached, but it still fails. Last error: %v", err)
	}
	return *new(T), err
}
