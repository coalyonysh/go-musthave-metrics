package storage

import (
	"os"
	"runtime/pprof"
	"testing"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
)

const profilePathResult = "/Users/salaga/Documents/_PROJECTS_/GOshkin_curss/go-musthave-metrics/profiles/result.pprof"

// TestProfileMemoryResult создаёт профиль памяти для хранилища после оптимизации
func TestProfileMemoryResult(t *testing.T) {
	// Запускаем тест с нагрузкой
	MemStorageProfile()

	// Создаём профиль памяти ПОСЛЕ нагрузки
	f, err := os.Create(profilePathResult)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if err := pprof.WriteHeapProfile(f); err != nil {
		t.Fatal(err)
	}

	t.Logf("Profile saved to %s", profilePathResult)
}

// MemStorageProfile профилирует основные операции хранилища
func MemStorageProfile() {
	s := NewMemStorage()

	// Заполняем хранилище данными
	for i := 0; i < 1000; i++ {
		s.SetGauge("gauge_"+string(rune(i%256)), float64(i))
		s.SetCounter("counter_"+string(rune(i%256)), int64(i))
	}

	// Много операций чтения/записи
	for i := 0; i < 10000; i++ {
		s.SetGauge("test_gauge", float64(i))
		s.GetGauge("test_gauge")
		s.SetCounter("test_counter", int64(i))
		s.GetCounter("test_counter")
	}

	// Много GetAll операций (высокие аллокации)
	for i := 0; i < 1000; i++ {
		s.GetAllGauges()
		s.GetAllCounters()
	}

	// Пакетная запись
	metrics := make([]models.Metric, 100)
	for i := 0; i < 100; i++ {
		val := float64(i)
		delta := int64(i)
		metrics[i] = models.Metric{
			ID:    "batch_metric_" + string(rune(i)),
			MType: models.Gauge,
			Value: &val,
		}
		if i%2 == 0 {
			metrics[i].MType = models.Counter
			metrics[i].Delta = &delta
			metrics[i].Value = nil
		}
	}
	for i := 0; i < 1000; i++ {
		s.SetMetricsBatch(metrics)
	}
}
