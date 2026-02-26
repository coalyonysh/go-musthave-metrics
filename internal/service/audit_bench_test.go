package service

import (
	"testing"
)

func BenchmarkAuditService_Log(b *testing.B) {
	audit := NewAuditService()
	metrics := []string{"metric1", "metric2", "metric3", "metric4", "metric5"}
	ip := "192.168.1.1"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		audit.Log(metrics, ip)
	}
}

func BenchmarkAuditService_Log_ManyMetrics(b *testing.B) {
	audit := NewAuditService()
	metrics := make([]string, 100)
	for i := 0; i < 100; i++ {
		metrics[i] = "metric_" + string(rune(i))
	}
	ip := "192.168.1.1"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		audit.Log(metrics, ip)
	}
}

func BenchmarkAuditService_Log_NoObservers(b *testing.B) {
	audit := NewAuditService()
	metrics := []string{"metric1", "metric2", "metric3"}
	ip := "192.168.1.1"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		audit.Log(metrics, ip)
	}
}

func BenchmarkAuditService_AddObserver(b *testing.B) {
	observer := NewFileObserver("/tmp/test_audit.log")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		audit := NewAuditService()
		audit.AddObserver(observer)
	}
}
