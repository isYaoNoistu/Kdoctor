package kafka

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type MetadataLatencyChecker struct {
	WarnMs int64
	CritMs int64
}

func (MetadataLatencyChecker) ID() string     { return "KFK-008" }
func (MetadataLatencyChecker) Name() string   { return "metadata_latency" }
func (MetadataLatencyChecker) Module() string { return "kafka" }

func (c MetadataLatencyChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Kafka == nil {
		return rule.NewSkip("KFK-008", "metadata_latency", "kafka", "当前没有可用的 Kafka metadata 快照，无法评估 metadata 延迟")
	}
	if c.WarnMs <= 0 {
		c.WarnMs = 500
	}
	if c.CritMs <= 0 {
		c.CritMs = 2000
	}

	evidence := []string{fmt.Sprintf("metadata_duration_ms=%d", snap.Kafka.MetadataDurationMs)}
	result := rule.NewPass("KFK-008", "metadata_latency", "kafka", "metadata 获取延迟处于安全范围")
	result.Evidence = evidence
	if snap.Kafka.MetadataDurationMs >= c.CritMs {
		result = rule.NewFail("KFK-008", "metadata_latency", "kafka", "metadata 获取延迟已经明显偏高，控制面或 broker 线程池可能存在压力")
		result.Evidence = evidence
		result.NextActions = []string{"结合 JVM 指标查看网络线程与请求线程压力", "排查 controller、broker 负载和日志中的超时迹象", "观察 metadata 请求是否反复超时或抖动"}
		return result
	}
	if snap.Kafka.MetadataDurationMs >= c.WarnMs {
		result = rule.NewWarn("KFK-008", "metadata_latency", "kafka", "metadata 获取延迟偏高，建议关注控制面和 broker 处理压力")
		result.Evidence = evidence
	}
	return result
}
