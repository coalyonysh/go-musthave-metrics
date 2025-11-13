package handlers

import (
	"fmt"
	"net/http"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
	"github.com/coalyonysh/go-musthave-metrics/internal/storage"
	"github.com/gorilla/mux"
)

type ValueHandler struct {
	storage storage.Storage
}

func NewValueHandler(storage storage.Storage) *ValueHandler {
	return &ValueHandler{
		storage: storage,
	}
}

func (h *ValueHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем параметры из URL через gorilla/mux
	vars := mux.Vars(r)
	metricType := vars["type"]
	metricName := vars["name"]

	// Получаем значение метрики в зависимости от типа
	switch metricType {
	case models.Gauge:
		value, exists := h.storage.GetGauge(metricName)
		if !exists {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%g", value)

	case models.Counter:
		value, exists := h.storage.GetCounter(metricName)
		if !exists {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%d", value)

	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}
}
