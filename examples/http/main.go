package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/amr8t/go-recur"
)

// Simple HTTP client with retry
type APIClient struct {
	baseURL string
	client  *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// GetUser fetches a user with retry logic
func (c *APIClient) GetUser(id int) (map[string]interface{}, error) {
	var user map[string]interface{}

	err := recur.Do(func() error {
		url := fmt.Sprintf("%s/users/%d", c.baseURL, id)
		resp, err := c.client.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 500 {
			return fmt.Errorf("server error: %d", resp.StatusCode)
		}

		if resp.StatusCode >= 400 {
			return fmt.Errorf("client error: %d", resp.StatusCode)
		}

		return json.NewDecoder(resp.Body).Decode(&user)
	}).
		WithMaxAttempts(5).
		WithBackoff(recur.Exponential(500 * time.Millisecond)).
		OnRetry(func(ctx context.Context, attempt int, err error, elapsed time.Duration) {
			log.Printf("Retry attempt %d: %v", attempt, err)
		}).
		Run()

	return user, err
}

// CreateOrder creates an order with retry
func (c *APIClient) CreateOrder(order map[string]interface{}) (string, error) {
	var orderID string

	err := recur.Do(func() error {
		body, _ := json.Marshal(order)
		url := fmt.Sprintf("%s/orders", c.baseURL)
		resp, err := c.client.Post(url, "application/json", bytes.NewReader(body))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 500 {
			return fmt.Errorf("server error: %d", resp.StatusCode)
		}

		var result struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return err
		}
		orderID = result.ID
		return nil
	}).
		WithMaxAttempts(3).
		WithBackoff(recur.Exponential(1 * time.Second)).
		WithTimeout(30 * time.Second).
		Run()

	return orderID, err
}

// BatchGetUsers fetches multiple users in one retry block
func (c *APIClient) BatchGetUsers(ids []int) ([]map[string]interface{}, error) {
	var users []map[string]interface{}

	err := recur.Do(func() error {
		users = make([]map[string]interface{}, 0, len(ids))

		for _, id := range ids {
			url := fmt.Sprintf("%s/users/%d", c.baseURL, id)
			resp, err := c.client.Get(url)
			if err != nil {
				return err
			}

			if resp.StatusCode != 200 {
				resp.Body.Close()
				return fmt.Errorf("failed to fetch user %d: status %d", id, resp.StatusCode)
			}

			var user map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
				resp.Body.Close()
				return err
			}
			resp.Body.Close()

			users = append(users, user)
		}
		return nil
	}).
		WithMaxAttempts(3).
		WithBackoff(recur.Exponential(200 * time.Millisecond)).
		Run()

	return users, err
}

func main() {
	fmt.Println("=== HTTP Client Examples ===")

	// Example 1: Simple GET with public API
	fmt.Println("--- Example 1: GET Request ---")
	var data map[string]interface{}
	err := recur.Do(func() error {
		resp, err := http.Get("https://httpbin.org/json")
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("status %d: %s", resp.StatusCode, body)
		}

		return json.NewDecoder(resp.Body).Decode(&data)
	}).
		WithMaxAttempts(3).
		WithBackoff(recur.Exponential(500 * time.Millisecond)).
		Run()

	if err != nil {
		log.Printf("Failed: %v\n", err)
	} else {
		fmt.Printf("Success! Got %d keys in response\n", len(data))
	}

	// Example 2: Check website availability
	fmt.Println("\n--- Example 2: Check Availability ---")
	err = recur.Do(func() error {
		resp, err := http.Get("https://httpbin.org/status/200")
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf("status: %d", resp.StatusCode)
		}
		return nil
	}).
		WithMaxAttempts(5).
		WithBackoff(recur.Constant(1 * time.Second)).
		OnRetry(func(ctx context.Context, attempt int, err error, elapsed time.Duration) {
			log.Printf("Attempt %d failed: %v", attempt, err)
		}).
		Run()

	if err != nil {
		log.Printf("Site unavailable: %v\n", err)
	} else {
		fmt.Println("Site is available!")
	}

	fmt.Println("\nDone!")
}
