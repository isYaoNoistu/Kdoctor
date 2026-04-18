package metrics

import (
	"context"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type OfflineLogDirChecker struct{}

func (OfflineLogDirChecker) ID() string     { return "MET-004" }
func (OfflineLogDirChecker) Name() string   { return "offline_log_directory_count" }
func (OfflineLogDirChecker) Module() string { return "metrics" }

func (OfflineLogDirChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if result, skip := skipIfUnavailable("MET-004", "offline_log_directory_count", "metrics", bundle); skip {
		return result
	}

	value, ok, evidence := aggregateMax(metricsSnap(bundle), "kafka_log_logmanager_offlinelogdirectorycount")
	if !ok {
		return rule.NewSkip("MET-004", "offline_log_directory_count", "metrics", "当前 JMX 指标中没有 OfflineLogDirectoryCount")
	}

	result := rule.NewPass("MET-004", "offline_log_directory_count", "metrics", "JMX 未发现离线 log directory")
	result.Evidence = evidence
	if value >= 1 {
		result = rule.NewFail("MET-004", "offline_log_directory_count", "metrics", "JMX 检测到 OfflineLogDirectoryCount 大于 0")
		result.Evidence = evidence
		result.NextActions = []string{"检查 broker 日志中的 log directory/offline 关键字", "确认目录权限、挂载和磁盘状态", "结合 STG-003/STG-005 判断是否为目录规划或挂载问题"}
	}
	return result
}
