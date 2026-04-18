package config

func Default() Config {
	return Config{
		Version:        2,
		DefaultProfile: "generic-bootstrap",
		Profiles:       map[string]ProfileConfig{},
		Docker: DockerConfig{
			Enabled:       true,
			InspectMounts: true,
		},
		Logs: LogConfig{
			Enabled:           true,
			TailLines:         300,
			LookbackMinutes:   15,
			MinLinesPerSource: 20,
			FreshnessWindow:   "15m",
			MaxFiles:          12,
			MaxBytesPerSource: 1024 * 1024,
		},
		Probe: ProbeConfig{
			Enabled:           true,
			Topic:             "_kdoctor_probe",
			GroupPrefix:       "kdoctor-probe",
			Timeout:           "15s",
			MessageBytes:      1024,
			ProduceCount:      1,
			Acks:              "all",
			EnableIdempotence: false,
			CleanupMode:       "disabled",
		},
		Execution: ExecutionConfig{
			Timeout:         "30s",
			MetadataTimeout: "5s",
			TCPTimeout:      "3s",
			AdminAPITimeout: "15s",
			JMXTimeout:      "5s",
		},
		JMX: JMXConfig{
			Enabled:       false,
			ScrapeTimeout: "5s",
			Path:          "/metrics",
			MetricSets:    []string{"kraft", "broker", "replica", "request", "quota", "jvm"},
		},
		Host: HostConfig{
			Enabled:         true,
			FDWarnPct:       70,
			FDCritPct:       85,
			ClockSkewWarnMs: 500,
		},
		Thresholds: ThresholdConfig{
			URPWarn:               1,
			UnderMinISRCrit:       1,
			NetworkIdleWarn:       0.3,
			RequestIdleWarn:       0.3,
			DiskWarnPct:           75,
			DiskCritPct:           85,
			InodeWarnPct:          80,
			ReplicaLagWarn:        10000,
			LeaderSkewWarnPct:     30,
			ConsumerLagWarn:       1000,
			ConsumerLagCrit:       10000,
			CertExpiryWarnDays:    30,
			ProduceThrottleWarnMs: 1,
			FetchThrottleWarnMs:   1,
			RequestLatencyWarnMs:  100,
			PurgatoryWarnCount:    1,
			HeapUsedWarnPct:       85,
			GCPauseWarnMs:         200,
		},
		Diagnosis: DiagnosisConfig{
			MaxRootCauses:              3,
			EnableConfidence:           true,
			SuppressDownstreamSymptoms: true,
			RulePacks:                  []string{"builtin"},
		},
	}
}

func Merge(base, override Config) Config {
	result := base
	if override.Version != 0 {
		result.Version = override.Version
	}
	if override.DefaultProfile != "" {
		result.DefaultProfile = override.DefaultProfile
	}
	if len(override.Profiles) > 0 {
		if result.Profiles == nil {
			result.Profiles = map[string]ProfileConfig{}
		}
		for name, profile := range override.Profiles {
			current := result.Profiles[name]
			result.Profiles[name] = mergeProfile(current, profile)
		}
	}
	result.Docker = mergeDocker(result.Docker, override.Docker)
	result.Logs = mergeLogs(result.Logs, override.Logs)
	result.Probe = mergeProbe(result.Probe, override.Probe)
	result.Execution = mergeExecution(result.Execution, override.Execution)
	result.JMX = mergeJMX(result.JMX, override.JMX)
	result.Host = mergeHost(result.Host, override.Host)
	result.Thresholds = mergeThresholds(result.Thresholds, override.Thresholds)
	result.Diagnosis = mergeDiagnosis(result.Diagnosis, override.Diagnosis)
	return result
}

func mergeProfile(base, override ProfileConfig) ProfileConfig {
	result := base
	if len(override.BootstrapExternal) > 0 {
		result.BootstrapExternal = append([]string(nil), override.BootstrapExternal...)
	}
	if len(override.BootstrapInternal) > 0 {
		result.BootstrapInternal = append([]string(nil), override.BootstrapInternal...)
	}
	if len(override.ControllerEndpoints) > 0 {
		result.ControllerEndpoints = append([]string(nil), override.ControllerEndpoints...)
	}
	if override.BrokerCount != 0 {
		result.BrokerCount = override.BrokerCount
	}
	if override.ExpectedMinISR != 0 {
		result.ExpectedMinISR = override.ExpectedMinISR
	}
	if override.ExpectedReplicationFactor != 0 {
		result.ExpectedReplicationFactor = override.ExpectedReplicationFactor
	}
	if override.ExecutionView != "" {
		result.ExecutionView = override.ExecutionView
	}
	if override.SecurityMode != "" {
		result.SecurityMode = override.SecurityMode
	}
	if override.SASLMechanism != "" {
		result.SASLMechanism = override.SASLMechanism
	}
	if len(override.GroupProbeTargets) > 0 {
		result.GroupProbeTargets = append([]GroupProbeTarget(nil), override.GroupProbeTargets...)
	}
	result.Producer = mergeProducerAudit(result.Producer, override.Producer)
	result.Consumer = mergeConsumerAudit(result.Consumer, override.Consumer)
	result.HostNetwork = result.HostNetwork || override.HostNetwork
	result.PlaintextExternal = result.PlaintextExternal || override.PlaintextExternal
	return result
}

