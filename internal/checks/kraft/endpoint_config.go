package kraft

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"kdoctor/internal/composeutil"
	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type EndpointConfigChecker struct{}

func (EndpointConfigChecker) ID() string     { return "KRF-005" }
func (EndpointConfigChecker) Name() string   { return "controller_endpoint_config" }
func (EndpointConfigChecker) Module() string { return "kraft" }

func (EndpointConfigChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Kafka == nil || snap.Compose == nil {
		return rule.NewSkip("KRF-005", "controller_endpoint_config", "kraft", "当前缺少 compose 或 Kafka metadata，无法评估 controller 端点配置异常")
	}
	if snap.Kafka.ControllerID == nil {
		return rule.NewSkip("KRF-005", "controller_endpoint_config", "kraft", "当前没有活动 controller 信息，暂不评估 controller 端点配置异常")
	}

	services := composeutil.KafkaServices(snap.Compose)
	if len(services) == 0 {
		return rule.NewSkip("KRF-005", "controller_endpoint_config", "kraft", "compose 中没有识别到 Kafka 服务")
	}

	voterSet := map[string]struct{}{}
	controllerListenerNames := map[string]struct{}{}
	evidence := []string{}
	for _, service := range services {
		for _, name := range composeutil.ParseCSV(service.Environment["KAFKA_CFG_CONTROLLER_LISTENER_NAMES"]) {
			controllerListenerNames[strings.TrimSpace(name)] = struct{}{}
		}
		voters, err := composeutil.ParseVoters(service.Environment["KAFKA_CFG_CONTROLLER_QUORUM_VOTERS"])
		if err != nil {
			continue
		}
		for _, address := range voters {
			voterSet[address] = struct{}{}
		}
	}

	controllerAddress := strings.TrimSpace(snap.Kafka.ControllerAddress)
	if controllerAddress == "" {
		return rule.NewSkip("KRF-005", "controller_endpoint_config", "kraft", "metadata 没有提供活动 controller 地址")
	}
	evidence = append(evidence, fmt.Sprintf("active_controller=%s", controllerAddress))

	if _, ok := voterSet[controllerAddress]; ok {
		names := make([]string, 0, len(controllerListenerNames))
		for name := range controllerListenerNames {
			names = append(names, name)
		}
		sort.Strings(names)
		result := rule.NewPass("KRF-005", "controller_endpoint_config", "kraft", "活动 controller 地址与 quorum voters 配置一致")
		if len(names) > 0 {
			result.Evidence = append(evidence, fmt.Sprintf("controller_listener_names=%v", names))
		} else {
			result.Evidence = evidence
		}
		return result
	}

	result := rule.NewFail("KRF-005", "controller_endpoint_config", "kraft", "活动 controller 地址不在 quorum voters 配置集合中，存在 controller 端点配置异常")
	result.Evidence = evidence
	result.NextActions = []string{"核对 controller.quorum.voters 与 controller.listener.names", "确认 metadata 报告的 controller 地址属于预期 listener", "检查 broker/controller 混用或迁移后的旧地址残留"}
	return result
}
