package storage

import (
	"context"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type PartialFailureChecker struct{}

func (PartialFailureChecker) ID() string     { return "STG-004" }
func (PartialFailureChecker) Name() string   { return "partial_logdir_failure" }
func (PartialFailureChecker) Module() string { return "storage" }

func (PartialFailureChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if bundle == nil || bundle.Metrics == nil || !bundle.Metrics.Available {
		return rule.NewSkip("STG-004", "partial_logdir_failure", "storage", "当前没有可用的 JMX 指标，无法评估部分目录故障")
	}

	offline := 0.0
	for _, endpoint := range bundle.Metrics.Endpoints {
		if value, ok := endpoint.Metrics["kafka_log_logmanager_offlinelogdirectorycount"]; ok {
			offline += value
		}
	}
	if offline == 0 {
		return rule.NewPass("STG-004", "partial_logdir_failure", "storage", "未见部分 log directory 故障迹象")
	}

	if bundle.Kafka != nil && len(bundle.Kafka.Brokers) > 0 {
		result := rule.NewWarn("STG-004", "partial_logdir_failure", "storage", "broker 仍在线，但已经出现 log directory 离线，疑似部分目录故障而非整机宕机")
		result.NextActions = []string{"检查 JBOD 或多目录场景下的单目录故障", "确认目录 failure timeout 与 broker 日志中的目录异常", "避免只根据 broker 存活误判为无存储问题"}
		return result
	}
	return rule.NewWarn("STG-004", "partial_logdir_failure", "storage", "检测到离线 log directory，可能存在部分目录故障")
}
