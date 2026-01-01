# go-recur

A simple, fluent retry library for Go with composable policies.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![CI](https://github.com/amr8t/go-recur/actions/workflows/ci.yml/badge.svg)](https://github.com/amr8t/go-recur/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/amr8t/go-recur)](https://goreportcard.com/report/github.com/amr8t/go-recur)
[![codecov](https://codecov.io/gh/amr8t/go-recur/branch/main/graph/badge.svg)](https://codecov.io/gh/amr8t/go-recur)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Why go-recur?

Unlike other retry libraries, go-recur offers:

- **Fluent Builder API** - Chain configuration methods naturally
- **Composable Policies** - Create and reuse retry strategies across your codebase
- **Rich Error Matching** - Built-in combinators (And, Or, Not) for complex retry logic
- **Cleaner Syntax** - Less visual noise compared to function options
- **Modern Codebase** - Written for Go 1.21+ with clean, maintainable code

### Comparison

```go
// Other libraries - function options
retry.Do(
    func() error { return operation() },
    retry.Attempts(3),
    retry.Delay(time.Second),
    retry.OnRetry(logFunc),
)

// go-recur - fluent builder
recur.Do(func() error {
    return operation()
}).
    WithMaxAttempts(3).
    WithBackoff(recur.Constant(time.Second)).
    OnRetry(logFunc).
    Run()
```

The builder pattern makes configuration more readable, especially with multiple options.

## Key Differentiators

### 1. Composable Policies

Define retry strategies once and reuse them everywhere:

```go
// Define once
standardRetry := recur.CombinePolicies(
    recur.MaxAttempts(5),
    recur.WithBackoff(recur.Exponential(100*time.Millisecond)),
    recur.Timeout(10*time.Second),
)

// Use everywhere
err := recur.Do(operation1).WithPolicy(standardRetry).Run()
err = recur.Do(operation2).WithPolicy(standardRetry).Run()
err = recur.Do(operation3).WithPolicy(standardRetry).Run()
```

Other libraries require repeating configuration or creating wrapper functions.

### 2. Rich Error Matching

Built-in boolean combinators for complex retry logic:

```go
err := recur.Do(func() error {
    return operation()
}).
    RetryIf(recur.Or(
        recur.MatchErrors(ErrTimeout, ErrRateLimited),
        recur.And(
            recur.MatchFunc(isNetworkError),
            recur.Not(recur.MatchErrors(ErrFatal)),
        ),
    )).
    Run()
```

Other libraries lack built-in error matching combinators.

### 3. Fluent Configuration

Method chaining is more readable than function option lists:

```go
// Clear hierarchy and flow
recur.Do(operation).
    WithMaxAttempts(5).
    WithBackoff(recur.Exponential(100*time.Millisecond)).
    WithTimeout(30*time.Second).
    RetryIf(recur.MatchErrors(ErrTemporary)).
    OnRetry(logFunc).
    Run()
```

### 4. Multiple Backoff Strategies

Five built-in strategies with easy customization:

```go
recur.Constant(500*time.Millisecond)                          // Fixed
recur.Exponential(100*time.Millisecond)                       // 100ms, 200ms, 400ms...
recur.Fibonacci(50*time.Millisecond)                          // 50ms, 50ms, 100ms, 150ms...
recur.Linear(100*time.Millisecond, 50*time.Millisecond)      // 100ms, 150ms, 200ms...
recur.NoDelay()                                                // Immediate
```

## Installation

```bash
go get github.com/amr8t/go-recur
```

Requires Go 1.21 or later.

## Quick Start

```go
import "github.com/amr8t/go-recur"

// Basic retry
err := recur.Do(func() error {
    return operation()
}).WithMaxAttempts(5).Run()

// With backoff and monitoring
err := recur.Do(func() error {
    return operation()
}).
    WithMaxAttempts(5).
    WithBackoff(recur.Exponential(100*time.Millisecond)).
    OnRetry(func(ctx context.Context, attempt int, err error, elapsed time.Duration) {
        log.Printf("Retry %d: %v", attempt, err)
    }).
    Run()

// Capture return values with closures
var result string
err := recur.Do(func() error {
    var e error
    result, e = fetchData()
    return e
}).WithMaxAttempts(3).Run()
```

## Usage

### Basic Retry

```go
err := recur.Do(func() error {
    return doSomething()
}).WithMaxAttempts(5).Run()
```

### With Return Values

```go
var user User
err := recur.Do(func() error {
    var e error
    user, e = db.GetUser(42)
    return e
}).
    WithMaxAttempts(3).
    WithBackoff(recur.Exponential(100*time.Millisecond)).
    Run()
```

### Composable Policies

```go
// Define reusable policies
aggressiveRetry := recur.CombinePolicies(
    recur.MaxAttempts(10),
    recur.WithBackoff(recur.Exponential(50*time.Millisecond)),
    recur.Timeout(30*time.Second),
)

conservativeRetry := recur.CombinePolicies(
    recur.MaxAttempts(3),
    recur.WithBackoff(recur.Constant(1*time.Second)),
    recur.Timeout(5*time.Second),
)

// Use across your codebase
err := recur.Do(criticalOperation).WithPolicy(aggressiveRetry).Run()
err = recur.Do(backgroundJob).WithPolicy(conservativeRetry).Run()
```

### Conditional Retry with Combinators

```go
var ErrTemporary = errors.New("temporary error")
var ErrRateLimited = errors.New("rate limited")
var ErrFatal = errors.New("fatal error")

err := recur.Do(func() error {
    return operation()
}).
    RetryIf(recur.Or(
        recur.MatchErrors(ErrTemporary, ErrRateLimited),
        recur.And(
            recur.MatchFunc(isNetworkError),
            recur.Not(recur.MatchErrors(ErrFatal)),
        ),
    )).
    Run()
```

### Monitoring

```go
err := recur.Do(func() error {
    return operation()
}).
    OnRetry(func(ctx context.Context, attempt int, err error, elapsed time.Duration) {
        log.Printf("[RETRY] Attempt %d after %v: %v", attempt, elapsed, err)
        metrics.IncrementRetryCount()
    }).
    Run()
```

### Context and Timeout

```go
// Overall timeout for all attempts
err := recur.Do(func() error {
    return operation()
}).
    WithTimeout(5*time.Second).
    Run()

// With custom context
err := recur.Do(func() error {
    return operation()
}).RunContext(ctx)
```

## Real-World Examples

### HTTP Client with Retry

```go
type APIClient struct {
    baseURL string
    retry   recur.Policy
}

func NewAPIClient(baseURL string) *APIClient {
    return &APIClient{
        baseURL: baseURL,
        retry: recur.CombinePolicies(
            recur.MaxAttempts(5),
            recur.WithBackoff(recur.Exponential(500*time.Millisecond)),
            recur.Timeout(30*time.Second),
        ),
    }
}

func (c *APIClient) GetUser(id int) (User, error) {
    var user User
    err := recur.Do(func() error {
        resp, e := http.Get(fmt.Sprintf("%s/users/%d", c.baseURL, id))
        if e != nil {
            return e
        }
        defer resp.Body.Close()
        
        if resp.StatusCode >= 500 {
            return fmt.Errorf("server error: %d", resp.StatusCode)
        }
        
        return json.NewDecoder(resp.Body).Decode(&user)
    }).
        WithPolicy(c.retry).
        RetryIf(recur.MatchFunc(isRetryableHTTPError)).
        Run()
    
    return user, err
}
```

### Database Operations

```go
var users []User
err := recur.Do(func() error {
    return db.Select(&users, "SELECT * FROM users WHERE active = ?", true)
}).
    WithMaxAttempts(3).
    WithBackoff(recur.Exponential(100*time.Millisecond)).
    RetryIf(recur.MatchFunc(func(err error) bool {
        return errors.Is(err, sql.ErrConnDone) || 
               errors.Is(err, driver.ErrBadConn)
    })).
    Run()
```

### Multiple Operations

```go
var user User
var orders []Order
var balance float64

err := recur.Do(func() error {
    u, e := getUser(userID)
    if e != nil {
        return e
    }
    user = u
    
    o, e := getOrders(user.ID)
    if e != nil {
        return e
    }
    orders = o
    
    balance, e = getBalance(user.ID)
    return e
}).
    WithMaxAttempts(5).
    WithBackoff(recur.Exponential(200*time.Millisecond)).
    Run()
```

## API Reference

### Main Function

```go
Do(operation func() error) *SimpleRetrier
```

Creates a new retrier for the given operation.

### Configuration Methods

- `WithMaxAttempts(int)` - Maximum retry attempts (default: 3)
- `WithBackoff(Backoff)` - Backoff strategy (default: 100ms constant)
- `WithTimeout(duration)` - Overall timeout for all attempts
- `WithPolicy(Policy)` - Apply a composable policy
- `RetryIf(ErrorMatcher)` - Only retry on specific errors
- `OnRetry(Hook)` - Add retry callback

### Execution Methods

- `Run()` - Execute with background context
- `RunContext(ctx)` - Execute with provided context

### Backoff Strategies

- `Constant(duration)` - Fixed delay
- `Exponential(initial)` - Exponential backoff (2x multiplier)
- `Fibonacci(initial)` - Fibonacci sequence delays
- `Linear(initial, increment)` - Linear increase
- `NoDelay()` - Immediate retry

### Error Matchers

- `MatchAny` - Match any error (default)
- `MatchErrors(...error)` - Match specific errors using `errors.Is`
- `MatchFunc(func(error) bool)` - Custom matcher function
- `And(...ErrorMatcher)` - Logical AND combinator
- `Or(...ErrorMatcher)` - Logical OR combinator
- `Not(ErrorMatcher)` - Logical NOT combinator

### Policy Composition

```go
CombinePolicies(policies ...Policy) Policy
```

Combine multiple policies into one reusable policy.

## When to Use go-recur

Choose go-recur if you:
- Prefer fluent/builder APIs
- Need composable retry policies across your codebase
- Want rich error matching with boolean combinators
- Value clean, readable configuration

Choose other libraries if you:
- Prefer function options pattern
- Want the most battle-tested solution (avast/retry-go)
- Need specialized backoff algorithms (cenkalti/backoff)

## Examples

See the [examples](examples/) directory for complete runnable examples:

- **[simple](examples/simple/)** - Basic retry patterns, backoff strategies, error matching, and monitoring
- **[http](examples/http/)** - HTTP client retry patterns with real API calls

Run them with:
```bash
cd examples/simple && go run main.go
cd examples/http && go run main.go
```

## Contributing

Contributions are welcome. Please open an issue or submit a pull request.

## License

MIT License. See [LICENSE](LICENSE) file for details.
