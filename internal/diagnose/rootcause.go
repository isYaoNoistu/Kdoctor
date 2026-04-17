package diagnose

import (
	"fmt"
	"sort"
	"strings"

	"kdoctor/pkg/model"
)

type RootCause struct{}

type correlatedCause struct {
	Priority int
	Summary  string
	Actions  []string
}

func (RootCause) Diagnose(report *model.Report) {
	if report == nil {
		return
	}

	if len(report.Checks) == 0 {
		report.Summary.Overview = "本次未执行任何检查项。"
		return
	}

	index := make(map[string]model.CheckResult, len(report.Checks))
	for _, check := range report.Checks {
		index[check.ID] = check
	}

	causes := inferRootCauses(index)
	report.Summary.RootCauses = nil
	report.Summary.RecommendedActions = nil

	if len(causes) == 0 {
		report.Summary.Overview = fmt.Sprintf("本次共执行 %d 项检查，最高状态为 %s，未识别出明确的单一主因，请结合各检查项逐条排查。", len(report.Checks), statusText(report.Summary.Status))
		appendFallbackCauses(report)
		return
	}

	sort.SliceStable(causes, func(i, j int) bool {
		return causes[i].Priority > causes[j].Priority
	})

	limit := min(3, len(causes))
	for i := 0; i < limit; i++ {
		report.Summary.RootCauses = append(report.Summary.RootCauses, causes[i].Summary)
		appendUniqueActions(&report.Summary.RecommendedActions, causes[i].Actions, 5)
	}

	report.Summary.Overview = fmt.Sprintf("本次共执行 %d 项检查，最高状态为 %s。已识别 %d 个优先级较高的主因，请优先按建议动作顺序处理。", len(report.Checks), statusText(report.Summary.Status), limit)
	appendFallbackCauses(report)
}

func inferRootCauses(index map[string]model.CheckResult) []correlatedCause {
	causes := []correlatedCause{}

	if isProblem(index["NET-001"]) {
		causes = append(causes, correlatedCause{
			Priority: 100,
			Summary:  "最可能主因：bootstrap 地址本身不可达，问题优先落在网络、防火墙、端口开放或 Kafka listener 绑定层。",
			Actions:  firstActions(index["NET-001"], "先确认 bootstrap 地址、端口、防火墙和 Kafka listener 绑定是否正确。"),
		})
	}

	if !isProblem(index["NET-001"]) && isProblem(index["NET-003"]) {
		causes = append(causes, correlatedCause{
			Priority: 95,
			Summary:  "最可能主因：metadata 返回的 broker 地址对当前客户端不可达，优先怀疑 advertised.listeners 配置、端口暴露或路由路径异常。",
			Actions:  firstActions(index["NET-003"], "对照 advertised.listeners、端口映射和当前客户端网络路径，确认 metadata 返回的地址真实可达。"),
		})
	}

	if isSeriousInternalTopicIssue(index["KFK-004"]) {
		causes = append(causes, correlatedCause{
			Priority: 92,
			Summary:  "高优先级主因：Kafka 内部主题异常，__consumer_offsets 缺失或副本不健康会直接影响消费组位点提交、协调器能力和部分客户端链路。",
			Actions:  firstActions(index["KFK-004"], "优先核对 controller、内部主题副本和 broker 日志，确认 __consumer_offsets 能被正常创建和加载。"),
		})
	}

	if isProblem(index["KRF-002"]) || isProblem(index["KRF-003"]) {
		causes = append(causes, correlatedCause{
			Priority: 90,
			Summary:  "高优先级主因：KRaft controller 或 quorum 存在异常，可能导致 metadata 不稳定、分区 leader 漂移或内部主题问题。",
			Actions:  mergeActionsWithFallback([]model.CheckResult{index["KRF-002"], index["KRF-003"]}, "优先确认 controller listener、quorum 多数派以及 controller 选举是否稳定。"),
		})
	}

	if isMissingTopicOrPartition(index["CLI-002"]) || isMissingTopicOrPartition(index["CLI-005"]) {
		causes = append(causes, correlatedCause{
			Priority: 88,
			Summary:  "高优先级主因：探针主题或对应分区在当前 broker 上不可用，可能是主题尚未创建、自动建主题未生效、metadata 尚未收敛或 leader 路由异常。",
			Actions:  mergeActionsWithFallback([]model.CheckResult{index["CLI-002"], index["CLI-005"]}, "优先核对探针主题是否存在、leader 是否可见，以及 broker 元数据是否已经收敛。"),
		})
	}

	if isProblem(index["TOP-003"]) || isProblem(index["TOP-004"]) || isProblem(index["TOP-005"]) {
		causes = append(causes, correlatedCause{
			Priority: 85,
			Summary:  "高优先级主因：Topic leader、ISR 或副本健康异常，已经足以影响生产、消费或 acks=all 写入成功率。",
			Actions:  mergeActionsWithFallback([]model.CheckResult{index["TOP-003"], index["TOP-004"], index["TOP-005"]}, "检查受影响分区的 leader、ISR、磁盘与 broker 副本状态。"),
		})
	}

	if isProblem(index["CLI-002"]) || isProblem(index["CLI-003"]) || isProblem(index["CLI-004"]) || isProblem(index["CLI-005"]) {
		causes = append(causes, correlatedCause{
			Priority: 80,
			Summary:  "业务链路主因：客户端端到端探针失败，说明问题已经影响真实的生产、消费或提交位点链路。",
			Actions:  mergeActionsWithFallback([]model.CheckResult{index["CLI-002"], index["CLI-003"], index["CLI-004"], index["CLI-005"]}, "先看失败阶段，再结合网络、KRaft、Topic 和 ISR 检查一起定位。"),
		})
	}

	if isProblem(index["HOST-004"]) || isProblem(index["HOST-006"]) || isProblem(index["DKR-002"]) || isProblem(index["DKR-003"]) || isProblem(index["DKR-004"]) {
		causes = append(causes, correlatedCause{
			Priority: 70,
			Summary:  "环境层主因：宿主机或 Docker 层存在异常，问题可能并非 Kafka 配置本身，而是资源、端口或挂载环境失真。",
			Actions:  mergeActionsWithFallback([]model.CheckResult{index["HOST-004"], index["HOST-006"], index["DKR-002"], index["DKR-003"], index["DKR-004"]}, "优先排查宿主机资源、容器运行态、端口监听与数据目录挂载。"),
		})
	}

	if isProblem(index["LOG-002"]) {
		causes = append(causes, correlatedCause{
			Priority: 65,
			Summary:  "辅助主因：日志已经命中关键错误指纹，可以把日志信号作为当前根因判断的加权证据。",
			Actions:  firstActions(index["LOG-002"], "结合命中的日志指纹和解释结果，优先沿对应故障链路继续深挖。"),
		})
	}

	return causes
}

