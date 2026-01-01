package recur

import (
	"math"
	"time"
)

// Backoff defines a strategy for calculating delay between retries
type Backoff interface {
	Next(attempt int) time.Duration
}

// ConstantBackoff returns a fixed delay between retries
type ConstantBackoff struct {
	delay time.Duration
}

// Constant creates a backoff that waits a fixed duration between retries
func Constant(delay time.Duration) Backoff {
	return &ConstantBackoff{delay: delay}
}

func (b *ConstantBackoff) Next(attempt int) time.Duration {
	return b.delay
}

// ExponentialBackoff increases delay exponentially
type ExponentialBackoff struct {
	initial time.Duration
	max     time.Duration
	factor  float64
}

// Exponential creates a backoff that increases exponentially
// delay = initial * (factor ^ attempt)
func Exponential(initial time.Duration) Backoff {
	return &ExponentialBackoff{
		initial: initial,
		max:     30 * time.Minute, // reasonable default max
		factor:  2.0,
	}
}

// WithMaxDelay sets maximum delay for exponential backoff
func (b *ExponentialBackoff) WithMaxDelay(max time.Duration) *ExponentialBackoff {
	b.max = max
	return b
}

// WithFactor sets the exponential factor (default 2.0)
func (b *ExponentialBackoff) WithFactor(factor float64) *ExponentialBackoff {
	b.factor = factor
	return b
}

func (b *ExponentialBackoff) Next(attempt int) time.Duration {
	delay := float64(b.initial) * math.Pow(b.factor, float64(attempt))
	if delay > float64(b.max) {
		return b.max
	}
	return time.Duration(delay)
}

// FibonacciBackoff uses fibonacci sequence for delays
type FibonacciBackoff struct {
	initial time.Duration
	max     time.Duration
}

// Fibonacci creates a backoff using fibonacci sequence
// delay = initial * fibonacci(attempt)
func Fibonacci(initial time.Duration) Backoff {
	return &FibonacciBackoff{
		initial: initial,
		max:     30 * time.Minute,
	}
}

// WithMaxDelay sets maximum delay for fibonacci backoff
func (b *FibonacciBackoff) WithMaxDelay(max time.Duration) *FibonacciBackoff {
	b.max = max
	return b
}

func (b *FibonacciBackoff) Next(attempt int) time.Duration {
	fib := fibonacci(attempt + 1)
	delay := time.Duration(fib) * b.initial
	if delay > b.max {
		return b.max
	}
	return delay
}

func fibonacci(n int) int {
	if n <= 1 {
		return 1
	}
	a, b := 1, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

// LinearBackoff increases delay linearly
type LinearBackoff struct {
	initial   time.Duration
	increment time.Duration
	max       time.Duration
}

// Linear creates a backoff that increases linearly
// delay = initial + (increment * attempt)
func Linear(initial, increment time.Duration) Backoff {
	return &LinearBackoff{
		initial:   initial,
		increment: increment,
		max:       30 * time.Minute,
	}
}

// WithMaxDelay sets maximum delay for linear backoff
func (b *LinearBackoff) WithMaxDelay(max time.Duration) *LinearBackoff {
	b.max = max
	return b
}

func (b *LinearBackoff) Next(attempt int) time.Duration {
	delay := b.initial + (b.increment * time.Duration(attempt))
	if delay > b.max {
		return b.max
	}
	return delay
}

// NoBackoff doesn't wait between retries
type NoBackoff struct{}

// NoDelay creates a backoff with no delay
func NoDelay() Backoff {
	return &NoBackoff{}
}

func (b *NoBackoff) Next(attempt int) time.Duration {
	return 0
}
