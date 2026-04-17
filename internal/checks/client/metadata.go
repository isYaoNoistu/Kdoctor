package client

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type MetadataChecker struct{}

func (MetadataChecker) ID() string     { return "CLI-001" }
func (MetadataChecker) Name() string   { return "bootstrap_metadata_probe" }
func (MetadataChecker) Module() string { return "client" }

func (MetadataChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Probe == nil {
		return rule.NewError("CLI-001", "bootstrap_metadata_probe", "client", "probe snapshot missing", "probe snapshot missing")
	}
	if snap.Probe.Skipped {
		result := rule.NewSkip("CLI-001", "bootstrap_metadata_probe", "client", snap.Probe.Reason)
		result.Evidence = []string{fmt.Sprintf("topic=%s", snap.Probe.Topic)}
		return result
	}

	result := rule.NewPass("CLI-001", "bootstrap_metadata_probe", "client", "bootstrap metadata probe succeeded")
	result.Evidence = []string{
		fmt.Sprintf("bootstrap=%s", snap.Probe.BootstrapAddress),
		fmt.Sprintf("metadata_duration_ms=%d", snap.Probe.MetadataDurationMs),
	}
	if !snap.Probe.MetadataOK {
		result = rule.NewFail("CLI-001", "bootstrap_metadata_probe", "client", "bootstrap metadata probe failed")
		result.Evidence = []string{
			fmt.Sprintf("bootstrap=%s", snap.Probe.BootstrapAddress),
			fmt.Sprintf("failure_stage=%s", snap.Probe.FailureStage),
			fmt.Sprintf("error=%s", snap.Probe.Error),
		}
		result.NextActions = []string{"verify bootstrap endpoints", "verify Kafka metadata requests are served", "check cluster and listener health"}
	}
	return result
}
