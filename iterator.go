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
		ctx, cancel := b.prepareContext()
		if cancel != nil {
			defer cancel()
		}

		state := &iteratorState{
			ctx:         ctx,
			builder:     b,
			startTime:   time.Now(),
			lastAttempt: nil,
		}

		for attempt := 1; attempt <= b.maxAttempts; attempt++ {
			if !state.checkContinue(attempt) {
				return
			}

			att := state.createAttempt(attempt)

			if !state.waitForBackoff(att) {
				return
			}

			state.operationStarted = true
			state.lastAttempt = att

			if !yield(att) {
				state.recordFinalMetrics()
				return
			}
		}

		state.recordExhaustedMetrics()
	}
}

// prepareContext sets up the context with timeout if configured
func (b *IteratorBuilder) prepareContext() (context.Context, context.CancelFunc) {
	if b.timeout > 0 {
		return context.WithTimeout(b.ctx, b.timeout)
	}
	return b.ctx, nil
}

// iteratorState holds the state during iteration
type iteratorState struct {
	ctx              context.Context
	builder          *IteratorBuilder
	startTime        time.Time
	lastAttempt      *Attempt
	operationStarted bool
}

// checkContinue checks if iteration should continue
func (s *iteratorState) checkContinue(attempt int) bool {
	// Check if previous attempt had non-retryable error
	if !s.shouldRetryLastAttempt() {
		s.recordFailureMetrics()
		return false
	}

	// Track retry metrics (not on first attempt)
	if s.builder.metrics != nil && attempt > 1 {
		s.builder.metrics.TotalRetries.Add(1)
	}

	// Check context cancellation
	if s.isContextDone() {
		s.recordFailureMetrics()
		return false
	}

	return true
}

// shouldRetryLastAttempt checks if the last attempt's error should be retried
func (s *iteratorState) shouldRetryLastAttempt() bool {
	if s.lastAttempt == nil {
		return true
	}
	if !s.lastAttempt.resultSet {
		return true
	}
	if s.lastAttempt.result == nil {
		return true
	}
	return s.builder.matcher(s.lastAttempt.result)
}

// isContextDone checks if context is cancelled
func (s *iteratorState) isContextDone() bool {
	select {
	case <-s.ctx.Done():
		return true
	default:
		return false
	}
}

// createAttempt creates a new Attempt for the given attempt number
func (s *iteratorState) createAttempt(attempt int) *Attempt {
	var delay time.Duration
	var lastErr error

	if attempt > 1 {
		delay = s.builder.backoff.Next(attempt - 1)
		if s.lastAttempt != nil {
			lastErr = s.lastAttempt.result
		}
	}

	return &Attempt{
		Number:   attempt,
		LastErr:  lastErr,
		Delay:    delay,
		ctx:      s.ctx,
		matcher:  s.builder.matcher,
		maxRetry: s.builder.maxAttempts,
	}
}

// waitForBackoff waits for the backoff delay or context cancellation
func (s *iteratorState) waitForBackoff(att *Attempt) bool {
	if att.Number <= 1 || att.Delay <= 0 {
		return true
	}

	select {
	case <-time.After(att.Delay):
		return true
	case <-s.ctx.Done():
		s.recordFailureMetrics()
		return false
	}
}

// recordFailureMetrics records failure metrics if enabled
func (s *iteratorState) recordFailureMetrics() {
	if s.builder.metrics != nil && s.operationStarted {
		s.builder.metrics.TotalAttempts.Add(1)
		s.builder.metrics.FailureCount.Add(1)
	}
}

// recordFinalMetrics records metrics when iterator is stopped by user
func (s *iteratorState) recordFinalMetrics() {
	if s.builder.metrics == nil {
		return
	}

	s.builder.metrics.TotalAttempts.Add(1)

	// Determine success based on last attempt result
	if s.lastAttempt != nil && s.lastAttempt.resultSet && s.lastAttempt.result != nil {
		s.builder.metrics.FailureCount.Add(1)
	} else {
		s.builder.metrics.SuccessCount.Add(1)
	}
}

// recordExhaustedMetrics records metrics when all attempts are exhausted
func (s *iteratorState) recordExhaustedMetrics() {
	if s.builder.metrics == nil || !s.operationStarted {
		return
	}

	s.builder.metrics.TotalAttempts.Add(1)

	// Determine success based on last attempt result
	if s.lastAttempt != nil && s.lastAttempt.resultSet && s.lastAttempt.result != nil {
		s.builder.metrics.FailureCount.Add(1)
	} else {
		s.builder.metrics.SuccessCount.Add(1)
	}
}

// Metrics returns the metrics collector if metrics are enabled
func (b *IteratorBuilder) Metrics() *MetricsCollector {
	return b.metrics
}
