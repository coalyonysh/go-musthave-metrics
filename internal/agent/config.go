package agent

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/coalyonysh/go-musthave-metrics/internal/config"
)

type Config struct {
	ServerURL      string
	PollInterval   time.Duration
	ReportInterval time.Duration
	Key            string
	CryptoKey      string
	RateLimit      int
}

// LoadConfig загружает конфигурацию из JSON файла, переменных окружения и флагов
// Приоритет: флаги > переменные окружения > JSON файл
func LoadConfig() (*Config, error) {
	// Сначала загружаем из JSON файла
	configFile := getConfigFile()
	jsonCfg, err := config.LoadAgentConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	// Значения по умолчанию
	cfg := &Config{
		ServerURL:      "localhost:8080",
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		Key:            "",
		CryptoKey:      "",
		RateLimit:      10,
	}

	// Применяем значения из JSON файла (низкий приоритет)
	if jsonCfg != nil {
		if jsonCfg.Address != "" {
			cfg.ServerURL = jsonCfg.Address
		}
		if jsonCfg.PollInterval != "" {
			d, err := config.ParseDuration(jsonCfg.PollInterval)
			if err != nil {
				return nil, fmt.Errorf("invalid poll_interval: %w", err)
			}
			cfg.PollInterval = d
		}
		if jsonCfg.ReportInterval != "" {
			d, err := config.ParseDuration(jsonCfg.ReportInterval)
			if err != nil {
				return nil, fmt.Errorf("invalid report_interval: %w", err)
			}
			cfg.ReportInterval = d
		}
		if jsonCfg.CryptoKey != "" {
			cfg.CryptoKey = jsonCfg.CryptoKey
		}
	}

	// Переменные окружения (средний приоритет)
	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		cfg.ServerURL = envAddr
	}
	if envPoll := os.Getenv("POLL_INTERVAL"); envPoll != "" {
		d, err := time.ParseDuration(envPoll)
		if err != nil {
			return nil, fmt.Errorf("invalid POLL_INTERVAL: %w", err)
		}
		cfg.PollInterval = d
	}
	if envReport := os.Getenv("REPORT_INTERVAL"); envReport != "" {
		d, err := time.ParseDuration(envReport)
		if err != nil {
			return nil, fmt.Errorf("invalid REPORT_INTERVAL: %w", err)
		}
		cfg.ReportInterval = d
	}
	if envKey := os.Getenv("KEY"); envKey != "" {
		cfg.Key = envKey
	}
	if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		cfg.CryptoKey = envCryptoKey
	}

	return cfg, nil
}

// Validate проверяет конфигурацию
func (c *Config) Validate() error {
	if c.ServerURL == "" {
		return fmt.Errorf("server URL cannot be empty")
	}

	if !strings.HasPrefix(c.ServerURL, "http") {
		c.ServerURL = "http://" + c.ServerURL
	}

	if c.PollInterval <= 0 {
		return fmt.Errorf("poll interval must be positive")
	}

	if c.ReportInterval <= 0 {
		return fmt.Errorf("report interval must be positive")
	}

	if c.RateLimit <= 0 {
		return fmt.Errorf("rate limit must be positive")
	}

	return nil
}

// getConfigFile возвращает путь к файлу конфигурации из переменной окружения или флага
func getConfigFile() string {
	// Сначала проверяем переменную окружения CONFIG
	if cfg := os.Getenv("CONFIG"); cfg != "" {
		return cfg
	}
	return ""
}
