package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/coalyonysh/go-musthave-metrics/internal/agent"
	"github.com/coalyonysh/go-musthave-metrics/internal/buildinfo"
)

// Глобальные переменные для информации о сборке
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt возвращает целочисленное значение переменной окружения или значение по умолчанию
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Invalid value for %s: %s, using default: %d", key, value, defaultValue)
	}
	return defaultValue
}

func main() {
	// Значения по умолчанию из переменных окружения
	defaultAddr := getEnv("ADDRESS", "localhost:8080")
	defaultReportSec := getEnvInt("REPORT_INTERVAL", 10)
	defaultPollSec := getEnvInt("POLL_INTERVAL", 2)
	defaultKeyFile := getEnv("KEY", "")
	defaultRateLimit := getEnvInt("RATE_LIMIT", 10)

	// Флаги (приоритет у переменных окружения)
	addrFlag := flag.String("a", defaultAddr, "server address")
	reportSec := flag.Int("r", defaultReportSec, "report interval in seconds")
	pollSec := flag.Int("p", defaultPollSec, "poll interval in seconds")
	keyFileFlag := flag.String("k", defaultKeyFile, "path to file containing hash key")
	rateLimitFlag := flag.Int("l", defaultRateLimit, "rate limit for concurrent requests")
	flag.Parse()

	buildinfo.PrintBuildInfo(buildVersion, buildDate, buildCommit)

	// Читаем ключ из файла, если указан
	var key string
	if *keyFileFlag != "" {
		keyBytes, err := os.ReadFile(*keyFileFlag)
		if err != nil {
			// Если файл не найден, используем значение флага как ключ напрямую
			key = *keyFileFlag
		} else {
			key = string(keyBytes)
		}
	}

	config := &agent.Config{
		ServerURL:      *addrFlag,
		PollInterval:   time.Duration(*pollSec) * time.Second,
		ReportInterval: time.Duration(*reportSec) * time.Second,
		Key:            key,
		RateLimit:      *rateLimitFlag,
	}

	if err := config.Validate(); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	log.Printf("Config: Server=%s, PollInterval=%v, ReportInterval=%v, Key=%s, RateLimit=%d",
		config.ServerURL, config.PollInterval, config.ReportInterval, config.Key, config.RateLimit)

	agent := agent.NewAgent(*config)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go agent.Start()

	<-stop
	log.Println("Shutting down agent...")
	agent.Stop()
	log.Println("Agent stopped")
}
