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

type UpdatesHandler struct {
	storage storage.Storage
}

func NewUpdatesHandler(storage storage.Storage) *UpdatesHandler {
	return &UpdatesHandler{
		storage: storage,
	}
}

func (h *UpdatesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем Content-Type
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(strings.ToLower(contentType), "application/json") {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	// Декодируем JSON из тела запроса
	var metrics []models.Metric
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&metrics); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Валидация метрик
	for _, metric := range metrics {
		if metric.ID == "" {
			http.Error(w, "Missing required field: id", http.StatusBadRequest)
			return
		}
		if metric.MType == "" {
			http.Error(w, "Missing required field: type", http.StatusBadRequest)
			return
		}
		if metric.MType == models.Gauge && metric.Value == nil {
			http.Error(w, "Missing required field: value for gauge metric", http.StatusBadRequest)
			return
		}
		if metric.MType == models.Counter && metric.Delta == nil {
			http.Error(w, "Missing required field: delta for counter metric", http.StatusBadRequest)
			return
		}
		if metric.MType != models.Gauge && metric.MType != models.Counter {
			http.Error(w, fmt.Sprintf("Invalid metric type: %s", metric.MType), http.StatusBadRequest)
			return
		}
	}

	// Обработка батча
	err := h.storage.SetMetricsBatch(metrics)
	if err != nil {
		log.Printf("Failed to set metrics batch: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Processed %d metrics batch", len(metrics))

	w.WriteHeader(http.StatusOK)
}
