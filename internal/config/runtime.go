package config

import (
	"log/slog"
	"time"
)

type Runtime struct {
	Mode                      string
	ProfileName               string
	Config                    Config
	SelectedProfile           ProfileConfig
	BootstrapInternal         []string
	BootstrapExternal         []string
	ControllerEndpoints       []string
	ComposePath               string
	LogDir                    string
	EnableDocker              bool
	EnableHost                bool
	EnableJMX                 bool
	LogFreshnessWindow        time.Duration
	LogMinLinesPerSource      int
	LogMaxFiles               int
	LogMaxBytesPerSource      int
	LogCustomPatternsDir      string
	ProbeTopic                string
	ProbeGroupPrefix          string
	ProbeTimeout              time.Duration
	ProbeMessageBytes         int
	ProbeProduceCount         int
	Timeout                   time.Duration
	MetadataTimeout           time.Duration
	TCPTimeout                time.Duration
	AdminAPITimeout           time.Duration
	JMXTimeout                time.Duration
	JMXScrapeTimeout          time.Duration
	JMXPath                   string
	JMXEndpoints              []string
	DiagnosisMaxRootCauses    int
	DiagnosisEnableConfidence bool
	MinimumOutputSeverity     string
	OutputMaxEvidenceItems    int
	OutputShowPassChecks      bool
	OutputShowSkipChecks      bool
	OutputVerbose             bool
	Logger                    *slog.Logger
}
