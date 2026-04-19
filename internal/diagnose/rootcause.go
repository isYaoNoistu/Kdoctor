package diagnose

import (
	"fmt"
	"sort"
	"strings"

	"kdoctor/pkg/model"
)

type RootCause struct {
	MaxCauses        int
	EnableConfidence bool
}

type correlatedCause struct {
	Priority    int
	Confidence  string
	Summary     string
	Actions     []string
	Limitations []string
}

func (d RootCause) Diagnose(report *model.Report) {
	if report == nil {
		return
	}

	maxCauses := d.MaxCauses
	if maxCauses <= 0 {
		maxCauses = 3
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
		report.Summary.Overview = fmt.Sprintf("本次共执行 %d 项检查，最高状态为 %s，未识别出足够明确的单一主因，请结合重点问题继续排查。", len(report.Checks), statusText(report.Summary.Status))
		appendFallbackCauses(report, maxCauses)
		return
	}

	sort.SliceStable(causes, func(i, j int) bool {
		return causes[i].Priority > causes[j].Priority
	})

	limit := min(maxCauses, len(causes))
	for i := 0; i < limit; i++ {
		report.Summary.RootCauses = append(report.Summary.RootCauses, d.renderCause(causes[i]))
		appendUniqueActions(&report.Summary.RecommendedActions, causes[i].Actions, 5)
	}

	report.Summary.Overview = fmt.Sprintf("本次共执行 %d 项检查，最高状态为 %s。已识别 %d 个优先级较高的主因，请优先按建议动作顺序处理。", len(report.Checks), statusText(report.Summary.Status), limit)
}

func (d RootCause) renderCause(cause correlatedCause) string {
	text := cause.Summary
	if d.EnableConfidence && cause.Confidence != "" {
		text = fmt.Sprintf("%s：%s", cause.Confidence, text)
	}
	limitations := dedupe(cause.Limitations)
	if len(limitations) > 0 {
		text = fmt.Sprintf("%s 反证/局限：%s。", text, strings.Join(limitations, "；"))
	}
	return text
}

