package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
	"github.com/coalyonysh/go-musthave-metrics/internal/storage"
)

type UpdateJSONHandler struct {
	storage storage.Storage
}

func NewUpdateJSONHandler(storage storage.Storage) *UpdateJSONHandler {
	return &UpdateJSONHandler{
		storage: storage,
	}
}

func (h *UpdateJSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем Content-Type (может содержать дополнительные параметры, например charset)
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(strings.ToLower(contentType), "application/json") {
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

	// Обрабатываем метрику в зависимости от типа
	var responseMetric models.Metric
	responseMetric.ID = metric.ID
	responseMetric.MType = metric.MType

	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			http.Error(w, "Missing required field: value for gauge metric", http.StatusBadRequest)
			return
		}
		h.storage.SetGauge(metric.ID, *metric.Value)
		// Получаем актуальное значение из storage
		value, _ := h.storage.GetGauge(metric.ID)
		responseMetric.Value = &value
		log.Printf("Metric of type GAUGE with name %s set with value %g", metric.ID, *metric.Value)

	case models.Counter:
		if metric.Delta == nil {
			http.Error(w, "Missing required field: delta for counter metric", http.StatusBadRequest)
			return
		}
		h.storage.SetCounter(metric.ID, *metric.Delta)
		// Получаем накопленное значение из storage
		delta, _ := h.storage.GetCounter(metric.ID)
		responseMetric.Delta = &delta
		log.Printf("Metric of type COUNTER with name %s set with delta %d", metric.ID, *metric.Delta)

	default:
		http.Error(w, fmt.Sprintf("Invalid metric type: %s", metric.MType), http.StatusBadRequest)
		return
	}

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Возвращаем обновленную метрику в JSON формате
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(responseMetric); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}
