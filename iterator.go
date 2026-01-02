package recur

import (
	"context"
	"iter"
	"sync/atomic"
	"time"
)

// Attempt represents a single retry attempt in an iterator
type Attempt struct {
	Number    int
	LastErr   error
	Delay     time.Duration
	ctx       context.Context
	matcher   ErrorMatcher
	maxRetry  int
	result    error
	resultSet bool
}

// MetricsCollector collects retry metrics
type MetricsCollector struct {
	TotalAttempts atomic.Int64
	SuccessCount  atomic.Int64
	FailureCount  atomic.Int64
	TotalRetries  atomic.Int64
	name          string
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(name string) *MetricsCollector {
	return &MetricsCollector{name: name}
}

// Name returns the collector's name
func (m *MetricsCollector) Name() string {
	return m.name
}

// Result tells the iterator about the operation result
// Call this after your operation to enable automatic retry control
func (a *Attempt) Result(err error) {
	a.result = err
	a.resultSet = true
}

// ShouldRetry returns true if the error should be retried
// Note: If you call Result(err), the iterator will automatically
// stop on non-retryable errors, making this method optional
func (a *Attempt) ShouldRetry(err error) bool {
	if err == nil {
		return false
	}
	if a.Number >= a.maxRetry {
		return false
	}
	return a.matcher(err)
}

// Context returns the attempt's context
func (a *Attempt) Context() context.Context {
	return a.ctx
}

// IteratorBuilder configures an iterator-based retrier
type IteratorBuilder struct {
	maxAttempts int
	backoff     Backoff
	matcher     ErrorMatcher
	timeout     time.Duration
	ctx         context.Context
	metrics     *MetricsCollector
}

// Iter creates a new iterator builder
func Iter() *IteratorBuilder {
	return &IteratorBuilder{
		maxAttempts: 3,
		backoff:     Constant(100 * time.Millisecond),
		matcher:     MatchAny,
		ctx:         context.Background(),
	}
}

// WithMaxAttempts sets the maximum number of attempts
func (b *IteratorBuilder) WithMaxAttempts(n int) *IteratorBuilder {
	b.maxAttempts = n
	return b
}

// WithBackoff sets the backoff strategy
func (b *IteratorBuilder) WithBackoff(backoff Backoff) *IteratorBuilder {
	b.backoff = backoff
	return b
}

// WithTimeout sets an overall timeout
func (b *IteratorBuilder) WithTimeout(d time.Duration) *IteratorBuilder {
	b.timeout = d
	return b
}

// RetryIf sets the error matcher
func (b *IteratorBuilder) RetryIf(matcher ErrorMatcher) *IteratorBuilder {
	b.matcher = matcher
	return b
}

// WithContext sets the context
func (b *IteratorBuilder) WithContext(ctx context.Context) *IteratorBuilder {
	b.ctx = ctx
	return b
}

// WithMetrics enables automatic metrics collection
func (b *IteratorBuilder) WithMetrics(name string) *IteratorBuilder {
	b.metrics = NewMetricsCollector(name)
	return b
}

// WithMetricsCollector uses an existing metrics collector
func (b *IteratorBuilder) WithMetricsCollector(m *MetricsCollector) *IteratorBuilder {
	b.metrics = m
	return b
}

// Seq returns an iterator for use in for...range loops
// If metrics are enabled, they are automatically tracked
func (b *IteratorBuilder) Seq() iter.Seq[*Attempt] {
	return func(yield func(*Attempt) bool) {
		ctx := b.ctx
		var cancel context.CancelFunc

		if b.timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, b.timeout)
			defer cancel()
		}

		var lastErr error
		var lastAttempt *Attempt
		startTime := time.Now()
		operationStarted := false

		for attempt := 1; attempt <= b.maxAttempts; attempt++ {
			// Check if previous attempt had non-retryable error
			if lastAttempt != nil && lastAttempt.resultSet {
				if lastAttempt.result != nil && !b.matcher(lastAttempt.result) {
					// Don't continue - previous error shouldn't be retried
					if b.metrics != nil && operationStarted {
						b.metrics.TotalAttempts.Add(1)
						b.metrics.FailureCount.Add(1)
					}
					return
				}
			}
			// Track retry metrics (not on first attempt)
			if b.metrics != nil && attempt > 1 {
				b.metrics.TotalRetries.Add(1)
			}

			// Check context cancellation
			select {
			case <-ctx.Done():
				if b.metrics != nil && operationStarted {
					b.metrics.TotalAttempts.Add(1)
					b.metrics.FailureCount.Add(1)
				}
				return
			default:
			}

			// Calculate delay for this attempt
			var delay time.Duration
			if attempt > 1 {
				delay = b.backoff.Next(attempt - 1)
			}

			// Create attempt
			att := &Attempt{
				Number:   attempt,
				LastErr:  lastErr,
				Delay:    delay,
				ctx:      ctx,
				matcher:  b.matcher,
				maxRetry: b.maxAttempts,
			}

			// Wait for backoff if not first attempt
			if attempt > 1 && delay > 0 {
				select {
				case <-time.After(delay):
				case <-ctx.Done():
					if b.metrics != nil && operationStarted {
						b.metrics.TotalAttempts.Add(1)
						b.metrics.FailureCount.Add(1)
					}
					return
				}
			}

			operationStarted = true
			lastAttempt = att

			// Yield the attempt
			if !yield(att) {
				// Iterator stopped - record final metrics
				if b.metrics != nil {
					duration := time.Since(startTime)
					b.metrics.TotalAttempts.Add(1)
					// Assume success if stopped with no error stored
					if lastErr == nil {
						b.metrics.SuccessCount.Add(1)
					} else {
						b.metrics.FailureCount.Add(1)
					}
					_ = duration // For future callback support
				}
				return
			}

			// Store the last error for next iteration
			// (will be set by user after they execute their operation)
		}

		// All attempts exhausted - record metrics
		if b.metrics != nil && operationStarted {
			b.metrics.TotalAttempts.Add(1)
			if lastErr == nil {
				b.metrics.SuccessCount.Add(1)
			} else {
				b.metrics.FailureCount.Add(1)
			}
		}
	}
}

// Metrics returns the metrics collector if metrics are enabled
func (b *IteratorBuilder) Metrics() *MetricsCollector {
	return b.metrics
}
