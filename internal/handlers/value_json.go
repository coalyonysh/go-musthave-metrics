package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
	"github.com/coalyonysh/go-musthave-metrics/internal/storage"
)

type ValueJSONHandler struct {
	storage storage.Storage
}

func NewValueJSONHandler(storage storage.Storage) *ValueJSONHandler {
	return &ValueJSONHandler{
		storage: storage,
	}
}

func (h *ValueJSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем Content-Type
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	// Декодируем JSON из тела запроса
	var metric models.Metric
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&metric); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Валидация обязательных полей
	if metric.ID == "" {
		http.Error(w, "Missing required field: id", http.StatusBadRequest)
		return
	}

	if metric.MType == "" {
		http.Error(w, "Missing required field: type", http.StatusBadRequest)
		return
	}

	// Получаем значение метрики в зависимости от типа
	var responseMetric models.Metric
	responseMetric.ID = metric.ID
	responseMetric.MType = metric.MType

	switch metric.MType {
	case models.Gauge:
		value, exists := h.storage.GetGauge(metric.ID)
		if !exists {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		responseMetric.Value = &value

	case models.Counter:
		value, exists := h.storage.GetCounter(metric.ID)
		if !exists {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		responseMetric.Delta = &value

	default:
		http.Error(w, fmt.Sprintf("Invalid metric type: %s", metric.MType), http.StatusBadRequest)
		return
	}

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Возвращаем метрику в JSON формате
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(responseMetric); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}
