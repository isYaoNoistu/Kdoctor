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
	if cfg.Execution.AdminAPITimeout != "" {
		if _, err := time.ParseDuration(cfg.Execution.AdminAPITimeout); err != nil {
			return fmt.Errorf("execution.admin_api_timeout: %w", err)
		}
	}
	if cfg.Execution.JMXTimeout != "" {
		if _, err := time.ParseDuration(cfg.Execution.JMXTimeout); err != nil {
			return fmt.Errorf("execution.jmx_timeout: %w", err)
		}
	}
	if cfg.Probe.Timeout != "" {
		if _, err := time.ParseDuration(cfg.Probe.Timeout); err != nil {
			return fmt.Errorf("probe.timeout: %w", err)
		}
	}
	if cfg.Logs.FreshnessWindow != "" {
		if _, err := time.ParseDuration(cfg.Logs.FreshnessWindow); err != nil {
			return fmt.Errorf("logs.freshness_window: %w", err)
		}
	}
	if cfg.Logs.MinLinesPerSource < 0 {
		return fmt.Errorf("logs.min_lines_per_source must be greater than or equal to 0")
	}
	if cfg.Logs.MaxFiles < 0 {
		return fmt.Errorf("logs.max_files must be greater than or equal to 0")
	}
	if cfg.Logs.MaxBytesPerSource < 0 {
		return fmt.Errorf("logs.max_bytes_per_source must be greater than or equal to 0")
	}
	if cfg.Diagnosis.MaxRootCauses < 0 {
		return fmt.Errorf("diagnosis.max_root_causes must be greater than or equal to 0")
	}
	return nil
}
