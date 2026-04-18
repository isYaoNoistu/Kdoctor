package docker

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type RestartChecker struct{}

func (RestartChecker) ID() string     { return "DKR-006" }
func (RestartChecker) Name() string   { return "container_restart_history" }
func (RestartChecker) Module() string { return "docker" }

func (RestartChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	docker := dockerSnap(bundle)
	if docker == nil || !docker.Collected {
		return rule.NewSkip("DKR-006", "container_restart_history", "docker", "docker runtime is not enabled in the current input mode")
	}
	if !docker.Available {
		result := rule.NewSkip("DKR-006", "container_restart_history", "docker", "docker runtime is not available on the current execution host")
		result.Evidence = append(result.Evidence, docker.Errors...)
		return result
	}

	restarted := 0
	evidence := []string{}
	for _, container := range docker.Containers {
		evidence = append(evidence, fmt.Sprintf("%s restart_count=%d", container.Name, container.RestartCount))
		if container.RestartCount > 0 {
			restarted++
		}
	}

	result := rule.NewPass("DKR-006", "container_restart_history", "docker", "Kafka 容器未见异常重启历史")
	result.Evidence = evidence
	if restarted > 0 {
		result = rule.NewWarn("DKR-006", "container_restart_history", "docker", "部分 Kafka 容器存在重启历史，单次 Up 不代表过去稳定")
		result.Evidence = evidence
		result.NextActions = []string{"检查最近重启的时间线与 broker 日志", "结合 OOM、端口冲突和宿主机资源一起判断", "确认是否存在反复拉起但短暂恢复的故障"}
	}
	return result
}
