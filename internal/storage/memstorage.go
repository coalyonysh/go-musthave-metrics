package storage

import (
	"sync"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
)

type Storage interface {
	SetGauge(name string, value float64)
	SetCounter(name string, value int64)
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
	GetAllGauges() map[string]float64
	GetAllCounters() map[string]int64
	SetMetricsBatch(metrics []models.Metric) error
}

type MemStorage struct {
	mu       sync.RWMutex
	gauges   map[string]float64
	counters map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (s *MemStorage) SetGauge(name string, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gauges[name] = value
}

func (s *MemStorage) SetCounter(name string, value int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// counter должен накапливать значения
	s.counters[name] += value
}

func (s *MemStorage) GetGauge(name string) (float64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.gauges[name]
	return value, exists
}

func (s *MemStorage) GetCounter(name string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.counters[name]
	return value, exists
}

func (s *MemStorage) GetAllGauges() map[string]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Создаем копию карты для безопасного возврата
	// Предварительно выделяем память для нужного количества элементов
	result := make(map[string]float64, len(s.gauges))
	for k, v := range s.gauges {
		result[k] = v
	}
	return result
}

func (s *MemStorage) SetMetricsBatch(metrics []models.Metric) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if metric.Value != nil {
				s.gauges[metric.ID] = *metric.Value
			}
		case models.Counter:
			if metric.Delta != nil {
				s.counters[metric.ID] += *metric.Delta
			}
		}
	}
	return nil
}

func (s *MemStorage) GetAllCounters() map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Создаем копию карты для безопасного возврата
	// Предварительно выделяем память для нужного количества элементов
	result := make(map[string]int64, len(s.counters))
	for k, v := range s.counters {
		result[k] = v
	}
	return result
}
