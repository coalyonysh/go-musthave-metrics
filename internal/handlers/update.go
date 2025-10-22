package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
	"github.com/coalyonysh/go-musthave-metrics/internal/storage"
)

type UpdateHandler struct {
	storage storage.Storage
}

func NewUpdateHandler(storage storage.Storage) *UpdateHandler {
	return &UpdateHandler{
		storage: storage,
	}
}

func (h *UpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Парсим URL path
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	// Проверяем формат пути: /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	if len(parts) != 4 || parts[0] != "update" {
		http.Error(w, "Metric name is required", http.StatusNotFound)
		return
	}

	metricType := parts[1]
	metricName := parts[2]
	metricValue := parts[3]

	// Обрабатываем метрику в зависимости от типа
	switch metricType {
	case models.Gauge:
		err := h.handleGauge(metricName, metricValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

	case models.Counter:
		err := h.handleCounter(metricName, metricValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleGauge обрабатывает gauge метрики
func (h *UpdateHandler) handleGauge(name, valueStr string) error {
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return fmt.Errorf("invalid gauge value: %v", err)
	}

	h.storage.SetGauge(name, value)
	log.Printf("Metric of type GAUGE with name %s set with value %s", name, valueStr)
	return nil
}

// handleCounter обрабатывает counter метрики
func (h *UpdateHandler) handleCounter(name, valueStr string) error {
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid counter value: %v", err)
	}

	h.storage.SetCounter(name, value)
	log.Printf("Metric of type COUNTER with name %s set with value %s", name, valueStr)
	return nil
}
