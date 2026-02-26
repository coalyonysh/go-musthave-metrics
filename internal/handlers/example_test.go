package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
	"github.com/coalyonysh/go-musthave-metrics/internal/service"
	"github.com/coalyonysh/go-musthave-metrics/internal/storage"
	"github.com/gorilla/mux"
)

// Пример работы с UpdateHandler - обновление метрики через URL
//
//nolint:errcheck
func ExampleUpdateHandler_ServeHTTP() {
	// Создаём хранилище и обработчик
	memStorage := storage.NewMemStorage()
	auditService := service.NewAuditService()
	handler := NewUpdateHandler(memStorage, auditService)

	// Создаём роутер с маршрутом
	router := mux.NewRouter()
	router.Handle("/update/{type}/{name}/{value}", handler).Methods("POST")

	// Тестируем обновление gauge метрики
	req := httptest.NewRequest(http.MethodPost, "/update/gauge/temperature/25.5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Проверяем результат
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Body: %s\n", w.Body.String())

	// Проверяем значение в хранилище
	val, _ := memStorage.GetGauge("temperature")
	fmt.Printf("Temperature: %f\n", val)

	// Тестируем обновление counter метрики
	req2 := httptest.NewRequest(http.MethodPost, "/update/counter/requests/1", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	counterVal, _ := memStorage.GetCounter("requests")
	fmt.Printf("Requests: %d\n", counterVal)

	// Output:
	// Status: 200
	// Body: OK
	// Temperature: 25.500000
	// Requests: 1
}

// Пример работы с ValueHandler - получение значения метрики
//
//nolint:errcheck
func ExampleValueHandler_ServeHTTP() {
	// Создаём хранилище с тестовыми данными
	memStorage := storage.NewMemStorage()
	memStorage.SetGauge("cpu_usage", 75.5)
	memStorage.SetCounter("request_count", 100)

	// Создаём обработчик
	handler := NewValueHandler(memStorage)

	// Создаём роутер
	router := mux.NewRouter()
	router.Handle("/value/{type}/{name}", handler).Methods("GET")

	// Тестируем получение gauge метрики
	req := httptest.NewRequest(http.MethodGet, "/value/gauge/cpu_usage", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Printf("Gauge - Status: %d, Value: %s\n", w.Code, w.Body.String())

	// Тестируем получение counter метрики
	req2 := httptest.NewRequest(http.MethodGet, "/value/counter/request_count", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	fmt.Printf("Counter - Status: %d, Value: %s\n", w2.Code, w2.Body.String())

	// Output:
	// Gauge - Status: 200, Value: 75.5
	// Counter - Status: 200, Value: 100
}

// Пример работы с UpdateJSONHandler - обновление метрики через JSON
//
//nolint:errcheck
func ExampleUpdateJSONHandler_ServeHTTP() {
	// Создаём хранилище и обработчик
	memStorage := storage.NewMemStorage()
	auditService := service.NewAuditService()
	handler := NewUpdateJSONHandler(memStorage, "", auditService)

	// Создаём JSON запрос
	val := float64(42.0)
	metric := models.Metric{
		ID:    "memory_used",
		MType: models.Gauge,
		Value: &val,
	}
	body, _ := json.Marshal(metric)

	req := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	fmt.Printf("Status: %d\n", w.Code)

	memVal, _ := memStorage.GetGauge("memory_used")
	fmt.Printf("Memory Used: %f\n", memVal)

	// Output:
	// Status: 200
	// Memory Used: 42.000000
}

// Пример работы с UpdatesHandler - пакетное обновление метрик
//
//nolint:errcheck
func ExampleUpdatesHandler_ServeHTTP() {
	// Создаём хранилище и обработчик
	memStorage := storage.NewMemStorage()
	auditService := service.NewAuditService()
	handler := NewUpdatesHandler(memStorage, "", auditService)

	// Создаём батч метрик
	val1 := float64(10.0)
	val2 := float64(20.0)
	delta1 := int64(5)
	delta2 := int64(15)

	metrics := []models.Metric{
		{ID: "metric1", MType: models.Gauge, Value: &val1},
		{ID: "metric2", MType: models.Gauge, Value: &val2},
		{ID: "counter1", MType: models.Counter, Delta: &delta1},
		{ID: "counter2", MType: models.Counter, Delta: &delta2},
	}
	body, _ := json.Marshal(metrics)

	req := httptest.NewRequest(http.MethodPost, "/updates", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	fmt.Printf("Status: %d\n", w.Code)

	m1, _ := memStorage.GetGauge("metric1")
	m2, _ := memStorage.GetGauge("metric2")
	c1, _ := memStorage.GetCounter("counter1")
	c2, _ := memStorage.GetCounter("counter2")

	fmt.Printf("metric1: %f, metric2: %f\n", m1, m2)
	fmt.Printf("counter1: %d, counter2: %d\n", c1, c2)

	// Output:
	// Status: 200
	// metric1: 10.000000, metric2: 20.000000
	// counter1: 5, counter2: 15
}

// Пример работы с IndexHandler - получение HTML страницы метрик
//
//nolint:errcheck
func ExampleIndexHandler_ServeHTTP() {
	// Создаём хранилище с тестовыми данными
	memStorage := storage.NewMemStorage()
	memStorage.SetGauge("gauge1", 100.5)
	memStorage.SetCounter("counter1", 50)

	// Создаём обработчик
	handler := NewIndexHandler(memStorage)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Content-Type: %s\n", w.Header().Get("Content-Type"))
	fmt.Printf("Body contains gauge1: %v\n", strings.Contains(w.Body.String(), "gauge1"))
	fmt.Printf("Body contains 100.5: %v\n", strings.Contains(w.Body.String(), "100.5"))

	// Output:
	// Status: 200
	// Content-Type: text/html
	// Body contains gauge1: true
	// Body contains 100.5: true
}
