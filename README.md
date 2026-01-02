# go-recur

**Iterator-native retry library for Go 1.23+**

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![CI](https://github.com/amr8t/go-recur/actions/workflows/ci.yml/badge.svg)](https://github.com/amr8t/go-recur/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/amr8t/go-recur)](https://goreportcard.com/report/github.com/amr8t/go-recur)
[![codecov](https://codecov.io/gh/amr8t/go-recur/branch/main/graph/badge.svg)](https://codecov.io/gh/amr8t/go-recur)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Why go-recur?

**Go 1.23's native iterators** felt like a good fit for retrying operations. A lot of the existing libraries don't take advantage of this feature.

```go
// iterator pattern with more control
for attempt := range recur.Iter().
    WithMaxAttempts(3).
    WithBackoff(recur.Exponential(100*time.Millisecond)).
    WithMetrics("api_call").
    Seq() {
    
    result, err := fetchData()
    attempt.Result(err)
    // Iterator automatically stops on success or non-retryable error
}
```

## Installation

```bash
go get github.com/amr8t/go-recur
```

**Requires Go 1.23 or later.**

## Quick Start

### Basic Retry

```go
for attempt := range recur.Iter().WithMaxAttempts(3).Seq() {
    err := operation()
    attempt.Result(err)
}
```

### With Backoff

```go
for attempt := range recur.Iter().
    WithMaxAttempts(5).
    WithBackoff(recur.Exponential(100*time.Millisecond)).
    Seq() {
    
    err := operation()
    attempt.Result(err)
}
```

### With Metrics

```go
builder := recur.Iter().
    WithMaxAttempts(3).
    WithMetrics("database_query")

for attempt := range builder.Seq() {
    err := db.Query()
    attempt.Result(err)
}

// Check metrics
metrics := builder.Metrics()
fmt.Printf("Success rate: %.2f%%\n", 
    float64(metrics.SuccessCount.Load()) / 
    float64(metrics.TotalAttempts.Load()) * 100)
```

### Error-Specific Handling

```go
for attempt := range recur.Iter().
    WithMaxAttempts(5).
    Seq() {
    
    err := apiCall()
    
    // Rate limit: wait longer
    if errors.Is(err, ErrRateLimited) {
        attempt.Result(err)
        time.Sleep(1 * time.Minute)
        continue
    }
    
    // Auth failure: stop immediately
    if errors.Is(err, ErrAuthFailed) {
        break
    }
    
    // Let iterator handle success/retry logic automatically
    attempt.Result(err)
}
```

## Backoff Strategies

```go
// Exponential: 100ms, 200ms, 400ms, 800ms...
recur.Exponential(100*time.Millisecond)

// Fibonacci: 100ms, 100ms, 200ms, 300ms, 500ms...
recur.Fibonacci(100*time.Millisecond)

// Linear: 100ms, 200ms, 300ms, 400ms...
recur.Linear(100*time.Millisecond, 100*time.Millisecond)

// Constant: 100ms, 100ms, 100ms...
recur.Constant(100*time.Millisecond)

// No delay: immediate retry
recur.NoDelay()
```

### Custom Backoff

```go
type CustomBackoff struct {
    delay time.Duration
}

func (b CustomBackoff) Next(attempt int) time.Duration {
    return b.delay * time.Duration(attempt*attempt)
}

// Use it
for attempt := range recur.Iter().
    WithBackoff(CustomBackoff{delay: 50*time.Millisecond}).
    Seq() {
    err := operation()
    attempt.Result(err)
}
```

## Error Matching

```go
// Match specific errors
for attempt := range recur.Iter().
    RetryIf(recur.MatchErrors(ErrTemporary, ErrTimeout)).
    Seq() {
    err := operation()
    attempt.Result(err)
}

// Custom matcher
for attempt := range recur.Iter().
    RetryIf(recur.MatchFunc(func(err error) bool {
        return errors.Is(err, sql.ErrConnDone) ||
               errors.Is(err, driver.ErrBadConn)
    })).
    Seq() {
    err := db.Query()
    attempt.Result(err)
}

// Combinators
for attempt := range recur.Iter().
    RetryIf(recur.Or(
        recur.MatchErrors(ErrNetwork),
        recur.And(
            recur.MatchFunc(isTransient),
            recur.Not(recur.MatchErrors(ErrFatal)),
        ),
    )).
    Seq() {
    err := operation()
    attempt.Result(err)
}
```

## Real-World Examples

### HTTP Client

```go
func FetchUser(id int) (*User, error) {
    var user *User
    
    for attempt := range recur.Iter().
        WithMaxAttempts(5).
        WithBackoff(recur.Exponential(500*time.Millisecond)).
        WithMetrics("fetch_user").
        Seq() {
        
        resp, err := http.Get(fmt.Sprintf("/users/%d", id))
        if err != nil {
            attempt.Result(err)
            continue
        }
        defer resp.Body.Close()
        
        if resp.StatusCode >= 500 {
            attempt.Result(fmt.Errorf("server error: %d", resp.StatusCode))
            continue
        }
        
        if resp.StatusCode >= 400 {
            break // Don't retry on client errors
        }
        
        err = json.NewDecoder(resp.Body).Decode(&user)
        attempt.Result(err)
    }
    
    if user == nil {
        return nil, errors.New("failed to fetch user")
    }
    return user, nil
}
```

### Database with Fallback

```go
func QueryWithFallback(query string) ([]Row, error) {
    var rows []Row
    
    for attempt := range recur.Iter().
        WithMaxAttempts(5).
        WithBackoff(recur.Exponential(100*time.Millisecond)).
        Seq() {
        
        var err error
        
        // Try primary database
        if attempt.Number <= 3 {
            rows, err = primaryDB.Query(query)
        } else {
            // Fallback to replica after 3 attempts
            log.Println("Falling back to replica")
            rows, err = replicaDB.Query(query)
        }
        
        attempt.Result(err)
    }
    
    return rows, nil
}
```

### With Circuit Breaker

```go
type CircuitBreaker struct {
    failures int
    open     bool
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    if cb.open {
        return errors.New("circuit breaker open")
    }
    
    for attempt := range recur.Iter().WithMaxAttempts(3).Seq() {
        err := fn()
        attempt.Result(err)
        
        if err == nil {
            cb.failures = 0
            return nil
        }
        
        cb.failures++
        if cb.failures >= 5 {
            cb.open = true
            break
        }
    }
    
    return errors.New("operation failed")
}
```

## API Reference

### Iterator Builder

```go
// Create iterator
Iter() *IteratorBuilder

// Configuration
WithMaxAttempts(n int) *IteratorBuilder
WithBackoff(b Backoff) *IteratorBuilder
WithTimeout(d time.Duration) *IteratorBuilder
WithContext(ctx context.Context) *IteratorBuilder
RetryIf(matcher ErrorMatcher) *IteratorBuilder

// Metrics
WithMetrics(name string) *IteratorBuilder
WithMetricsCollector(m *MetricsCollector) *IteratorBuilder
Metrics() *MetricsCollector

// Execute
Seq() iter.Seq[*Attempt]
```

### Attempt

```go
type Attempt struct {
    Number   int           // Attempt number (1-based)
    LastErr  error         // Error from previous attempt
    Delay    time.Duration // Delay before this attempt
}

func (a *Attempt) Result(err error)            // Tell iterator the result; automatically stops on success/non-retryable error
func (a *Attempt) ShouldRetry(err error) bool  // Check if error should be retried (optional if using Result)
func (a *Attempt) Context() context.Context
```

### Metrics

```go
type MetricsCollector struct {
    TotalAttempts atomic.Int64  // Total operations started
    SuccessCount  atomic.Int64  // Successful completions
    FailureCount  atomic.Int64  // Failed operations
    TotalRetries  atomic.Int64  // Total retry attempts
}
```

## When to Use go-recur

**Choose go-recur if you:**
- Use Go 1.23+ and want native iterators
- Need fine-grained control over retry logic
- Want explicit, readable code over callbacks
- Need automatic metrics tracking
- Have complex logic between retries

**Reconsider go-recur if you:**
- Need to support Go < 1.23
- Want function option patterns
- Don't need iterator-based control

## Examples

See the [examples](examples/) directory for complete runnable examples:

- **[simple](examples/simple/)** - Iterator patterns, backoff strategies, metrics, error handling
- **[http](examples/http/)** - HTTP client retry patterns with real API calls

```bash
cd examples/simple && go run main.go
cd examples/http && go run main.go
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License. See [LICENSE](LICENSE) file for details.

---
