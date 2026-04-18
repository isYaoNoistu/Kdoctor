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
		return rule.NewSkip("HOST-011", "host_memory_pressure", "host", "host memory evidence is not available in the current input mode")
	}
	if c.WarnPct <= 0 {
		c.WarnPct = 85
	}

	memory := bundle.Host.Memory
	evidence := []string{
		fmt.Sprintf("used_pct=%.1f total_bytes=%d available_bytes=%d", memory.UsedPercent, memory.TotalBytes, memory.AvailableBytes),
	}
	result := rule.NewPass("HOST-011", "host_memory_pressure", "host", "host memory headroom looks acceptable for the current Kafka runtime view")
	result.Evidence = evidence
	if memory.UsedPercent >= c.WarnPct {
		result = rule.NewWarn("HOST-011", "host_memory_pressure", "host", "host memory pressure is high and may amplify JVM or container instability")
		result.Evidence = evidence
		result.NextActions = []string{"check whether host memory pressure aligns with JVM heap growth or page cache pressure", "review container limits and Kafka heap sizing together", "watch for swap activity or recent OOM events before load increases further"}
	}
	return result
}
