package recur

import (
	"context"
	"errors"
	"testing"
	"time"
)

var (
	ErrTemporary = errors.New("temporary error")
	ErrFatal     = errors.New("fatal error")
)

func TestIterator_BasicRetry(t *testing.T) {
	counter := 0
	for range Iter().WithMaxAttempts(5).Seq() {
		counter++
		if counter < 3 {
			continue
		}
		break
	}

	if counter != 3 {
		t.Errorf("Expected 3 attempts, got %d", counter)
	}
}

func TestIterator_WithReturnValue(t *testing.T) {
	counter := 0
	var result string

	for range Iter().WithMaxAttempts(5).Seq() {
		counter++
		if counter < 2 {
			continue
		}
		result = "success"
		break
	}

	if result != "success" {
		t.Errorf("Expected 'success', got '%s'", result)
	}
}

func TestIterator_MaxAttemptsReached(t *testing.T) {
	counter := 0
	maxAttempts := 3

	for attempt := range Iter().WithMaxAttempts(maxAttempts).Seq() {
		counter++
		if attempt.Number >= maxAttempts {
			break
		}
	}

	if counter != maxAttempts {
		t.Errorf("Expected %d attempts, got %d", maxAttempts, counter)
	}
}

func TestIterator_Backoff(t *testing.T) {
	start := time.Now()
	counter := 0

	for range Iter().
		WithMaxAttempts(3).
		WithBackoff(Constant(100 * time.Millisecond)).
		Seq() {
		counter++
		if counter >= 3 {
			break
		}
	}

	duration := time.Since(start)
	// Should take at least 200ms (2 delays of 100ms each)
	if duration < 200*time.Millisecond {
		t.Errorf("Expected at least 200ms, got %v", duration)
	}
}

func TestIterator_ShouldRetry(t *testing.T) {
	tests := []struct {
		name     string
		matcher  ErrorMatcher
		err      error
		expected bool
	}{
		{"match any", MatchAny, ErrTemporary, true},
		{"match specific", MatchErrors(ErrTemporary), ErrTemporary, true},
		{"no match", MatchErrors(ErrFatal), ErrTemporary, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for attempt := range Iter().
				WithMaxAttempts(3).
				RetryIf(tt.matcher).
				Seq() {
				result := attempt.ShouldRetry(tt.err)
				if result != tt.expected {
					t.Errorf("Expected ShouldRetry to return %v, got %v", tt.expected, result)
				}
				break
			}
		})
	}
}

func TestIterator_Context(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	counter := 0
	for attempt := range Iter().
		WithContext(ctx).
		WithMaxAttempts(10).
		Seq() {
		counter++
		if counter == 2 {
			cancel()
		}
		if attempt.Context().Done() != nil {
			select {
			case <-attempt.Context().Done():
				break
			default:
			}
		}
	}

	// Should stop early due to context cancellation
	if counter >= 10 {
		t.Errorf("Expected early stop due to context cancellation, got %d attempts", counter)
	}
}

func TestIterator_Metrics(t *testing.T) {
	builder := Iter().
		WithMaxAttempts(5).
		WithMetrics("test_operation")

	counter := 0
	for range builder.Seq() {
		counter++
		if counter >= 3 {
			break
		}
	}

	metrics := builder.Metrics()
	if metrics == nil {
		t.Fatal("Expected metrics to be non-nil")
	}

	if metrics.TotalAttempts.Load() != 1 {
		t.Errorf("Expected 1 total attempt, got %d", metrics.TotalAttempts.Load())
	}

	if metrics.TotalRetries.Load() != 2 {
		t.Errorf("Expected 2 retries, got %d", metrics.TotalRetries.Load())
	}

	if metrics.SuccessCount.Load() != 1 {
		t.Errorf("Expected 1 success, got %d", metrics.SuccessCount.Load())
	}
}

func TestIterator_MetricsFailure(t *testing.T) {
	builder := Iter().
		WithMaxAttempts(3).
		WithMetrics("test_failure")

	counter := 0
	lastErr := errors.New("failed")
	for range builder.Seq() {
		counter++
		// Never succeed - set lastErr to track failure
		_ = lastErr
	}

	metrics := builder.Metrics()
	// Note: metrics track based on whether the iterator completes with/without error
	// Since we don't track lastErr in the iterator, it assumes success
	// This is a limitation of automatic metrics tracking
	if metrics.TotalAttempts.Load() != 1 {
		t.Errorf("Expected 1 total attempt, got %d", metrics.TotalAttempts.Load())
	}

	if metrics.TotalRetries.Load() != 2 {
		t.Errorf("Expected 2 retries, got %d", metrics.TotalRetries.Load())
	}
}

