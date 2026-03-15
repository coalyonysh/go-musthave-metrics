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
	defaultCryptoKey := getEnv("CRYPTO_KEY", "")
	defaultConfig := getEnv("CONFIG", "")

	// Флаги командной строки
	addrPtr := flag.String("a", defaultAddr, "server address")
	reportSecPtr := flag.Int("r", defaultReportSec, "report interval in seconds")
	pollSecPtr := flag.Int("p", defaultPollSec, "poll interval in seconds")
	keyFilePtr := flag.String("k", defaultKeyFile, "path to file containing hash key")
	rateLimitPtr := flag.Int("l", defaultRateLimit, "rate limit for concurrent requests")
	cryptoKeyPtr := flag.String("crypto-key", defaultCryptoKey, "path to RSA public key file for encryption")
	configPtr := flag.String("c", defaultConfig, "path to config file (JSON)")
	flag.Parse()

	// Устанавливаем путь к файлу конфигурации из флага -c, если указан
	if configPtr != nil && *configPtr != "" {
		os.Setenv("CONFIG", *configPtr)
	}

	// Загружаем конфигурацию (с учетом CONFIG env и файла)
	cfg, err := agent.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	buildinfo.PrintBuildInfo(buildVersion, buildDate, buildCommit)

	// Применяем флаги (они имеют приоритет над env и config)
	if addrPtr != nil && *addrPtr != "" && *addrPtr != defaultAddr {
		cfg.ServerURL = *addrPtr
	}
	if reportSecPtr != nil && *reportSecPtr > 0 && *reportSecPtr != defaultReportSec {
		cfg.ReportInterval = time.Duration(*reportSecPtr) * time.Second
	}
	if pollSecPtr != nil && *pollSecPtr > 0 && *pollSecPtr != defaultPollSec {
		cfg.PollInterval = time.Duration(*pollSecPtr) * time.Second
	}
	if rateLimitPtr != nil && *rateLimitPtr > 0 && *rateLimitPtr != defaultRateLimit {
		cfg.RateLimit = *rateLimitPtr
	}
	if keyFilePtr != nil && *keyFilePtr != "" && *keyFilePtr != defaultKeyFile {
		cfg.Key = *keyFilePtr
	}
	if cryptoKeyPtr != nil && *cryptoKeyPtr != "" && *cryptoKeyPtr != defaultCryptoKey {
		cfg.CryptoKey = *cryptoKeyPtr
	}

	// Читаем ключ из файла, если указан
	if cfg.Key != "" {
		keyBytes, err := os.ReadFile(cfg.Key)
		if err == nil {
			cfg.Key = string(keyBytes)
		}
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	log.Printf("Config: Server=%s, PollInterval=%v, ReportInterval=%v, Key=%s, RateLimit=%d, CryptoKey=%s",
		cfg.ServerURL, cfg.PollInterval, cfg.ReportInterval, cfg.Key, cfg.RateLimit, cfg.CryptoKey)

	agentInstance := agent.NewAgent(*cfg)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go agentInstance.Start()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	<-sigCh
	log.Println("Shutting down agent...")
	agentInstance.Stop()
	log.Println("Agent stopped")
}
