package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
	"github.com/coalyonysh/go-musthave-metrics/internal/service"
	"github.com/coalyonysh/go-musthave-metrics/internal/storage"
	"github.com/gorilla/mux"
)

func TestUpdateHandler_ServeHTTP_Gauge(t *testing.T) {
	memStorage := storage.NewMemStorage()
	auditService := service.NewAuditService()
	handler := NewUpdateHandler(memStorage, auditService)

	router := mux.NewRouter()
	router.Handle("/update/{type}/{name}/{value}", handler).Methods("POST")

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/test_gauge/123.45", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	val, ok := memStorage.GetGauge("test_gauge")
	if !ok || val != 123.45 {
		t.Errorf("expected gauge=123.45, got %f", val)
	}
}

func TestUpdateHandler_ServeHTTP_Counter(t *testing.T) {
	memStorage := storage.NewMemStorage()
	auditService := service.NewAuditService()
	handler := NewUpdateHandler(memStorage, auditService)

	router := mux.NewRouter()
	router.Handle("/update/{type}/{name}/{value}", handler).Methods("POST")

	req := httptest.NewRequest(http.MethodPost, "/update/counter/test_counter/100", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	val, ok := memStorage.GetCounter("test_counter")
	if !ok || val != 100 {
		t.Errorf("expected counter=100, got %d", val)
	}
}

func TestUpdateHandler_ServeHTTP_InvalidType(t *testing.T) {
	memStorage := storage.NewMemStorage()
	auditService := service.NewAuditService()
	handler := NewUpdateHandler(memStorage, auditService)

	router := mux.NewRouter()
	router.Handle("/update/{type}/{name}/{value}", handler).Methods("POST")

	req := httptest.NewRequest(http.MethodPost, "/update/invalid/test_metric/100", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestUpdateHandler_ServeHTTP_InvalidValue(t *testing.T) {
	memStorage := storage.NewMemStorage()
	auditService := service.NewAuditService()
	handler := NewUpdateHandler(memStorage, auditService)

	router := mux.NewRouter()
	router.Handle("/update/{type}/{name}/{value}", handler).Methods("POST")

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/test_gauge/not_a_number", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestValueHandler_ServeHTTP_Gauge(t *testing.T) {
	memStorage := storage.NewMemStorage()
	memStorage.SetGauge("test_gauge", 123.45)
	handler := NewValueHandler(memStorage)

	router := mux.NewRouter()
	router.Handle("/value/{type}/{name}", handler).Methods("GET")

	req := httptest.NewRequest(http.MethodGet, "/value/gauge/test_gauge", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "123.45" {
		t.Errorf("expected body '123.45', got '%s'", w.Body.String())
	}
}

func TestValueHandler_ServeHTTP_Counter(t *testing.T) {
	memStorage := storage.NewMemStorage()
	memStorage.SetCounter("test_counter", 100)
	handler := NewValueHandler(memStorage)

	router := mux.NewRouter()
	router.Handle("/value/{type}/{name}", handler).Methods("GET")

	req := httptest.NewRequest(http.MethodGet, "/value/counter/test_counter", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "100" {
		t.Errorf("expected body '100', got '%s'", w.Body.String())
	}
}

func TestValueHandler_ServeHTTP_NotFound(t *testing.T) {
	memStorage := storage.NewMemStorage()
	handler := NewValueHandler(memStorage)

	router := mux.NewRouter()
	router.Handle("/value/{type}/{name}", handler).Methods("GET")

	req := httptest.NewRequest(http.MethodGet, "/value/gauge/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestIndexHandler_ServeHTTP(t *testing.T) {
	memStorage := storage.NewMemStorage()
	memStorage.SetGauge("gauge1", 100.5)
	memStorage.SetCounter("counter1", 50)
	handler := NewIndexHandler(memStorage)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if len(body) == 0 {
		t.Error("expected non-empty body")
	}
}

func TestUpdateJSONHandler_ServeHTTP(t *testing.T) {
	memStorage := storage.NewMemStorage()
	auditService := service.NewAuditService()
	handler := NewUpdateJSONHandler(memStorage, "", auditService)

	val := float64(123.45)
	metric := models.Metric{
		ID:    "test_gauge",
		MType: models.Gauge,
		Value: &val,
	}
	body, _ := json.Marshal(metric)

	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	gaugeVal, ok := memStorage.GetGauge("test_gauge")
	if !ok || gaugeVal != 123.45 {
		t.Errorf("expected gauge=123.45, got %f", gaugeVal)
	}
}

func TestUpdateJSONHandler_ServeHTTP_WithHash(t *testing.T) {
	memStorage := storage.NewMemStorage()
	auditService := service.NewAuditService()
	handler := NewUpdateJSONHandler(memStorage, "test_key", auditService)

	val := float64(123.45)
	metric := models.Metric{
		ID:    "test_gauge",
		MType: models.Gauge,
		Value: &val,
	}
	body, _ := json.Marshal(metric)

	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HashSHA256", "invalid_hash")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid hash, got %d", w.Code)
	}
}

func TestUpdatesHandler_ServeHTTP(t *testing.T) {
	memStorage := storage.NewMemStorage()
	auditService := service.NewAuditService()
	handler := NewUpdatesHandler(memStorage, "", auditService)

	val1 := float64(100.5)
	val2 := float64(200.5)
	delta1 := int64(50)
	metrics := []models.Metric{
		{ID: "gauge1", MType: models.Gauge, Value: &val1},
		{ID: "counter1", MType: models.Counter, Delta: &delta1},
		{ID: "gauge2", MType: models.Gauge, Value: &val2},
	}
	body, _ := json.Marshal(metrics)

	req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	gaugeVal, _ := memStorage.GetGauge("gauge1")
	if gaugeVal != 100.5 {
		t.Errorf("expected gauge1=100.5, got %f", gaugeVal)
	}

	counterVal, _ := memStorage.GetCounter("counter1")
	if counterVal != 50 {
		t.Errorf("expected counter1=50, got %d", counterVal)
	}
}

func TestUpdatesHandler_ServeHTTP_InvalidJSON(t *testing.T) {
	memStorage := storage.NewMemStorage()
	auditService := service.NewAuditService()
	handler := NewUpdatesHandler(memStorage, "", auditService)

	req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestPingHandler_ServeHTTP_NoDB(t *testing.T) {
	handler := NewPingHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestValueJSONHandler_ServeHTTP_Gauge(t *testing.T) {
	memStorage := storage.NewMemStorage()
	memStorage.SetGauge("test_gauge", 123.45)
	handler := NewValueJSONHandler(memStorage, "")

	metric := models.Metric{ID: "test_gauge", MType: models.Gauge}
	body, _ := json.Marshal(metric)

	req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var result models.Metric
	json.Unmarshal(w.Body.Bytes(), &result)
	if result.Value == nil || *result.Value != 123.45 {
		t.Errorf("expected value=123.45, got %v", result.Value)
	}
}

func TestGetClientIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4")

	ip := getClientIP(req)
	if ip != "1.2.3.4" {
		t.Errorf("expected X-Forwarded-For, got %s", ip)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "5.6.7.8:1234"

	ip2 := getClientIP(req2)
	if ip2 != "5.6.7.8" {
		t.Errorf("expected remote addr, got %s", ip2)
	}
}

type mockStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (m *mockStorage) SetGauge(name string, value float64) {
	m.gauges[name] = value
}

func (m *mockStorage) SetCounter(name string, value int64) {
	m.counters[name] += value
}

func (m *mockStorage) GetGauge(name string) (float64, bool) {
	val, ok := m.gauges[name]
	return val, ok
}

func (m *mockStorage) GetCounter(name string) (int64, bool) {
	val, ok := m.counters[name]
	return val, ok
}

func (m *mockStorage) GetAllGauges() map[string]float64 {
	return m.gauges
}

func (m *mockStorage) GetAllCounters() map[string]int64 {
	return m.counters
}

func (m *mockStorage) SetMetricsBatch(metrics []models.Metric) error {
	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if metric.Value != nil {
				m.SetGauge(metric.ID, *metric.Value)
			}
		case models.Counter:
			if metric.Delta != nil {
				m.SetCounter(metric.ID, *metric.Delta)
			}
		}
	}
	return nil
}

// Test with mock
func TestHandlers_WithMockStorage(t *testing.T) {
	s := newMockStorage()
	s.SetGauge("test", 10.0)

	val, ok := s.GetGauge("test")
	if !ok || val != 10.0 {
		t.Error("mock storage failed")
	}
}

func init() {
	// For context key types
	_ = context.Background()
}
