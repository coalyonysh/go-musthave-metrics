package main

import (
	"os"
	"testing"
)

// TestGetEnvWithDefault tests getEnv function with no environment variable set
func TestGetEnvWithDefault(t *testing.T) {
	// Make sure the env var is not set
	os.Unsetenv("TEST_VAR")

	result := getEnv("TEST_VAR", "default")
	if result != "default" {
		t.Errorf("Expected 'default', got '%s'", result)
	}
}

// TestGetEnvWithValue tests getEnv function with environment variable set
func TestGetEnvWithValue(t *testing.T) {
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	result := getEnv("TEST_VAR", "default")
	if result != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", result)
	}
}

// TestGetEnvWithEmptyValue tests getEnv function with empty environment variable
// Note: Empty string in env var is treated as "not set" by getEnv, so default is returned
func TestGetEnvWithEmptyValue(t *testing.T) {
	os.Setenv("TEST_EMPTY", "")
	defer os.Unsetenv("TEST_EMPTY")

	// Empty string from env should return default (empty string is treated as not set)
	result := getEnv("TEST_EMPTY", "default")
	if result != "default" {
		t.Errorf("Expected 'default', got '%s'", result)
	}
}

// TestGetEnvIntWithDefault tests getEnvInt function with no environment variable set
func TestGetEnvIntWithDefault(t *testing.T) {
	os.Unsetenv("TEST_INT_VAR")

	result := getEnvInt("TEST_INT_VAR", 100)
	if result != 100 {
		t.Errorf("Expected 100, got %d", result)
	}
}

// TestGetEnvIntWithValue tests getEnvInt function with valid integer
func TestGetEnvIntWithValue(t *testing.T) {
	os.Setenv("TEST_INT_VAR", "200")
	defer os.Unsetenv("TEST_INT_VAR")

	result := getEnvInt("TEST_INT_VAR", 100)
	if result != 200 {
		t.Errorf("Expected 200, got %d", result)
	}
}

// TestGetEnvIntWithInvalidValue tests getEnvInt function with invalid integer
func TestGetEnvIntWithInvalidValue(t *testing.T) {
	os.Setenv("TEST_INT_VAR", "invalid")
	defer os.Unsetenv("TEST_INT_VAR")

	result := getEnvInt("TEST_INT_VAR", 100)
	if result != 100 {
		t.Errorf("Expected 100 (default), got %d", result)
	}
}

// TestGetEnvIntWithNegative tests getEnvInt function with negative value
func TestGetEnvIntWithNegative(t *testing.T) {
	os.Setenv("TEST_INT_VAR", "-5")
	defer os.Unsetenv("TEST_INT_VAR")

	result := getEnvInt("TEST_INT_VAR", 100)
	if result != -5 {
		t.Errorf("Expected -5, got %d", result)
	}
}
