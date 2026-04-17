package config

func Default() Config {
	return Config{
		Version:        1,
		DefaultProfile: "generic-bootstrap",
		Profiles:       map[string]ProfileConfig{},
		Docker: DockerConfig{
			Enabled: true,
		},
		Logs: LogConfig{
			Enabled:         true,
			TailLines:       300,
			LookbackMinutes: 15,
		},
		Probe: ProbeConfig{
			Enabled:      true,
			Topic:        "_kdoctor_probe",
			GroupPrefix:  "kdoctor-probe",
			Timeout:      "15s",
			MessageBytes: 1024,
			ProduceCount: 1,
		},
		Execution: ExecutionConfig{
			Timeout:         "30s",
			MetadataTimeout: "5s",
			TCPTimeout:      "3s",
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
	result.HostNetwork = result.HostNetwork || override.HostNetwork
	result.PlaintextExternal = result.PlaintextExternal || override.PlaintextExternal
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
	return result
}
