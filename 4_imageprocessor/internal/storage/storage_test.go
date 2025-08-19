package storage

import (
	"errors"
	"testing"
)

func TestErrNotFound(t *testing.T) {
	// Test that ErrNotFound is not nil
	if ErrNotFound == nil {
		t.Error("Expected ErrNotFound to be defined, got nil")
	}

	// Test that ErrNotFound is an error
	var err error = ErrNotFound
	if err == nil {
		t.Error("Expected ErrNotFound to implement error interface")
	}

	// Test error message
	expectedMessage := "not found"
	if ErrNotFound.Error() != expectedMessage {
		t.Errorf("Expected error message '%s', got '%s'", expectedMessage, ErrNotFound.Error())
	}

	// Test error comparison
	if !errors.Is(ErrNotFound, ErrNotFound) {
		t.Error("Expected errors.Is to return true for same error")
	}

	// Test that it's not equal to a different error
	differentError := errors.New("different error")
	if errors.Is(ErrNotFound, differentError) {
		t.Error("Expected errors.Is to return false for different error")
	}
}

func TestErrorConstants(t *testing.T) {
	// Test that all error constants are defined
	errors := []error{
		ErrNotFound,
	}

	for i, err := range errors {
		if err == nil {
			t.Errorf("Error at index %d is nil", i)
		}
	}

	// Test that all errors have non-empty messages
	for i, err := range errors {
		if err.Error() == "" {
			t.Errorf("Error at index %d has empty message", i)
		}
	}
}