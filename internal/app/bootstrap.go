package app

import (
	"fmt"
	"log/slog"
	"os"

	"kdoctor/internal/config"
)

func Bootstrap(opts Options) (*config.Runtime, error) {
	fileCfg, err := config.LoadFile(opts.ConfigPath)
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
	probeTimeout, err := parseDurationOrDefault("", cfg.Probe.Timeout)
	if err != nil {
		return nil, fmt.Errorf("resolve probe timeout: %w", err)
	}

	composePath := cfg.Docker.ComposeFile
	if opts.ComposePath != "" {
		composePath = opts.ComposePath
	}

	logDir := cfg.Logs.LogDir
	if opts.LogDir != "" {
		logDir = opts.LogDir
	}

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	return &config.Runtime{
		Mode:                  opts.Mode,
		ProfileName:           selectedProfile,
		Config:                cfg,
		SelectedProfile:       profileCfg,
		BootstrapInternal:     internalBootstraps,
		BootstrapExternal:     externalBootstraps,
		ControllerEndpoints:   profileCfg.ControllerEndpoints,
		ComposePath:           composePath,
		LogDir:                logDir,
		EnableDocker:          cfg.Docker.Enabled,
		EnableHost:            true,
		EnableJMX:             false,
		ProbeTopic:            cfg.Probe.Topic,
		ProbeGroupPrefix:      cfg.Probe.GroupPrefix,
		ProbeTimeout:          probeTimeout,
		ProbeMessageBytes:     cfg.Probe.MessageBytes,
		ProbeProduceCount:     cfg.Probe.ProduceCount,
		Timeout:               timeout,
		MetadataTimeout:       metadataTimeout,
		TCPTimeout:            tcpTimeout,
		MinimumOutputSeverity: opts.Severity,
		Logger:                logger,
	}, nil
}
