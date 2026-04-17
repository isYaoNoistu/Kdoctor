package logs

import (
	"context"
	"testing"

	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

func TestFingerprintCheckerUsesHighestSeverity(t *testing.T) {
	result := FingerprintChecker{}.Run(context.Background(), &snapshot.Bundle{
		Logs: &snapshot.LogSnapshot{
			Collected: true,
			Sources:   []string{"docker:kafka1"},
			Matches: []snapshot.LogPatternMatch{
				{ID: "LOG-LEADER", Severity: "fail", Count: 2},
				{ID: "LOG-OOM", Severity: "crit", Count: 1},
			},
		},
	})

	if result.Status != model.StatusCrit {
		t.Fatalf("expected CRIT, got %s", result.Status)
	}
}
