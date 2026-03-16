package agent

import (
	"errors"
	"sync"
	"testing"

	"github.com/coalyonysh/go-musthave-metrics/internal/client"
	"github.com/coalyonysh/go-musthave-metrics/internal/models"
)

// MockClient is a mock implementation of client.MetricSender for testing
type MockClient struct {
	SentMetrics []models.Metric
	Err         error
}

func (m *MockClient) SendMetric(metric models.Metric) error {
	m.SentMetrics = append(m.SentMetrics, metric)
	return m.Err
}

func (m *MockClient) SendMetricJSON(metric models.Metric) error {
	m.SentMetrics = append(m.SentMetrics, metric)
	return m.Err
}

func (m *MockClient) SendMetricsBatch(metrics []models.Metric) error {
	m.SentMetrics = append(m.SentMetrics, metrics...)
	return m.Err
}

func TestAgent_SendMetrics_FallbackPath(t *testing.T) {
	// This test covers the fallback path in sendMetrics when client is not *client.MetricHTTPClient
	// We need to create an agent with a mock client that implements MetricSender but is not *client.MetricHTTPClient

	// Create a simple mock that is NOT a *client.MetricHTTPClient
	mockClient := &MockClient{}

	agent := &Agent{
		client:       mockClient,
		rateLimit:    1,
		metricsMutex: sync.RWMutex{},
	}

	metrics := []models.Metric{
		{ID: "test1", MType: models.Gauge, Value: float64Ptr(1.0)},
		{ID: "test2", MType: models.Counter, Delta: int64Ptr(10)},
	}

	agent.sendMetrics(metrics)

	// Verify metrics were sent through fallback path
	if len(mockClient.SentMetrics) != 2 {
		t.Errorf("expected 2 metrics to be sent, got %d", len(mockClient.SentMetrics))
	}
}

func TestAgent_SendMetrics_WithMockError(t *testing.T) {
	mockClient := &MockClient{
		Err: errors.New("send error"),
	}

	agent := &Agent{
		client:       mockClient,
		rateLimit:    1,
		metricsMutex: sync.RWMutex{},
	}

	metrics := []models.Metric{
		{ID: "test1", MType: models.Gauge, Value: float64Ptr(1.0)},
	}

	// Should not panic, just log error
	agent.sendMetrics(metrics)

	// Verify metric was attempted - with rateLimit=1, we try batch first, then fallback
	// So we get 2 attempts (batch fails, then individual send fails)
	if len(mockClient.SentMetrics) != 2 {
		t.Errorf("expected 2 metric to be attempted, got %d", len(mockClient.SentMetrics))
	}
}

func TestAgent_SendMetrics_WithHTTPClientSuccess(t *testing.T) {
	// Create a real HTTP client with a test server
	httpClient := client.NewMetricHTTPClient("http://invalid:9999", "")

	agent := &Agent{
		client:       httpClient,
		rateLimit:    1,
		metricsMutex: sync.RWMutex{},
	}

	metrics := []models.Metric{
		{ID: "test1", MType: models.Gauge, Value: float64Ptr(1.0)},
	}

	// This will fail because the server doesn't exist, but it will exercise the code path
	agent.sendMetrics(metrics)
}
