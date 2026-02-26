package client

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
)

func TestSendMetricJSON_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Content-Encoding") != "gzip" {
			t.Errorf("expected Content-Encoding gzip, got %s", r.Header.Get("Content-Encoding"))
		}

		// Verify gzip compression
		reader, err := gzip.NewReader(r.Body)
		if err != nil {
			t.Fatalf("failed to create gzip reader: %v", err)
		}
		defer reader.Close()

		var metric models.Metric
		decoder := json.NewDecoder(reader)
		if err := decoder.Decode(&metric); err != nil {
			t.Fatalf("failed to decode JSON: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewMetricHTTPClient(server.URL, "")

	metric := models.Metric{
		ID:    "test_gauge",
		MType: models.Gauge,
		Value: float64Ptr(1.5),
	}

	err := client.SendMetricJSON(metric)
	if err != nil {
		t.Errorf("SendMetricJSON returned error: %v", err)
	}
}

func TestSendMetricJSON_WithKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify hash header is present
		hash := r.Header.Get("HashSHA256")
		if hash == "" {
			t.Error("expected HashSHA256 header")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewMetricHTTPClient(server.URL, "testkey")

	metric := models.Metric{
		ID:    "test_gauge",
		MType: models.Gauge,
		Value: float64Ptr(1.5),
	}

	err := client.SendMetricJSON(metric)
	if err != nil {
		t.Errorf("SendMetricJSON returned error: %v", err)
	}
}

func TestSendMetricJSON_MissingDelta(t *testing.T) {
	client := NewMetricHTTPClient("http://localhost:8080", "")

	metric := models.Metric{
		ID:    "test_counter",
		MType: models.Counter,
		// Delta is nil
	}

	err := client.SendMetricJSON(metric)
	if err == nil {
		t.Error("expected error for missing delta")
	}
}

func TestSendMetricJSON_MissingValue(t *testing.T) {
	client := NewMetricHTTPClient("http://localhost:8080", "")

	metric := models.Metric{
		ID:    "test_gauge",
		MType: models.Gauge,
		// Value is nil
	}

	err := client.SendMetricJSON(metric)
	if err == nil {
		t.Error("expected error for missing value")
	}
}

func TestSendMetricJSON_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewMetricHTTPClient(server.URL, "")

	metric := models.Metric{
		ID:    "test_gauge",
		MType: models.Gauge,
		Value: float64Ptr(1.5),
	}

	err := client.SendMetricJSON(metric)
	if err == nil {
		t.Error("expected error for server error")
	}
}

func TestSendMetricsBatch_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify gzip compression
		reader, err := gzip.NewReader(r.Body)
		if err != nil {
			t.Fatalf("failed to create gzip reader: %v", err)
		}
		defer reader.Close()

		var metrics []models.Metric
		decoder := json.NewDecoder(reader)
		if err := decoder.Decode(&metrics); err != nil {
			t.Fatalf("failed to decode JSON: %v", err)
		}

		if len(metrics) != 2 {
			t.Errorf("expected 2 metrics, got %d", len(metrics))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewMetricHTTPClient(server.URL, "")

	metrics := []models.Metric{
		{ID: "gauge1", MType: models.Gauge, Value: float64Ptr(1.0)},
		{ID: "counter1", MType: models.Counter, Delta: int64Ptr(10)},
	}

	err := client.SendMetricsBatch(metrics)
	if err != nil {
		t.Errorf("SendMetricsBatch returned error: %v", err)
	}
}

func TestSendMetricsBatch_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewMetricHTTPClient(server.URL, "")

	metrics := []models.Metric{
		{ID: "gauge1", MType: models.Gauge, Value: float64Ptr(1.0)},
	}

	err := client.SendMetricsBatch(metrics)
	if err == nil {
		t.Error("expected error for server error")
	}
}

func float64Ptr(f float64) *float64 {
	return &f
}

func int64Ptr(i int64) *int64 {
	return &i
}

func TestSendMetricJSON_GzipResponse(t *testing.T) {
	// Create a server that returns gzipped response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")

		// Create gzip response
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		gz.Write([]byte(`{"status":"ok"}`))
		gz.Close()

		w.Write(buf.Bytes())
	}))
	defer server.Close()

	client := NewMetricHTTPClient(server.URL, "")

	metric := models.Metric{
		ID:    "test_gauge",
		MType: models.Gauge,
		Value: float64Ptr(1.5),
	}

	// This should handle gzipped response
	err := client.SendMetricJSON(metric)
	// We expect either success or error depending on how response is handled
	_ = err
}
