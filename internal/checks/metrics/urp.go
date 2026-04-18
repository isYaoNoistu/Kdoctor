package metrics

import (
	"context"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type UnderReplicatedChecker struct {
	WarnCount int
}

func (UnderReplicatedChecker) ID() string     { return "MET-001" }
func (UnderReplicatedChecker) Name() string   { return "under_replicated_partitions" }
func (UnderReplicatedChecker) Module() string { return "metrics" }

func (c UnderReplicatedChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if result, skip := skipIfUnavailable("MET-001", "under_replicated_partitions", "metrics", bundle); skip {
		return result
	}

	if c.WarnCount <= 0 {
		c.WarnCount = 1
	}
	value, ok, evidence := aggregateMax(metricsSnap(bundle), "kafka_server_replicamanager_underreplicatedpartitions")
	if !ok {
		return rule.NewSkip("MET-001", "under_replicated_partitions", "metrics", "当前 JMX 指标中没有 UnderReplicatedPartitions")
	}

	result := rule.NewPass("MET-001", "under_replicated_partitions", "metrics", "JMX 未发现 UnderReplicatedPartitions")
	result.Evidence = evidence
	if int(value) >= c.WarnCount {
		result = rule.NewWarn("MET-001", "under_replicated_partitions", "metrics", "JMX 检测到 UnderReplicatedPartitions 大于 0")
		result.Evidence = evidence
		result.NextActions = []string{"结合 TOP-004/TOP-005 判断受影响分区", "检查 follower broker、磁盘与复制链路", "观察是否伴随 controller 或网络异常"}
	}
	return result
}