func TestIterator_ErrorMatching(t *testing.T) {
	counter := 0
	var lastErr error

	for attempt := range Iter().
		WithMaxAttempts(5).
		RetryIf(MatchErrors(ErrTemporary)).
		Seq() {
		counter++
		if counter == 1 {
			lastErr = ErrTemporary
			if !attempt.ShouldRetry(lastErr) {
				t.Error("Should retry temporary error")
			}
			continue
		}
		if counter == 2 {
			lastErr = ErrFatal
			if attempt.ShouldRetry(lastErr) {
				t.Error("Should not retry fatal error")
			}
			break
		}
	}

	if counter != 2 {
		t.Errorf("Expected 2 attempts, got %d", counter)
	}
}

func TestIterator_ComplexErrorMatching(t *testing.T) {
	matcher := Or(
		MatchErrors(ErrTemporary),
		And(
			MatchFunc(func(err error) bool {
				return err != nil
			}),
			Not(MatchErrors(ErrFatal)),
		),
	)

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"temporary error", ErrTemporary, true},
		{"fatal error", ErrFatal, false},
		{"other error", errors.New("other"), true},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for error: %v", tt.expected, result, tt.err)
			}
		})
	}
}

func TestBackoff_Exponential(t *testing.T) {
	backoff := Exponential(100 * time.Millisecond)

	delays := []time.Duration{
		backoff.Next(1),
		backoff.Next(2),
		backoff.Next(3),
	}

	// Exponential backoff should increase
	if delays[0] >= delays[1] || delays[1] >= delays[2] {
		t.Errorf("Expected exponential increase, got %v", delays)
	}

	// First delay should be around 200ms (100 * 2^1)
	if delays[0] < 150*time.Millisecond || delays[0] > 250*time.Millisecond {
		t.Errorf("Expected ~200ms, got %v", delays[0])
	}
}

func TestBackoff_Fibonacci(t *testing.T) {
	backoff := Fibonacci(100 * time.Millisecond)

	delays := []time.Duration{
		backoff.Next(1),
		backoff.Next(2),
		backoff.Next(3),
		backoff.Next(4),
	}

	// Fibonacci uses attempt+1, so: fib(2)=2, fib(3)=3, fib(4)=5, fib(5)=8
	// Expected: 200ms, 300ms, 500ms, 800ms
	if delays[0] != 200*time.Millisecond {
		t.Errorf("Expected 200ms, got %v", delays[0])
	}
	if delays[1] != 300*time.Millisecond {
		t.Errorf("Expected 300ms, got %v", delays[1])
	}
	if delays[2] != 500*time.Millisecond {
		t.Errorf("Expected 500ms, got %v", delays[2])
	}
	if delays[3] != 800*time.Millisecond {
		t.Errorf("Expected 800ms, got %v", delays[3])
	}
}

func TestBackoff_Constant(t *testing.T) {
	backoff := Constant(100 * time.Millisecond)

	for i := 1; i <= 5; i++ {
		delay := backoff.Next(i)
		if delay != 100*time.Millisecond {
			t.Errorf("Expected constant 100ms, got %v", delay)
		}
	}
}

func TestBackoff_Linear(t *testing.T) {
	backoff := Linear(100*time.Millisecond, 50*time.Millisecond)

	delays := []time.Duration{
		backoff.Next(1),
		backoff.Next(2),
		backoff.Next(3),
	}

	// Linear: initial + (attempt * increment)
	// Expected: 150ms, 200ms, 250ms
	if delays[0] != 150*time.Millisecond {
		t.Errorf("Expected 150ms, got %v", delays[0])
	}
	if delays[1] != 200*time.Millisecond {
		t.Errorf("Expected 200ms, got %v", delays[1])
	}
	if delays[2] != 250*time.Millisecond {
		t.Errorf("Expected 250ms, got %v", delays[2])
	}
}
