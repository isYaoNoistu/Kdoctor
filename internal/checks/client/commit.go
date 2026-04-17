package client

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type CommitChecker struct{}

func (CommitChecker) ID() string     { return "CLI-004" }
func (CommitChecker) Name() string   { return "consumer_group_commit_probe" }
func (CommitChecker) Module() string { return "client" }

func (CommitChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if result, skipped := skipIfProbeStageNotExecuted("CLI-004", "consumer_group_commit_probe", pickProbe(snap), snapshot.ProbeStageCommit); skipped {
		return result
	}

	result := rule.NewPass("CLI-004", "consumer_group_commit_probe", "client", "consumer group commit probe succeeded")
	result.Evidence = []string{
		fmt.Sprintf("group_id=%s", snap.Probe.GroupID),
		fmt.Sprintf("executed_stage=%s", snap.Probe.ExecutedStage),
		fmt.Sprintf("commit_duration_ms=%d", snap.Probe.CommitDurationMs),
	}
	if !snap.Probe.CommitOK {
		result = rule.NewFail("CLI-004", "consumer_group_commit_probe", "client", "consumer group commit probe failed")
		result.Evidence = mergeEvidence([]string{fmt.Sprintf("group_id=%s", snap.Probe.GroupID)}, probeEvidence(snap.Probe))
		result.NextActions = []string{"verify __consumer_offsets health", "verify coordinator health", "check controller and internal topic replicas"}
	}
	return result
}
