package kraft

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type MajorityChecker struct{}

func (MajorityChecker) ID() string     { return "KRF-004" }
func (MajorityChecker) Name() string   { return "controller_quorum_majority_evidence" }
func (MajorityChecker) Module() string { return "kraft" }

func (MajorityChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if bundle == nil || bundle.Network == nil || len(bundle.Network.ControllerChecks) == 0 {
		return rule.NewSkip("KRF-004", "controller_quorum_majority_evidence", "kraft", "当前没有可用的 controller quorum 端点，无法评估多数派")
	}

	reachable := 0
	evidence := make([]string, 0, len(bundle.Network.ControllerChecks)+1)
	for _, check := range bundle.Network.ControllerChecks {
		if check.Reachable {
			reachable++
			evidence = append(evidence, fmt.Sprintf("controller 节点=%s 可达 duration_ms=%d", check.Address, check.DurationMs))
			continue
		}
		evidence = append(evidence, fmt.Sprintf("controller 节点=%s 不可达：%s", check.Address, check.Error))
	}

	majority := len(bundle.Network.ControllerChecks)/2 + 1
	evidence = append(evidence, fmt.Sprintf("reachable=%d majority=%d", reachable, majority))

	result := rule.NewPass("KRF-004", "controller_quorum_majority_evidence", "kraft", "controller quorum 多数派证据正常")
	result.Evidence = evidence
	switch {
	case reachable < majority:
		result = rule.NewCrit("KRF-004", "controller_quorum_majority_evidence", "kraft", "controller quorum 当前缺少多数派证据")
		result.Evidence = evidence
		result.NextActions = []string{"确认 quorum voter 之间的 controller listener 是否互通", "检查 controller 进程与最近的选举错误", "确认 controller.quorum.voters 仍与当前拓扑一致"}
	case reachable < len(bundle.Network.ControllerChecks):
		result = rule.NewWarn("KRF-004", "controller_quorum_majority_evidence", "kraft", "controller quorum 仍有多数派，但部分 voter 不可达")
		result.Evidence = evidence
		result.NextActions = []string{"在再次故障前先恢复不可达的 controller listener", "比较当前执行主机与 broker 主机上的可达性差异", "检查 controller 日志中的间歇性网络或追加失败"}
	}
	return result
}
