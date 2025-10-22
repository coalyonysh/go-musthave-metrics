package agent

import (
	"log"
	"time"

	"github.com/coalyonysh/go-musthave-metrics/internal/client"
	"github.com/coalyonysh/go-musthave-metrics/internal/models"
)

type Agent struct {
	collector      *MetricsCollector
	client         client.HTTPClient
	pollInterval   time.Duration
	reportInterval time.Duration
	done           chan bool
}

func NewAgent(serverURL string, pollInterval, reportInterval time.Duration) *Agent {
	return &Agent{
		collector:      NewMetricsCollector(),
		client:         client.NewMetricHTTPClient(serverURL),
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		done:           make(chan bool),
	}
}

func (a *Agent) Start() {
	log.Printf("Starting metrics agent")
	log.Printf("Poll interval: %v, Report interval: %v", a.pollInterval, a.reportInterval)

	pollTicker := time.NewTicker(a.pollInterval)
	reportTicker := time.NewTicker(a.reportInterval)

	var metrics []models.Metric

	for {
		select {
		case <-pollTicker.C:
			metrics = a.collector.CollectMetrics()
			log.Printf("Collected %d metrics", len(metrics))

		case <-reportTicker.C:
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
	a.done <- true
}

func (a *Agent) sendMetrics(metrics []models.Metric) {
	for _, metric := range metrics {
		err := a.client.SendMetric(metric)
		if err != nil {
			log.Printf("Failed to send metric %s: %v", metric.ID, err)
		}
	}
	log.Printf("Sent %d metrics to server", len(metrics))
}
