package network

import (
	"testing"

	"kdoctor/internal/config"
)

func TestBootstrapTargetsFallsBackToInternal(t *testing.T) {
	env := &config.Runtime{
		BootstrapInternal: []string{"10.0.0.10:9092"},
	}
	targets := bootstrapTargets(env)
	if len(targets) != 1 || targets[0] != "10.0.0.10:9092" {
		t.Fatalf("expected internal bootstrap fallback, got %#v", targets)
	}
}
