package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coalyonysh/go-musthave-metrics/internal/agent"
)

func main() {
	serverURL := getEnv("ADDRESS", "http://localhost:8080")
	pollInterval := getEnvAsDuration("POLL_INTERVAL", 2*time.Second)
	reportInterval := getEnvAsDuration("REPORT_INTERVAL", 10*time.Second)

	log.Printf("Config: Server=%s, PollInterval=%v, ReportInterval=%v",
		serverURL, pollInterval, reportInterval)

	agent := agent.NewAgent(serverURL, pollInterval, reportInterval)

	// Обработка сигналов для graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go agent.Start()

	<-stop
	log.Println("Shutting down agent...")
	agent.Stop()
	log.Println("Agent stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
