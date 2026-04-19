package config

import (
	"os"
	"path/filepath"
	"testing"
)

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

func TestLoadFilePreservesExplicitFalseForEnabledFlags(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "kdoctor.yaml")
	content := []byte(`
version: 2
docker:
  enabled: false
logs:
  enabled: false
probe:
  enabled: false
host:
  enabled: false
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	cfg, err := LoadFile(path, true)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	merged := Merge(Default(), cfg)
	if merged.Docker.Enabled {
		t.Fatal("expected docker.enabled=false to override default")
	}
	if merged.Logs.Enabled {
		t.Fatal("expected logs.enabled=false to override default")
	}
	if merged.Probe.Enabled {
		t.Fatal("expected probe.enabled=false to override default")
	}
	if merged.Host.Enabled {
		t.Fatal("expected host.enabled=false to override default")
	}
}

func TestDefaultProfileIsGenericBootstrap(t *testing.T) {
	cfg := Default()
	if cfg.DefaultProfile != "generic-bootstrap" {
		t.Fatalf("expected generic-bootstrap default profile, got %q", cfg.DefaultProfile)
	}
}

func TestLoadFileReturnsErrorWhenExplicitPathIsMissing(t *testing.T) {
	_, err := LoadFile("definitely-missing-kdoctor.yaml", true)
	if err == nil {
		t.Fatal("expected explicit missing config path to return an error")
	}
}

func TestLoadFileIgnoresMissingDefaultPath(t *testing.T) {
	cfg, err := LoadFile("definitely-missing-kdoctor.yaml", false)
	if err != nil {
		t.Fatalf("expected missing default config path to be ignored, got %v", err)
	}
	if cfg.Version != 0 {
		t.Fatalf("expected zero-value config, got version=%d", cfg.Version)
	}
}

func TestLoadFileReadsYamlContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "kdoctor.yaml")
	content := []byte("version: 1\ndefault_profile: test-profile\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	cfg, err := LoadFile(path, true)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.DefaultProfile != "test-profile" {
		t.Fatalf("expected default_profile to be loaded, got %q", cfg.DefaultProfile)
	}
}

func TestValidateRejectsInvalidFreshnessWindow(t *testing.T) {
	cfg := Default()
	cfg.Logs.FreshnessWindow = "not-a-duration"

	if err := Validate(cfg); err == nil {
		t.Fatal("expected invalid logs.freshness_window to fail validation")
	}
}

func TestValidateRejectsNegativeDiagnosisLimit(t *testing.T) {
	cfg := Default()
	cfg.Diagnosis.MaxRootCauses = -1

	if err := Validate(cfg); err == nil {
		t.Fatal("expected negative diagnosis.max_root_causes to fail validation")
	}
}
