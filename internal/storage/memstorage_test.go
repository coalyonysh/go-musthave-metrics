package storage

import (
	"testing"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
)

func TestMemStorage_SetGauge(t *testing.T) {
	s := NewMemStorage()
	s.SetGauge("test_gauge", 123.45)

	value, ok := s.GetGauge("test_gauge")
	if !ok {
		t.Error("expected gauge to exist")
	}
	if value != 123.45 {
		t.Errorf("expected 123.45, got %f", value)
	}
}

func TestMemStorage_SetGauge_Update(t *testing.T) {
	s := NewMemStorage()
	s.SetGauge("test_gauge", 100.0)
	s.SetGauge("test_gauge", 200.0)

	value, ok := s.GetGauge("test_gauge")
	if !ok {
		t.Error("expected gauge to exist")
	}
	if value != 200.0 {
		t.Errorf("expected 200.0, got %f", value)
	}
}

func TestMemStorage_SetCounter(t *testing.T) {
	s := NewMemStorage()
	s.SetCounter("test_counter", 10)
	s.SetCounter("test_counter", 20)

	value, ok := s.GetCounter("test_counter")
	if !ok {
		t.Error("expected counter to exist")
	}
	if value != 30 {
		t.Errorf("expected 30, got %d", value)
	}
}

func TestMemStorage_GetGauge_NotFound(t *testing.T) {
	s := NewMemStorage()
	_, ok := s.GetGauge("nonexistent")
	if ok {
		t.Error("expected gauge to not exist")
	}
}

func TestMemStorage_GetCounter_NotFound(t *testing.T) {
	s := NewMemStorage()
	_, ok := s.GetCounter("nonexistent")
	if ok {
		t.Error("expected counter to not exist")
	}
}

func TestMemStorage_GetAllGauges(t *testing.T) {
	s := NewMemStorage()
	s.SetGauge("gauge1", 1.0)
	s.SetGauge("gauge2", 2.0)
	s.SetGauge("gauge3", 3.0)

	gauges := s.GetAllGauges()
	if len(gauges) != 3 {
		t.Errorf("expected 3 gauges, got %d", len(gauges))
	}
	if gauges["gauge1"] != 1.0 {
		t.Errorf("expected gauge1=1.0, got %f", gauges["gauge1"])
	}
	if gauges["gauge2"] != 2.0 {
		t.Errorf("expected gauge2=2.0, got %f", gauges["gauge2"])
	}
	if gauges["gauge3"] != 3.0 {
		t.Errorf("expected gauge3=3.0, got %f", gauges["gauge3"])
	}
}

func TestMemStorage_GetAllCounters(t *testing.T) {
	s := NewMemStorage()
	s.SetCounter("counter1", 10)
	s.SetCounter("counter2", 20)
	s.SetCounter("counter3", 30)

	counters := s.GetAllCounters()
	if len(counters) != 3 {
		t.Errorf("expected 3 counters, got %d", len(counters))
	}
	if counters["counter1"] != 10 {
		t.Errorf("expected counter1=10, got %d", counters["counter1"])
	}
	if counters["counter2"] != 20 {
		t.Errorf("expected counter2=20, got %d", counters["counter2"])
	}
	if counters["counter3"] != 30 {
		t.Errorf("expected counter3=30, got %d", counters["counter3"])
	}
}

func TestMemStorage_GetAllGauges_Empty(t *testing.T) {
	s := NewMemStorage()
	gauges := s.GetAllGauges()
	if len(gauges) != 0 {
		t.Errorf("expected 0 gauges, got %d", len(gauges))
	}
}

func TestMemStorage_GetAllCounters_Empty(t *testing.T) {
	s := NewMemStorage()
	counters := s.GetAllCounters()
	if len(counters) != 0 {
		t.Errorf("expected 0 counters, got %d", len(counters))
	}
}

func TestMemStorage_SetMetricsBatch(t *testing.T) {
	s := NewMemStorage()
	val1 := float64(100.5)
	val2 := float64(200.5)
	delta1 := int64(50)
	delta2 := int64(60)

	metrics := []models.Metric{
		{ID: "gauge1", MType: models.Gauge, Value: &val1},
		{ID: "counter1", MType: models.Counter, Delta: &delta1},
		{ID: "gauge2", MType: models.Gauge, Value: &val2},
		{ID: "counter2", MType: models.Counter, Delta: &delta2},
	}

	err := s.SetMetricsBatch(metrics)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	gaugeVal, ok := s.GetGauge("gauge1")
	if !ok || gaugeVal != 100.5 {
		t.Errorf("expected gauge1=100.5, got %f", gaugeVal)
	}

	counterVal, ok := s.GetCounter("counter1")
	if !ok || counterVal != 50 {
		t.Errorf("expected counter1=50, got %d", counterVal)
	}
}

func TestMemStorage_SetMetricsBatch_AccumulateCounters(t *testing.T) {
	s := NewMemStorage()
	delta1 := int64(50)
	delta2 := int64(60)

	metrics := []models.Metric{
		{ID: "counter1", MType: models.Counter, Delta: &delta1},
	}
	s.SetMetricsBatch(metrics)

	metrics = []models.Metric{
		{ID: "counter1", MType: models.Counter, Delta: &delta2},
	}
	s.SetMetricsBatch(metrics)

	counterVal, _ := s.GetCounter("counter1")
	if counterVal != 110 {
		t.Errorf("expected counter1=110, got %d", counterVal)
	}
}

func TestMemStorage_SetMetricsBatch_Empty(t *testing.T) {
	s := NewMemStorage()
	err := s.SetMetricsBatch([]models.Metric{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMemStorage_SetGauge_Concurrent(t *testing.T) {
	s := NewMemStorage()
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(n int) {
			s.SetGauge("concurrent_gauge", float64(n))
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	_, ok := s.GetGauge("concurrent_gauge")
	if !ok {
		t.Error("expected gauge to exist after concurrent writes")
	}
}

func TestMemStorage_SetCounter_Concurrent(t *testing.T) {
	s := NewMemStorage()
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(n int) {
			s.SetCounter("concurrent_counter", int64(n))
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	_, ok := s.GetCounter("concurrent_counter")
	if !ok {
		t.Error("expected counter to exist after concurrent writes")
	}
}

func TestNewMemStorage(t *testing.T) {
	s := NewMemStorage()
	if s == nil {
		t.Error("expected non-nil storage")
	}
	// Check that maps are accessible
	gauges := s.GetAllGauges()
	if gauges == nil {
		t.Error("expected gauges map to be non-nil")
	}
	counters := s.GetAllCounters()
	if counters == nil {
		t.Error("expected counters map to be non-nil")
	}
}