func mergeProducerAudit(base, override ProducerAuditConfig) ProducerAuditConfig {
	result := base
	if override.Acks != "" {
		result.Acks = override.Acks
	}
	result.EnableIdempotence = result.EnableIdempotence || override.EnableIdempotence
	if override.Retries != 0 {
		result.Retries = override.Retries
	}
	if override.MaxInFlight != 0 {
		result.MaxInFlight = override.MaxInFlight
	}
	if override.DeliveryTimeoutMs != 0 {
		result.DeliveryTimeoutMs = override.DeliveryTimeoutMs
	}
	if override.RequestTimeoutMs != 0 {
		result.RequestTimeoutMs = override.RequestTimeoutMs
	}
	if override.LingerMs != 0 {
		result.LingerMs = override.LingerMs
	}
	if override.ExpectedDurability != "" {
		result.ExpectedDurability = override.ExpectedDurability
	}
	if override.TransactionalID != "" {
		result.TransactionalID = override.TransactionalID
	}
	if override.TransactionTimeoutMs != 0 {
		result.TransactionTimeoutMs = override.TransactionTimeoutMs
	}
	return result
}

func mergeConsumerAudit(base, override ConsumerAuditConfig) ConsumerAuditConfig {
	result := base
	if override.MaxPollIntervalMs != 0 {
		result.MaxPollIntervalMs = override.MaxPollIntervalMs
	}
	if override.SessionTimeoutMs != 0 {
		result.SessionTimeoutMs = override.SessionTimeoutMs
	}
	if override.HeartbeatIntervalMs != 0 {
		result.HeartbeatIntervalMs = override.HeartbeatIntervalMs
	}
	if override.AutoOffsetReset != "" {
		result.AutoOffsetReset = override.AutoOffsetReset
	}
	if override.IsolationLevel != "" {
		result.IsolationLevel = override.IsolationLevel
	}
	return result
}

func mergeDocker(base, override DockerConfig) DockerConfig {
	result := base
	result.Enabled = result.Enabled || override.Enabled
	if override.ComposeFile != "" {
		result.ComposeFile = override.ComposeFile
	}
	if len(override.ContainerNames) > 0 {
		result.ContainerNames = append([]string(nil), override.ContainerNames...)
	}
	result.InspectMounts = result.InspectMounts || override.InspectMounts
	return result
}

func mergeLogs(base, override LogConfig) LogConfig {
	result := base
	result.Enabled = result.Enabled || override.Enabled
	if override.LogDir != "" {
		result.LogDir = override.LogDir
	}
	if override.TailLines != 0 {
		result.TailLines = override.TailLines
	}
	if override.LookbackMinutes != 0 {
		result.LookbackMinutes = override.LookbackMinutes
	}
	if override.MinLinesPerSource != 0 {
		result.MinLinesPerSource = override.MinLinesPerSource
	}
	if override.FreshnessWindow != "" {
		result.FreshnessWindow = override.FreshnessWindow
	}
	if override.MaxFiles != 0 {
		result.MaxFiles = override.MaxFiles
	}
	if override.MaxBytesPerSource != 0 {
		result.MaxBytesPerSource = override.MaxBytesPerSource
	}
	if override.CustomPatternsDir != "" {
		result.CustomPatternsDir = override.CustomPatternsDir
	}
	return result
}

func mergeProbe(base, override ProbeConfig) ProbeConfig {
	result := base
	result.Enabled = result.Enabled || override.Enabled
	if override.Topic != "" {
		result.Topic = override.Topic
	}
	if override.GroupPrefix != "" {
		result.GroupPrefix = override.GroupPrefix
	}
	if override.Timeout != "" {
		result.Timeout = override.Timeout
	}
	if override.MessageBytes != 0 {
		result.MessageBytes = override.MessageBytes
	}
	if override.ProduceCount != 0 {
		result.ProduceCount = override.ProduceCount
	}
	if override.Acks != "" {
		result.Acks = override.Acks
	}
	if override.CleanupMode != "" {
		result.CleanupMode = override.CleanupMode
	}
	result.EnableIdempotence = result.EnableIdempotence || override.EnableIdempotence
	result.TXProbeEnabled = result.TXProbeEnabled || override.TXProbeEnabled
	result.Cleanup = result.Cleanup || override.Cleanup
	return result
}

