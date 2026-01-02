package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/amr8t/go-recur"
)

var (
	ErrTemporary   = errors.New("temporary error")
	ErrRateLimited = errors.New("rate limited")
	ErrAuthFailed  = errors.New("authentication failed")
	ErrNetwork     = errors.New("network error")
)

func main() {
	fmt.Println("=== go-recur Examples (Iterator-Only) ===\n")

	example1_BasicIterator()
	example2_WithBackoff()
	example3_WithMetrics()
	example4_ErrorSpecificHandling()
	example5_ContextAndTimeout()
}

// Example 1: Basic iterator pattern
func example1_BasicIterator() {
	fmt.Println("--- Example 1: Basic Iterator Pattern ---")

	counter := 0
	var result string

	for attempt := range recur.Iter().WithMaxAttempts(4).Seq() {
		counter++
		fmt.Printf("  Attempt %d\n", attempt.Number)

		// Execute operation
		var err error
		if counter < 3 {
			err = ErrTemporary
		}

		attempt.Result(err) // Tell iterator the result
		if err == nil {
			result = "success"
			break
		}
		// Iterator automatically stops if error shouldn't be retried
	}

	fmt.Printf("✓ Success! Result: %s\n\n", result)
}

// Example 2: Iterator with exponential backoff
func example2_WithBackoff() {
	fmt.Println("--- Example 2: Exponential Backoff ---")

	counter := 0
	for attempt := range recur.Iter().
		WithMaxAttempts(4).
		WithBackoff(recur.Exponential(100 * time.Millisecond)).
		Seq() {

		counter++
		fmt.Printf("  Attempt %d (delay: %v)\n", attempt.Number, attempt.Delay)

		var err error
		if counter < 3 {
			err = ErrTemporary
			fmt.Println("    → Temporary error, retrying...")
		} else {
			fmt.Println("    → Success!")
		}

		attempt.Result(err)
		if err == nil {
			break
		}
	}
	fmt.Println()
}

// Example 3: Automated metrics
func example3_WithMetrics() {
	fmt.Println("--- Example 3: Automated Metrics ---")

	builder := recur.Iter().
		WithMaxAttempts(4).
		WithBackoff(recur.Constant(50 * time.Millisecond)).
		WithMetrics("api_operation")

	counter := 0
	for attempt := range builder.Seq() {
		counter++
		var err error
		if counter < 3 {
			err = ErrTemporary
		}
		attempt.Result(err)
		if err == nil {
			break
		}
	}

	// Metrics are automatically tracked
	metrics := builder.Metrics()
	fmt.Printf("  Total Operations: %d\n", metrics.TotalAttempts.Load())
	fmt.Printf("  Total Retries:    %d\n", metrics.TotalRetries.Load())
	fmt.Printf("  Success Count:    %d\n", metrics.SuccessCount.Load())
	fmt.Printf("  Success Rate:     %.0f%%\n",
		float64(metrics.SuccessCount.Load())/float64(metrics.TotalAttempts.Load())*100)
	fmt.Println()
}

// Example 4: Error-specific handling
func example4_ErrorSpecificHandling() {
	fmt.Println("--- Example 4: Error-Specific Handling ---")

	counter := 0
	for attempt := range recur.Iter().
		WithMaxAttempts(5).
		WithBackoff(recur.Constant(100 * time.Millisecond)).
		Seq() {

		counter++
		fmt.Printf("  Attempt %d: ", attempt.Number)

		var err error
		switch counter {
		case 1:
			err = ErrRateLimited
			fmt.Println("Rate limited!")
			attempt.Result(err)
			// Custom handling: wait longer for rate limits
			time.Sleep(1 * time.Second)
			continue
		case 2:
			err = ErrNetwork
			fmt.Println("Network error")
			attempt.Result(err)
			continue
		case 3:
			fmt.Println("Success!")
			attempt.Result(nil)
			break
		}
	}
	fmt.Println()
}

// Example 5: Context and timeout
func example5_ContextAndTimeout() {
	fmt.Println("--- Example 5: Context and Timeout ---")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	counter := 0
	for attempt := range recur.Iter().
		WithContext(ctx).
		WithMaxAttempts(10).
		WithBackoff(recur.Constant(500 * time.Millisecond)).
		Seq() {

		counter++
		fmt.Printf("  Attempt %d\n", counter)

		select {
		case <-attempt.Context().Done():
			fmt.Println("  ✗ Context timeout")
			return
		default:
			var err error
			if counter < 5 {
				err = ErrTemporary
			}
			attempt.Result(err)
			if err == nil {
				fmt.Println("  ✓ Success!")
				return
			}
		}
	}
}
