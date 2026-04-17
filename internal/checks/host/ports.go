package host

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type PortChecker struct{}

func (PortChecker) ID() string     { return "HOST-006" }
func (PortChecker) Name() string   { return "listener_port_occupation" }
func (PortChecker) Module() string { return "host" }

func (PortChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Host == nil || !snap.Host.Collected || len(snap.Host.PortChecks) == 0 {
		return rule.NewSkip("HOST-006", "listener_port_occupation", "host", "host listener ports are not available in the current input mode")
	}

	unreachable := 0
	evidence := []string{}
	for _, check := range snap.Host.PortChecks {
		if check.Reachable {
			evidence = append(evidence, fmt.Sprintf("%s reachable in %dms", check.Address, check.DurationMs))
			continue
		}
		unreachable++
		evidence = append(evidence, fmt.Sprintf("%s unreachable: %s", check.Address, check.Error))
	}

	result := rule.NewPass("HOST-006", "listener_port_occupation", "host", "expected Kafka listener ports are reachable from the host execution view")
	result.Evidence = evidence
	if unreachable > 0 {
		result = rule.NewFail("HOST-006", "listener_port_occupation", "host", "some expected Kafka listener ports are not reachable from the host execution view")
		result.Evidence = evidence
		result.NextActions = []string{"verify broker processes are listening on the expected ports", "check docker host network or service binding", "compare compose listener settings with the active process state"}
	}
	return result
}
