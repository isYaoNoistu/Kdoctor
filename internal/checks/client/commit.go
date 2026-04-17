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
	if snap == nil || snap.Probe == nil {
		return rule.NewError("CLI-004", "consumer_group_commit_probe", "client", "probe snapshot missing", "probe snapshot missing")
	}
	if snap.Probe.Skipped {
		return rule.NewSkip("CLI-004", "consumer_group_commit_probe", "client", snap.Probe.Reason)
	}

	result := rule.NewPass("CLI-004", "consumer_group_commit_probe", "client", "consumer group commit probe succeeded")
	result.Evidence = []string{
		fmt.Sprintf("group_id=%s", snap.Probe.GroupID),
		fmt.Sprintf("commit_duration_ms=%d", snap.Probe.CommitDurationMs),
	}
	if !snap.Probe.CommitOK {
		result = rule.NewFail("CLI-004", "consumer_group_commit_probe", "client", "consumer group commit probe failed")
		result.Evidence = []string{
			fmt.Sprintf("group_id=%s", snap.Probe.GroupID),
			fmt.Sprintf("failure_stage=%s", snap.Probe.FailureStage),
			fmt.Sprintf("error=%s", snap.Probe.Error),
		}
		result.NextActions = []string{"verify __consumer_offsets health", "verify coordinator health", "check controller and internal topic replicas"}
	}
	return result
}
