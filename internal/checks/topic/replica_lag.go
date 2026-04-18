package topic

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type ReplicaLagChecker struct {
	Warn int64
}

func (ReplicaLagChecker) ID() string     { return "TOP-010" }
func (ReplicaLagChecker) Name() string   { return "replica_lag" }
func (ReplicaLagChecker) Module() string { return "topic" }

func (c ReplicaLagChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if bundle == nil || bundle.Metrics == nil || !bundle.Metrics.Available {
		return rule.NewSkip("TOP-010", "replica_lag", "topic", "当前没有可用的 JMX 指标，无法评估副本同步 lag")
	}
	if c.Warn <= 0 {
		c.Warn = 10000
	}

	maxLag := float64(0)
	found := false
	evidence := []string{}
	for _, endpoint := range bundle.Metrics.Endpoints {
		for _, name := range []string{
			"kafka_server_replicafetchermanager_maxlag",
			"kafka_server_replicafetchermanager_consumerlag",
		} {
			if value, ok := endpoint.Metrics[name]; ok {
				if !found || value > maxLag {
					maxLag = value
				}
				found = true
				evidence = append(evidence, fmt.Sprintf("endpoint=%s metric=%s value=%.0f", endpoint.Address, name, value))
			}
		}
	}
	if !found {
		return rule.NewSkip("TOP-010", "replica_lag", "topic", "当前 JMX 指标中没有副本 lag 指标")
	}

	result := rule.NewPass("TOP-010", "replica_lag", "topic", "未发现明显的副本同步 lag")
	result.Evidence = evidence
	if int64(maxLag) >= c.Warn {
		result = rule.NewWarn("TOP-010", "replica_lag", "topic", "副本同步 lag 已偏高，即使 ISR 仍完整也建议提前关注")
		result.Evidence = evidence
		result.NextActions = []string{"检查 follower broker 的磁盘和网络压力", "结合 ISR、UnderReplicated 与请求高峰一起判断", "在业务流量继续放大前优先恢复复制余量"}
	}
	return result
}
