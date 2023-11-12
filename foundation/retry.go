package foundation

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff"
)

func DoRetry(ctx context.Context, exec func(timeoutCtx context.Context) error, maxTimes int, execTimeout time.Duration) error {
	logger := Logger()

	backoffConfig := backoff.WithMaxRetries(backoff.NewConstantBackOff(100*time.Millisecond), uint64(maxTimes))

	err := backoff.Retry(func() error {
		if ctx.Err() == context.DeadlineExceeded || ctx.Err() == context.Canceled {
			return nil
		}

		timeoutCtx, timeoutCancel := context.WithTimeout(ctx, execTimeout)
		insideErr := exec(timeoutCtx)
		timeoutCancel()

		return insideErr
	}, backoffConfig)

	if err != nil {
		err = fmt.Errorf("failed retry eventually: %v", err)
		logger.Error(err)
		return err
	}

	return nil
}
