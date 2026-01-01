package recur

import (
	"errors"
	"fmt"
)

// MaxAttemptsExceededError is returned when all retry attempts have been exhausted
type MaxAttemptsExceededError struct {
	Attempts int
	LastErr  error
}

func (e *MaxAttemptsExceededError) Error() string {
	return fmt.Sprintf("max attempts (%d) exceeded: %v", e.Attempts, e.LastErr)
}

func (e *MaxAttemptsExceededError) Unwrap() error {
	return e.LastErr
}

// IsMaxAttemptsExceeded checks if the error is a MaxAttemptsExceededError
func IsMaxAttemptsExceeded(err error) bool {
	var e *MaxAttemptsExceededError
	return errors.As(err, &e)
}

// ErrorMatcher is a function that determines if an error should trigger a retry
type ErrorMatcher func(error) bool

// MatchAny matches any non-nil error (default behavior)
func MatchAny(err error) bool {
	return err != nil
}

// MatchNone matches no errors (effectively disables retry)
func MatchNone(err error) bool {
	return false
}

// MatchErrors creates a matcher that retries only for specific error values
func MatchErrors(targets ...error) ErrorMatcher {
	return func(err error) bool {
		for _, target := range targets {
			if errors.Is(err, target) {
				return true
			}
		}
		return false
	}
}

// MatchTypes creates a matcher that retries only for specific error types.
// Note: Due to Go's type system limitations, this uses errors.Is for comparison.
// For custom type matching, use MatchFunc with type assertions.
//
// Example:
//
//	var netErr *net.OpError
//	matcher := MatchFunc(func(err error) bool {
//	    return errors.As(err, &netErr)
//	})
func MatchTypes(targets ...error) ErrorMatcher {
	return func(err error) bool {
		for _, target := range targets {
			// Use Is instead of As since we can't properly use As with interface{} targets
			if errors.Is(err, target) {
				return true
			}
		}
		return false
	}
}

// MatchFunc creates a matcher from a custom function
func MatchFunc(fn func(error) bool) ErrorMatcher {
	return fn
}

// Not inverts an error matcher
func Not(matcher ErrorMatcher) ErrorMatcher {
	return func(err error) bool {
		return !matcher(err)
	}
}

// And combines multiple matchers with AND logic
func And(matchers ...ErrorMatcher) ErrorMatcher {
	return func(err error) bool {
		for _, matcher := range matchers {
			if !matcher(err) {
				return false
			}
		}
		return true
	}
}

// Or combines multiple matchers with OR logic
func Or(matchers ...ErrorMatcher) ErrorMatcher {
	return func(err error) bool {
		for _, matcher := range matchers {
			if matcher(err) {
				return true
			}
		}
		return false
	}
}
