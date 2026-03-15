package main

import (
	"testing"
)

// TestGenerateLoad tests the generateLoad function
// This test requires a running server, so it is skipped by default
func TestGenerateLoad(t *testing.T) {
	t.Skip("Skipping test: requires running server at localhost:8080")
}

// TestCollectProfile tests the collectProfile function
// This test requires a running server, so it is skipped by default
func TestCollectProfile(t *testing.T) {
	t.Skip("Skipping test: requires running server at localhost:8080")
}

// TestProfileURLConstant tests the profileURL constant
func TestProfileURLConstant(t *testing.T) {
	expected := "http://localhost:8080/debug/pprof/heap"
	if profileURL != expected {
		t.Errorf("Expected profileURL to be %s, got %s", expected, profileURL)
	}
}

// TestUpdatesURLConstant tests the updatesURL constant
func TestUpdatesURLConstant(t *testing.T) {
	expected := "http://localhost:8080/updates"
	if updatesURL != expected {
		t.Errorf("Expected updatesURL to be %s, got %s", expected, updatesURL)
	}
}

// TestValueURLConstant tests the valueURL constant
func TestValueURLConstant(t *testing.T) {
	expected := "http://localhost:8080/value/gauge/metric_0"
	if valueURL != expected {
		t.Errorf("Expected valueURL to be %s, got %s", expected, valueURL)
	}
}

// TestIndexURLConstant tests the indexURL constant
func TestIndexURLConstant(t *testing.T) {
	expected := "http://localhost:8080/"
	if indexURL != expected {
		t.Errorf("Expected indexURL to be %s, got %s", expected, indexURL)
	}
}

// TestServerAddrConstant tests the serverAddr constant
func TestServerAddrConstant(t *testing.T) {
	expected := "localhost:8080"
	if serverAddr != expected {
		t.Errorf("Expected serverAddr to be %s, got %s", expected, serverAddr)
	}
}
