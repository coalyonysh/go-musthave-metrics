package agent

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
)

func TestGetConfigFile_WithEnvVar(t *testing.T) {
	// Set CONFIG environment variable
	os.Setenv("CONFIG", "/test/config.json")
	defer os.Unsetenv("CONFIG")

	result := getConfigFile()

	if result != "/test/config.json" {
		t.Errorf("expected /test/config.json, got %s", result)
	}
}

func TestGetConfigFile_WithoutEnvVar(t *testing.T) {
	// Ensure CONFIG is not set
	os.Unsetenv("CONFIG")

	result := getConfigFile()

	if result != "" {
		t.Errorf("expected empty string, got %s", result)
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	// Set a non-existent config file - this should NOT return an error
	// because LoadAgentConfig returns nil when file doesn't exist
	os.Setenv("CONFIG", "/nonexistent/config.json")
	defer os.Unsetenv("CONFIG")

	cfg, err := LoadConfig()

	// Should NOT return error - missing config file is silently ignored
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should return default config
	if cfg == nil {
		t.Error("expected config to be returned with defaults")
	}
}

func TestLoadConfig_WithEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("ADDRESS", "localhost:9090")
	os.Setenv("POLL_INTERVAL", "5s")
	os.Setenv("REPORT_INTERVAL", "15s")
	os.Setenv("KEY", "test_key")
	os.Setenv("CRYPTO_KEY", "test_crypto_key")
	os.Setenv("CONFIG", "") // Clear config file
	defer func() {
		os.Unsetenv("ADDRESS")
		os.Unsetenv("POLL_INTERVAL")
		os.Unsetenv("REPORT_INTERVAL")
		os.Unsetenv("KEY")
		os.Unsetenv("CRYPTO_KEY")
		os.Unsetenv("CONFIG")
	}()

	cfg, err := LoadConfig()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if cfg.ServerURL != "localhost:9090" {
		t.Errorf("expected ServerURL localhost:9090, got %s", cfg.ServerURL)
	}

	if cfg.PollInterval != 5*time.Second {
		t.Errorf("expected PollInterval 5s, got %v", cfg.PollInterval)
	}

	if cfg.ReportInterval != 15*time.Second {
		t.Errorf("expected ReportInterval 15s, got %v", cfg.ReportInterval)
	}

	if cfg.Key != "test_key" {
		t.Errorf("expected Key test_key, got %s", cfg.Key)
	}

	if cfg.CryptoKey != "test_crypto_key" {
		t.Errorf("expected CryptoKey test_crypto_key, got %s", cfg.CryptoKey)
	}
}

func TestLoadConfig_InvalidPollInterval(t *testing.T) {
	os.Setenv("POLL_INTERVAL", "invalid")
	os.Setenv("CONFIG", "")
	defer func() {
		os.Unsetenv("POLL_INTERVAL")
		os.Unsetenv("CONFIG")
	}()

	_, err := LoadConfig()

	if err == nil {
		t.Error("expected error for invalid POLL_INTERVAL")
	}
}

