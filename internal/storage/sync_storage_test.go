package storage

import (
	"testing"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
)

func TestSyncStorage_SetGauge(t *testing.T) {
	memStorage := NewMemStorage()
	s := NewSyncStorage(memStorage, nil)

	s.SetGauge("test_gauge", 123.45)

	value, ok := memStorage.GetGauge("test_gauge")
	if !ok {
		t.Error("expected gauge to exist")
	}
	if value != 123.45 {
		t.Errorf("expected 123.45, got %f", value)
	}
}

func TestSyncStorage_SetCounter(t *testing.T) {
	memStorage := NewMemStorage()
	s := NewSyncStorage(memStorage, nil)

	s.SetCounter("test_counter", 10)
	s.SetCounter("test_counter", 20)

	value, ok := memStorage.GetCounter("test_counter")
	if !ok {
		t.Error("expected counter to exist")
	}
	if value != 30 {
		t.Errorf("expected 30, got %d", value)
	}
}

func TestSyncStorage_GetGauge(t *testing.T) {
	memStorage := NewMemStorage()
	memStorage.SetGauge("test_gauge", 100.0)
	s := NewSyncStorage(memStorage, nil)

	value, ok := s.GetGauge("test_gauge")
	if !ok {
		t.Error("expected gauge to exist")
	}
	if value != 100.0 {
		t.Errorf("expected 100.0, got %f", value)
	}
}

func TestSyncStorage_GetCounter(t *testing.T) {
	memStorage := NewMemStorage()
	memStorage.SetCounter("test_counter", 50)
	s := NewSyncStorage(memStorage, nil)

	value, ok := s.GetCounter("test_counter")
	if !ok {
		t.Error("expected counter to exist")
	}
	if value != 50 {
		t.Errorf("expected 50, got %d", value)
	}
}

func TestSyncStorage_GetAllGauges(t *testing.T) {
	memStorage := NewMemStorage()
	memStorage.SetGauge("gauge1", 1.0)
	memStorage.SetGauge("gauge2", 2.0)
	s := NewSyncStorage(memStorage, nil)

	gauges := s.GetAllGauges()
	if len(gauges) != 2 {
		t.Errorf("expected 2 gauges, got %d", len(gauges))
	}
	if gauges["gauge1"] != 1.0 {
		t.Errorf("expected gauge1=1.0, got %f", gauges["gauge1"])
	}
}

func TestSyncStorage_GetAllCounters(t *testing.T) {
	memStorage := NewMemStorage()
	memStorage.SetCounter("counter1", 10)
	memStorage.SetCounter("counter2", 20)
	s := NewSyncStorage(memStorage, nil)

	counters := s.GetAllCounters()
	if len(counters) != 2 {
		t.Errorf("expected 2 counters, got %d", len(counters))
	}
	if counters["counter1"] != 10 {
		t.Errorf("expected counter1=10, got %d", counters["counter1"])
	}
}

func TestSyncStorage_SetMetricsBatch(t *testing.T) {
	memStorage := NewMemStorage()
	s := NewSyncStorage(memStorage, nil)
	val := float64(100.5)
	delta := int64(50)

	metrics := []models.Metric{
		{ID: "gauge1", MType: models.Gauge, Value: &val},
		{ID: "counter1", MType: models.Counter, Delta: &delta},
	}

	err := s.SetMetricsBatch(metrics)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	gaugeVal, ok := memStorage.GetGauge("gauge1")
	if !ok || gaugeVal != 100.5 {
		t.Errorf("expected gauge1=100.5, got %f", gaugeVal)
	}

	counterVal, ok := memStorage.GetCounter("counter1")
	if !ok || counterVal != 50 {
		t.Errorf("expected counter1=50, got %d", counterVal)
	}
}

func TestSyncStorage_NewSyncStorage(t *testing.T) {
	memStorage := NewMemStorage()
	s := NewSyncStorage(memStorage, nil)
	if s == nil {
		t.Error("expected non-nil sync storage")
	}
	// Verify storage is accessible through public methods
	_, ok := s.GetGauge("nonexistent")
	if ok {
		t.Error("gauge should not exist")
	}
}

func TestSyncStorage_SetGauge_WithSaveFunc(t *testing.T) {
	memStorage := NewMemStorage()
	saveCalled := false
	s := NewSyncStorage(memStorage, func() {
		saveCalled = true
	})

	s.SetGauge("test_gauge", 100.0)

	if !saveCalled {
		t.Error("expected save function to be called")
	}
}
