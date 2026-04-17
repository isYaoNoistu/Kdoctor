package host

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type DiskChecker struct{}

func (DiskChecker) ID() string     { return "HOST-004" }
func (DiskChecker) Name() string   { return "disk_space" }
func (DiskChecker) Module() string { return "host" }

func (DiskChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Host == nil || !snap.Host.Collected || len(snap.Host.DiskUsages) == 0 {
		return rule.NewSkip("HOST-004", "disk_space", "host", "host disk usage is not available in the current input mode")
	}

	evidence := []string{}
	result := rule.NewPass("HOST-004", "disk_space", "host", "host disk usage is within safe range")
	warnCount := 0
	failCount := 0
	for _, usage := range snap.Host.DiskUsages {
		evidence = append(evidence, fmt.Sprintf("%s used=%.1f%% free=%d bytes", usage.Path, usage.UsedPercent, usage.AvailableBytes))
		switch {
		case usage.UsedPercent >= 95:
			failCount++
		case usage.UsedPercent >= 85:
			warnCount++
		}
	}

	result.Evidence = evidence
	if failCount > 0 {
		result = rule.NewFail("HOST-004", "disk_space", "host", "some Kafka disk paths are critically full")
		result.Evidence = evidence
		result.NextActions = []string{"free disk space immediately", "review retention and cleanup settings", "check whether a broker or metadata directory stopped writing"}
		return result
	}
	if warnCount > 0 {
		result = rule.NewWarn("HOST-004", "disk_space", "host", "some Kafka disk paths are nearing capacity")
		result.Evidence = evidence
		result.NextActions = []string{"plan disk cleanup or capacity expansion", "review retention and segment sizing", "monitor growth before the next traffic spike"}
	}
	return result
}
