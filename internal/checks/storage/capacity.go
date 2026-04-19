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
		return rule.NewSkip("STG-001", "disk_and_inode_capacity", "storage", "当前输入模式下没有可用的磁盘和 inode 证据")
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

	result := rule.NewPass("STG-001", "disk_and_inode_capacity", "storage", "Kafka 磁盘和 inode 余量正常")
	result.Evidence = evidence
	switch {
	case fail > 0:
		result = rule.NewFail("STG-001", "disk_and_inode_capacity", "storage", "至少有一个 Kafka 存储路径已经非常接近磁盘耗尽")
		result.Evidence = evidence
		result.NextActions = []string{"立即释放磁盘空间或扩容受影响卷", "在下一次流量高峰前复核保留策略和 segment 大小", "检查是否已有 broker 路径因容量压力停止写入"}
	case warn > 0:
		result = rule.NewWarn("STG-001", "disk_and_inode_capacity", "storage", "Kafka 存储在磁盘空间或 inode 上的余量开始变紧")
		result.Evidence = evidence
		result.NextActions = []string{"在写入压力增加前规划清理或扩容", "复核 Kafka 数据路径和 metadata 路径的 inode 消耗", "将容量增长与近期 topic 或 partition 变化关联起来"}
	}
	return result
}
