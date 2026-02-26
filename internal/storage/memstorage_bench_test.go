package storage

import (
	"testing"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
)

func BenchmarkMemStorage_SetGauge(b *testing.B) {
	s := NewMemStorage()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.SetGauge("test_metric", float64(i))
	}
}

func BenchmarkMemStorage_SetCounter(b *testing.B) {
	s := NewMemStorage()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.SetCounter("test_counter", int64(i))
	}
}

func BenchmarkMemStorage_GetGauge(b *testing.B) {
	s := NewMemStorage()
	for i := 0; i < 1000; i++ {
		s.SetGauge("test_metric", float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.GetGauge("test_metric")
	}
}

func BenchmarkMemStorage_GetCounter(b *testing.B) {
	s := NewMemStorage()
	for i := 0; i < 1000; i++ {
		s.SetCounter("test_counter", int64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.GetCounter("test_counter")
	}
}

func BenchmarkMemStorage_GetAllGauges(b *testing.B) {
	s := NewMemStorage()
	for i := 0; i < 1000; i++ {
		s.SetGauge("test_metric_"+string(rune(i)), float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.GetAllGauges()
	}
}

func BenchmarkMemStorage_GetAllCounters(b *testing.B) {
	s := NewMemStorage()
	for i := 0; i < 1000; i++ {
		s.SetCounter("test_counter_"+string(rune(i)), int64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.GetAllCounters()
	}
}

func BenchmarkMemStorage_SetMetricsBatch(b *testing.B) {
	s := NewMemStorage()
	metrics := make([]models.Metric, 100)
	for i := 0; i < 100; i++ {
		val := float64(i)
		delta := int64(i)
		metrics[i] = models.Metric{
			ID:    "metric_" + string(rune(i)),
			MType: models.Gauge,
			Value: &val,
		}
		if i%2 == 0 {
			metrics[i].MType = models.Counter
			metrics[i].Delta = &delta
			metrics[i].Value = nil
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.SetMetricsBatch(metrics)
	}
}

// Benchmark for concurrent operations
func BenchmarkMemStorage_ConcurrentSetGauge(b *testing.B) {
	s := NewMemStorage()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			s.SetGauge("test_metric", float64(i))
			i++
		}
	})
}

func BenchmarkMemStorage_ConcurrentGetGauge(b *testing.B) {
	s := NewMemStorage()
	for i := 0; i < 100; i++ {
		s.SetGauge("test_metric", float64(i))
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s.GetGauge("test_metric")
		}
	})
}
