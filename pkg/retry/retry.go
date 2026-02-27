package retry

import (
	"context"
	"fmt"
	"time"
)

// Do executes a function with exponential backoff.
func Do(ctx context.Context, maxRetries int, operation func() error) error {
	var err error
	baseDelay := 200 * time.Millisecond

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err = operation(); err == nil {
			return nil // Success
		}

		if attempt == maxRetries {
			break // Don't sleep on the last attempt
		}

		delay := baseDelay * (1 << attempt) // Exponential: 200ms, 400ms, 800ms...

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled during retry: %w", ctx.Err())
		}
	}
	return fmt.Errorf("max retries reached: %w", err)
}
