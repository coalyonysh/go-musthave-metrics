package storage

import (
	"sync"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
)

// Storage определяет интерфейс хранилища метрик.
// Интерфейс позволяет хранить и извлекать метрики типа gauge и counter.
type Storage interface {
	SetGauge(name string, value float64)
	SetCounter(name string, value int64)
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
	GetAllGauges() map[string]float64
	GetAllCounters() map[string]int64
	SetMetricsBatch(metrics []models.Metric) error
}

// MemStorage представляет in-memory хранилище метрик.
// Использует sync.RWMutex для безопасного доступа из горутин.
type MemStorage struct {
	mu       sync.RWMutex
	gauges   map[string]float64
	counters map[string]int64
}

// NewMemStorage создаёт новое in-memory хранилище.
// Возвращает инициализированный MemStorage с пустыми картами для gauge и counter метрик.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

// SetGauge устанавливает значение gauge метрики.
// Gauge метрика может принимать произвольное дробное значение.
func (s *MemStorage) SetGauge(name string, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gauges[name] = value
}

// SetCounter устанавливает/увеличивает значение counter метрики.
// Counter метрика аккумулирует значения (суммирует).
func (s *MemStorage) SetCounter(name string, value int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// counter должен накапливать значения
	s.counters[name] += value
}

// GetGauge возвращает значение gauge метрики по имени.
// Возвращает значение и флаг существования.
func (s *MemStorage) GetGauge(name string) (float64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.gauges[name]
	return value, exists
}

// GetCounter возвращает значение counter метрики по имени.
// Возвращает значение и флаг существования.
func (s *MemStorage) GetCounter(name string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.counters[name]
	return value, exists
}

// GetAllGauges возвращает копию всех gauge метрик.
// Возвращает новую карту для избежания race condition.
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

// SetMetricsBatch устанавливает пакет метрик за один вызов.
// Более эффективно чем установка по одной метрике.
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

// GetAllCounters возвращает копию всех counter метрик.
// Возвращает новую карту для избежания race condition.
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
