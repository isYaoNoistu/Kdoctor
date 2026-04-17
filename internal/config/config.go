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
	Diagnosis      DiagnosisConfig          `json:"diagnosis" yaml:"diagnosis"`
}

type ProfileConfig struct {
	BootstrapExternal         []string `json:"bootstrap_external" yaml:"bootstrap_external"`
	BootstrapInternal         []string `json:"bootstrap_internal" yaml:"bootstrap_internal"`
	ControllerEndpoints       []string `json:"controller_endpoints" yaml:"controller_endpoints"`
	BrokerCount               int      `json:"broker_count" yaml:"broker_count"`
	ExpectedMinISR            int      `json:"expected_min_isr" yaml:"expected_min_isr"`
	ExpectedReplicationFactor int      `json:"expected_replication_factor" yaml:"expected_replication_factor"`
	HostNetwork               bool     `json:"host_network" yaml:"host_network"`
	PlaintextExternal         bool     `json:"plaintext_external" yaml:"plaintext_external"`
}

type DockerConfig struct {
	Enabled        bool     `json:"enabled" yaml:"enabled"`
	ComposeFile    string   `json:"compose_file" yaml:"compose_file"`
	ContainerNames []string `json:"container_names" yaml:"container_names"`
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
	Enabled      bool   `json:"enabled" yaml:"enabled"`
	Topic        string `json:"topic" yaml:"topic"`
	GroupPrefix  string `json:"group_prefix" yaml:"group_prefix"`
	Timeout      string `json:"timeout" yaml:"timeout"`
	MessageBytes int    `json:"message_bytes" yaml:"message_bytes"`
	ProduceCount int    `json:"produce_count" yaml:"produce_count"`
	Cleanup      bool   `json:"cleanup" yaml:"cleanup"`
}

type ExecutionConfig struct {
	Timeout         string `json:"timeout" yaml:"timeout"`
	MetadataTimeout string `json:"metadata_timeout" yaml:"metadata_timeout"`
	TCPTimeout      string `json:"tcp_timeout" yaml:"tcp_timeout"`
	AdminAPITimeout string `json:"admin_api_timeout" yaml:"admin_api_timeout"`
	JMXTimeout      string `json:"jmx_timeout" yaml:"jmx_timeout"`
}

type DiagnosisConfig struct {
	MaxRootCauses    int  `json:"max_root_causes" yaml:"max_root_causes"`
	EnableConfidence bool `json:"enable_confidence" yaml:"enable_confidence"`
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
