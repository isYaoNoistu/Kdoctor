package storage

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

func (CapacityChecker) ID() string     { return "STG-001" }
func (CapacityChecker) Name() string   { return "disk_and_inode_capacity" }
func (CapacityChecker) Module() string { return "storage" }

func (c CapacityChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if bundle == nil || bundle.Host == nil || !bundle.Host.Collected || len(bundle.Host.DiskUsages) == 0 {
		return rule.NewSkip("STG-001", "disk_and_inode_capacity", "storage", "disk and inode evidence is not available in the current input mode")
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

	result := rule.NewPass("STG-001", "disk_and_inode_capacity", "storage", "Kafka disk and inode headroom look acceptable")
	result.Evidence = evidence
	switch {
	case fail > 0:
		result = rule.NewFail("STG-001", "disk_and_inode_capacity", "storage", "at least one Kafka storage path is critically close to disk exhaustion")
		result.Evidence = evidence
		result.NextActions = []string{"free disk space immediately or expand the affected volume", "review retention and segment sizing before the next traffic peak", "check whether any broker path stopped writing because of capacity pressure"}
	case warn > 0:
		result = rule.NewWarn("STG-001", "disk_and_inode_capacity", "storage", "Kafka storage headroom is getting tight on disk space or inodes")
		result.Evidence = evidence
		result.NextActions = []string{"plan cleanup or capacity expansion before write pressure increases", "review inode consumption on Kafka data and metadata paths", "correlate capacity growth with recent topic or partition changes"}
	}
	return result
}
