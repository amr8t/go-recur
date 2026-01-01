package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/amr8t/go-recur"
)

var (
	ErrTemporary = errors.New("temporary error")
	ErrFatal     = errors.New("fatal error")
)

func main() {
	fmt.Println("=== go-recur Simple Examples ===")

	example1_BasicRetry()
	example2_WithReturnValue()
	example3_WithArguments()
	example4_MultipleReturns()
	example5_Backoff()
	example6_ConditionalRetry()
	example7_WithHooks()
	example8_RealWorld()
}

// Example 1: Basic retry without return values
func example1_BasicRetry() {
	fmt.Println("--- Example 1: Basic Retry ---")

	counter := 0
	err := recur.Do(func() error {
		counter++
		fmt.Printf("Attempt %d\n", counter)
		if counter < 3 {
			return ErrTemporary
		}
		return nil
	}).WithMaxAttempts(5).Run()

	if err != nil {
		log.Printf("Failed: %v\n", err)
	} else {
		fmt.Println("Success!")
	}
	fmt.Println()
}

// Example 2: With return value
func example2_WithReturnValue() {
	fmt.Println("--- Example 2: With Return Value ---")

	counter := 0
	var result string

	err := recur.Do(func() error {
		counter++
		if counter < 2 {
			return ErrTemporary
		}
		result = "Hello, World!"
		return nil
	}).WithMaxAttempts(3).Run()

	if err != nil {
		log.Printf("Failed: %v\n", err)
	} else {
		fmt.Printf("Got result: %s\n", result)
	}
	fmt.Println()
}

// Example 3: With arguments captured in closure
func example3_WithArguments() {
	fmt.Println("--- Example 3: With Arguments ---")

	userID := 42
	userName := "Alice"
	var greeting string

	err := recur.Do(func() error {
		greeting = fmt.Sprintf("Hello, %s (ID: %d)!", userName, userID)
		return nil
	}).WithMaxAttempts(3).Run()

	if err != nil {
		log.Printf("Failed: %v\n", err)
	} else {
		fmt.Printf("Greeting: %s\n", greeting)
	}
	fmt.Println()
}

// Example 4: Multiple return values
func example4_MultipleReturns() {
	fmt.Println("--- Example 4: Multiple Return Values ---")

	var name string
	var age int
	var active bool

	err := recur.Do(func() error {
		// Simulate fetching user data
		name = "Bob"
		age = 30
		active = true
		return nil
	}).WithMaxAttempts(3).Run()

	if err != nil {
		log.Printf("Failed: %v\n", err)
	} else {
		fmt.Printf("User: %s, Age: %d, Active: %v\n", name, age, active)
	}
	fmt.Println()
}

// Example 5: Different backoff strategies
func example5_Backoff() {
	fmt.Println("--- Example 5: Backoff Strategies ---")

	fmt.Println("Exponential backoff:")
	counter := 0
	recur.Do(func() error {
		counter++
		fmt.Printf("Attempt %d\n", counter)
		return ErrTemporary
	}).
		WithMaxAttempts(4).
		WithBackoff(recur.Exponential(50 * time.Millisecond)).
		Run()

	fmt.Println("\nFibonacci backoff:")
	counter = 0
	recur.Do(func() error {
		counter++
		fmt.Printf("Attempt %d\n", counter)
		return ErrTemporary
	}).
		WithMaxAttempts(4).
		WithBackoff(recur.Fibonacci(30 * time.Millisecond)).
		Run()

	fmt.Println()
}

// Example 6: Conditional retry
func example6_ConditionalRetry() {
	fmt.Println("--- Example 6: Conditional Retry ---")

	counter := 0
	err := recur.Do(func() error {
		counter++
		if counter == 1 {
			fmt.Println("Returning temporary error (will retry)")
			return ErrTemporary
		}
		fmt.Println("Returning fatal error (won't retry)")
		return ErrFatal
	}).
		WithMaxAttempts(5).
		RetryIf(recur.MatchErrors(ErrTemporary)).
		Run()

	if err != nil {
		log.Printf("Stopped on: %v\n", err)
	}
	fmt.Println()
}

// Example 7: With monitoring hooks
func example7_WithHooks() {
	fmt.Println("--- Example 7: With Monitoring Hooks ---")

	counter := 0
	err := recur.Do(func() error {
		counter++
		if counter < 3 {
			return ErrTemporary
		}
		return nil
	}).
		WithMaxAttempts(5).
		WithBackoff(recur.Constant(100 * time.Millisecond)).
		OnRetry(func(ctx context.Context, attempt int, err error, elapsed time.Duration) {
			log.Printf("[RETRY] Attempt %d failed after %v: %v", attempt, elapsed, err)
		}).
		Run()

	if err != nil {
		log.Printf("Final failure: %v\n", err)
	} else {
		fmt.Println("Success after retries!")
	}
	fmt.Println()
}

// Example 8: Real-world API simulation
func example8_RealWorld() {
	fmt.Println("--- Example 8: Real-World API Call ---")

	// Simulate API call with retry
	userID := 123
	var userData map[string]interface{}

	err := recur.Do(func() error {
		// Simulate API call
		fmt.Printf("Calling API for user %d...\n", userID)

		// Simulate success on first try
		userData = map[string]interface{}{
			"id":    userID,
			"name":  "John Doe",
			"email": "john@example.com",
		}
		return nil
	}).
		WithMaxAttempts(5).
		WithBackoff(recur.Exponential(200 * time.Millisecond)).
		WithTimeout(5 * time.Second).
		OnRetry(func(ctx context.Context, attempt int, err error, elapsed time.Duration) {
			log.Printf("API call failed, retrying: %v", err)
		}).
		Run()

	if err != nil {
		log.Printf("API call failed: %v\n", err)
	} else {
		fmt.Printf("Got user data: %v\n", userData)
	}
	fmt.Println()
}
