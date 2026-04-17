package client

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type EndToEndChecker struct{}

func (EndToEndChecker) ID() string     { return "CLI-005" }
func (EndToEndChecker) Name() string   { return "end_to_end_probe" }
func (EndToEndChecker) Module() string { return "client" }

func (EndToEndChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Probe == nil {
		return rule.NewError("CLI-005", "end_to_end_probe", "client", "probe snapshot missing", "probe snapshot missing")
	}
	if snap.Probe.Skipped {
		return rule.NewSkip("CLI-005", "end_to_end_probe", "client", snap.Probe.Reason)
	}

	allOK := snap.Probe.MetadataOK && snap.Probe.ProduceOK && snap.Probe.ConsumeOK && snap.Probe.CommitOK
	result := rule.NewPass("CLI-005", "end_to_end_probe", "client", "end-to-end probe succeeded")
	result.Evidence = []string{
		fmt.Sprintf("topic=%s", snap.Probe.Topic),
		fmt.Sprintf("group_id=%s", snap.Probe.GroupID),
		fmt.Sprintf("end_to_end_duration_ms=%d", snap.Probe.EndToEndDurationMs),
	}
	if !allOK {
		result = rule.NewFail("CLI-005", "end_to_end_probe", "client", "end-to-end probe failed")
		result.Evidence = []string{
			fmt.Sprintf("topic=%s", snap.Probe.Topic),
			fmt.Sprintf("failure_stage=%s", snap.Probe.FailureStage),
			fmt.Sprintf("error=%s", snap.Probe.Error),
		}
		result.NextActions = []string{"check the failing stage first", "verify probe topic and broker reachability", "correlate with network, ISR and controller checks"}
	}
	return result
}
