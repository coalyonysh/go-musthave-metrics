package agent

import (
	"log"
	"sync"
	"time"

	"github.com/coalyonysh/go-musthave-metrics/internal/client"
	"github.com/coalyonysh/go-musthave-metrics/internal/models"
)

type Agent struct {
	collector      *MetricsCollector
	client         client.MetricSender
	pollInterval   time.Duration
	reportInterval time.Duration
	done           chan struct{}
	stopOnce       sync.Once
	metricsMutex   sync.RWMutex
	metrics        []models.Metric
}

func NewAgent(serverURL string, pollInterval, reportInterval time.Duration) *Agent {
	return &Agent{
		collector:      NewMetricsCollector(),
		client:         client.NewMetricHTTPClient(serverURL),
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		done:           make(chan struct{}),
		metricsMutex:   sync.RWMutex{},
		metrics:        nil,
	}
}

func (a *Agent) Start() {
	log.Printf("Starting metrics agent")
	log.Printf("Poll interval: %v, Report interval: %v", a.pollInterval, a.reportInterval)

	pollTicker := time.NewTicker(a.pollInterval)
	reportTicker := time.NewTicker(a.reportInterval)

	for {
		select {
		case <-pollTicker.C:
			metrics := a.collector.CollectMetrics()
			a.setMetrics(metrics)
			log.Printf("Collected %d metrics", len(metrics))

		case <-reportTicker.C:
			metrics := a.getMetrics()
			if metrics != nil {
				a.sendMetrics(metrics)
			}

		case <-a.done:
			pollTicker.Stop()
			reportTicker.Stop()
			return
		}
	}
}

func (a *Agent) Stop() {
	a.stopOnce.Do(func() {
		close(a.done)
	})
}

func (a *Agent) sendMetrics(metrics []models.Metric) {
	// Используем JSON клиент для отправки метрик через POST /update
	jsonClient, ok := a.client.(*client.MetricHTTPClient)
	if !ok {
		// Fallback на старый метод, если клиент не MetricHTTPClient
		for _, metric := range metrics {
			err := a.client.SendMetric(metric)
			if err != nil {
				log.Printf("Failed to send metric %s: %v", metric.ID, err)
			}
		}
		log.Printf("Sent %d metrics to server", len(metrics))
		return
	}

	for _, metric := range metrics {
		err := jsonClient.SendMetricJSON(metric)
		if err != nil {
			log.Printf("Failed to send metric %s: %v", metric.ID, err)
		}
	}
	log.Printf("Sent %d metrics to server", len(metrics))
}

func (a *Agent) setMetrics(metrics []models.Metric) {
	a.metricsMutex.Lock()
	defer a.metricsMutex.Unlock()
	a.metrics = metrics
}

func (a *Agent) getMetrics() []models.Metric {
	a.metricsMutex.RLock()
	defer a.metricsMutex.RUnlock()
	return a.metrics
}
