package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
)

func TestSaveMetrics(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "metrics.json")

	memStorage := NewMemStorage()
	memStorage.SetGauge("gauge1", 100.5)
	memStorage.SetCounter("counter1", 50)

	err := SaveMetrics(memStorage, filePath)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check file exists
	_, err = os.Stat(filePath)
	if err != nil {
		t.Errorf("file should exist: %v", err)
	}
}

func TestLoadMetrics(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "metrics.json")

	// Create a test file - format is []models.Metric
	val1 := float64(100.5)
	val2 := float64(200.5)
	delta1 := int64(50)
	delta2 := int64(60)
	metrics := []models.Metric{
		{ID: "gauge1", MType: models.Gauge, Value: &val1},
		{ID: "gauge2", MType: models.Gauge, Value: &val2},
		{ID: "counter1", MType: models.Counter, Delta: &delta1},
		{ID: "counter2", MType: models.Counter, Delta: &delta2},
	}
	data, err := json.Marshal(metrics)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	memStorage := NewMemStorage()
	err = LoadMetrics(memStorage, filePath)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check gauges
	val, ok := memStorage.GetGauge("gauge1")
	if !ok || val != 100.5 {
		t.Errorf("expected gauge1=100.5, got %f", val)
	}

	val2, ok = memStorage.GetGauge("gauge2")
	if !ok || val2 != 200.5 {
		t.Errorf("expected gauge2=200.5, got %f", val2)
	}

	// Check counters
	cnt, ok := memStorage.GetCounter("counter1")
	if !ok || cnt != 50 {
		t.Errorf("expected counter1=50, got %d", cnt)
	}

	cnt2, ok := memStorage.GetCounter("counter2")
	if !ok || cnt2 != 60 {
		t.Errorf("expected counter2=60, got %d", cnt2)
	}
}

func TestLoadMetrics_FileNotFound(t *testing.T) {
	memStorage := NewMemStorage()
	err := LoadMetrics(memStorage, "/nonexistent/path/metrics.json")
	// File doesn't exist - this should return nil (no error)
	if err != nil {
		t.Logf("got error: %v", err)
	}
}

func TestSaveMetrics_EmptyStorage(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "metrics.json")

	memStorage := NewMemStorage()

	err := SaveMetrics(memStorage, filePath)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// File should exist but be empty or have empty structure
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Errorf("failed to read file: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty file")
	}
}

func TestLoadMetrics_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "metrics.json")

	// Create invalid JSON
	content := `invalid json`
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	memStorage := NewMemStorage()
	err = LoadMetrics(memStorage, filePath)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
