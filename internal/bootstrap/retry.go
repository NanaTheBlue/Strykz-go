package bootstrap

import (
	"context"
	"log"
	"time"
)

func retryWithBackoff(
	ctx context.Context,
	attempts int,
	initialDelay time.Duration,
	fn func() error,
) error {
	delay := initialDelay

	var err error
	for i := 1; i <= attempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		log.Printf(
			"attempt %d/%d failed: %v — retrying in %v",
			i,
			attempts,
			err,
			delay,
		)

		if i == attempts {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):

			delay *= 2

			if delay > 5*time.Second {
				delay = 5 * time.Second
			}
		}
	}

	return err
}
