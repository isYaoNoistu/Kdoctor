package client

import (
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

func skipIfProbeStageNotExecuted(id string, name string, probe *snapshot.ProbeSnapshot, stage string) (model.CheckResult, bool) {
	if probe == nil {
		return rule.NewError(id, name, "client", "probe snapshot missing", "probe snapshot missing"), true
	}
	if probe.Skipped {
		return rule.NewSkip(id, name, "client", probe.Reason), true
	}
	if stageExecuted(probe, stage) {
		return model.CheckResult{}, false
	}

	reason := probe.StageSkipReason(stage)
	if reason == "" {
		reason = fmt.Sprintf("%s stage was not executed", stage)
	}
	result := rule.NewSkip(id, name, "client", reason)
	result.Evidence = probeEvidence(probe)
	return result, true
}

func pickProbe(snap *snapshot.Bundle) *snapshot.ProbeSnapshot {
	if snap == nil {
		return nil
	}
	return snap.Probe
}

func stageExecuted(probe *snapshot.ProbeSnapshot, stage string) bool {
	switch stage {
	case snapshot.ProbeStageMetadata:
		return probe.MetadataExecuted
	case snapshot.ProbeStageProduce:
		return probe.ProduceExecuted
	case snapshot.ProbeStageConsume:
		return probe.ConsumeExecuted
	case snapshot.ProbeStageCommit:
		return probe.CommitExecuted
	default:
		return false
	}
}

func probeEvidence(probe *snapshot.ProbeSnapshot) []string {
	if probe == nil {
		return nil
	}

	evidence := []string{}
	if probe.Topic != "" {
		evidence = append(evidence, fmt.Sprintf("topic=%s", probe.Topic))
	}
	if probe.GroupID != "" {
		evidence = append(evidence, fmt.Sprintf("group_id=%s", probe.GroupID))
	}
	if probe.ExecutedStage != "" {
		evidence = append(evidence, fmt.Sprintf("executed_stage=%s", probe.ExecutedStage))
	}
	if probe.FailureStage != "" {
		evidence = append(evidence, fmt.Sprintf("failure_stage=%s", probe.FailureStage))
	}
	if probe.TopicReadyReason != "" {
		evidence = append(evidence, fmt.Sprintf("topic_ready_reason=%s", probe.TopicReadyReason))
	}
	if probe.Error != "" {
		evidence = append(evidence, fmt.Sprintf("error=%s", probe.Error))
	}
	return evidence
}

func mergeEvidence(chunks ...[]string) []string {
	seen := map[string]struct{}{}
	merged := []string{}
	for _, chunk := range chunks {
		for _, item := range chunk {
			if item == "" {
				continue
			}
			if _, ok := seen[item]; ok {
				continue
			}
			seen[item] = struct{}{}
			merged = append(merged, item)
		}
	}
	return merged
}
