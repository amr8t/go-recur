# go-recur Examples

This directory contains runnable examples demonstrating go-recur's **iterator-first** approach to retry logic.

## Running Examples

Each example is in its own directory with a separate `go.mod` file, so they can be run independently.

### Simple Examples

Showcases iterator-first approach with all features:

```bash
cd simple
go run main.go
```

**Features:**
- **Iterator pattern** - Primary approach using `for...range` (Go 1.23)
- Backoff strategies (Exponential, Fibonacci, Linear, Constant)
- **Error-specific handling** - Different strategies per error
- **Metrics collection** - Built-in observability with callbacks
- Traditional `Do()` API - When you don't need fine-grained control

### HTTP Client Examples

HTTP-specific retry patterns:

```bash
cd http
go run main.go
```

Demonstrates:
- GET requests with retry
- Checking website availability
- Structured API client with retry
- Retrying on specific HTTP status codes
- POST requests with retry
- Batch operations with retry

## Key Features

### 1. Iterator Pattern

```go
// The recommended way to use go-recur
for attempt := range recur.Iter().
    WithMaxAttempts(3).
    WithBackoff(recur.Exponential(100*time.Millisecond)).
    Seq() {
    
    err := operation()
    if err == nil {
        break
    }
    
    if !attempt.ShouldRetry(err) {
        log.Printf("Won't retry: %v", err)
        break
    }
    
    // Add custom logic between retries
    log.Printf("Attempt %d failed after %v", attempt.Number, attempt.Delay)
}
```

**Why iterators:**
- More idiomatic Go (native `for...range`)
- Fine-grained control over retry loop
- Access to attempt context and metadata
- Easy to add custom logic between retries
- No callback nesting

### 2. Error-Specific Handling

```go
// With iterator (full control)
for attempt := range recur.Iter().WithMaxAttempts(3).Seq() {
    err := operation()
    
    if errors.Is(err, ErrRateLimited) {
        time.Sleep(1 * time.Minute)
        continue
    }
    
    if errors.Is(err, ErrAuthFailed) {
        break // Stop immediately
    }
    
    if err == nil || !attempt.ShouldRetry(err) {
        break
    }
}

// With Do() API (declarative)
err := recur.Do(operation).
    On(recur.MatchErrors(ErrRateLimited), recur.Wait(1*time.Minute)).
    On(recur.MatchErrors(ErrAuthFailed), recur.Stop()).
    WithMaxAttempts(3).
    Run()
```

### 3. Metrics Collection

```go
metrics := recur.NewMetricsCollector("api_calls").
    WithCallback(func(name string, attempts, retries int64, success bool, duration time.Duration) {
        // Send to Prometheus, StatsD, or any observability platform
        prometheus.RecordRetry(name, attempts, retries, success, duration)
    })

err := recur.Do(operation).
    WithMetricsCollector(metrics).
    Run()

// Access metrics directly
fmt.Printf("Total attempts: %d\n", metrics.TotalAttempts.Load())
fmt.Printf("Success count: %d\n", metrics.SuccessCount.Load())
fmt.Printf("Failure count: %d\n", metrics.FailureCount.Load())
```

**Metrics tracked:**
- `TotalAttempts` - Operations started
- `SuccessCount` - Successful completions
- `FailureCount` - Failed operations
- `TotalRetries` - Retry attempts made

### 4. Backoff Strategies

```go
// Exponential: 100ms, 200ms, 400ms...
recur.Exponential(100*time.Millisecond)

// Fibonacci: 100ms, 100ms, 200ms, 300ms...
recur.Fibonacci(100*time.Millisecond)

// Linear: 100ms, 150ms, 200ms...
recur.Linear(100*time.Millisecond, 50*time.Millisecond)

// Constant: 100ms, 100ms, 100ms...
recur.Constant(100*time.Millisecond)
```

### 5. Composable Policies (Work in progress)

```go
standardRetry := recur.CombinePolicies(
    recur.MaxAttempts(5),
    recur.WithBackoff(recur.Exponential(100*time.Millisecond)),
    recur.Timeout(10*time.Second),
)

// Reuse across operations
recur.Do(op1).WithPolicy(standardRetry).Run()
recur.Do(op2).WithPolicy(standardRetry).Run()
```

### 4. Traditional Do() API

```go
// Use when you don't need fine-grained control
err := recur.Do(func() error {
    return operation()
}).
    WithMaxAttempts(3).
    WithBackoff(recur.Exponential(100*time.Millisecond)).
    Run()
```

## Iterator-First Examples

### Basic Retry with Iterator

```go
var result string
for attempt := range recur.Iter().WithMaxAttempts(3).Seq() {
    var err error
    result, err = fetchData()
    if err == nil || !attempt.ShouldRetry(err) {
        break
    }
}
```

### With Context and Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

for attempt := range recur.Iter().
    WithContext(ctx).
    WithMaxAttempts(10).
    Seq() {
    
    select {
    case <-attempt.Context().Done():
        log.Println("Timeout or cancellation")
        return
    default:
        err := operation()
        if err == nil || !attempt.ShouldRetry(err) {
            break
        }
    }
}
```


## Adding New Examples

To add a new example:

1. Create a new directory: `examples/myexample/`
2. Add a `go.mod` file with the replace directive pointing to `../../`
3. Add your `main.go` file
4. Update this README

## Philosophy

**go-recur is designed iterator-first:**
- Iterators are the **primary pattern** (not an afterthought)
- Use `Do()` API only when you don't need fine-grained control
- Leverages Go 1.23's native iterator support for clean, idiomatic code

## Notes

- **Requires Go 1.23 or later** for iterator support
- All examples use the `replace` directive in `go.mod` to point to the local go-recur module
- Examples are self-contained and don't depend on each other


## Quick Tips

- **Start with iterators** - They're the primary pattern, not an afterthought
- Use `Do()` API only when you don't need control between retries
- Start with `simple/` to see iterator-first approach
- Use `http/` for real-world HTTP retry patterns
