package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version        int                      `json:"version" yaml:"version"`
	DefaultProfile string                   `json:"default_profile" yaml:"default_profile"`
	Profiles       map[string]ProfileConfig `json:"profiles" yaml:"profiles"`
	Docker         DockerConfig             `json:"docker" yaml:"docker"`
	Logs           LogConfig                `json:"logs" yaml:"logs"`
	Probe          ProbeConfig              `json:"probe" yaml:"probe"`
	Execution      ExecutionConfig          `json:"execution" yaml:"execution"`
	JMX            JMXConfig                `json:"jmx" yaml:"jmx"`
	Host           HostConfig               `json:"host" yaml:"host"`
	Thresholds     ThresholdConfig          `json:"thresholds" yaml:"thresholds"`
	Diagnosis      DiagnosisConfig          `json:"diagnosis" yaml:"diagnosis"`
}

type ProfileConfig struct {
	BootstrapExternal         []string            `json:"bootstrap_external" yaml:"bootstrap_external"`
	BootstrapInternal         []string            `json:"bootstrap_internal" yaml:"bootstrap_internal"`
	ControllerEndpoints       []string            `json:"controller_endpoints" yaml:"controller_endpoints"`
	BrokerCount               int                 `json:"broker_count" yaml:"broker_count"`
	ExpectedMinISR            int                 `json:"expected_min_isr" yaml:"expected_min_isr"`
	ExpectedReplicationFactor int                 `json:"expected_replication_factor" yaml:"expected_replication_factor"`
	HostNetwork               bool                `json:"host_network" yaml:"host_network"`
	PlaintextExternal         bool                `json:"plaintext_external" yaml:"plaintext_external"`
	ExecutionView             string              `json:"execution_view" yaml:"execution_view"`
	SecurityMode              string              `json:"security_mode" yaml:"security_mode"`
	SASLMechanism             string              `json:"sasl_mechanism" yaml:"sasl_mechanism"`
	GroupProbeTargets         []GroupProbeTarget  `json:"group_probe_targets" yaml:"group_probe_targets"`
	Producer                  ProducerAuditConfig `json:"producer" yaml:"producer"`
	Consumer                  ConsumerAuditConfig `json:"consumer" yaml:"consumer"`
}

type ProducerAuditConfig struct {
	Acks                 string `json:"acks" yaml:"acks"`
	EnableIdempotence    bool   `json:"enable_idempotence" yaml:"enable_idempotence"`
	Retries              int    `json:"retries" yaml:"retries"`
	MaxInFlight          int    `json:"max_in_flight" yaml:"max_in_flight"`
	DeliveryTimeoutMs    int    `json:"delivery_timeout_ms" yaml:"delivery_timeout_ms"`
	RequestTimeoutMs     int    `json:"request_timeout_ms" yaml:"request_timeout_ms"`
	LingerMs             int    `json:"linger_ms" yaml:"linger_ms"`
	ExpectedDurability   string `json:"expected_durability" yaml:"expected_durability"`
	TransactionalID      string `json:"transactional_id" yaml:"transactional_id"`
	TransactionTimeoutMs int    `json:"transaction_timeout_ms" yaml:"transaction_timeout_ms"`
}

type ConsumerAuditConfig struct {
	MaxPollIntervalMs   int    `json:"max_poll_interval_ms" yaml:"max_poll_interval_ms"`
	SessionTimeoutMs    int    `json:"session_timeout_ms" yaml:"session_timeout_ms"`
	HeartbeatIntervalMs int    `json:"heartbeat_interval_ms" yaml:"heartbeat_interval_ms"`
	AutoOffsetReset     string `json:"auto_offset_reset" yaml:"auto_offset_reset"`
	IsolationLevel      string `json:"isolation_level" yaml:"isolation_level"`
}

type GroupProbeTarget struct {
	Name    string `json:"name" yaml:"name"`
	GroupID string `json:"group_id" yaml:"group_id"`
	Topic   string `json:"topic" yaml:"topic"`
	LagWarn int64  `json:"lag_warn" yaml:"lag_warn"`
	LagCrit int64  `json:"lag_crit" yaml:"lag_crit"`
}

type DockerConfig struct {
	Enabled        bool     `json:"enabled" yaml:"enabled"`
	ComposeFile    string   `json:"compose_file" yaml:"compose_file"`
	ContainerNames []string `json:"container_names" yaml:"container_names"`
	InspectMounts  bool     `json:"inspect_mounts" yaml:"inspect_mounts"`
}

type LogConfig struct {
	Enabled           bool   `json:"enabled" yaml:"enabled"`
	LogDir            string `json:"log_dir" yaml:"log_dir"`
	TailLines         int    `json:"tail_lines" yaml:"tail_lines"`
	LookbackMinutes   int    `json:"lookback_minutes" yaml:"lookback_minutes"`
	MinLinesPerSource int    `json:"min_lines_per_source" yaml:"min_lines_per_source"`
	FreshnessWindow   string `json:"freshness_window" yaml:"freshness_window"`
	MaxFiles          int    `json:"max_files" yaml:"max_files"`
	MaxBytesPerSource int    `json:"max_bytes_per_source" yaml:"max_bytes_per_source"`
	CustomPatternsDir string `json:"custom_patterns_dir" yaml:"custom_patterns_dir"`
}

