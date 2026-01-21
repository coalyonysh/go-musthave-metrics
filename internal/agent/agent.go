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
	rateLimit      int
	done           chan struct{}
	stopOnce       sync.Once
	wg             sync.WaitGroup
	metricsMutex   sync.RWMutex
	metrics        []models.Metric
	metricsChan    chan []models.Metric
}

func NewAgent(config Config) *Agent {
	return &Agent{
		collector:      NewMetricsCollector(),
		client:         client.NewMetricHTTPClient(config.ServerURL, config.Key),
		pollInterval:   config.PollInterval,
		reportInterval: config.ReportInterval,
		rateLimit:      config.RateLimit,
		done:           make(chan struct{}),
		wg:             sync.WaitGroup{},
		metricsMutex:   sync.RWMutex{},
		metrics:        nil,
		metricsChan:    make(chan []models.Metric, config.RateLimit*2), // буфер для избежания блокировки
	}
}

func (a *Agent) Start() {
	log.Printf("Starting metrics agent")
	log.Printf("Poll interval: %v, Report interval: %v, Rate limit: %d", a.pollInterval, a.reportInterval, a.rateLimit)

	// Start worker pool
	a.startWorkerPool()

	// Start poll goroutine
	a.wg.Add(1)
	go a.pollMetrics()

	// Start report ticker
	reportTicker := time.NewTicker(a.reportInterval)
	defer reportTicker.Stop()

	for {
		select {
		case <-reportTicker.C:
			metrics := a.getMetrics()
			if metrics != nil {
				a.metricsChan <- metrics
				log.Printf("Sent %d metrics to worker pool", len(metrics))
			}

		case <-a.done:
			return
		}
	}
}

func (a *Agent) pollMetrics() {
	defer a.wg.Done()
	pollTicker := time.NewTicker(a.pollInterval)
	defer pollTicker.Stop()

	for {
		select {
		case <-pollTicker.C:
			runtimeMetrics := a.collector.CollectMetrics()
			gopsutilMetrics := a.collector.CollectGopsutilMetrics()
			allMetrics := append(runtimeMetrics, gopsutilMetrics...)
			a.setMetrics(allMetrics)
			log.Printf("Collected %d metrics", len(allMetrics))

		case <-a.done:
			return
		}
	}
}

func (a *Agent) startWorkerPool() {
	a.wg.Add(a.rateLimit)
	for i := 0; i < a.rateLimit; i++ {
		go func(workerID int) {
			defer a.wg.Done()
			log.Printf("Starting worker %d", workerID)
			for metrics := range a.metricsChan {
				a.sendMetrics(metrics)
			}
			log.Printf("Stopping worker %d", workerID)
		}(i)
	}
}

func (a *Agent) Stop() {
	a.stopOnce.Do(func() {
		close(a.done)
		close(a.metricsChan)
		a.wg.Wait()
	})
}

func (a *Agent) sendMetrics(metrics []models.Metric) {
	// Используем JSON клиент для отправки батча метрик через POST /updates
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

	// Отправляем батч
	err := jsonClient.SendMetricsBatch(metrics)
	if err != nil {
		log.Printf("Failed to send metrics batch: %v", err)
		// Fallback: отправляем по одной
		for _, metric := range metrics {
			err := jsonClient.SendMetricJSON(metric)
			if err != nil {
				log.Printf("Failed to send metric %s: %v", metric.ID, err)
			}
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
