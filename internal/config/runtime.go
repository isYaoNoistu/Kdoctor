package config

import (
	"log/slog"
	"time"
)

type Runtime struct {
	Mode                  string
	ProfileName           string
	Config                Config
	SelectedProfile       ProfileConfig
	BootstrapInternal     []string
	BootstrapExternal     []string
	ControllerEndpoints   []string
	ComposePath           string
	LogDir                string
	EnableDocker          bool
	EnableHost            bool
	EnableJMX             bool
	ProbeTopic            string
	ProbeGroupPrefix      string
	ProbeTimeout          time.Duration
	ProbeMessageBytes     int
	ProbeProduceCount     int
	Timeout               time.Duration
	MetadataTimeout       time.Duration
	TCPTimeout            time.Duration
	MinimumOutputSeverity string
	Logger                *slog.Logger
}
