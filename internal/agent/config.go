package agent

import (
	"fmt"
	"strings"
	"time"
)

type Config struct {
	ServerURL      string
	PollInterval   time.Duration
	ReportInterval time.Duration
}

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

	return nil
}