type ProbeConfig struct {
	Enabled           bool   `json:"enabled" yaml:"enabled"`
	Topic             string `json:"topic" yaml:"topic"`
	GroupPrefix       string `json:"group_prefix" yaml:"group_prefix"`
	Timeout           string `json:"timeout" yaml:"timeout"`
	MessageBytes      int    `json:"message_bytes" yaml:"message_bytes"`
	ProduceCount      int    `json:"produce_count" yaml:"produce_count"`
	Cleanup           bool   `json:"cleanup" yaml:"cleanup"`
	Acks              string `json:"acks" yaml:"acks"`
	EnableIdempotence bool   `json:"enable_idempotence" yaml:"enable_idempotence"`
	CleanupMode       string `json:"cleanup_mode" yaml:"cleanup_mode"`
	TXProbeEnabled    bool   `json:"tx_probe_enabled" yaml:"tx_probe_enabled"`
}

type ExecutionConfig struct {
	Timeout         string `json:"timeout" yaml:"timeout"`
	MetadataTimeout string `json:"metadata_timeout" yaml:"metadata_timeout"`
	TCPTimeout      string `json:"tcp_timeout" yaml:"tcp_timeout"`
	AdminAPITimeout string `json:"admin_api_timeout" yaml:"admin_api_timeout"`
	JMXTimeout      string `json:"jmx_timeout" yaml:"jmx_timeout"`
}

type JMXConfig struct {
	Enabled       bool     `json:"enabled" yaml:"enabled"`
	ScrapeTimeout string   `json:"scrape_timeout" yaml:"scrape_timeout"`
	Path          string   `json:"path" yaml:"path"`
	Endpoints     []string `json:"endpoints" yaml:"endpoints"`
	MetricSets    []string `json:"metric_sets" yaml:"metric_sets"`
}

type HostConfig struct {
	Enabled         bool     `json:"enabled" yaml:"enabled"`
	DiskPaths       []string `json:"disk_paths" yaml:"disk_paths"`
	CheckPorts      []int    `json:"check_ports" yaml:"check_ports"`
	FDWarnPct       int      `json:"fd_warn_pct" yaml:"fd_warn_pct"`
	FDCritPct       int      `json:"fd_crit_pct" yaml:"fd_crit_pct"`
	ClockSkewWarnMs int      `json:"clock_skew_warn_ms" yaml:"clock_skew_warn_ms"`
}

type ThresholdConfig struct {
	URPWarn               int     `json:"urp_warn" yaml:"urp_warn"`
	UnderMinISRCrit       int     `json:"under_min_isr_crit" yaml:"under_min_isr_crit"`
	NetworkIdleWarn       float64 `json:"network_idle_warn" yaml:"network_idle_warn"`
	RequestIdleWarn       float64 `json:"request_idle_warn" yaml:"request_idle_warn"`
	DiskWarnPct           float64 `json:"disk_warn_pct" yaml:"disk_warn_pct"`
	DiskCritPct           float64 `json:"disk_crit_pct" yaml:"disk_crit_pct"`
	InodeWarnPct          float64 `json:"inode_warn_pct" yaml:"inode_warn_pct"`
	ReplicaLagWarn        int64   `json:"replica_lag_warn" yaml:"replica_lag_warn"`
	LeaderSkewWarnPct     float64 `json:"leader_skew_warn_pct" yaml:"leader_skew_warn_pct"`
	ConsumerLagWarn       int64   `json:"consumer_lag_warn" yaml:"consumer_lag_warn"`
	ConsumerLagCrit       int64   `json:"consumer_lag_crit" yaml:"consumer_lag_crit"`
	CertExpiryWarnDays    int     `json:"cert_expiry_warn_days" yaml:"cert_expiry_warn_days"`
	ProduceThrottleWarnMs float64 `json:"produce_throttle_warn_ms" yaml:"produce_throttle_warn_ms"`
	FetchThrottleWarnMs   float64 `json:"fetch_throttle_warn_ms" yaml:"fetch_throttle_warn_ms"`
	RequestLatencyWarnMs  float64 `json:"request_latency_warn_ms" yaml:"request_latency_warn_ms"`
	PurgatoryWarnCount    float64 `json:"purgatory_warn_count" yaml:"purgatory_warn_count"`
	HeapUsedWarnPct       float64 `json:"heap_used_warn_pct" yaml:"heap_used_warn_pct"`
	GCPauseWarnMs         float64 `json:"gc_pause_warn_ms" yaml:"gc_pause_warn_ms"`
}

type DiagnosisConfig struct {
	MaxRootCauses              int      `json:"max_root_causes" yaml:"max_root_causes"`
	EnableConfidence           bool     `json:"enable_confidence" yaml:"enable_confidence"`
	SuppressDownstreamSymptoms bool     `json:"suppress_downstream_symptoms" yaml:"suppress_downstream_symptoms"`
	RulePacks                  []string `json:"rule_packs" yaml:"rule_packs"`
}

func NormalizeInputPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	return strings.ReplaceAll(path, "\\", string(os.PathSeparator))
}

func LoadFile(path string, strict bool) (Config, error) {
	path = NormalizeInputPath(path)
	if path == "" {
		return Config{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if strict {
				return Config{}, fmt.Errorf("config file not found: %s", path)
			}
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("read config file: %w", err)
	}

	cfg := Config{}
	switch {
	case strings.HasSuffix(path, ".json"):
		if err := json.Unmarshal(data, &cfg); err != nil {
			return Config{}, fmt.Errorf("parse json config: %w", err)
		}
	default:
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return Config{}, fmt.Errorf("parse yaml config: %w", err)
		}
	}
	return cfg, nil
}
