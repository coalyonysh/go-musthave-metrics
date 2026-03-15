package main

import (
	"sync"
	"testing"
)

func TestNew(t *testing.T) {
	factory := func() *Request {
		return &Request{
			Headers: make(map[string]string),
		}
	}

	pool := New(factory)

	if pool == nil {
		t.Fatal("New() returned nil")
	}
}

func TestPoolGetPut(t *testing.T) {
	factory := func() *Request {
		return &Request{
			Headers: make(map[string]string),
		}
	}

	pool := New(factory)

	// Get an object from the pool
	req1 := pool.Get()
	req1.ID = 1
	req1.Method = "GET"
	req1.URL = "https://example.com"
	req1.Headers["Content-Type"] = "application/json"
	req1.Body = []byte(`{"key": "value"}`)

	// Put the object back (this should reset it)
	pool.Put(req1)

	// Get another object - should be reset
	req2 := pool.Get()

	if req2.ID != 0 {
		t.Errorf("Expected ID to be 0 after reset, got %d", req2.ID)
	}
	if req2.Method != "" {
		t.Errorf("Expected Method to be empty after reset, got %s", req2.Method)
	}
	if req2.URL != "" {
		t.Errorf("Expected URL to be empty after reset, got %s", req2.URL)
	}
	if len(req2.Headers) != 0 {
		t.Errorf("Expected Headers to be empty after reset, got %v", req2.Headers)
	}
	if len(req2.Body) != 0 {
		t.Errorf("Expected Body to be empty after reset, got %v", req2.Body)
	}
}

func TestRequestReset(t *testing.T) {
	req := &Request{
		ID:      42,
		Method:  "POST",
		URL:     "https://example.com/submit",
		Headers: map[string]string{"Authorization": "Bearer token"},
		Body:    []byte(`{"data": "test"}`),
	}

	// Reset the request
	req.Reset()

	if req.ID != 0 {
		t.Errorf("Expected ID to be 0 after reset, got %d", req.ID)
	}
	if req.Method != "" {
		t.Errorf("Expected Method to be empty after reset, got %s", req.Method)
	}
	if req.URL != "" {
		t.Errorf("Expected URL to be empty after reset, got %s", req.URL)
	}
	if req.Headers == nil {
		t.Error("Expected Headers to not be nil after reset")
	}
	if len(req.Headers) != 0 {
		t.Errorf("Expected Headers to be empty after reset, got %v", req.Headers)
	}
	if req.Body == nil {
		t.Error("Expected Body to not be nil after reset")
	}
	if len(req.Body) != 0 {
		t.Errorf("Expected Body length to be 0 after reset, got %d", len(req.Body))
	}
}

func TestPoolConcurrentAccess(t *testing.T) {
	factory := func() *Request {
		return &Request{
			Headers: make(map[string]string),
		}
	}

	pool := New(factory)

	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Get an object
			req := pool.Get()
			req.ID = id
			req.Method = "GET"

			// Put it back
			pool.Put(req)
		}(i)
	}

	wg.Wait()

	// Verify the pool works after concurrent access
	req := pool.Get()
	if req == nil {
		t.Error("Failed to get request from pool after concurrent access")
	}
}

func TestPoolMultipleGets(t *testing.T) {
	factory := func() *Request {
		return &Request{
			Headers: make(map[string]string),
		}
	}

	pool := New(factory)

	// Get multiple objects
	req1 := pool.Get()
	req1.ID = 1
	pool.Put(req1)

	req2 := pool.Get()
	req2.ID = 2
	pool.Put(req2)

	req3 := pool.Get()

	// After putting back req1 and req2, getting req3 should give us a reset object
	if req3.ID != 0 {
		t.Errorf("Expected ID to be 0, got %d", req3.ID)
	}
}
