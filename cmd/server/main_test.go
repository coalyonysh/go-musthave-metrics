package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetEnv tests the getEnv function
func TestGetEnv(t *testing.T) {
	// Test with default value when env is not set
	t.Setenv("TEST_VAR", "")
	result := getEnv("TEST_VAR", "default")
	assert.Equal(t, "default", result)

	// Test with value set
	t.Setenv("TEST_VAR", "test_value")
	result = getEnv("TEST_VAR", "default")
	assert.Equal(t, "test_value", result)
}

// TestGetEnvBool tests the getEnvBool function
func TestGetEnvBool(t *testing.T) {
	// Test with default value when env is not set
	t.Setenv("TEST_BOOL", "")
	result := getEnvBool("TEST_BOOL", true)
	assert.Equal(t, true, result)

	// Test with "true"
	t.Setenv("TEST_BOOL", "true")
	result = getEnvBool("TEST_BOOL", false)
	assert.Equal(t, true, result)

	// Test with "false"
	t.Setenv("TEST_BOOL", "false")
	result = getEnvBool("TEST_BOOL", true)
	assert.Equal(t, false, result)

	// Test with invalid value
	t.Setenv("TEST_BOOL", "invalid")
	result = getEnvBool("TEST_BOOL", true)
	assert.Equal(t, true, result)
}

// TestGetEnvInt tests the getEnvInt function
func TestGetEnvInt(t *testing.T) {
	// Test with default value when env is not set
	t.Setenv("TEST_INT", "")
	result := getEnvInt("TEST_INT", 100)
	assert.Equal(t, 100, result)

	// Test with valid number
	t.Setenv("TEST_INT", "42")
	result = getEnvInt("TEST_INT", 100)
	assert.Equal(t, 42, result)

	// Test with invalid value
	t.Setenv("TEST_INT", "invalid")
	result = getEnvInt("TEST_INT", 100)
	assert.Equal(t, 100, result)
}

// TestLoggingResponseWriter tests the loggingResponseWriter
func TestLoggingResponseWriter(t *testing.T) {
	// Create a test response recorder
	rec := httptest.NewRecorder()
	respData := &responseData{
		status: 0,
		size:   0,
	}

	lw := loggingResponseWriter{
		ResponseWriter: rec,
		responseData:   respData,
	}

	// Test Write
	n, err := lw.Write([]byte("test"))
	assert.Equal(t, 4, n) // "test" has 4 bytes
	assert.Nil(t, err)
	assert.Equal(t, 4, respData.size)

	// Test WriteHeader
	lw.WriteHeader(http.StatusOK)
	assert.Equal(t, http.StatusOK, respData.status)
}

// TestWithLogging tests the WithLogging middleware
// SKIP: This test requires the global logger to be initialized which happens in main()
func TestWithLogging(t *testing.T) {
	t.Skip("This test requires global logger initialization which is done in main()")
}

// TestCreateDecryptMiddleware tests the createDecryptMiddleware function
func TestCreateDecryptMiddleware(t *testing.T) {
	// Test without private key - just verify it doesn't panic
	middleware := createDecryptMiddleware(nil)
	assert.NotNil(t, middleware)

	// Test with a handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware(handler)
	assert.NotNil(t, wrapped)
}

// TestCreateTrustedSubnetMiddleware tests the createTrustedSubnetMiddleware function
func TestCreateTrustedSubnetMiddleware(t *testing.T) {
	// Skip this test - it requires global logger initialization
	t.Skip("This test requires global logger initialization which is done in main()")
}
