package storage

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type OfflineLogDirChecker struct{}

func (OfflineLogDirChecker) ID() string     { return "STG-002" }
func (OfflineLogDirChecker) Name() string   { return "offline_log_directory" }
func (OfflineLogDirChecker) Module() string { return "storage" }

func (OfflineLogDirChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if bundle == nil || bundle.Metrics == nil || !bundle.Metrics.Available {
		return rule.NewSkip("STG-002", "offline_log_directory", "storage", "当前没有可用的 JMX 指标，无法评估离线 log directory")
	}

	total := 0.0
	evidence := []string{}
	for _, endpoint := range bundle.Metrics.Endpoints {
		if value, ok := endpoint.Metrics["kafka_log_logmanager_offlinelogdirectorycount"]; ok {
			total += value
			evidence = append(evidence, fmt.Sprintf("endpoint=%s offline_log_dir=%.0f", endpoint.Address, value))
		}
	}
	if len(evidence) == 0 {
		return rule.NewSkip("STG-002", "offline_log_directory", "storage", "当前 JMX 指标中没有 OfflineLogDirectoryCount")
	}

	result := rule.NewPass("STG-002", "offline_log_directory", "storage", "未发现离线 log directory")
	result.Evidence = evidence
	if total > 0 {
		result = rule.NewFail("STG-002", "offline_log_directory", "storage", "检测到离线 log directory，属于高危存储故障")
		result.Evidence = evidence
		result.NextActions = []string{"优先检查目录权限、挂载和磁盘状态", "查看 broker 日志中的 log directory/offline 报错", "结合 STG-003/STG-005 判断是否为目录规划或挂载问题"}
	}
	return result
}
