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
			SourceStats: []snapshot.LogSourceStat{
				{Source: "docker:kafka1", Kind: "docker", Lines: 120, Bytes: 4096, Fresh: true, SufficientLines: true},
			},
			Matches: []snapshot.LogPatternMatch{
				{ID: "LOG-LEADER", Library: "builtin", Severity: "fail", Count: 2},
				{ID: "LOG-OOM", Library: "builtin", Severity: "crit", Count: 1},
			},
		},
	})

	if result.Status != model.StatusCrit {
		t.Fatalf("expected CRIT, got %s", result.Status)
	}
}

func TestFingerprintCheckerUsesCautiousPassWordingWhenNoMatches(t *testing.T) {
	result := FingerprintChecker{}.Run(context.Background(), &snapshot.Bundle{
		Logs: &snapshot.LogSnapshot{
			Collected:           true,
			Sources:             []string{"file:/tmp/server.log"},
			BuiltinPatternCount: 15,
			SourceStats: []snapshot.LogSourceStat{
				{Source: "file:/tmp/server.log", Kind: "file", Lines: 64, Bytes: 2048, Fresh: true, SufficientLines: true},
			},
		},
	})

	if result.Status != model.StatusPass {
		t.Fatalf("expected PASS, got %s", result.Status)
	}
	if result.Summary != "近期日志未命中已知错误指纹" {
		t.Fatalf("unexpected summary: %q", result.Summary)
	}
}