func appendFallbackCauses(report *model.Report) {
	if report == nil {
		return
	}
	for _, check := range report.Checks {
		if !check.IsProblem() {
			continue
		}
		if len(report.Summary.RootCauses) >= 3 {
			break
		}
		report.Summary.RootCauses = append(report.Summary.RootCauses, fmt.Sprintf("%s：%s", check.ID, check.Summary))
		appendUniqueActions(&report.Summary.RecommendedActions, check.NextActions, 5)
	}
}

func firstActions(check model.CheckResult, fallback string) []string {
	if len(check.NextActions) > 0 {
		return dedupe(check.NextActions)
	}
	if fallback == "" {
		return nil
	}
	return []string{fallback}
}

func mergeActions(checks ...model.CheckResult) []string {
	actions := []string{}
	for _, check := range checks {
		actions = append(actions, check.NextActions...)
	}
	if len(actions) == 0 {
		return nil
	}
	return dedupe(actions)
}

func mergeActionsWithFallback(checks []model.CheckResult, fallback string) []string {
	actions := mergeActions(checks...)
	if len(actions) > 0 {
		return actions
	}
	if fallback == "" {
		return nil
	}
	return []string{fallback}
}

func appendUniqueActions(target *[]string, actions []string, limit int) {
	if target == nil || len(actions) == 0 {
		return
	}
	current := append([]string(nil), (*target)...)
	current = append(current, actions...)
	current = dedupe(current)
	if limit > 0 && len(current) > limit {
		current = current[:limit]
	}
	*target = current
}

func dedupe(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func isProblem(check model.CheckResult) bool {
	return check.IsProblem()
}

func isSeriousInternalTopicIssue(check model.CheckResult) bool {
	if !check.IsProblem() {
		return false
	}
	switch check.Status {
	case model.StatusFail, model.StatusCrit, model.StatusError, model.StatusTimeout:
		return containsAny(strings.ToLower(check.Summary), "__consumer_offsets", "internal kafka topics are unhealthy")
	default:
		return false
	}
}

func isMissingTopicOrPartition(check model.CheckResult) bool {
	if !check.IsProblem() {
		return false
	}
	texts := append([]string{check.Summary, check.ErrorMessage}, check.Evidence...)
	for _, text := range texts {
		if containsAny(strings.ToLower(text), "topic or partition that does not exist", "unknown_topic_or_partition") {
			return true
		}
	}
	return false
}

func containsAny(text string, fragments ...string) bool {
	for _, fragment := range fragments {
		if strings.Contains(text, strings.ToLower(fragment)) {
			return true
		}
	}
	return false
}

func statusText(status model.CheckStatus) string {
	switch status {
	case model.StatusPass:
		return "通过"
	case model.StatusWarn:
		return "告警"
	case model.StatusFail:
		return "失败"
	case model.StatusCrit:
		return "严重"
	case model.StatusError:
		return "错误"
	case model.StatusTimeout:
		return "超时"
	case model.StatusSkip:
		return "跳过"
	default:
		return string(status)
	}
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
