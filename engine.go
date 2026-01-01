package recur

import (
	"context"
	"time"
)

// Hook is a callback function that is called during retry lifecycle
type Hook func(ctx context.Context, attempt int, err error, elapsed time.Duration)

// execute is the core retry engine that runs the operation with all policies applied
func execute(ctx context.Context, operation func() error, policies []Policy, hooks []Hook) error {
	// Apply default configuration
	config := &executionConfig{
		maxAttempts: 3,
		backoff:     Constant(100 * time.Millisecond),
		matcher:     MatchAny,
		timeout:     0,
	}

	// Apply all policies to build final configuration
	for _, policy := range policies {
		policy.apply(config)
	}

	// Create execution context with timeout if specified
	execCtx := ctx
	var cancel context.CancelFunc
	if config.timeout > 0 {
		execCtx, cancel = context.WithTimeout(ctx, config.timeout)
		defer cancel()
	}

	var lastErr error
	startTime := time.Now()

	for attempt := 1; attempt <= config.maxAttempts; attempt++ {
		// Check for context cancellation
		select {
		case <-execCtx.Done():
			return &MaxAttemptsExceededError{
				Attempts: attempt - 1,
				LastErr:  lastErr,
			}
		default:
		}

		// Execute hooks before attempt
		elapsed := time.Since(startTime)
		for _, hook := range hooks {
			hook(execCtx, attempt, lastErr, elapsed)
		}

		// Execute the operation
		err := operation()

		// Success - we're done!
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if we should retry this error
		if !config.matcher(err) {
			return err
		}

		// Last attempt - don't wait, just return
		if attempt >= config.maxAttempts {
			break
		}

		// Calculate backoff delay
		delay := config.backoff.Next(attempt)

		// Wait for backoff or context cancellation
		select {
		case <-time.After(delay):
			// Continue to next attempt
		case <-execCtx.Done():
			return &MaxAttemptsExceededError{
				Attempts: attempt,
				LastErr:  lastErr,
			}
		}
	}

	// All attempts exhausted
	return &MaxAttemptsExceededError{
		Attempts: config.maxAttempts,
		LastErr:  lastErr,
	}
}

// executionConfig holds the runtime configuration for retry execution
type executionConfig struct {
	maxAttempts int
	backoff     Backoff
	matcher     ErrorMatcher
	timeout     time.Duration
}

// Policy is a function that modifies execution configuration
type Policy func(*executionConfig)

func (p Policy) apply(config *executionConfig) {
	p(config)
}

// MaxAttempts creates a policy that sets maximum retry attempts
func MaxAttempts(max int) Policy {
	return func(config *executionConfig) {
		config.maxAttempts = max
	}
}

// Timeout creates a policy that sets overall timeout for all attempts
func Timeout(d time.Duration) Policy {
	return func(config *executionConfig) {
		config.timeout = d
	}
}

// WithBackoff creates a policy that sets the backoff strategy
func WithBackoff(b Backoff) Policy {
	return func(config *executionConfig) {
		config.backoff = b
	}
}

// OnlyRetryIf creates a policy that sets the error matcher
func OnlyRetryIf(matcher ErrorMatcher) Policy {
	return func(config *executionConfig) {
		config.matcher = matcher
	}
}

// CombinePolicies merges multiple policies into one
func CombinePolicies(policies ...Policy) Policy {
	return func(config *executionConfig) {
		for _, p := range policies {
			p.apply(config)
		}
	}
}
