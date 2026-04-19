package app

import (
	"fmt"
	"log/slog"
	"os"

	"kdoctor/internal/config"
)

func Bootstrap(opts Options) (*config.Runtime, error) {
	fileCfg, err := config.LoadFile(opts.ConfigPath, opts.ConfigPathExplicit)
	if err != nil {
		return nil, err
	}

	selectedProfile := availableProfileName(opts, fileCfg)
	cfg := config.Default()
	cfg = mergeProfileConfig(cfg, selectedProfile)
	cfg = config.Merge(cfg, fileCfg)
	cfg = mergeProfileConfig(cfg, selectedProfile)

	if err := config.Validate(cfg); err != nil {
		return nil, err
	}

	profileCfg := cfg.Profiles[selectedProfile]
	internalBootstraps := profileCfg.BootstrapInternal
	externalBootstraps := profileCfg.BootstrapExternal
	if override := parseCSV(opts.Bootstrap); len(override) > 0 {
		externalBootstraps = override
	}
	if override := parseCSV(opts.BootstrapInternal); len(override) > 0 {
		internalBootstraps = override
	}
	if override := parseCSV(opts.BootstrapExternal); len(override) > 0 {
		externalBootstraps = override
	}

	timeout, err := parseDurationOrDefault(opts.Timeout, cfg.Execution.Timeout)
	if err != nil {
		return nil, fmt.Errorf("resolve execution timeout: %w", err)
	}
	metadataTimeout, err := parseDurationOrDefault("", cfg.Execution.MetadataTimeout)
	if err != nil {
		return nil, fmt.Errorf("resolve metadata timeout: %w", err)
	}
	tcpTimeout, err := parseDurationOrDefault("", cfg.Execution.TCPTimeout)
	if err != nil {
		return nil, fmt.Errorf("resolve tcp timeout: %w", err)
	}
	adminAPITimeout, err := parseDurationOrDefault("", cfg.Execution.AdminAPITimeout)
	if err != nil {
		return nil, fmt.Errorf("resolve admin api timeout: %w", err)
	}
	jmxTimeout, err := parseDurationOrDefault("", cfg.Execution.JMXTimeout)
	if err != nil {
		return nil, fmt.Errorf("resolve jmx timeout: %w", err)
	}
	jmxScrapeTimeout, err := parseDurationOrDefault("", cfg.JMX.ScrapeTimeout)
	if err != nil {
		return nil, fmt.Errorf("resolve jmx scrape timeout: %w", err)
	}
	probeTimeout, err := parseDurationOrDefault("", cfg.Probe.Timeout)
	if err != nil {
		return nil, fmt.Errorf("resolve probe timeout: %w", err)
	}
	logFreshnessWindow, err := parseDurationOrDefault("", cfg.Logs.FreshnessWindow)
	if err != nil {
		return nil, fmt.Errorf("resolve logs freshness window: %w", err)
	}

	composePath := config.NormalizeInputPath(cfg.Docker.ComposeFile)
	if opts.ComposePath != "" {
		composePath = config.NormalizeInputPath(opts.ComposePath)
	}

	logDir := config.NormalizeInputPath(cfg.Logs.LogDir)
	if opts.LogDir != "" {
		logDir = config.NormalizeInputPath(opts.LogDir)
	}
	logCustomPatternsDir := config.NormalizeInputPath(cfg.Logs.CustomPatternsDir)

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	outputVerbose := cfg.Output.Verbose || opts.Verbose

	return &config.Runtime{
		Mode:                      opts.Mode,
		ProfileName:               selectedProfile,
		Config:                    cfg,
		SelectedProfile:           profileCfg,
		BootstrapInternal:         internalBootstraps,
		BootstrapExternal:         externalBootstraps,
		ControllerEndpoints:       profileCfg.ControllerEndpoints,
		ComposePath:               composePath,
		LogDir:                    logDir,
		EnableDocker:              cfg.Docker.Enabled,
		EnableHost:                cfg.Host.Enabled,
		EnableJMX:                 cfg.JMX.Enabled,
		LogFreshnessWindow:        logFreshnessWindow,
		LogMinLinesPerSource:      cfg.Logs.MinLinesPerSource,
		LogMaxFiles:               cfg.Logs.MaxFiles,
		LogMaxBytesPerSource:      cfg.Logs.MaxBytesPerSource,
		LogCustomPatternsDir:      logCustomPatternsDir,
		ProbeTopic:                cfg.Probe.Topic,
		ProbeGroupPrefix:          cfg.Probe.GroupPrefix,
		ProbeTimeout:              probeTimeout,
		ProbeMessageBytes:         cfg.Probe.MessageBytes,
		ProbeProduceCount:         cfg.Probe.ProduceCount,
		Timeout:                   timeout,
		MetadataTimeout:           metadataTimeout,
		TCPTimeout:                tcpTimeout,
		AdminAPITimeout:           adminAPITimeout,
		JMXTimeout:                jmxTimeout,
		JMXScrapeTimeout:          jmxScrapeTimeout,
		JMXPath:                   cfg.JMX.Path,
		JMXEndpoints:              append([]string(nil), cfg.JMX.Endpoints...),
		DiagnosisMaxRootCauses:    cfg.Diagnosis.MaxRootCauses,
		DiagnosisEnableConfidence: cfg.Diagnosis.EnableConfidence,
		MinimumOutputSeverity:     opts.Severity,
		OutputMaxEvidenceItems:    cfg.Output.MaxEvidenceItems,
		OutputShowPassChecks:      cfg.Output.ShowPassChecks,
		OutputShowSkipChecks:      cfg.Output.ShowSkipChecks,
		OutputVerbose:             outputVerbose,
		Logger:                    logger,
	}, nil
}
