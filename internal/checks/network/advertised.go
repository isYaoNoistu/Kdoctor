package network

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type AdvertisedPrivateChecker struct{}

func (AdvertisedPrivateChecker) ID() string     { return "NET-006" }
func (AdvertisedPrivateChecker) Name() string   { return "advertised_private_endpoints" }
func (AdvertisedPrivateChecker) Module() string { return "network" }

func (AdvertisedPrivateChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Kafka == nil {
		return rule.NewSkip("NET-006", "advertised_private_endpoints", "network", "当前没有可用的 Kafka metadata，无法评估返回地址是否暴露错误")
	}
	if !isExternalProbeView(snap) {
		return rule.NewSkip("NET-006", "advertised_private_endpoints", "network", "当前不是外部执行视角，暂不评估私网地址暴露问题")
	}

	privateEndpoints := make([]string, 0)
	for _, broker := range snap.Kafka.Brokers {
		if isPrivateEndpoint(broker.Address) {
			privateEndpoints = append(privateEndpoints, broker.Address)
		}
	}

	result := rule.NewPass("NET-006", "advertised_private_endpoints", "network", "当前外部执行视角下，metadata 返回的 broker 地址可路由")
	if len(privateEndpoints) > 0 {
		result = rule.NewFail("NET-006", "advertised_private_endpoints", "network", "外部执行视角下，metadata 返回了私网 broker 地址")
		result.Evidence = []string{fmt.Sprintf("private_endpoints=%v", privateEndpoints)}
		result.NextActions = []string{"优先检查 advertised.listeners 是否为外部可路由地址", "确认公网客户端不会收到 RFC1918 地址", "如需内外网双视角，请为不同 listener 明确区分路由地址"}
	}
	return result
}
