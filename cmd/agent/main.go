package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/coalyonysh/go-musthave-metrics/internal/agent"
)

func main() {
	addrFlag := flag.String("a", "localhost:8080", "server address")
	reportSec := flag.Int("r", 10, "report interval in seconds")
	pollSec := flag.Int("p", 2, "poll interval in seconds")
	flag.Parse()

	serverURL := *addrFlag

	if !hasProtocol(serverURL) {
		serverURL = "http://" + serverURL
	}

	pollInterval := time.Duration(*pollSec) * time.Second
	reportInterval := time.Duration(*reportSec) * time.Second

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

func hasProtocol(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}
