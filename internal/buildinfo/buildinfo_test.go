package buildinfo

import (
	"testing"
)

func TestPrintBuildInfo(t *testing.T) {
	// Тест на вывод с заполненными значениями
	t.Run("with values", func(t *testing.T) {
		// Просто проверяем, что функция не паникует
		PrintBuildInfo("v1.0.0", "2024-01-01", "abc123")
	})

	// Тест на вывод с пустыми значениями
	t.Run("with empty values", func(t *testing.T) {
		PrintBuildInfo("", "", "")
	})

	// Тест на вывод с частично заполненными значениями
	t.Run("with partial values", func(t *testing.T) {
		PrintBuildInfo("v1.0.0", "", "")
	})
}

func TestNormalizeValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"empty string", "", "N/A"},
		{"non-empty string", "test", "test"},
		{"version", "v1.0.0", "v1.0.0"},
		{"date", "2024-01-01", "2024-01-01"},
		{"commit", "abc123", "abc123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeValue(tt.value)
			if result != tt.expected {
				t.Errorf("normalizeValue(%q) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}