func mergeExecution(base, override ExecutionConfig) ExecutionConfig {
	result := base
	if override.Timeout != "" {
		result.Timeout = override.Timeout
	}
	if override.MetadataTimeout != "" {
		result.MetadataTimeout = override.MetadataTimeout
	}
	if override.TCPTimeout != "" {
		result.TCPTimeout = override.TCPTimeout
	}
	if override.AdminAPITimeout != "" {
		result.AdminAPITimeout = override.AdminAPITimeout
	}
	if override.JMXTimeout != "" {
		result.JMXTimeout = override.JMXTimeout
	}
	return result
}

func mergeJMX(base, override JMXConfig) JMXConfig {
	result := base
	result.Enabled = result.Enabled || override.Enabled
	if override.ScrapeTimeout != "" {
		result.ScrapeTimeout = override.ScrapeTimeout
	}
	if override.Path != "" {
		result.Path = override.Path
	}
	if len(override.Endpoints) > 0 {
		result.Endpoints = append([]string(nil), override.Endpoints...)
	}
	if len(override.MetricSets) > 0 {
		result.MetricSets = append([]string(nil), override.MetricSets...)
	}
	return result
}

func mergeHost(base, override HostConfig) HostConfig {
	result := base
	result.Enabled = result.Enabled || override.Enabled
	if len(override.DiskPaths) > 0 {
		result.DiskPaths = append([]string(nil), override.DiskPaths...)
	}
	if len(override.CheckPorts) > 0 {
		result.CheckPorts = append([]int(nil), override.CheckPorts...)
	}
	if override.FDWarnPct != 0 {
		result.FDWarnPct = override.FDWarnPct
	}
	if override.FDCritPct != 0 {
		result.FDCritPct = override.FDCritPct
	}
	if override.ClockSkewWarnMs != 0 {
		result.ClockSkewWarnMs = override.ClockSkewWarnMs
	}
	return result
}

func mergeThresholds(base, override ThresholdConfig) ThresholdConfig {
	result := base
	if override.URPWarn != 0 {
		result.URPWarn = override.URPWarn
	}
	if override.UnderMinISRCrit != 0 {
		result.UnderMinISRCrit = override.UnderMinISRCrit
	}
	if override.NetworkIdleWarn != 0 {
		result.NetworkIdleWarn = override.NetworkIdleWarn
	}
	if override.RequestIdleWarn != 0 {
		result.RequestIdleWarn = override.RequestIdleWarn
	}
	if override.DiskWarnPct != 0 {
		result.DiskWarnPct = override.DiskWarnPct
	}
	if override.DiskCritPct != 0 {
		result.DiskCritPct = override.DiskCritPct
	}
	if override.InodeWarnPct != 0 {
		result.InodeWarnPct = override.InodeWarnPct
	}
	if override.ReplicaLagWarn != 0 {
		result.ReplicaLagWarn = override.ReplicaLagWarn
	}
	if override.LeaderSkewWarnPct != 0 {
		result.LeaderSkewWarnPct = override.LeaderSkewWarnPct
	}
	if override.ConsumerLagWarn != 0 {
		result.ConsumerLagWarn = override.ConsumerLagWarn
	}
	if override.ConsumerLagCrit != 0 {
		result.ConsumerLagCrit = override.ConsumerLagCrit
	}
	if override.CertExpiryWarnDays != 0 {
		result.CertExpiryWarnDays = override.CertExpiryWarnDays
	}
	if override.ProduceThrottleWarnMs != 0 {
		result.ProduceThrottleWarnMs = override.ProduceThrottleWarnMs
	}
	if override.FetchThrottleWarnMs != 0 {
		result.FetchThrottleWarnMs = override.FetchThrottleWarnMs
	}
	if override.RequestLatencyWarnMs != 0 {
		result.RequestLatencyWarnMs = override.RequestLatencyWarnMs
	}
	if override.PurgatoryWarnCount != 0 {
		result.PurgatoryWarnCount = override.PurgatoryWarnCount
	}
	if override.HeapUsedWarnPct != 0 {
		result.HeapUsedWarnPct = override.HeapUsedWarnPct
	}
	if override.GCPauseWarnMs != 0 {
		result.GCPauseWarnMs = override.GCPauseWarnMs
	}
	return result
}

func mergeDiagnosis(base, override DiagnosisConfig) DiagnosisConfig {
	result := base
	if override.MaxRootCauses != 0 {
		result.MaxRootCauses = override.MaxRootCauses
	}
	result.EnableConfidence = result.EnableConfidence || override.EnableConfidence
	result.SuppressDownstreamSymptoms = result.SuppressDownstreamSymptoms || override.SuppressDownstreamSymptoms
	if len(override.RulePacks) > 0 {
		result.RulePacks = append([]string(nil), override.RulePacks...)
	}
	return result
}
