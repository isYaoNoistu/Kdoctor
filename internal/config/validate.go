package config

import (
	"fmt"
	"time"
)

func Validate(cfg Config) error {
	if cfg.Version == 0 {
		return fmt.Errorf("config version must be greater than 0")
	}
	if cfg.Execution.Timeout != "" {
		if _, err := time.ParseDuration(cfg.Execution.Timeout); err != nil {
			return fmt.Errorf("execution.timeout: %w", err)
		}
	}
	if cfg.Execution.MetadataTimeout != "" {
		if _, err := time.ParseDuration(cfg.Execution.MetadataTimeout); err != nil {
			return fmt.Errorf("execution.metadata_timeout: %w", err)
		}
	}
	if cfg.Execution.TCPTimeout != "" {
		if _, err := time.ParseDuration(cfg.Execution.TCPTimeout); err != nil {
			return fmt.Errorf("execution.tcp_timeout: %w", err)
		}
	}
	if cfg.Probe.Timeout != "" {
		if _, err := time.ParseDuration(cfg.Probe.Timeout); err != nil {
			return fmt.Errorf("probe.timeout: %w", err)
		}
	}
	return nil
}
