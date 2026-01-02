package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/amr8t/go-recur"
)

func main() {
	fmt.Println("=== HTTP Client Examples (Iterator-Only) ===\n")

	example1_SimpleGET()
	example2_CheckAvailability()
	example3_WithReturnValue()
}

// Example 1: Simple GET request with retry
func example1_SimpleGET() {
	fmt.Println("--- Example 1: Simple GET Request ---")

	var data map[string]interface{}

	for attempt := range recur.Iter().
		WithMaxAttempts(3).
		WithBackoff(recur.Exponential(500 * time.Millisecond)).
		Seq() {

		resp, err := http.Get("https://httpbin.org/json")
		if err != nil {
			attempt.Result(err)
			log.Printf("  Attempt %d failed: %v", attempt.Number, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			err = fmt.Errorf("status %d", resp.StatusCode)
			attempt.Result(err)
			log.Printf("  Attempt %d: bad status %d", attempt.Number, resp.StatusCode)
			continue
		}

		err = json.NewDecoder(resp.Body).Decode(&data)
		attempt.Result(err)
		if err == nil {
			fmt.Printf("✓ Success! Got %d keys in response\n", len(data))
			break
		}
	}
	fmt.Println()
}

// Example 2: Check website availability
func example2_CheckAvailability() {
	fmt.Println("--- Example 2: Check Availability ---")

	for attempt := range recur.Iter().
		WithMaxAttempts(5).
		WithBackoff(recur.Constant(500 * time.Millisecond)).
		Seq() {

		resp, err := http.Get("https://httpbin.org/status/200")
		if err != nil {
			attempt.Result(err)
			log.Printf("  Attempt %d: request failed", attempt.Number)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			err = fmt.Errorf("status: %d", resp.StatusCode)
			attempt.Result(err)
			log.Printf("  Attempt %d: bad status", attempt.Number)
			continue
		}

		attempt.Result(nil)
		fmt.Println("✓ Site is available!")
		break
	}
	fmt.Println()
}

// Example 3: Fetch user data with return value
func example3_WithReturnValue() {
	fmt.Println("--- Example 3: Fetch User Data ---")

	user, err := fetchUser(1)
	if err != nil {
		log.Printf("Failed to fetch user: %v\n", err)
	} else {
		fmt.Printf("✓ Got user data: %v\n", user)
	}
	fmt.Println()
}

func fetchUser(id int) (map[string]interface{}, error) {
	var user map[string]interface{}

	for attempt := range recur.Iter().
		WithMaxAttempts(3).
		WithBackoff(recur.Exponential(300 * time.Millisecond)).
		WithMetrics("fetch_user").
		Seq() {

		url := fmt.Sprintf("https://httpbin.org/json")
		resp, err := http.Get(url)
		if err != nil {
			attempt.Result(err)
			continue
		}
		defer resp.Body.Close()

		// Retry on server errors
		if resp.StatusCode >= 500 {
			err = fmt.Errorf("server error: %d", resp.StatusCode)
			attempt.Result(err)
			continue
		}

		// Don't retry on client errors
		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("client error: %d", resp.StatusCode)
		}

		err = json.NewDecoder(resp.Body).Decode(&user)
		attempt.Result(err)
		if err == nil {
			return user, nil
		}
	}

	return nil, errors.New("max retries exceeded")
}
