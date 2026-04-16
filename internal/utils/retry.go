package utils

import (
	"context"
	"log/slog"
	"math/rand"
	"time"
)

type RetryEffector = func(ctx context.Context) (bool, error)

func Retry(ctx context.Context, effector RetryEffector, timeout time.Duration, maxRetries int, maxRetryTime time.Duration) error {
	var (
		retry bool
		err   error
	)
	for i := range maxRetries {
		retry, err = func() (bool, error) {
			dctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			return effector(dctx)
		}()
		if err == nil {
			return nil
		}
		if !retry {
			return err
		}
		slog.Info("retry failed", "attempt", i+1, "error", err)
		jitter := time.Duration(rand.Intn(100)) * time.Millisecond
		currentWaitTime := (1<<i)*time.Second + jitter
		waitTime := min(currentWaitTime, maxRetryTime)
		time.Sleep(waitTime)
	}
	return err
}
