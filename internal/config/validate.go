package config

import (
	"fmt"
	"strings"
	"time"
)

func Validate(cfg Config) error {
	if cfg.Version == 0 {
		return fmt.Errorf("config version must be greater than 0")
	}
	for name, profile := range cfg.Profiles {
		if err := validateProfile(name, profile); err != nil {
			return err
		}
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
	if cfg.JMX.ScrapeTimeout != "" {
		if _, err := time.ParseDuration(cfg.JMX.ScrapeTimeout); err != nil {
			return fmt.Errorf("jmx.scrape_timeout: %w", err)
		}
	}
	if strings.TrimSpace(cfg.JMX.Path) != "" && !strings.HasPrefix(strings.TrimSpace(cfg.JMX.Path), "/") {
		return fmt.Errorf("jmx.path must start with /")
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
	if cfg.Host.FDWarnPct < 0 || cfg.Host.FDWarnPct > 100 {
		return fmt.Errorf("host.fd_warn_pct must be between 0 and 100")
	}
	if cfg.Host.FDCritPct < 0 || cfg.Host.FDCritPct > 100 {
		return fmt.Errorf("host.fd_crit_pct must be between 0 and 100")
	}
	if cfg.Host.ClockSkewWarnMs < 0 {
		return fmt.Errorf("host.clock_skew_warn_ms must be greater than or equal to 0")
	}
	if cfg.Thresholds.DiskWarnPct < 0 || cfg.Thresholds.DiskWarnPct > 100 {
		return fmt.Errorf("thresholds.disk_warn_pct must be between 0 and 100")
	}
	if cfg.Thresholds.DiskCritPct < 0 || cfg.Thresholds.DiskCritPct > 100 {
		return fmt.Errorf("thresholds.disk_crit_pct must be between 0 and 100")
	}
	if cfg.Thresholds.InodeWarnPct < 0 || cfg.Thresholds.InodeWarnPct > 100 {
		return fmt.Errorf("thresholds.inode_warn_pct must be between 0 and 100")
	}
	if cfg.Thresholds.ConsumerLagWarn < 0 {
		return fmt.Errorf("thresholds.consumer_lag_warn must be greater than or equal to 0")
	}
	if cfg.Thresholds.ConsumerLagCrit < 0 {
		return fmt.Errorf("thresholds.consumer_lag_crit must be greater than or equal to 0")
	}
	if cfg.Thresholds.CertExpiryWarnDays < 0 {
		return fmt.Errorf("thresholds.cert_expiry_warn_days must be greater than or equal to 0")
	}
	if cfg.Thresholds.ProduceThrottleWarnMs < 0 {
		return fmt.Errorf("thresholds.produce_throttle_warn_ms must be greater than or equal to 0")
	}
	if cfg.Thresholds.FetchThrottleWarnMs < 0 {
		return fmt.Errorf("thresholds.fetch_throttle_warn_ms must be greater than or equal to 0")
	}
	if cfg.Thresholds.RequestLatencyWarnMs < 0 {
		return fmt.Errorf("thresholds.request_latency_warn_ms must be greater than or equal to 0")
	}
	if cfg.Thresholds.PurgatoryWarnCount < 0 {
		return fmt.Errorf("thresholds.purgatory_warn_count must be greater than or equal to 0")
	}
	if cfg.Thresholds.HeapUsedWarnPct < 0 || cfg.Thresholds.HeapUsedWarnPct > 100 {
		return fmt.Errorf("thresholds.heap_used_warn_pct must be between 0 and 100")
	}
	if cfg.Thresholds.GCPauseWarnMs < 0 {
		return fmt.Errorf("thresholds.gc_pause_warn_ms must be greater than or equal to 0")
	}
	if cfg.Diagnosis.MaxRootCauses < 0 {
		return fmt.Errorf("diagnosis.max_root_causes must be greater than or equal to 0")
	}
	return nil
}

func validateProfile(name string, profile ProfileConfig) error {
	switch strings.ToLower(strings.TrimSpace(profile.ExecutionView)) {
	case "", "auto", "internal", "external", "host-network", "docker-container", "bastion":
	default:
		return fmt.Errorf("profiles.%s.execution_view has unsupported value", name)
	}

	switch strings.ToLower(strings.TrimSpace(profile.SecurityMode)) {
	case "", "plaintext", "ssl", "tls", "sasl", "sasl_plaintext", "sasl_ssl":
	default:
		return fmt.Errorf("profiles.%s.security_mode has unsupported value", name)
	}

	switch strings.ToLower(strings.TrimSpace(profile.Producer.Acks)) {
	case "", "0", "1", "all":
	default:
		return fmt.Errorf("profiles.%s.producer.acks has unsupported value", name)
	}

	switch strings.ToLower(strings.TrimSpace(profile.Producer.ExpectedDurability)) {
	case "", "best_effort", "strong":
	default:
		return fmt.Errorf("profiles.%s.producer.expected_durability has unsupported value", name)
	}

	switch strings.ToLower(strings.TrimSpace(profile.Consumer.AutoOffsetReset)) {
	case "", "latest", "earliest", "none", "by_duration":
	default:
		return fmt.Errorf("profiles.%s.consumer.auto_offset_reset has unsupported value", name)
	}

	switch strings.ToLower(strings.TrimSpace(profile.Consumer.IsolationLevel)) {
	case "", "read_uncommitted", "read_committed":
	default:
		return fmt.Errorf("profiles.%s.consumer.isolation_level has unsupported value", name)
	}
	return nil
}
