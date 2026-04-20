package kraft

import (
	"context"
	"fmt"
	"sort"
	"strconv"
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

	controllerID := int(*snap.Kafka.ControllerID)
	controllerAddress := strings.TrimSpace(snap.Kafka.ControllerAddress)
	evidence := []string{
		fmt.Sprintf("active_controller_id=%d", controllerID),
	}
	if controllerAddress != "" {
		evidence = append(evidence, fmt.Sprintf("active_controller_broker_address=%s", controllerAddress))
	}

	serviceByNodeID := map[int]composeutil.KafkaService{}
	voterByNodeID := map[int]string{}
	for _, service := range services {
		rawNodeID := strings.TrimSpace(service.Environment["KAFKA_CFG_NODE_ID"])
		if rawNodeID != "" {
			if nodeID, err := strconv.Atoi(rawNodeID); err == nil {
				serviceByNodeID[nodeID] = service
			}
		}
		voters, err := composeutil.ParseVoters(service.Environment["KAFKA_CFG_CONTROLLER_QUORUM_VOTERS"])
		if err != nil {
			continue
		}
		for nodeID, address := range voters {
			if existing, ok := voterByNodeID[nodeID]; !ok || existing == "" {
				voterByNodeID[nodeID] = address
			}
		}
	}

	service, ok := serviceByNodeID[controllerID]
	if !ok {
		result := rule.NewFail("KRF-005", "controller_endpoint_config", "kraft", "活动 controller ID 没有映射到 compose 中的 Kafka 节点")
		result.Evidence = evidence
		result.NextActions = []string{"核对 compose 中的 node.id 是否完整", "确认 metadata 返回的 controller ID 对应当前运行节点", "检查 broker 重建或拓扑变更后是否存在旧配置残留"}
		return result
	}

	if !serviceHasRole(service, "controller") {
		result := rule.NewFail("KRF-005", "controller_endpoint_config", "kraft", "活动 controller 所属节点在 compose 中未声明 controller 角色")
		result.Evidence = append(evidence, fmt.Sprintf("service=%s node_id=%d roles=%s", service.ServiceName, controllerID, strings.TrimSpace(service.Environment["KAFKA_CFG_PROCESS_ROLES"])))
		result.NextActions = []string{"核对 process.roles 是否包含 controller", "检查 node.id 与 controller 角色配置是否一致", "确认最近是否存在角色迁移或旧配置残留"}
		return result
	}

	expectedVoter, ok := voterByNodeID[controllerID]
	if !ok || strings.TrimSpace(expectedVoter) == "" {
		result := rule.NewFail("KRF-005", "controller_endpoint_config", "kraft", "活动 controller ID 不在 quorum voters 配置集合中")
		result.Evidence = append(evidence, fmt.Sprintf("service=%s node_id=%d", service.ServiceName, controllerID))
		result.NextActions = []string{"核对 controller.quorum.voters 是否覆盖当前 controller 节点", "确认所有 node.id 都正确出现在 quorum voters 中", "检查控制面配置是否有节点遗漏"}
		return result
	}

	controllerListeners, listenerEvidence, matched := activeControllerListenerEvidence(service, controllerID, expectedVoter)
	evidence = append(evidence, fmt.Sprintf("service=%s node_id=%d expected_voter=%s", service.ServiceName, controllerID, expectedVoter))
	evidence = append(evidence, listenerEvidence...)

	if matched {
		result := rule.NewPass("KRF-005", "controller_endpoint_config", "kraft", "活动 controller 节点与 quorum voters 中的 controller listener 配置一致")
		result.Evidence = evidence
		if len(controllerListeners) > 0 && controllerAddress != "" {
			result.Evidence = append(result.Evidence, "metadata 返回的是活动 controller 所属 broker 地址；它不要求等于 quorum voters 中的 CONTROLLER 端点")
		}
		return result
	}

	result := rule.NewFail("KRF-005", "controller_endpoint_config", "kraft", "活动 controller 节点存在 controller listener 配置异常")
	result.Evidence = evidence
	result.NextActions = []string{"核对 controller.quorum.voters 与 controller.listener.names", "确认活动 controller 节点的 CONTROLLER listener 与 voter 地址一致", "检查 broker/controller 混用或迁移后的旧地址残留"}
	return result
}

func serviceHasRole(service composeutil.KafkaService, role string) bool {
	for _, item := range composeutil.ParseCSV(service.Environment["KAFKA_CFG_PROCESS_ROLES"]) {
		if strings.EqualFold(strings.TrimSpace(item), strings.TrimSpace(role)) {
			return true
		}
	}
	return false
}

func activeControllerListenerEvidence(service composeutil.KafkaService, controllerID int, expectedVoter string) ([]string, []string, bool) {
	names := composeutil.ParseCSV(service.Environment["KAFKA_CFG_CONTROLLER_LISTENER_NAMES"])
	sort.Strings(names)
	if len(names) == 0 {
		return nil, []string{fmt.Sprintf("service=%s node_id=%d controller_listener_names missing", service.ServiceName, controllerID)}, false
	}

	listeners, err := composeutil.ParseListeners(service.Environment["KAFKA_CFG_LISTENERS"])
	if err != nil {
		return nil, []string{fmt.Sprintf("service=%s node_id=%d listeners parse error=%v", service.ServiceName, controllerID, err)}, false
	}

	controllerListeners := make([]string, 0, len(names))
	evidence := make([]string, 0, len(names)+1)
	evidence = append(evidence, fmt.Sprintf("controller_listener_names=%v", names))
	matched := false
	for _, name := range names {
		listener, ok := listeners[name]
		if !ok {
			evidence = append(evidence, fmt.Sprintf("controller_listener=%s missing_in_listeners", name))
			continue
		}
		address := fmt.Sprintf("%s:%d", listener.Host, listener.Port)
		controllerListeners = append(controllerListeners, address)
		evidence = append(evidence, fmt.Sprintf("controller_listener=%s address=%s", listener.Name, address))
		if strings.EqualFold(strings.TrimSpace(address), strings.TrimSpace(expectedVoter)) {
			matched = true
		}
	}
	return controllerListeners, evidence, matched
}