func TestLoadConfig_InvalidReportInterval(t *testing.T) {
	os.Setenv("REPORT_INTERVAL", "invalid")
	os.Setenv("CONFIG", "")
	defer func() {
		os.Unsetenv("REPORT_INTERVAL")
		os.Unsetenv("CONFIG")
	}()

	_, err := LoadConfig()

	if err == nil {
		t.Error("expected error for invalid REPORT_INTERVAL")
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear all env vars
	os.Unsetenv("ADDRESS")
	os.Unsetenv("POLL_INTERVAL")
	os.Unsetenv("REPORT_INTERVAL")
	os.Unsetenv("KEY")
	os.Unsetenv("CRYPTO_KEY")
	os.Unsetenv("CONFIG")

	cfg, err := LoadConfig()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check defaults - note: Validate() is NOT called in LoadConfig()
	// so ServerURL stays as "localhost:8080" (not prefixed with http://)
	if cfg.ServerURL != "localhost:8080" {
		t.Errorf("expected ServerURL localhost:8080, got %s", cfg.ServerURL)
	}

	if cfg.PollInterval != 2*time.Second {
		t.Errorf("expected PollInterval 2s, got %v", cfg.PollInterval)
	}

	if cfg.ReportInterval != 10*time.Second {
		t.Errorf("expected ReportInterval 10s, got %v", cfg.ReportInterval)
	}

	if cfg.RateLimit != 10 {
		t.Errorf("expected RateLimit 10, got %d", cfg.RateLimit)
	}
}

func TestConfig_Validate_EmptyServerURL(t *testing.T) {
	cfg := &Config{
		ServerURL:      "",
		PollInterval:   time.Second,
		ReportInterval: time.Second,
		RateLimit:      1,
	}

	err := cfg.Validate()

	if err == nil {
		t.Error("expected error for empty ServerURL")
	}
}

func TestConfig_Validate_AutoPrefixHTTP(t *testing.T) {
	cfg := &Config{
		ServerURL:      "localhost:8080",
		PollInterval:   time.Second,
		ReportInterval: time.Second,
		RateLimit:      1,
	}

	err := cfg.Validate()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if cfg.ServerURL != "http://localhost:8080" {
		t.Errorf("expected ServerURL http://localhost:8080, got %s", cfg.ServerURL)
	}
}

func TestConfig_Validate_InvalidPollInterval(t *testing.T) {
	cfg := &Config{
		ServerURL:      "http://localhost:8080",
		PollInterval:   0,
		ReportInterval: time.Second,
		RateLimit:      1,
	}

	err := cfg.Validate()

	if err == nil {
		t.Error("expected error for invalid PollInterval")
	}
}

func TestConfig_Validate_NegativePollInterval(t *testing.T) {
	cfg := &Config{
		ServerURL:      "http://localhost:8080",
		PollInterval:   -1,
		ReportInterval: time.Second,
		RateLimit:      1,
	}

	err := cfg.Validate()

	if err == nil {
		t.Error("expected error for negative PollInterval")
	}
}

func TestConfig_Validate_InvalidReportInterval(t *testing.T) {
	cfg := &Config{
		ServerURL:      "http://localhost:8080",
		PollInterval:   time.Second,
		ReportInterval: 0,
		RateLimit:      1,
	}

	err := cfg.Validate()

	if err == nil {
		t.Error("expected error for invalid ReportInterval")
	}
}

func TestConfig_Validate_InvalidRateLimit(t *testing.T) {
	cfg := &Config{
		ServerURL:      "http://localhost:8080",
		PollInterval:   time.Second,
		ReportInterval: time.Second,
		RateLimit:      0,
	}

	err := cfg.Validate()

	if err == nil {
		t.Error("expected error for invalid RateLimit")
	}
}

func TestConfig_Validate_NegativeRateLimit(t *testing.T) {
	cfg := &Config{
		ServerURL:      "http://localhost:8080",
		PollInterval:   time.Second,
		ReportInterval: time.Second,
		RateLimit:      -1,
	}

	err := cfg.Validate()

	if err == nil {
		t.Error("expected error for negative RateLimit")
	}
}

func TestConfig_Validate_Success(t *testing.T) {
	cfg := &Config{
		ServerURL:      "http://localhost:8080",
		PollInterval:   time.Second,
		ReportInterval: time.Second,
		RateLimit:      1,
	}

	err := cfg.Validate()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// Agent tests

func TestAgent_SetMetrics(t *testing.T) {
	agent := &Agent{
		metricsMutex: sync.RWMutex{},
		metrics:      nil,
	}

	testMetrics := []models.Metric{
		{ID: "test1", MType: "gauge", Value: func() *float64 { v := 1.0; return &v }()},
	}

	agent.setMetrics(testMetrics)

	retrieved := agent.getMetrics()
	if len(retrieved) != 1 {
		t.Errorf("expected 1 metric, got %d", len(retrieved))
	}
	if retrieved[0].ID != "test1" {
		t.Errorf("expected metric ID test1, got %s", retrieved[0].ID)
	}
}

func TestAgent_GetMetrics_Empty(t *testing.T) {
	agent := &Agent{
		metricsMutex: sync.RWMutex{},
		metrics:      nil,
	}

	retrieved := agent.getMetrics()
	if retrieved != nil {
		t.Errorf("expected nil metrics, got %v", retrieved)
	}
}

func TestAgent_NewAgent_WithCryptoKeyFallback(t *testing.T) {
	// This tests creating agent with invalid crypto key - should fallback to regular client
	// Note: We're testing the code path, not the actual crypto behavior
	os.Setenv("CONFIG", "")
	defer os.Unsetenv("CONFIG")

	config := Config{
		ServerURL:      "http://localhost:8080",
		PollInterval:   time.Second,
		ReportInterval: time.Second,
		Key:            "test_key",
		CryptoKey:      "/nonexistent/key.pem", // Invalid path - should fallback
		RateLimit:      1,
	}

	// This will create agent but with fallback client since crypto key is invalid
	agent := NewAgent(config)

	if agent == nil {
		t.Error("expected agent to be created")
	}

	if agent == nil {
		t.Error("expected agent to be created")
		return
	}

	if agent.pollInterval != time.Second {
		t.Errorf("expected poll interval 1s, got %v", agent.pollInterval)
	}

	if agent.reportInterval != time.Second {
		t.Errorf("expected report interval 1s, got %v", agent.reportInterval)
	}

	if agent.rateLimit != 1 {
		t.Errorf("expected rate limit 1, got %d", agent.rateLimit)
	}
}

func TestAgent_NewAgent_WithoutCryptoKey(t *testing.T) {
	os.Setenv("CONFIG", "")
	defer os.Unsetenv("CONFIG")

	config := Config{
		ServerURL:      "http://localhost:8080",
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		Key:            "test_key",
		CryptoKey:      "",
		RateLimit:      5,
	}

	agent := NewAgent(config)

	if agent == nil {
		t.Error("expected agent to be created")
	}

	if agent == nil {
		t.Error("expected agent to be created")
		return
	}

	if agent.pollInterval != 2*time.Second {
		t.Errorf("expected poll interval 2s, got %v", agent.pollInterval)
	}

	if agent.reportInterval != 10*time.Second {
		t.Errorf("expected report interval 10s, got %v", agent.reportInterval)
	}

	if agent.rateLimit != 5 {
		t.Errorf("expected rate limit 5, got %d", agent.rateLimit)
	}

	if agent.metricsChan == nil {
		t.Error("expected metricsChan to be created")
	}

	if agent.done == nil {
		t.Error("expected done channel to be created")
	}
}
