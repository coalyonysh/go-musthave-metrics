package main

import (
	"sync"
	"testing"
)

// TestRequest is a test struct that implements Resetter interface
type TestRequest struct {
	ID     int
	Status string
}

// Reset implements Resetter interface
func (r *TestRequest) Reset() {
	r.ID = 0
	r.Status = ""
}

// TestNewPool tests the New constructor
func TestNewPool(t *testing.T) {
	factory := func() *TestRequest {
		return &TestRequest{}
	}

	pool := New(factory)
	if pool == nil {
		t.Fatal("Expected non-nil pool")
	}

	// Get an object from pool
	obj := pool.Get()
	if obj == nil {
		t.Error("Expected non-nil object from pool")
	}
}

// TestPoolGetPut tests getting and putting objects in the pool
func TestPoolGetPut(t *testing.T) {
	factory := func() *TestRequest {
		return &TestRequest{Status: "initialized"}
	}

	pool := New(factory)

	// Get an object
	obj1 := pool.Get()
	obj1.ID = 1
	obj1.Status = "modified"

	// Put it back
	pool.Put(obj1)

	// Get again - should be reset
	obj2 := pool.Get()
	if obj2.ID != 0 {
		t.Errorf("Expected ID to be 0 after reset, got %d", obj2.ID)
	}
	if obj2.Status != "" {
		t.Errorf("Expected Status to be empty after reset, got %s", obj2.Status)
	}
}

// TestPoolMultipleGets tests multiple gets from the pool
func TestPoolMultipleGets(t *testing.T) {
	factory := func() *TestRequest {
		return &TestRequest{}
	}

	pool := New(factory)

	// Get multiple objects
	obj1 := pool.Get()
	obj1.ID = 100
	pool.Put(obj1)

	obj2 := pool.Get()
	obj2.ID = 200
	pool.Put(obj2)

	obj3 := pool.Get()
	if obj3.ID != 0 {
		t.Errorf("Expected ID to be 0, got %d", obj3.ID)
	}
}

// TestPoolConcurrentAccess tests concurrent access to the pool
func TestPoolConcurrentAccess(t *testing.T) {
	factory := func() *TestRequest {
		return &TestRequest{}
	}

	pool := New(factory)

	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Get from pool
			obj := pool.Get()
			obj.ID = id

			// Verify we can set ID
			if obj.ID != id {
				t.Errorf("Expected ID %d, got %d", id, obj.ID)
			}

			// Put back
			pool.Put(obj)
		}(i)
	}

	wg.Wait()
}

// TestRequestReset tests the Reset method of Request
func TestRequestReset(t *testing.T) {
	req := &Request{
		ID:     42,
		Method: "POST",
		URL:    "https://example.com",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: []byte("test body"),
	}

	req.Reset()

	if req.ID != 0 {
		t.Errorf("Expected ID to be 0, got %d", req.ID)
	}
	if req.Method != "" {
		t.Errorf("Expected Method to be empty, got %s", req.Method)
	}
	if req.URL != "" {
		t.Errorf("Expected URL to be empty, got %s", req.URL)
	}
	if req.Headers == nil {
		t.Error("Expected Headers to be initialized")
	}
	if len(req.Headers) != 0 {
		t.Errorf("Expected Headers to be empty, got %d items", len(req.Headers))
	}
	if len(req.Body) != 0 {
		t.Errorf("Expected Body length to be 0, got %d", len(req.Body))
	}
}

// TestRequestResetWithLargeBody tests Reset with large body capacity
func TestRequestResetWithLargeBody(t *testing.T) {
	req := &Request{
		Body: make([]byte, 0, 1000),
	}

	// Add some data
	req.Body = append(req.Body, []byte("some data")...)

	req.Reset()

	// Body should be empty but retain capacity
	if len(req.Body) != 0 {
		t.Errorf("Expected Body length to be 0, got %d", len(req.Body))
	}
	if cap(req.Body) != 1000 {
		t.Errorf("Expected Body capacity to be 1000, got %d", cap(req.Body))
	}
}
