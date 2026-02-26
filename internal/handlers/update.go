package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
	"github.com/coalyonysh/go-musthave-metrics/internal/service"
	"github.com/coalyonysh/go-musthave-metrics/internal/storage"
	"github.com/gorilla/mux"
)

// UpdateHandler обрабатывает запросы на обновление метрик через URL.
// Пример запроса: POST /update/gauge/metric_name/123.45
type UpdateHandler struct {
	storage      storage.Storage
	auditService *service.AuditService
}

// NewUpdateHandler создаёт новый обработчик обновления метрик.
// Принимает storage для сохранения метрик и опциональный auditService для логирования.
func NewUpdateHandler(storage storage.Storage, auditService *service.AuditService) *UpdateHandler {
	return &UpdateHandler{
		storage:      storage,
		auditService: auditService,
	}
}

// ServeHTTP обрабатывает HTTP запрос на обновление метрики.
// Ожидает POST запрос с параметрами type, name и value в пути URL.
func (h *UpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем параметры из URL через gorilla/mux
	vars := mux.Vars(r)
	metricType := vars["type"]
	metricName := vars["name"]
	metricValue := vars["value"]

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

	// Аудит после успешной обработки
	ip := getClientIP(r)
	h.auditService.Log([]string{metricName}, ip)

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
