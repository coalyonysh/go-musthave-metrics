package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
	"github.com/coalyonysh/go-musthave-metrics/internal/storage"
	"github.com/coalyonysh/go-musthave-metrics/pkg/signature"
)

type ValueJSONHandler struct {
	storage storage.Storage
	key     string
}

func NewValueJSONHandler(storage storage.Storage, key string) *ValueJSONHandler {
	return &ValueJSONHandler{
		storage: storage,
		key:     key,
	}
}

func (h *ValueJSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	// Сериализуем ответ
	responseBytes, err := json.Marshal(responseMetric)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal response: %v", err), http.StatusInternalServerError)
		return
	}

	// Вычисляем хеш ответа, если ключ задан
	if h.key != "" {
		hash := signature.CalculateHash(responseBytes, h.key)
		w.Header().Set("HashSHA256", hash)
	}

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Возвращаем метрику в JSON формате
	if _, err := w.Write(responseBytes); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write response: %v", err), http.StatusInternalServerError)
		return
	}
}
