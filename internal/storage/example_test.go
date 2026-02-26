package storage

import (
	"fmt"
	"sort"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
)

// Пример работы с MemStorage - базовое использование
func ExampleMemStorage_basic() {
	// Создаём новое хранилище
	s := NewMemStorage()

	// Устанавливаем gauge метрику
	s.SetGauge("temperature", 25.5)
	s.SetGauge("humidity", 60.0)

	// Устанавливаем counter метрику
	s.SetCounter("requests", 1)
	s.SetCounter("requests", 2) // Counter аккумулирует значение

	// Получаем значения
	temp, _ := s.GetGauge("temperature")
	reqCount, _ := s.GetCounter("requests")

	fmt.Printf("Temperature: %f\n", temp)
	fmt.Printf("Requests: %d\n", reqCount)

	// Output:
	// Temperature: 25.500000
	// Requests: 3
}

// Пример работы с GetAllGauges
func ExampleMemStorage_GetAllGauges() {
	s := NewMemStorage()
	s.SetGauge("cpu", 80.5)
	s.SetGauge("memory", 60.0)
	s.SetGauge("disk", 45.0)

	// Получаем все gauge метрики
	allGauges := s.GetAllGauges()

	fmt.Printf("Number of gauges: %d\n", len(allGauges))
	// Сортируем ключи для детерминированного вывода
	keys := make([]string, 0, len(allGauges))
	for k := range allGauges {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("%s: %f\n", k, allGauges[k])
	}

	// Output:
	// Number of gauges: 3
	// cpu: 80.500000
	// disk: 45.000000
	// memory: 60.000000
}

// Пример работы с GetAllCounters
func ExampleMemStorage_GetAllCounters() {
	s := NewMemStorage()
	s.SetCounter("hits", 10)
	s.SetCounter("visits", 5)

	// Получаем все counter метрики
	allCounters := s.GetAllCounters()

	fmt.Printf("Number of counters: %d\n", len(allCounters))
	// Сортируем ключи для детерминированного вывода
	keys := make([]string, 0, len(allCounters))
	for k := range allCounters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("%s: %d\n", k, allCounters[k])
	}

	// Output:
	// Number of counters: 2
	// hits: 10
	// visits: 5
}

// Пример работы с SetMetricsBatch
func ExampleMemStorage_SetMetricsBatch() {
	s := NewMemStorage()

	// Создаём пакет метрик
	val1 := float64(100.0)
	val2 := float64(200.0)
	delta1 := int64(50)
	delta2 := int64(75)

	metrics := []models.Metric{
		{ID: "gauge1", MType: models.Gauge, Value: &val1},
		{ID: "gauge2", MType: models.Gauge, Value: &val2},
		{ID: "counter1", MType: models.Counter, Delta: &delta1},
		{ID: "counter2", MType: models.Counter, Delta: &delta2},
	}

	// Устанавливаем все метрики за один вызов
	err := s.SetMetricsBatch(metrics)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	g1, _ := s.GetGauge("gauge1")
	g2, _ := s.GetGauge("gauge2")
	c1, _ := s.GetCounter("counter1")
	c2, _ := s.GetCounter("counter2")

	fmt.Printf("gauge1: %f, gauge2: %f\n", g1, g2)
	fmt.Printf("counter1: %d, counter2: %d\n", c1, c2)

	// Output:
	// gauge1: 100.000000, gauge2: 200.000000
	// counter1: 50, counter2: 75
}

// Пример работы с SyncStorage
func ExampleSyncStorage() {
	// Создаём базовое хранилище
	memStorage := NewMemStorage()

	// Флаг для проверки вызова функции сохранения
	saved := false
	saveFunc := func() {
		saved = true
	}

	// Создаём синхронизированное хранилище
	s := NewSyncStorage(memStorage, saveFunc)

	// Устанавливаем метрику - должна вызваться функция сохранения
	s.SetGauge("test_metric", 10.0)

	val, _ := memStorage.GetGauge("test_metric")
	fmt.Printf("Saved: %v\n", saved)
	fmt.Printf("Value: %f\n", val)

	// Output:
	// Saved: true
	// Value: 10.000000
}
