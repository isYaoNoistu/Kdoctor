package host

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type MemoryChecker struct {
	WarnPct float64
}

func (MemoryChecker) ID() string     { return "HOST-011" }
func (MemoryChecker) Name() string   { return "host_memory_pressure" }
func (MemoryChecker) Module() string { return "host" }

func (c MemoryChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if bundle == nil || bundle.Host == nil || !bundle.Host.Collected || bundle.Host.Memory == nil {
		return rule.NewSkip("HOST-011", "host_memory_pressure", "host", "当前输入模式下没有可用的宿主机内存证据")
	}
	if c.WarnPct <= 0 {
		c.WarnPct = 85
	}

	memory := bundle.Host.Memory
	evidence := []string{
		fmt.Sprintf("used_pct=%.1f total_bytes=%d available_bytes=%d", memory.UsedPercent, memory.TotalBytes, memory.AvailableBytes),
	}
	result := rule.NewPass("HOST-011", "host_memory_pressure", "host", "当前 Kafka 运行视角下的宿主机内存余量正常")
	result.Evidence = evidence
	if memory.UsedPercent >= c.WarnPct {
		result = rule.NewWarn("HOST-011", "host_memory_pressure", "host", "宿主机内存压力偏高，可能放大 JVM 或容器不稳定问题")
		result.Evidence = evidence
		result.NextActions = []string{"检查宿主机内存压力是否与 JVM 堆增长或页缓存压力一致", "一起复核容器限制与 Kafka 堆大小", "在负载继续上升前关注 swap 活动或近期 OOM 事件"}
	}
	return result
}
