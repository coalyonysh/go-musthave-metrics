package client

import (
	"context"
	"errors"
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
)

func TestIsRetriableError_Nil(t *testing.T) {
	result := isRetriableError(nil)
	if result {
		t.Error("expected false for nil error")
	}
}

func TestIsRetriableError_NetOpError(t *testing.T) {
	err := &net.OpError{
		Op:  "dial",
		Err: errors.New("connection refused"),
	}
	result := isRetriableError(err)
	if !result {
		t.Error("expected true for net.OpError")
	}
}

func TestIsRetriableError_Timeout(t *testing.T) {
	err := context.DeadlineExceeded
	result := isRetriableError(err)
	if !result {
		t.Error("expected true for context.DeadlineExceeded")
	}
}

func TestIsRetriableError_URLError(t *testing.T) {
	err := &url.Error{
		Op:  "Get",
		URL: "http://localhost",
		Err: errors.New("timeout"),
	}
	result := isRetriableError(err)
	if !result {
		t.Error("expected true for url.Error")
	}
}

func TestIsRetriableError_Other(t *testing.T) {
	err := errors.New("some error")
	result := isRetriableError(err)
	if result {
		t.Error("expected false for other errors")
	}
}

func TestRetrySend_Success(t *testing.T) {
	callCount := 0
	err := retrySend(func() error {
		callCount++
		return nil
	})

	if err != nil {
		t.Errorf("retrySend returned error: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
}

func TestRetrySend_NonRetriableError(t *testing.T) {
	callCount := 0
	err := retrySend(func() error {
		callCount++
		return errors.New("non-retriable error")
	})

	if err == nil {
		t.Error("expected error for non-retriable error")
	}
	if callCount != 1 {
		t.Errorf("expected 1 call (no retries for non-retriable), got %d", callCount)
	}
}

func TestRetrySend_RetriableError(t *testing.T) {
	callCount := 0
	err := retrySend(func() error {
		callCount++
		if callCount < 3 {
			return &net.OpError{Err: errors.New("connection refused")}
		}
		return nil
	})

	if err != nil {
		t.Errorf("retrySend returned error after retries: %v", err)
	}
	if callCount != 3 {
		t.Errorf("expected 3 calls (2 retries + 1 success), got %d", callCount)
	}
}

func TestRetrySend_AllRetriable(t *testing.T) {
	callCount := 0
	err := retrySend(func() error {
		callCount++
		return &net.OpError{Err: errors.New("connection refused")}
	})

	// All attempts fail, but we still make the final retry
	if err == nil {
		t.Error("expected error after all retries failed")
	}
	// 3 retries + 1 final = 4 calls
	if callCount != 4 {
		t.Errorf("expected 4 calls, got %d", callCount)
	}
}

func TestSendMetricsBatch_Empty(t *testing.T) {
	client := NewMetricHTTPClient("http://localhost:8080", "")

	// Empty batch should succeed
	err := client.SendMetricsBatch([]models.Metric{})
	if err != nil {
		t.Errorf("SendMetricsBatch returned error for empty batch: %v", err)
	}
}

func TestNewMetricHTTPClient(t *testing.T) {
	client := NewMetricHTTPClient("http://localhost:8080", "testkey")

	if client == nil {
		t.Fatal("NewMetricHTTPClient returned nil")
	}
	if client.baseURL != "http://localhost:8080" {
		t.Errorf("expected baseURL http://localhost:8080, got %s", client.baseURL)
	}
	if client.key != "testkey" {
		t.Errorf("expected key testkey, got %s", client.key)
	}
	if client.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
	if client.httpClient.Timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", client.httpClient.Timeout)
	}
}
