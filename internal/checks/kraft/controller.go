package kraft

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type ControllerChecker struct{}

func (ControllerChecker) ID() string     { return "KRF-002" }
func (ControllerChecker) Name() string   { return "active_controller" }
func (ControllerChecker) Module() string { return "kraft" }

func (ControllerChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Kafka == nil {
		return rule.NewError("KRF-002", "active_controller", "kraft", "active controller cannot be evaluated", "kafka snapshot missing")
	}
	if snap.Kafka.ControllerID == nil {
		result := rule.NewFail("KRF-002", "active_controller", "kraft", "metadata did not report an active controller")
		result.NextActions = []string{"verify controller quorum majority", "check controller listener reachability", "inspect broker logs for controller election errors"}
		return result
	}

	controllerID := *snap.Kafka.ControllerID
	result := rule.NewPass("KRF-002", "active_controller", "kraft", "metadata reports an active controller")
	result.Evidence = []string{fmt.Sprintf("controller id=%d address=%s", controllerID, snap.Kafka.ControllerAddress)}

	if !controllerIsRegistered(controllerID, snap.Kafka.Brokers) {
		result = rule.NewFail("KRF-002", "active_controller", "kraft", "active controller is not present in the broker registration set")
		result.Evidence = []string{fmt.Sprintf("controller id=%d address=%s", controllerID, snap.Kafka.ControllerAddress)}
		result.NextActions = []string{"verify broker registration", "verify controller election completed", "inspect broker logs for registration failures"}
		return result
	}

	if snap.Network == nil || len(snap.Network.ControllerChecks) == 0 {
		result.Evidence = append(result.Evidence, "metadata 返回的是活动 controller 所属 broker 地址；当前输入模式未直接探测 CONTROLLER listener")
		return result
	}

	for _, check := range snap.Network.ControllerChecks {
		if check.Address != snap.Kafka.ControllerAddress {
			continue
		}
		if check.Reachable {
			result.Evidence = append(result.Evidence, fmt.Sprintf("controller listener reachable in %dms", check.DurationMs))
			return result
		}
		if isExternalProbeView(snap) && isPrivateEndpoint(check.Address) {
			result.Evidence = append(result.Evidence, "controller listener is private and was not directly reachable from the current external view")
			return result
		}
		result = rule.NewWarn("KRF-002", "active_controller", "kraft", "metadata reports an active controller but its listener is not reachable from the current execution view")
		result.Evidence = []string{
			fmt.Sprintf("controller id=%d address=%s", controllerID, snap.Kafka.ControllerAddress),
			fmt.Sprintf("reachability error=%s", check.Error),
		}
		result.NextActions = []string{"verify controller listener binding and exposure", "run kdoctor from the Kafka host or private network", "check controller logs for election churn"}
		return result
	}

	result.Evidence = append(result.Evidence, "metadata 返回的是活动 controller 所属 broker 地址；它不要求等于 controller.quorum.voters 中的 CONTROLLER 端点")
	return result
}

func controllerIsRegistered(controllerID int32, brokers []snapshot.BrokerSnapshot) bool {
	for _, broker := range brokers {
		if broker.ID == controllerID {
			return true
		}
	}
	return false
}
