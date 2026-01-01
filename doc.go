// Package recur provides a simple, type-safe, decorator-style retry mechanism for Go functions.
//
// go-recur makes it easy to add retry logic to any Go function while preserving type safety
// through generics. It offers a fluent, decorator-style API similar to Python decorators.
//
// # Basic Usage
//
// Wrap any function with retry logic:
//
//	unreliableFunc := func() error {
//	    return doSomethingRisky()
//	}
//
//	retryFunc := recur.Func0(unreliableFunc).
//	    WithMaxAttempts(5).
//	    WithBackoff(recur.Exponential(100 * time.Millisecond)).
//	    Build()
//
//	err := retryFunc()
//
// # Type-Safe Return Values
//
// Functions with return values maintain their type signature:
//
//	fetchData := func() (string, error) {
//	    return callAPI()
//	}
//
//	retryFetch := recur.FuncR(fetchData).
//	    WithMaxAttempts(3).
//	    Build()
//
//	result, err := retryFetch() // Returns (string, error)
//
// # Functions with Arguments
//
// Support for functions with 0-2 arguments:
//
//	processItem := func(item string) error {
//	    return process(item)
//	}
//
//	retryProcess := recur.Func1(processItem).
//	    WithMaxAttempts(3).
//	    Build()
//
//	err := retryProcess("my-item")
//
// # Backoff Strategies
//
// Multiple backoff strategies are available:
//
//	// Constant delay
//	recur.Constant(500 * time.Millisecond)
//
//	// Exponential backoff
//	recur.Exponential(100 * time.Millisecond)
//
//	// Fibonacci sequence
//	recur.Fibonacci(50 * time.Millisecond)
//
//	// Linear increase
//	recur.Linear(100*time.Millisecond, 50*time.Millisecond)
//
//	// No delay
//	recur.NoDelay()
//
// # Conditional Retry
//
// Retry only on specific errors:
//
//	var ErrTemporary = errors.New("temporary error")
//
//	retryFunc := recur.Func0(fn).
//	    RetryIf(recur.MatchErrors(ErrTemporary)).
//	    Build()
//
// # Hooks for Monitoring
//
// Add hooks to monitor retry behavior:
//
//	retryFunc := recur.Func0(fn).
//	    OnRetry(func(ctx context.Context, attempt int, err error, elapsed time.Duration) {
//	        log.Printf("[RETRY] Attempt %d: %v", attempt, err)
//	    }).
//	    Build()
//
// # Composable Policies
//
// Create reusable policy combinations:
//
//	standardRetry := recur.CombinePolicies(
//	    recur.MaxAttempts(5),
//	    recur.WithBackoff(recur.Exponential(100 * time.Millisecond)),
//	    recur.Timeout(10 * time.Second),
//	)
//
//	retryFunc := recur.Func0(fn).WithPolicy(standardRetry).Build()
//
// # Context Support
//
// Full context and timeout support:
//
//	retryFunc := recur.Func0(fn).
//	    WithTimeout(5 * time.Second).
//	    BuildContext()
//
//	err := retryFunc(context.Background())
package recur
