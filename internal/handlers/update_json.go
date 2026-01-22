package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
	"github.com/coalyonysh/go-musthave-metrics/internal/service"
	"github.com/coalyonysh/go-musthave-metrics/internal/storage"
	"github.com/coalyonysh/go-musthave-metrics/pkg/signature"
)

type UpdateJSONHandler struct {
	storage      storage.Storage
	key          string
	auditService *service.AuditService
}

func NewUpdateJSONHandler(storage storage.Storage, key string, auditService *service.AuditService) *UpdateJSONHandler {
	return &UpdateJSONHandler{
		storage:      storage,
		key:          key,
		auditService: auditService,
	}
}

func (h *UpdateJSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем Content-Type (может содержать дополнительные параметры, например charset)
	// Если Content-Type не указан, пытаемся обработать как JSON
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(strings.ToLower(contentType), "application/json") {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	// Читаем тело запроса
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
		return
	}

	// Проверяем хеш, если ключ задан
	if h.key != "" {
		providedHash := r.Header.Get("HashSHA256")
		if !signature.VerifyHash(bodyBytes, h.key, providedHash) {
			http.Error(w, "Invalid hash", http.StatusBadRequest)
			return
		}
	}

	// Декодируем JSON из тела запроса
	var metric models.Metric
	decoder := json.NewDecoder(bytes.NewReader(bodyBytes))
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

	// Аудит после успешной обработки
	ip := getClientIP(r)
	h.auditService.Log([]string{metric.ID}, ip)

	// Сериализуем ответ
	responseBytes, err := json.Marshal(responseMetric)
	if err != nil {
		log.Printf("Failed to marshal response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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

	// Возвращаем обновленную метрику в JSON формате
	if _, err := w.Write(responseBytes); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}
