package recur

import (
	"errors"
	"testing"
)

func TestSimpleDo(t *testing.T) {
	t.Run("BasicRetry", func(t *testing.T) {
		counter := 0
		err := Do(func() error {
			counter++
			if counter < 3 {
				return errors.New("retry")
			}
			return nil
		}).WithMaxAttempts(5).Run()
		
		if err != nil {
			t.Errorf("Expected success, got: %v", err)
		}
		if counter != 3 {
			t.Errorf("Expected 3 attempts, got %d", counter)
		}
	})
	
	t.Run("WithReturnValue", func(t *testing.T) {
		counter := 0
		var result string
		
		err := Do(func() error {
			counter++
			if counter < 2 {
				return errors.New("retry")
			}
			result = "success"
			return nil
		}).WithMaxAttempts(3).Run()
		
		if err != nil {
			t.Errorf("Expected success, got: %v", err)
		}
		if result != "success" {
			t.Errorf("Expected 'success', got '%s'", result)
		}
	})
	
	t.Run("AnyArgs", func(t *testing.T) {
		// User captures their own variables
		a, b := 10, 20
		var result int
		
		err := Do(func() error {
			result = a + b
			return nil
		}).Run()
		
		if err != nil {
			t.Errorf("Expected success, got: %v", err)
		}
		if result != 30 {
			t.Errorf("Expected 30, got %d", result)
		}
	})
}
