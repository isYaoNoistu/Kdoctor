package kafka

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type ReturnedRouteChecker struct{}

func (ReturnedRouteChecker) ID() string     { return "KFK-005" }
func (ReturnedRouteChecker) Name() string   { return "returned_broker_route" }
func (ReturnedRouteChecker) Module() string { return "kafka" }

func (ReturnedRouteChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Network == nil || snap.Kafka == nil {
		return rule.NewSkip("KFK-005", "returned_broker_route", "kafka", "当前缺少完整的网络或 metadata 快照，无法评估返回 broker 路径")
	}
	if len(snap.Network.MetadataChecks) == 0 {
		return rule.NewSkip("KFK-005", "returned_broker_route", "kafka", "尚未执行 metadata 返回地址探测")
	}

	unreachable := 0
	evidence := make([]string, 0, len(snap.Network.MetadataChecks))
	for _, check := range snap.Network.MetadataChecks {
		if check.Reachable {
			evidence = append(evidence, fmt.Sprintf("%s reachable", check.Address))
			continue
		}
		unreachable++
		evidence = append(evidence, fmt.Sprintf("%s unreachable: %s", check.Address, check.Error))
	}

	result := rule.NewPass("KFK-005", "returned_broker_route", "kafka", "metadata 返回的 broker 路径可达")
	result.Evidence = evidence
	if unreachable > 0 {
		result = rule.NewWarn("KFK-005", "returned_broker_route", "kafka", "metadata 返回的 broker 路径存在不可达节点，客户端可能在后续阶段失败")
		if unreachable == len(snap.Network.MetadataChecks) {
			result = rule.NewFail("KFK-005", "returned_broker_route", "kafka", "metadata 返回的所有 broker 路径都不可达，客户端很可能在 metadata 之后立刻失败")
		}
		result.Evidence = evidence
		result.NextActions = []string{"核对 broker 返回地址与当前客户端网络是否匹配", "检查 advertised.listeners、端口暴露与 NAT", "结合 NET-005/NET-006 一起判断入口与返回路径是否分裂"}
	}
	return result
}
