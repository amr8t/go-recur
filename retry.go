package recur

import (
	"context"
	"time"
)

// Do executes any function with retry logic.
// The function must return an error as its last return value.
// All other return values are captured and returned.
//
// Example:
//   result, err := recur.Do(func() (string, error) {
//       return fetchData()
//   }).WithMaxAttempts(3).Run()
func Do(operation func() error) *SimpleRetrier {
	return &SimpleRetrier{
		operation:   operation,
		maxAttempts: 3,
		backoff:     Constant(100 * time.Millisecond),
		matcher:     MatchAny,
	}
}

// SimpleRetrier provides a simple retry interface for any function
type SimpleRetrier struct {
	operation   func() error
	maxAttempts int
	backoff     Backoff
	matcher     ErrorMatcher
	timeout     time.Duration
	hooks       []Hook
}

func (r *SimpleRetrier) WithMaxAttempts(max int) *SimpleRetrier {
	r.maxAttempts = max
	return r
}

func (r *SimpleRetrier) WithBackoff(b Backoff) *SimpleRetrier {
	r.backoff = b
	return r
}

func (r *SimpleRetrier) WithTimeout(d time.Duration) *SimpleRetrier {
	r.timeout = d
	return r
}

func (r *SimpleRetrier) RetryIf(matcher ErrorMatcher) *SimpleRetrier {
	r.matcher = matcher
	return r
}

func (r *SimpleRetrier) OnRetry(hook Hook) *SimpleRetrier {
	r.hooks = append(r.hooks, hook)
	return r
}

func (r *SimpleRetrier) Run() error {
	return r.RunContext(context.Background())
}

func (r *SimpleRetrier) RunContext(ctx context.Context) error {
	policies := []Policy{
		MaxAttempts(r.maxAttempts),
		WithBackoff(r.backoff),
		OnlyRetryIf(r.matcher),
	}
	
	if r.timeout > 0 {
		policies = append(policies, Timeout(r.timeout))
	}
	
	return execute(ctx, r.operation, policies, r.hooks)
}
