package agent

import (
	"log"
	"math/rand"
	"runtime"
	"strconv"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type MetricsCollector struct {
	pollCount   int64
	randomValue float64
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{}
}

func (mc *MetricsCollector) CollectMetrics() []models.Metric {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	mc.pollCount++
	mc.randomValue = rand.Float64()

	metrics := []models.Metric{
		{ID: "Alloc", MType: models.Gauge, Value: float64Ptr(float64(stats.Alloc))},
		{ID: "BuckHashSys", MType: models.Gauge, Value: float64Ptr(float64(stats.BuckHashSys))},
		{ID: "Frees", MType: models.Gauge, Value: float64Ptr(float64(stats.Frees))},
		{ID: "GCCPUFraction", MType: models.Gauge, Value: float64Ptr(stats.GCCPUFraction)},
		{ID: "GCSys", MType: models.Gauge, Value: float64Ptr(float64(stats.GCSys))},
		{ID: "HeapAlloc", MType: models.Gauge, Value: float64Ptr(float64(stats.HeapAlloc))},
		{ID: "HeapIdle", MType: models.Gauge, Value: float64Ptr(float64(stats.HeapIdle))},
		{ID: "HeapInuse", MType: models.Gauge, Value: float64Ptr(float64(stats.HeapInuse))},
		{ID: "HeapObjects", MType: models.Gauge, Value: float64Ptr(float64(stats.HeapObjects))},
		{ID: "HeapReleased", MType: models.Gauge, Value: float64Ptr(float64(stats.HeapReleased))},
		{ID: "HeapSys", MType: models.Gauge, Value: float64Ptr(float64(stats.HeapSys))},
		{ID: "LastGC", MType: models.Gauge, Value: float64Ptr(float64(stats.LastGC))},
		{ID: "Lookups", MType: models.Gauge, Value: float64Ptr(float64(stats.Lookups))},
		{ID: "MCacheInuse", MType: models.Gauge, Value: float64Ptr(float64(stats.MCacheInuse))},
		{ID: "MCacheSys", MType: models.Gauge, Value: float64Ptr(float64(stats.MCacheSys))},
		{ID: "MSpanInuse", MType: models.Gauge, Value: float64Ptr(float64(stats.MSpanInuse))},
		{ID: "MSpanSys", MType: models.Gauge, Value: float64Ptr(float64(stats.MSpanSys))},
		{ID: "Mallocs", MType: models.Gauge, Value: float64Ptr(float64(stats.Mallocs))},
		{ID: "NextGC", MType: models.Gauge, Value: float64Ptr(float64(stats.NextGC))},
		{ID: "NumForcedGC", MType: models.Gauge, Value: float64Ptr(float64(stats.NumForcedGC))},
		{ID: "NumGC", MType: models.Gauge, Value: float64Ptr(float64(stats.NumGC))},
		{ID: "OtherSys", MType: models.Gauge, Value: float64Ptr(float64(stats.OtherSys))},
		{ID: "PauseTotalNs", MType: models.Gauge, Value: float64Ptr(float64(stats.PauseTotalNs))},
		{ID: "StackInuse", MType: models.Gauge, Value: float64Ptr(float64(stats.StackInuse))},
		{ID: "StackSys", MType: models.Gauge, Value: float64Ptr(float64(stats.StackSys))},
		{ID: "Sys", MType: models.Gauge, Value: float64Ptr(float64(stats.Sys))},
		{ID: "TotalAlloc", MType: models.Gauge, Value: float64Ptr(float64(stats.TotalAlloc))},
		{ID: "PollCount", MType: models.Counter, Delta: int64Ptr(1)},
		{ID: "RandomValue", MType: models.Gauge, Value: float64Ptr(mc.randomValue)},
	}

	return metrics
}

func (mc *MetricsCollector) CollectGopsutilMetrics() []models.Metric {
	var metrics []models.Metric

	// TotalMemory
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("Failed to get virtual memory: %v", err)
	} else {
		metrics = append(metrics, models.Metric{
			ID:    "TotalMemory",
			MType: models.Gauge,
			Value: float64Ptr(float64(vmStat.Total)),
		})
		metrics = append(metrics, models.Metric{
			ID:    "FreeMemory",
			MType: models.Gauge,
			Value: float64Ptr(float64(vmStat.Available)), // Available is free + buffers/cache
		})
	}

	// CPUutilization
	percents, err := cpu.Percent(0, true)
	if err != nil {
		log.Printf("Failed to get CPU percent: %v", err)
	} else {
		for i, percent := range percents {
			metrics = append(metrics, models.Metric{
				ID:    "CPUutilization" + strconv.Itoa(i+1),
				MType: models.Gauge,
				Value: float64Ptr(percent),
			})
		}
	}

	return metrics
}

// Вспомогательные функции для создания указателей
func float64Ptr(f float64) *float64 {
	return &f
}

func int64Ptr(i int64) *int64 {
	return &i
}
