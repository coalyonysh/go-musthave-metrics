package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
)

func TestMetricHTTPClient_SendMetric_Gauge(t *testing.T) {
	// Создаем тестовый сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/update/gauge/testMetric/123.45" {
			t.Errorf("Expected path /update/gauge/testMetric/123.45, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "text/plain" {
			t.Errorf("Expected Content-Type text/plain, got %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewMetricHTTPClient(server.URL)

	value := 123.45
	metric := models.Metric{
		ID:    "testMetric",
		MType: models.Gauge,
		Value: &value,
	}

	err := client.SendMetric(metric)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestMetricHTTPClient_SendMetric_Counter(t *testing.T) {
	// Создаем тестовый сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/update/counter/testCounter/42" {
			t.Errorf("Expected path /update/counter/testCounter/42, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "text/plain" {
			t.Errorf("Expected Content-Type text/plain, got %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewMetricHTTPClient(server.URL)

	delta := int64(42)
	metric := models.Metric{
		ID:    "testCounter",
		MType: models.Counter,
		Delta: &delta,
	}

	err := client.SendMetric(metric)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestMetricHTTPClient_SendMetric_InvalidMetric(t *testing.T) {
	client := NewMetricHTTPClient("http://localhost:8080")

	// Тест с gauge без значения
	metric := models.Metric{
		ID:    "testMetric",
		MType: models.Gauge,
		Value: nil,
	}

	err := client.SendMetric(metric)
	if err == nil {
		t.Error("Expected error for invalid gauge metric, got nil")
	}

	// Тест с counter без значения
	metric = models.Metric{
		ID:    "testCounter",
		MType: models.Counter,
		Delta: nil,
	}

	err = client.SendMetric(metric)
	if err == nil {
		t.Error("Expected error for invalid counter metric, got nil")
	}
}

func TestMetricHTTPClient_SendMetric_ServerError(t *testing.T) {
	// Создаем тестовый сервер, который возвращает ошибку
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewMetricHTTPClient(server.URL)

	value := 123.45
	metric := models.Metric{
		ID:    "testMetric",
		MType: models.Gauge,
		Value: &value,
	}

	err := client.SendMetric(metric)
	if err == nil {
		t.Error("Expected error for server error, got nil")
	}
}
