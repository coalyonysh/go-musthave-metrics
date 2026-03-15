package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadAgentConfig(t *testing.T) {
	t.Run("empty filename", func(t *testing.T) {
		cfg, err := LoadAgentConfig("")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if cfg != nil {
			t.Errorf("expected nil config, got %v", cfg)
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		cfg, err := LoadAgentConfig("/non/existent/path.json")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if cfg != nil {
			t.Errorf("expected nil config for non-existent file, got %v", cfg)
		}
	})

	t.Run("valid config file", func(t *testing.T) {
		// Create temp file
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "agent.json")
		content := `{"address": "localhost:8080", "report_interval": "10s", "poll_interval": "2s", "crypto_key": "/path/to/key"}`
		err := os.WriteFile(configFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("failed to write temp config: %v", err)
		}

		cfg, err := LoadAgentConfig(configFile)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if cfg == nil {
			t.Fatal("expected config, got nil")
		}
		if cfg.Address != "localhost:8080" {
			t.Errorf("expected address 'localhost:8080', got %q", cfg.Address)
		}
		if cfg.ReportInterval != "10s" {
			t.Errorf("expected report_interval '10s', got %q", cfg.ReportInterval)
		}
		if cfg.PollInterval != "2s" {
			t.Errorf("expected poll_interval '2s', got %q", cfg.PollInterval)
		}
		if cfg.CryptoKey != "/path/to/key" {
			t.Errorf("expected crypto_key '/path/to/key', got %q", cfg.CryptoKey)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "invalid.json")
		err := os.WriteFile(configFile, []byte("not valid json"), 0644)
		if err != nil {
			t.Fatalf("failed to write temp config: %v", err)
		}

		cfg, err := LoadAgentConfig(configFile)
		if err == nil {
			t.Error("expected error for invalid JSON, got nil")
		}
		if cfg != nil {
			t.Error("expected nil config for invalid JSON")
		}
	})
}

func TestLoadServerConfig(t *testing.T) {
	t.Run("empty filename", func(t *testing.T) {
		cfg, err := LoadServerConfig("")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if cfg != nil {
			t.Errorf("expected nil config, got %v", cfg)
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		cfg, err := LoadServerConfig("/non/existent/path.json")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if cfg != nil {
			t.Errorf("expected nil config for non-existent file, got %v", cfg)
		}
	})

	t.Run("valid config file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "server.json")
		content := `{"address": "localhost:8080", "restore": true, "store_interval": "300s", "store_file": "/tmp/metrics.db", "database_dsn": "postgres://localhost:5432/metrics", "crypto_key": "/path/to/key"}`
		err := os.WriteFile(configFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("failed to write temp config: %v", err)
		}

		cfg, err := LoadServerConfig(configFile)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if cfg == nil {
			t.Fatal("expected config, got nil")
		}
		if cfg.Address != "localhost:8080" {
			t.Errorf("expected address 'localhost:8080', got %q", cfg.Address)
		}
		if cfg.Restore != true {
			t.Errorf("expected restore true, got %v", cfg.Restore)
		}
		if cfg.StoreInterval != "300s" {
			t.Errorf("expected store_interval '300s', got %q", cfg.StoreInterval)
		}
		if cfg.StoreFile != "/tmp/metrics.db" {
			t.Errorf("expected store_file '/tmp/metrics.db', got %q", cfg.StoreFile)
		}
		if cfg.DatabaseDSN != "postgres://localhost:5432/metrics" {
			t.Errorf("expected database_dsn 'postgres://localhost:5432/metrics', got %q", cfg.DatabaseDSN)
		}
		if cfg.CryptoKey != "/path/to/key" {
			t.Errorf("expected crypto_key '/path/to/key', got %q", cfg.CryptoKey)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "invalid.json")
		err := os.WriteFile(configFile, []byte("not valid json"), 0644)
		if err != nil {
			t.Fatalf("failed to write temp config: %v", err)
		}

		cfg, err := LoadServerConfig(configFile)
		if err == nil {
			t.Error("expected error for invalid JSON, got nil")
		}
		if cfg != nil {
			t.Error("expected nil config for invalid JSON")
		}
	})
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{"empty string", "", 0, false},
		{"valid duration", "10s", 10 * time.Second, false},
		{"valid duration minutes", "5m", 5 * time.Minute, false},
		{"invalid duration", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}
