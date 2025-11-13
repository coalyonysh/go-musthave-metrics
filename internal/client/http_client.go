package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
)

type MetricSender interface {
	SendMetric(metric models.Metric) error
}

type MetricHTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewMetricHTTPClient(baseURL string) *MetricHTTPClient {
	return &MetricHTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *MetricHTTPClient) SendMetric(metric models.Metric) error {
	// Валидация метрики
	if metric.MType == models.Counter && metric.Delta == nil {
		return fmt.Errorf("invalid metric: missing delta for counter type")
	}
	if metric.MType == models.Gauge && metric.Value == nil {
		return fmt.Errorf("invalid metric: missing value for gauge type")
	}

	// Создаем JSON тело запроса
	jsonData, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("failed to marshal metric to JSON: %w", err)
	}

	url := fmt.Sprintf("%s/update", c.baseURL)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned non-200 status: %d", resp.StatusCode)
	}

	var valueStr string
	if metric.MType == models.Counter && metric.Delta != nil {
		valueStr = fmt.Sprintf("%d", *metric.Delta)
	} else if metric.MType == models.Gauge && metric.Value != nil {
		valueStr = fmt.Sprintf("%g", *metric.Value)
	}

	log.Printf("Successfully sent metric: %s/%s/%s", metric.MType, metric.ID, valueStr)
	return nil
}
