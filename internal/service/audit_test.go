package service

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
)

// MockObserver implements Observer interface for testing
type MockObserver struct {
	events []models.AuditEvent
	err    error
}

func (m *MockObserver) Notify(event models.AuditEvent) error {
	m.events = append(m.events, event)
	return m.err
}

func TestNewAuditService(t *testing.T) {
	as := NewAuditService()
	if as == nil {
		t.Fatal("NewAuditService returned nil")
	}
	if as.observers == nil {
		t.Error("observers slice should be initialized")
	}
	if len(as.observers) != 0 {
		t.Errorf("expected empty observers, got %d", len(as.observers))
	}
}

func TestAuditService_AddObserver(t *testing.T) {
	as := NewAuditService()
	observer := &MockObserver{}

	as.AddObserver(observer)

	if len(as.observers) != 1 {
		t.Errorf("expected 1 observer, got %d", len(as.observers))
	}

	// Add second observer
	observer2 := &MockObserver{}
	as.AddObserver(observer2)

	if len(as.observers) != 2 {
		t.Errorf("expected 2 observers, got %d", len(as.observers))
	}
}

func TestAuditService_Log_NoObservers(t *testing.T) {
	as := NewAuditService()

	// Should not panic when no observers
	as.Log([]string{"metric1", "metric2"}, "192.168.1.1")
}

func TestAuditService_Log_WithObservers(t *testing.T) {
	observer := &MockObserver{}
	as := NewAuditService()
	as.AddObserver(observer)

	metrics := []string{"metric1", "metric2"}
	ip := "192.168.1.1"

	as.Log(metrics, ip)

	if len(observer.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(observer.events))
	}

	event := observer.events[0]
	if len(event.Metrics) != 2 {
		t.Errorf("expected 2 metrics, got %d", len(event.Metrics))
	}
	if event.IPAddress != ip {
		t.Errorf("expected IP %s, got %s", ip, event.IPAddress)
	}
	if event.Timestamp == 0 {
		t.Error("timestamp should not be zero")
	}
}

func TestAuditService_Log_MultipleObservers(t *testing.T) {
	observer1 := &MockObserver{}
	observer2 := &MockObserver{}

	as := NewAuditService()
	as.AddObserver(observer1)
	as.AddObserver(observer2)

	as.Log([]string{"metric1"}, "127.0.0.1")

	if len(observer1.events) != 1 {
		t.Errorf("expected observer1 to receive 1 event, got %d", len(observer1.events))
	}
	if len(observer2.events) != 1 {
		t.Errorf("expected observer2 to receive 1 event, got %d", len(observer2.events))
	}
}

func TestAuditService_Log_ObserverError(t *testing.T) {
	observer := &MockObserver{err: fmt.Errorf("test error")}
	as := NewAuditService()
	as.AddObserver(observer)

	// Should not panic when observer returns error
	as.Log([]string{"metric1"}, "127.0.0.1")

	if len(observer.events) != 1 {
		t.Errorf("expected event to be sent even with error, got %d", len(observer.events))
	}
}

func TestNewFileObserver(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_audit.log")

	fo := NewFileObserver(filePath)
	if fo == nil {
		t.Fatal("NewFileObserver returned nil")
	}
	if fo.filePath != filePath {
		t.Errorf("expected filePath %s, got %s", filePath, fo.filePath)
	}
}

func TestFileObserver_Notify(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_audit.log")

	fo := NewFileObserver(filePath)
	event := models.AuditEvent{
		Timestamp: 1234567890,
		Metrics:   []string{"metric1", "metric2"},
		IPAddress: "192.168.1.1",
	}

	err := fo.Notify(event)
	if err != nil {
		t.Errorf("Notify returned error: %v", err)
	}

	// Check file was created
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		t.Error("audit file was not created")
	}
}

func TestFileObserver_Notify_InvalidPath(t *testing.T) {
	// Use non-writable path
	fo := NewFileObserver("/nonexistent/path/that/does/not/exist/audit.log")
	event := models.AuditEvent{
		Timestamp: 1234567890,
		Metrics:   []string{"metric1"},
		IPAddress: "127.0.0.1",
	}

	err := fo.Notify(event)
	if err == nil {
		t.Error("Notify should return error for invalid path")
	}
}

func TestNewURLOobserver(t *testing.T) {
	url := "http://localhost:8080/audit"
	uo := NewURLOobserver(url)
	if uo == nil {
		t.Fatal("NewURLOobserver returned nil")
	}
	if uo.url != url {
		t.Errorf("expected url %s, got %s", url, uo.url)
	}
}
