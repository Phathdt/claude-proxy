package retry

import (
	"context"
	"fmt"
	"math"
	"time"
)

// Config holds retry configuration
type Config struct {
	MaxRetries int
	RetryDelay time.Duration
}

// Do executes a function with retry logic and exponential backoff
func Do(ctx context.Context, cfg Config, fn func() error) error {
	var lastErr error
	
	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		// Execute the function
		if err := fn(); err != nil {
			lastErr = err
			
			// If this was the last attempt, return the error
			if attempt == cfg.MaxRetries {
				return fmt.Errorf("all retry attempts failed: %w", lastErr)
			}
			
			// Calculate exponential backoff delay
			delay := time.Duration(math.Pow(2, float64(attempt))) * cfg.RetryDelay
			
			// Wait before retrying (with context cancellation support)
			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				return fmt.Errorf("retry cancelled: %w", ctx.Err())
			}
		}
		
		// Success
		return nil
	}
	
	return lastErr
}

// DoWithResult executes a function with retry logic and returns the result
func DoWithResult[T any](ctx context.Context, cfg Config, fn func() (T, error)) (T, error) {
	var result T
	var lastErr error
	
	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		// Execute the function
		res, err := fn()
		if err != nil {
			lastErr = err
			
			// If this was the last attempt, return the error
			if attempt == cfg.MaxRetries {
				return result, fmt.Errorf("all retry attempts failed: %w", lastErr)
			}
			
			// Calculate exponential backoff delay
			delay := time.Duration(math.Pow(2, float64(attempt))) * cfg.RetryDelay
			
			// Wait before retrying (with context cancellation support)
			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				return result, fmt.Errorf("retry cancelled: %w", ctx.Err())
			}
		}
		
		// Success
		return res, nil
	}
	
	return result, lastErr
}
