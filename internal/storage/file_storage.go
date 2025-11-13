package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
)

// SaveMetrics сохраняет метрики в JSON файл
func SaveMetrics(s Storage, filePath string) error {
	if filePath == "" {
		return nil // Если путь не указан, не сохраняем
	}

	// Получаем все метрики
	gauges := s.GetAllGauges()
	counters := s.GetAllCounters()

	// Преобразуем в формат для сохранения
	var metrics []models.Metric

	// Добавляем gauge метрики
	for name, value := range gauges {
		metrics = append(metrics, models.Metric{
			ID:    name,
			MType: models.Gauge,
			Value: &value,
		})
	}

	// Добавляем counter метрики
	for name, delta := range counters {
		metrics = append(metrics, models.Metric{
			ID:    name,
			MType: models.Counter,
			Delta: &delta,
		})
	}

	// Сериализуем в JSON
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	// Сохраняем в файл
	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// LoadMetrics загружает метрики из JSON файла
func LoadMetrics(s Storage, filePath string) error {
	if filePath == "" {
		return nil // Если путь не указан, не загружаем
	}

	// Проверяем существование файла
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // Файл не существует, это нормально
	}

	// Читаем файл
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Десериализуем JSON
	var metrics []models.Metric
	if err := json.Unmarshal(data, &metrics); err != nil {
		return fmt.Errorf("failed to unmarshal metrics: %w", err)
	}

	// Загружаем метрики в storage
	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if metric.Value != nil {
				s.SetGauge(metric.ID, *metric.Value)
			}
		case models.Counter:
			if metric.Delta != nil {
				s.SetCounter(metric.ID, *metric.Delta)
			}
		}
	}

	return nil
}

