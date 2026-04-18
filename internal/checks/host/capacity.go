package host

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type CapacityChecker struct {
	DiskWarnPct  float64
	DiskCritPct  float64
	InodeWarnPct float64
}

func (CapacityChecker) ID() string     { return "HOST-007" }
func (CapacityChecker) Name() string   { return "host_disk_and_inode_capacity" }
func (CapacityChecker) Module() string { return "host" }

func (c CapacityChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if bundle == nil || bundle.Host == nil || !bundle.Host.Collected || len(bundle.Host.DiskUsages) == 0 {
		return rule.NewSkip("HOST-007", "host_disk_and_inode_capacity", "host", "host disk and inode evidence is not available in the current input mode")
	}
	if c.DiskWarnPct <= 0 {
		c.DiskWarnPct = 75
	}
	if c.DiskCritPct <= 0 {
		c.DiskCritPct = 85
	}
	if c.InodeWarnPct <= 0 {
		c.InodeWarnPct = 80
	}

	warn := 0
	fail := 0
	evidence := []string{}
	for _, usage := range bundle.Host.DiskUsages {
		evidence = append(evidence, fmt.Sprintf("path=%s used_pct=%.1f inode_used_pct=%.1f available_bytes=%d available_inodes=%d", usage.Path, usage.UsedPercent, usage.UsedInodePct, usage.AvailableBytes, usage.AvailableInodes))
		switch {
		case usage.UsedPercent >= c.DiskCritPct:
			fail++
		case usage.UsedPercent >= c.DiskWarnPct:
			warn++
		}
		if usage.TotalInodes > 0 && usage.UsedInodePct >= c.InodeWarnPct {
			warn++
		}
	}

	result := rule.NewPass("HOST-007", "host_disk_and_inode_capacity", "host", "host disk and inode headroom for Kafka paths looks acceptable")
	result.Evidence = evidence
	switch {
	case fail > 0:
		result = rule.NewFail("HOST-007", "host_disk_and_inode_capacity", "host", "some Kafka host paths are critically close to disk exhaustion")
		result.Evidence = evidence
		result.NextActions = []string{"free disk space immediately on the affected host path", "review inode and retention growth before the next write peak", "check whether the host-level mount or filesystem is already impairing broker writes"}
	case warn > 0:
		result = rule.NewWarn("HOST-007", "host_disk_and_inode_capacity", "host", "host disk or inode headroom for Kafka paths is getting tight")
		result.Evidence = evidence
		result.NextActions = []string{"plan host-level cleanup or capacity expansion", "review inode usage on Kafka data and metadata directories", "track which broker path is growing fastest before it becomes a write outage"}
	}
	return result
}
