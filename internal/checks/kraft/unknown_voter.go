package kraft

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type UnknownVoterChecker struct{}

func (UnknownVoterChecker) ID() string     { return "KRF-007" }
func (UnknownVoterChecker) Name() string   { return "unknown_voter_connections" }
func (UnknownVoterChecker) Module() string { return "kraft" }

func (UnknownVoterChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if result, skip := metricsSkip("KRF-007", "unknown_voter_connections", "kraft", bundle); skip {
		return result
	}

	value, ok, evidence := metricsAggregateMax(bundle,
		"kafka_server_raftmanager_numberofunknownvoterconnections",
		"kafka_server_raftmanager_unknownvoterconnections",
	)
	if !ok {
		return rule.NewSkip("KRF-007", "unknown_voter_connections", "kraft", "当前 JMX 指标中没有 UnknownVoterConnections")
	}

	result := rule.NewPass("KRF-007", "unknown_voter_connections", "kraft", "JMX 未发现 unknown voter connection")
	result.Evidence = evidence
	if value > 0 {
		result = rule.NewWarn("KRF-007", "unknown_voter_connections", "kraft", "JMX 检测到 unknown voter connection，controller quorum 信任链可能存在异常")
		result.Evidence = evidence
		result.NextActions = []string{"检查 quorum voter 配置是否一致", "查看 controller 日志中的连接或复制异常", "确认 controller 网络与端口没有被错误漂移"}
	}
	return result
}

func metricsSkip(id string, name string, module string, bundle *snapshot.Bundle) (model.CheckResult, bool) {
	if bundle == nil || bundle.Metrics == nil || !bundle.Metrics.Collected {
		return rule.NewSkip(id, name, module, "当前输入模式未启用 JMX 指标采集"), true
	}
	if !bundle.Metrics.Available {
		result := rule.NewSkip(id, name, module, "当前没有可用的 JMX 指标来源")
		result.Evidence = append(result.Evidence, bundle.Metrics.Errors...)
		return result, true
	}
	return model.CheckResult{}, false
}

func metricsAggregateMax(bundle *snapshot.Bundle, names ...string) (float64, bool, []string) {
	if bundle == nil || bundle.Metrics == nil {
		return 0, false, nil
	}
	best := 0.0
	found := false
	evidence := []string{}
	for _, endpoint := range bundle.Metrics.Endpoints {
		for _, name := range names {
			if value, ok := endpoint.Metrics[name]; ok {
				if !found || value > best {
					best = value
				}
				found = true
				evidence = append(evidence, fmt.Sprintf("endpoint=%s metric=%s value=%.0f", endpoint.Address, name, value))
			}
		}
	}
	return best, found, evidence
}
