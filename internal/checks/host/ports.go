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
		return rule.NewSkip("HOST-006", "listener_port_occupation", "host", "当前输入模式下没有可用的宿主机 listener 端口证据")
	}

	unreachable := 0
	evidence := []string{}
	for _, check := range snap.Host.PortChecks {
		if check.Reachable {
			evidence = append(evidence, fmt.Sprintf("%s 可达，耗时 %dms", check.Address, check.DurationMs))
			continue
		}
		unreachable++
		evidence = append(evidence, fmt.Sprintf("%s 不可达：%s", check.Address, check.Error))
	}

	result := rule.NewPass("HOST-006", "listener_port_occupation", "host", "从宿主机执行视角看，期望的 Kafka listener 端口可达")
	result.Evidence = evidence
	if unreachable > 0 {
		result = rule.NewFail("HOST-006", "listener_port_occupation", "host", "从宿主机执行视角看，部分期望的 Kafka listener 端口不可达")
		result.Evidence = evidence
		result.NextActions = []string{"确认 broker 进程正在预期端口监听", "检查 Docker host network 或服务绑定", "对比 compose listener 配置与实际进程状态"}
	}
	return result
}
