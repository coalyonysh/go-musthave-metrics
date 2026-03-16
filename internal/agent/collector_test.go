package agent

import (
	"testing"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
)

func TestMetricsCollector_CollectGopsutilMetrics(t *testing.T) {
	collector := NewMetricsCollector()

	metrics := collector.CollectGopsutilMetrics()

	// Should return some metrics (may be empty on some systems)
	// Just verify the function runs without panic
	_ = metrics
}

func TestMetricsCollector_PollCountIncrements(t *testing.T) {
	collector := NewMetricsCollector()

	if collector.pollCount != 0 {
		t.Errorf("expected initial pollCount to be 0, got %d", collector.pollCount)
	}

	collector.CollectMetrics()

	if collector.pollCount != 1 {
		t.Errorf("expected pollCount to be 1 after first collect, got %d", collector.pollCount)
	}

	collector.CollectMetrics()

	if collector.pollCount != 2 {
		t.Errorf("expected pollCount to be 2 after second collect, got %d", collector.pollCount)
	}
}

func TestMetricsCollector_AllRuntimeMetrics(t *testing.T) {
	collector := NewMetricsCollector()

	metrics := collector.CollectMetrics()

	// Check that we have the expected number of metrics
	// Runtime metrics + PollCount + RandomValue
	if len(metrics) < 25 {
		t.Errorf("expected at least 25 runtime metrics, got %d", len(metrics))
	}

	// Check for specific metrics
	metricIDs := make(map[string]bool)
	for _, m := range metrics {
		metricIDs[m.ID] = true
	}

	// Check for important runtime metrics
	importantMetrics := []string{
		"Alloc", "TotalAlloc", "HeapAlloc", "Sys",
		"PollCount", "RandomValue",
	}

	for _, id := range importantMetrics {
		if !metricIDs[id] {
			t.Errorf("expected metric %s to be present", id)
		}
	}
}

func TestMetricsCollector_MetricTypes(t *testing.T) {
	collector := NewMetricsCollector()

	metrics := collector.CollectMetrics()

	hasGauge := false
	hasCounter := false

	for _, m := range metrics {
		if m.MType == models.Gauge {
			hasGauge = true
			if m.Value == nil {
				t.Error("gauge metric should have Value")
			}
		}
		if m.MType == models.Counter {
			hasCounter = true
			if m.Delta == nil {
				t.Error("counter metric should have Delta")
			}
		}
	}

	if !hasGauge {
		t.Error("expected at least one gauge metric")
	}
	if !hasCounter {
		t.Error("expected at least one counter metric")
	}
}
