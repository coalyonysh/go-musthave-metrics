package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coalyonysh/go-musthave-metrics/internal/agent"
)

func main() {
	addrFlag := flag.String("a", "localhost:8080", "server address")
	reportSec := flag.Int("r", 10, "report interval in seconds")
	pollSec := flag.Int("p", 2, "poll interval in seconds")
	flag.Parse()

	config := &agent.Config{
		ServerURL:      *addrFlag,
		PollInterval:   time.Duration(*pollSec) * time.Second,
		ReportInterval: time.Duration(*reportSec) * time.Second,
	}

	if err := config.Validate(); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	log.Printf("Config: Server=%s, PollInterval=%v, ReportInterval=%v",
		config.ServerURL, config.PollInterval, config.ReportInterval)

	agent := agent.NewAgent(config.ServerURL, config.PollInterval, config.ReportInterval)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go agent.Start()

	<-stop
	log.Println("Shutting down agent...")
	agent.Stop()
	log.Println("Agent stopped")
}
