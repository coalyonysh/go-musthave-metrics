package agent

import (
	"testing"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
)

func TestMetricsCollector_CollectMetrics(t *testing.T) {
	collector := NewMetricsCollector()

	metrics := collector.CollectMetrics()

	// Проверяем, что все метрики собраны
	expectedMetrics := []string{
		"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys",
		"HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased",
		"HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
		"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
		"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys",
		"Sys", "TotalAlloc", "PollCount", "RandomValue",
	}

	if len(metrics) != len(expectedMetrics) {
		t.Errorf("Expected %d metrics, got %d", len(expectedMetrics), len(metrics))
	}

	// Проверяем наличие всех ожидаемых метрик
	metricMap := make(map[string]models.Metric)
	for _, metric := range metrics {
		metricMap[metric.ID] = metric
	}

	for _, expectedID := range expectedMetrics {
		if _, exists := metricMap[expectedID]; !exists {
			t.Errorf("Expected metric %s not found", expectedID)
		}
	}

	// Проверяем типы метрик
	for _, metric := range metrics {
		if metric.ID == "PollCount" {
			if metric.MType != models.Counter {
				t.Errorf("PollCount should be Counter type, got %s", metric.MType)
			}
			if metric.Delta == nil {
				t.Error("PollCount should have Delta value")
			}
		} else if metric.ID == "RandomValue" {
			if metric.MType != models.Gauge {
				t.Errorf("RandomValue should be Gauge type, got %s", metric.MType)
			}
			if metric.Value == nil {
				t.Error("RandomValue should have Value")
			}
		} else {
			if metric.MType != models.Gauge {
				t.Errorf("Metric %s should be Gauge type, got %s", metric.ID, metric.MType)
			}
			if metric.Value == nil {
				t.Errorf("Metric %s should have Value", metric.ID)
			}
		}
	}
}

func TestMetricsCollector_PollCountIncrement(t *testing.T) {
	collector := NewMetricsCollector()

	// Первый вызов
	metrics1 := collector.CollectMetrics()
	pollCount1 := findMetricByID(metrics1, "PollCount")
	if pollCount1 == nil || pollCount1.Delta == nil {
		t.Fatal("PollCount not found or has no Delta")
	}

	// Второй вызов
	metrics2 := collector.CollectMetrics()
	pollCount2 := findMetricByID(metrics2, "PollCount")
	if pollCount2 == nil || pollCount2.Delta == nil {
		t.Fatal("PollCount not found or has no Delta")
	}

	// PollCount должен увеличиваться на 1 каждый раз
	if *pollCount2.Delta != 1 {
		t.Errorf("Expected PollCount delta to be 1, got %d", *pollCount2.Delta)
	}
}

func TestMetricsCollector_RandomValue(t *testing.T) {
	collector := NewMetricsCollector()

	// Собираем метрики несколько раз
	values := make([]float64, 5)
	for i := 0; i < 5; i++ {
		metrics := collector.CollectMetrics()
		randomValue := findMetricByID(metrics, "RandomValue")
		if randomValue == nil || randomValue.Value == nil {
			t.Fatal("RandomValue not found or has no Value")
		}
		values[i] = *randomValue.Value
	}

	// Проверяем, что значения разные (с высокой вероятностью)
	allSame := true
	for i := 1; i < len(values); i++ {
		if values[i] != values[0] {
			allSame = false
			break
		}
	}

	if allSame {
		t.Error("RandomValue should generate different values")
	}
}

func findMetricByID(metrics []models.Metric, id string) *models.Metric {
	for _, metric := range metrics {
		if metric.ID == id {
			return &metric
		}
	}
	return nil
}

func TestMetricsCollector_NewMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector()

	if collector == nil {
		t.Fatal("NewMetricsCollector returned nil")
	}

	// Check initial state
	if collector.pollCount != 0 {
		t.Errorf("Expected initial pollCount to be 0, got %d", collector.pollCount)
	}

	if collector.randomValue != 0.0 {
		t.Errorf("Expected initial randomValue to be 0.0, got %f", collector.randomValue)
	}
}