func inferRootCauses(index map[string]model.CheckResult) []correlatedCause {
	causes := []correlatedCause{}

	netBootstrap := index["NET-001"]
	netMetadata := index["NET-003"]
	netRouteMismatch := index["NET-005"]
	netAdvertisedPrivate := index["NET-006"]
	netProtocolMismatch := index["NET-009"]
	internalTopics := index["KFK-004"]
	kafkaRoute := index["KFK-005"]
	kraftController := index["KRF-002"]
	kraftQuorum := index["KRF-003"]
	kraftMajority := index["KRF-004"]
	kraftEndpointConfig := index["KRF-005"]
	topicLeader := index["TOP-003"]
	topicReplica := index["TOP-004"]
	topicISR := index["TOP-005"]
	topicURP := index["TOP-006"]
	topicUnderMinISR := index["TOP-007"]
	topicOffline := index["TOP-008"]
	topicLeaderSkew := index["TOP-009"]
	topicPlanning := index["TOP-011"]
	clientProduce := index["CLI-002"]
	clientCommit := index["CLI-004"]
	clientEndToEnd := index["CLI-005"]
	consumerLag := index["CSM-001"]
	consumerRebalance := index["CSM-002"]
	consumerCoordinator := index["CSM-006"]
	securityListener := index["SEC-001"]
	securitySASL := index["SEC-002"]
	securityTLS := index["SEC-003"]
	securityAuthorization := index["SEC-004"]
	securityAuthorizer := index["SEC-005"]
	storageCapacity := index["STG-001"]
	storageLayout := index["STG-003"]
	storageMounts := index["STG-005"]
	storageTiered := index["STG-006"]
	producerAcks := index["PRD-001"]
	producerIdempotence := index["PRD-002"]
	producerTimeout := index["PRD-003"]
	producerMessageSize := index["PRD-004"]
	producerTxnTimeout := index["PRD-006"]
	transactionContext := index["TXN-001"]
	transactionRequired := index["TXN-002"]
	transactionTimeout := index["TXN-003"]
	transactionIsolation := index["TXN-004"]
	transactionOutcome := index["TXN-005"]
	hostDisk := index["HOST-004"]
	hostCapacity := index["HOST-007"]
	hostFD := index["HOST-008"]
	hostListener := index["HOST-010"]
	hostMemory := index["HOST-011"]
	dockerExistence := index["DKR-001"]
	dockerRunning := index["DKR-002"]
	dockerOOM := index["DKR-003"]
	dockerRestart := index["DKR-004"]
	dockerMemory := index["DKR-006"]
	dockerMounts := index["DKR-007"]
	logs := index["LOG-002"]
	logStorm := index["LOG-005"]
	logExplain := index["LOG-007"]

	if isProblem(netBootstrap) {
		causes = append(causes, correlatedCause{
			Priority:   100,
			Confidence: "最可能主因",
			Summary:    "bootstrap 地址本身不可达，问题优先落在网络、防火墙、端口开放或 Kafka listener 绑定层。",
			Actions:    firstActions(netBootstrap, "先确认 bootstrap 地址、端口、防火墙和 Kafka listener 绑定是否正确。"),
		})
	}

	if !isProblem(netBootstrap) && (isProblem(netMetadata) || isProblem(netRouteMismatch) || isProblem(netAdvertisedPrivate) || isProblem(kafkaRoute)) {
		causes = append(causes, correlatedCause{
			Priority:   95,
			Confidence: "高优先级主因",
			Summary:    "metadata 返回地址与当前客户端视角不匹配，更像是 advertised.listeners、路由、NAT 或端口暴露设计错位。",
			Actions:    mergeActionsWithFallback([]model.CheckResult{netMetadata, netRouteMismatch, netAdvertisedPrivate, kafkaRoute}, "优先核对 advertised.listeners、端口暴露、负载均衡入口和当前客户端网络路径。"),
			Limitations: []string{
				"bootstrap 仍可达，说明问题更偏向返回地址与后续路由，而不一定是整个集群完全离线",
			},
		})
	}

	if isProblem(netProtocolMismatch) {
		causes = append(causes, correlatedCause{
			Priority:   92,
			Confidence: "高优先级主因",
			Summary:    "端口层面可连通，但 Kafka 协议握手失败，优先考虑 listener 安全协议、代理转发或端口后端服务错配。",
			Actions:    firstActions(netProtocolMismatch, "优先核对 listener.security.protocol.map、SASL/SSL 和代理转发配置。"),
		})
	}

	if isProblem(kraftController) || isProblem(kraftQuorum) || isProblem(kraftMajority) || isProblem(kraftEndpointConfig) {
		causes = append(causes, correlatedCause{
			Priority:   90,
			Confidence: "高优先级主因",
			Summary:    "KRaft controller 或 quorum 存在异常，可能导致 metadata 不稳定、分区 leader 漂移或内部主题问题。",
			Actions:    mergeActionsWithFallback([]model.CheckResult{kraftController, kraftQuorum, kraftMajority, kraftEndpointConfig}, "优先确认 controller listener、quorum 多数派以及 controller 选举是否稳定。"),
		})
	}

	if isProblem(topicLeader) || isProblem(topicReplica) || isProblem(topicISR) || isProblem(topicURP) || isProblem(topicUnderMinISR) || isProblem(topicOffline) || isProblem(topicLeaderSkew) {
		causes = append(causes, correlatedCause{
			Priority:   86,
			Confidence: "高优先级主因",
			Summary:    "Topic leader、ISR 或副本健康异常，已经足以影响生产、消费或 acks=all 写入成功率。",
			Actions:    mergeActionsWithFallback([]model.CheckResult{topicLeader, topicReplica, topicISR, topicURP, topicUnderMinISR, topicOffline, topicLeaderSkew}, "检查受影响分区的 leader、ISR、磁盘与 broker 副本状态。"),
			Limitations: []string{
				"如果 metadata 返回地址本身不可达，部分 leader 或 ISR 异常也可能被客户端视角问题放大",
			},
		})
	}

	if isProblem(clientProduce) || isProblem(clientCommit) || isProblem(clientEndToEnd) {
		causes = append(causes, correlatedCause{
			Priority:   82,
			Confidence: "业务链路主因",
			Summary:    "客户端端到端探针失败，说明问题已经影响真实生产、消费或位点提交链路。",
			Actions:    mergeActionsWithFallback([]model.CheckResult{clientProduce, clientCommit, clientEndToEnd}, "先检查失败阶段本身，再结合网络、KRaft 和 Topic 检查一起定位。"),
		})
	}

	if isProblem(consumerLag) || isProblem(consumerRebalance) || isProblem(consumerCoordinator) {
		causes = append(causes, correlatedCause{
			Priority:   79,
			Confidence: "高优先级主因",
			Summary:    "消费组 lag、rebalance 或 coordinator 视图异常，问题已经进入真实消费链路，而不只是探针侧抖动。",
			Actions:    mergeActionsWithFallback([]model.CheckResult{consumerLag, consumerRebalance, consumerCoordinator}, "优先核对消费组 lag、成员稳定性、coordinator 和 __consumer_offsets 状态。"),
		})
	}

	if isSeriousInternalTopicIssue(internalTopics) || isProblem(transactionRequired) || isProblem(transactionIsolation) || isProblem(transactionOutcome) {
		causes = append(causes, correlatedCause{
			Priority:   78,
			Confidence: "高优先级主因",
			Summary:    "内部主题或事务链路已经异常，尤其是 __consumer_offsets / __transaction_state 不健康时，会直接影响提交、协调和事务可见性。",
			Actions:    mergeActionsWithFallback([]model.CheckResult{internalTopics, transactionRequired, transactionIsolation, transactionOutcome}, "优先核对 controller、内部主题副本和 broker 日志，确认内部主题能正常创建与加载。"),
		})
	}

	if isContextOnlyTransactionHint(internalTopics, transactionContext) {
		causes = append(causes, correlatedCause{
			Priority:   62,
			Confidence: "上下文提示",
			Summary:    "当前没有看到事务使用证据，__transaction_state 缺失更像是未启用事务场景的背景提示，不应优先当成真实故障。",
			Actions:    mergeActionsWithFallback([]model.CheckResult{transactionContext}, "如果环境未使用事务，这条通常可以降级处理；如果准备启用事务，再单独核对事务主题。"),
		})
	}

	if isProblem(securityListener) || isProblem(securitySASL) || isProblem(securityTLS) || isProblem(securityAuthorization) || isProblem(securityAuthorizer) {
		causes = append(causes, correlatedCause{
			Priority:   75,
			Confidence: "高优先级主因",
			Summary:    "安全协议、SASL/TLS 或授权配置与当前执行视角不一致，这类问题常表现为端口可达但 metadata、认证或 ACL 行为异常。",
			Actions:    mergeActionsWithFallback([]model.CheckResult{securityListener, securitySASL, securityTLS, securityAuthorization, securityAuthorizer}, "优先核对 listener 安全协议、SASL/TLS 证书和 Authorizer 配置是否与当前客户端视角一致。"),
		})
	}

	if isProblem(storageCapacity) || isProblem(storageLayout) || isProblem(storageMounts) || isProblem(storageTiered) || isProblem(hostDisk) || isProblem(hostCapacity) || isProblem(hostFD) || isProblem(hostListener) || isProblem(hostMemory) || isProblem(dockerExistence) || isProblem(dockerRunning) || isProblem(dockerOOM) || isProblem(dockerRestart) || isProblem(dockerMemory) || isProblem(dockerMounts) {
		causes = append(causes, correlatedCause{
			Priority:   72,
			Confidence: "环境侧主因",
			Summary:    "宿主机、Docker、磁盘目录或挂载规划存在异常，当前问题不一定是 Kafka 配置本身，而可能是运行环境失真。",
			Actions:    mergeActionsWithFallback([]model.CheckResult{storageCapacity, storageLayout, storageMounts, hostDisk, hostCapacity, hostFD, hostListener, hostMemory, dockerExistence, dockerRunning, dockerOOM, dockerRestart, dockerMemory, dockerMounts}, "优先排查宿主机资源、容器运行状态、数据目录挂载和端口监听情况。"),
		})
	}

	if isProblem(producerAcks) || isProblem(producerIdempotence) || isProblem(producerTimeout) || isProblem(producerMessageSize) || isProblem(producerTxnTimeout) {
		causes = append(causes, correlatedCause{
			Priority:   68,
			Confidence: "配置侧主因",
			Summary:    "生产端参数组合存在一致性、幂等或超时风险，即使集群健康也可能出现重复、乱序或事务初始化失败。",
			Actions:    mergeActionsWithFallback([]model.CheckResult{producerAcks, producerIdempotence, producerTimeout, producerMessageSize, producerTxnTimeout}, "优先核对 producer acks、幂等、消息大小和超时组合是否匹配当前集群策略。"),
		})
	}

	if isProblem(topicPlanning) {
		causes = append(causes, correlatedCause{
			Priority:   60,
			Confidence: "规划风险",
			Summary:    "部分 topic 的分区或副本规划不理想，当前未必立刻引发故障，但会放大热点、扩容和恢复风险。",
			Actions:    firstActions(topicPlanning, "复核关键业务 topic 的 partitions 和 replication factor，避免长期规划偏差。"),
		})
	}

	if isProblem(logs) || isProblem(logStorm) || isProblem(logExplain) {
		causes = append(causes, correlatedCause{
			Priority:   58,
			Confidence: "辅助证据",
			Summary:    "近期日志已命中关键错误指纹，可以作为当前根因判断的辅助证据，但仍需与 metadata、探针和副本状态交叉验证。",
			Actions:    mergeActionsWithFallback([]model.CheckResult{logs, logStorm, logExplain}, "结合日志指纹、出现频次和对应检查项，沿着命中的故障链路继续深挖。"),
		})
	}

	if isProblem(transactionTimeout) {
		causes = append(causes, correlatedCause{
			Priority:   55,
			Confidence: "配置风险",
			Summary:    "事务超时配置与 broker 上限不一致，事务型生产者会在初始化或提交阶段直接失败。",
			Actions:    firstActions(transactionTimeout, "核对 transaction.timeout.ms 与 broker transaction.max.timeout.ms 的关系。"),
		})
	}

	return causes
}

func appendFallbackCauses(report *model.Report, limit int) {
	if report == nil {
		return
	}
	for _, check := range report.Checks {
		if !check.IsProblem() {
			continue
		}
		if len(report.Summary.RootCauses) >= limit {
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
	text := strings.ToLower(check.Summary + " " + check.ErrorMessage + " " + strings.Join(check.Evidence, " "))
	return containsAny(text, "__consumer_offsets", "__transaction_state", "内部主题", "internal kafka topics")
}

func isContextOnlyTransactionHint(internalTopics model.CheckResult, transactionContext model.CheckResult) bool {
	if !isProblem(transactionContext) {
		return false
	}
	if isProblem(internalTopics) {
		text := strings.ToLower(internalTopics.Summary + " " + strings.Join(internalTopics.Evidence, " "))
		if containsAny(text, "__transaction_state", "transaction") && !containsAny(text, "__consumer_offsets") {
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
