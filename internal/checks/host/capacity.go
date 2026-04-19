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
		return rule.NewSkip("HOST-007", "host_disk_and_inode_capacity", "host", "当前输入模式下没有可用的宿主机磁盘和 inode 证据")
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

	result := rule.NewPass("HOST-007", "host_disk_and_inode_capacity", "host", "Kafka 相关宿主机路径的磁盘和 inode 余量正常")
	result.Evidence = evidence
	switch {
	case fail > 0:
		result = rule.NewFail("HOST-007", "host_disk_and_inode_capacity", "host", "部分 Kafka 宿主机路径已经非常接近磁盘耗尽")
		result.Evidence = evidence
		result.NextActions = []string{"立即释放受影响宿主机路径的磁盘空间", "在下一次写入高峰前复核 inode 与保留策略增长情况", "检查宿主机挂载点或文件系统是否已经影响 broker 写入"}
	case warn > 0:
		result = rule.NewWarn("HOST-007", "host_disk_and_inode_capacity", "host", "Kafka 宿主机路径的磁盘或 inode 余量开始变紧")
		result.Evidence = evidence
		result.NextActions = []string{"规划宿主机层面的清理或容量扩展", "复核 Kafka 数据目录和 metadata 目录的 inode 使用情况", "找出增长最快的 broker 路径，避免演变成写入故障"}
	}
	return result
}
