package agent

import (
	"sync"
	"testing"
	"time"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
)

// Mock HTTP клиент для тестирования
type mockHTTPClient struct {
	sentMetrics []models.Metric
	shouldError bool
}

func (m *mockHTTPClient) SendMetric(metric models.Metric) error {
	if m.shouldError {
		return &mockError{}
	}
	m.sentMetrics = append(m.sentMetrics, metric)
	return nil
}

type mockError struct{}

func (e *mockError) Error() string {
	return "mock error"
}

func TestAgent_NewAgent(t *testing.T) {
	config := Config{
		ServerURL:      "http://localhost:8080",
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		Key:            "",
		RateLimit:      10,
	}

	agent := NewAgent(config)

	if agent == nil {
		t.Fatal("Expected agent to be created, got nil")
	}

	if agent.pollInterval != config.PollInterval {
		t.Errorf("Expected pollInterval %v, got %v", config.PollInterval, agent.pollInterval)
	}

	if agent.reportInterval != config.ReportInterval {
		t.Errorf("Expected reportInterval %v, got %v", config.ReportInterval, agent.reportInterval)
	}

	if agent.rateLimit != config.RateLimit {
		t.Errorf("Expected rateLimit %d, got %d", config.RateLimit, agent.rateLimit)
	}

	if agent.collector == nil {
		t.Error("Expected collector to be initialized")
	}

	if agent.client == nil {
		t.Error("Expected client to be initialized")
	}

	if agent.done == nil {
		t.Error("Expected done channel to be initialized")
	}

	if agent.metricsChan == nil {
		t.Error("Expected metricsChan to be initialized")
	}
}

func TestAgent_SendMetrics(t *testing.T) {
	mockClient := &mockHTTPClient{}
	agent := &Agent{
		client: mockClient,
	}

	metrics := []models.Metric{
		{ID: "test1", MType: models.Gauge, Value: float64Ptr(1.0)},
		{ID: "test2", MType: models.Counter, Delta: int64Ptr(2)},
	}

	agent.sendMetrics(metrics)

	if len(mockClient.sentMetrics) != 2 {
		t.Errorf("Expected 2 metrics to be sent, got %d", len(mockClient.sentMetrics))
	}

	// Проверяем, что метрики были отправлены
	if mockClient.sentMetrics[0].ID != "test1" {
		t.Errorf("Expected first metric ID to be 'test1', got %s", mockClient.sentMetrics[0].ID)
	}

	if mockClient.sentMetrics[1].ID != "test2" {
		t.Errorf("Expected second metric ID to be 'test2', got %s", mockClient.sentMetrics[1].ID)
	}
}

func TestAgent_SendMetrics_WithError(t *testing.T) {
	mockClient := &mockHTTPClient{shouldError: true}
	agent := &Agent{
		client: mockClient,
	}

	metrics := []models.Metric{
		{ID: "test1", MType: models.Gauge, Value: float64Ptr(1.0)},
	}

	// Функция не должна паниковать при ошибке
	agent.sendMetrics(metrics)

	// Метрики не должны быть добавлены в список при ошибке
	if len(mockClient.sentMetrics) != 0 {
		t.Errorf("Expected 0 metrics to be sent due to error, got %d", len(mockClient.sentMetrics))
	}
}

func TestAgent_Stop(t *testing.T) {
	agent := &Agent{
		done:        make(chan struct{}, 1),
		metricsChan: make(chan []models.Metric, 1),
		wg:          sync.WaitGroup{},
	}

	// Тест не блокируется
	agent.Stop()

	// Проверяем, что сигнал был отправлен
	select {
	case <-agent.done:
		// Ожидаемое поведение
	default:
		t.Error("Expected done signal to be sent")
	}
}
