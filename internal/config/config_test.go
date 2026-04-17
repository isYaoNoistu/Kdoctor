package config

import "testing"

func TestMergePreservesDefaultsAndOverrides(t *testing.T) {
	base := Default()
	override := Config{
		DefaultProfile: "custom",
		Docker: DockerConfig{
			ComposeFile: "docker-compose.yml",
		},
		Profiles: map[string]ProfileConfig{
			"custom": {
				BrokerCount:               5,
				ExpectedMinISR:            3,
				ExpectedReplicationFactor: 5,
			},
		},
	}

	merged := Merge(base, override)
	if merged.DefaultProfile != "custom" {
		t.Fatalf("expected default profile to be overridden, got %q", merged.DefaultProfile)
	}
	if merged.Docker.ComposeFile != "docker-compose.yml" {
		t.Fatalf("expected compose file override, got %q", merged.Docker.ComposeFile)
	}
	if merged.Profiles["custom"].BrokerCount != 5 {
		t.Fatalf("expected broker count override, got %d", merged.Profiles["custom"].BrokerCount)
	}
	if merged.Probe.Topic == "" {
		t.Fatal("expected default probe topic to be preserved")
	}
}

func TestDefaultProfileIsGenericBootstrap(t *testing.T) {
	cfg := Default()
	if cfg.DefaultProfile != "generic-bootstrap" {
		t.Fatalf("expected generic-bootstrap default profile, got %q", cfg.DefaultProfile)
	}
}
