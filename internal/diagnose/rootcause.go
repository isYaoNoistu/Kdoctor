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
		report.Summary.Overview = fmt.Sprintf("本次共执行 %d 项检查，最高状态为 %s，未识别出明确的单一主因，请结合各检查项逐条排查。", len(report.Checks), statusText(report.Summary.Status))
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
	netDNSDrift := index["NET-008"]
	netProtocolMismatch := index["NET-009"]
	internalTopics := index["KFK-004"]
	kafkaRoute := index["KFK-005"]
	kafkaMetadataLatency := index["KFK-008"]
	kraftController := index["KRF-002"]
	kraftQuorum := index["KRF-003"]
	kraftMajority := index["KRF-004"]
	kraftEndpointConfig := index["KRF-005"]
	kraftEpoch := index["KRF-006"]
	kraftUnknownVoter := index["KRF-007"]
	kraftFinalization := index["KRF-008"]
	topicLeader := index["TOP-003"]
	topicReplica := index["TOP-004"]
	topicISR := index["TOP-005"]
	topicURP := index["TOP-006"]
	topicUnderMinISR := index["TOP-007"]
	topicOffline := index["TOP-008"]
	topicLeaderSkew := index["TOP-009"]
	topicReplicaLag := index["TOP-010"]
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
	metricURP := index["MET-001"]
	metricMinISR := index["MET-002"]
	metricReplicaLag := index["MET-003"]
	metricOfflineLogDir := index["MET-004"]
	metricNetworkIdle := index["MET-005"]
	metricRequestIdle := index["MET-006"]
	jvmNetworkIdle := index["JVM-001"]
	jvmRequestIdle := index["JVM-002"]
	jvmRequestPressure := index["JVM-003"]
	jvmHeapGC := index["JVM-004"]
	producerAcks := index["PRD-001"]
	producerIdempotence := index["PRD-002"]
	producerTimeout := index["PRD-003"]
	producerMessageSize := index["PRD-004"]
	producerThrottle := index["PRD-005"]
	producerTxnTimeout := index["PRD-006"]
	transactionTopicRequired := index["TXN-002"]
	transactionIsolation := index["TXN-004"]
	transactionOutcome := index["TXN-005"]
	upgradeVersion := index["UPG-001"]
	upgradeFeature := index["UPG-002"]
	quotaProduce := index["QTA-001"]
	quotaFetch := index["QTA-002"]
	quotaRequest := index["QTA-003"]
	quotaBackpressure := index["QTA-004"]
	hostCapacity := index["HOST-007"]
	hostFD := index["HOST-008"]
	hostClock := index["HOST-009"]
	hostListener := index["HOST-010"]
	hostMemory := index["HOST-011"]
	dockerMountRuntime := index["DKR-005"]
	logs := index["LOG-002"]

	if isProblem(netBootstrap) {
		causes = append(causes, correlatedCause{
			Priority:   100,
			Confidence: "高置信度主因",
			Summary:    "bootstrap 地址本身不可达，问题优先落在网络、防火墙、端口开放或 Kafka listener 绑定层。",
			Actions:    firstActions(netBootstrap, "先确认 bootstrap 地址、端口、防火墙和 Kafka listener 绑定是否正确。"),
		})
	}

	if !isProblem(netBootstrap) && isProblem(netMetadata) {
		causes = append(causes, correlatedCause{
			Priority:   95,
			Confidence: "高置信度主因",
			Summary:    "metadata 返回的 broker 地址对当前客户端不可达，更像是 advertised.listeners、端口暴露或路由视角问题，而不是整个集群完全离线。",
			Actions:    firstActions(netMetadata, "对照 advertised.listeners、端口映射和当前客户端网络路径，确认 metadata 返回的地址真实可达。"),
			Limitations: []string{
				"bootstrap 地址仍然可达，因此不像整组 broker 全部宕机",
			},
		})
	}

	if isProblem(netRouteMismatch) || isProblem(netAdvertisedPrivate) || isProblem(netDNSDrift) || isProblem(kafkaRoute) {
		causes = append(causes, correlatedCause{
			Priority:   94,
			Confidence: "高置信度主因",
			Summary:    "入口 bootstrap 与 metadata 返回地址之间存在明显分裂，问题更像是 advertised.listeners、路由或 NAT/LB 设计错位。",
			Actions:    mergeActionsWithFallback([]model.CheckResult{netRouteMismatch, netAdvertisedPrivate, netDNSDrift, kafkaRoute}, "优先核对 advertised.listeners、返回地址路由与负载均衡入口设计。"),
		})
	}

	if isProblem(netProtocolMismatch) {
		causes = append(causes, correlatedCause{
			Priority:   93,
			Confidence: "高置信度主因",
			Summary:    "端口层面可连通，但 Kafka 协议握手失败，这更像是 security protocol、代理转发或监听端口本身的协议错配。",
			Actions:    firstActions(netProtocolMismatch, "优先核对 listener 安全协议、SASL/SSL 与端口后端的真实服务类型。"),
		})
	}

	if isSeriousInternalTopicIssue(internalTopics) {
		causes = append(causes, correlatedCause{
			Priority:   92,
			Confidence: "高置信度主因",
			Summary:    "Kafka 内部主题异常，尤其是 __consumer_offsets 缺失或副本不健康，会直接影响消费组位点提交、协调器能力和部分客户端链路。",
			Actions:    firstActions(internalTopics, "优先核对 controller、内部主题副本和 broker 日志，确认 __consumer_offsets 能被正常创建和加载。"),
		})
	}

	if isProblem(kraftController) || isProblem(kraftQuorum) || isProblem(kraftMajority) || isProblem(kraftEndpointConfig) || isProblem(kraftEpoch) || isProblem(kraftUnknownVoter) || isProblem(kraftFinalization) {
		causes = append(causes, correlatedCause{
			Priority:   90,
			Confidence: "中高置信度主因",
			Summary:    "KRaft controller 或 quorum 存在异常，可能导致 metadata 不稳定、分区 leader 漂移或内部主题问题。",
			Actions:    mergeActionsWithFallback([]model.CheckResult{kraftController, kraftQuorum, kraftMajority, kraftEndpointConfig, kraftEpoch, kraftUnknownVoter, kraftFinalization}, "优先确认 controller listener、quorum 多数派以及 controller 选举是否稳定。"),
		})
	}

	if isMissingTopicOrPartition(clientProduce) || isMissingTopicOrPartition(clientEndToEnd) {
		causes = append(causes, correlatedCause{
			Priority:   88,
			Confidence: "中高置信度主因",
			Summary:    "探针主题或对应分区在当前 broker 上不可用，可能是主题尚未创建、自动建主题未生效、metadata 尚未收敛或 leader 路由异常。",
			Actions:    mergeActionsWithFallback([]model.CheckResult{clientProduce, clientEndToEnd}, "优先核对探针主题是否存在、leader 是否可见，以及 broker 元数据是否已经收敛。"),
		})
	}

	if isProblem(topicLeader) || isProblem(topicReplica) || isProblem(topicISR) || isProblem(topicURP) || isProblem(topicUnderMinISR) || isProblem(topicOffline) || isProblem(topicLeaderSkew) || isProblem(topicReplicaLag) || isProblem(metricReplicaLag) {
		limitations := []string{}
		if isProblem(netMetadata) {
			limitations = append(limitations, "若 metadata 返回端点本身不可达，部分 leader/ISR 异常也可能掺杂客户端视角问题")
		}
		causes = append(causes, correlatedCause{
			Priority:    85,
			Confidence:  "中高置信度主因",
			Summary:     "Topic leader、ISR 或副本健康异常，已经足以影响生产、消费或 acks=all 写入成功率。",
			Actions:     mergeActionsWithFallback([]model.CheckResult{topicLeader, topicReplica, topicISR, topicURP, topicUnderMinISR, topicOffline, topicLeaderSkew, topicReplicaLag}, "检查受影响分区的 leader、ISR、磁盘与 broker 副本状态。"),
			Limitations: limitations,
		})
	}

	if isProblem(clientProduce) || isProblem(clientCommit) || isProblem(clientEndToEnd) {
		causes = append(causes, correlatedCause{
			Priority:   80,
			Confidence: "业务链路主因",
			Summary:    "客户端端到端探针失败，说明问题已经影响真实的生产、消费或位点提交链路。",
			Actions:    mergeActionsWithFallback([]model.CheckResult{clientProduce, clientCommit, clientEndToEnd}, "先看失败阶段，再结合网络、KRaft、Topic 和 ISR 检查一起定位。"),
		})
	}

	if isProblem(consumerLag) || isProblem(consumerRebalance) || isProblem(consumerCoordinator) {
		causes = append(causes, correlatedCause{
			Priority:   78,
			Confidence: "中高置信度主因",
			Summary:    "消费组堆积、rebalance 或 coordinator 视图异常，说明问题已经进入真实消费链路，而不只是探针层面的瞬时抖动。",
			Actions:    mergeActionsWithFallback([]model.CheckResult{consumerLag, consumerRebalance, consumerCoordinator}, "优先核对消费组 lag、成员稳定性、coordinator 与 __consumer_offsets 健康状态。"),
			Limitations: []string{
				"若消费组本身是新建或空闲状态，部分 missing offsets 需要结合业务预期进一步确认",
			},
		})
	}

	if isProblem(producerAcks) || isProblem(producerIdempotence) || isProblem(producerTimeout) || isProblem(producerMessageSize) || isProblem(producerThrottle) || isProblem(producerTxnTimeout) {
		causes = append(causes, correlatedCause{
			Priority:   77,
			Confidence: "配置侧主因",
			Summary:    "producer 参数组合本身存在一致性、幂等或超时风险，即使集群健康也可能出现重复、乱序或事务初始化失败。",
			Actions:    mergeActionsWithFallback([]model.CheckResult{producerAcks, producerIdempotence, producerTimeout, producerMessageSize, producerThrottle, producerTxnTimeout}, "优先核对 producer acks、幂等、消息大小、限流与超时组合。"),
		})
	}

	if isProblem(transactionTopicRequired) || isProblem(transactionIsolation) || isProblem(transactionOutcome) {
		causes = append(causes, correlatedCause{
			Priority:   76,
			Confidence: "事务链路主因",
			Summary:    "事务主题或 read_committed 前提不满足，事务型生产/消费链路目前不可信。",
			Actions:    mergeActionsWithFallback([]model.CheckResult{transactionTopicRequired, transactionIsolation, transactionOutcome}, "优先修复 __transaction_state 与事务 coordinator，再继续验证事务读写链路。"),
		})
	}

	if isProblem(securityListener) || isProblem(securitySASL) || isProblem(securityTLS) || isProblem(securityAuthorization) || isProblem(securityAuthorizer) {
		causes = append(causes, correlatedCause{
			Priority:   76,
			Confidence: "中高置信度主因",
			Summary:    "安全协议、SASL 机制或 Authorizer 配置与当前执行视角不一致，这类问题会表现为 TCP 可达但 metadata、认证或 ACL 行为异常。",
			Actions:    mergeActionsWithFallback([]model.CheckResult{securityListener, securitySASL, securityTLS, securityAuthorization, securityAuthorizer}, "优先核对 listener 安全协议、SASL/TLS 证书与 Authorizer 配置是否和当前客户端视角一致。"),
			Limitations: []string{
				"当前这批判断主要来自 compose 静态配置，后续仍需结合真实认证与授权探针继续验证",
			},
		})
	}

	if isProblem(storageCapacity) || isProblem(storageLayout) || isProblem(storageMounts) || isProblem(storageTiered) || isProblem(hostCapacity) || isProblem(dockerMountRuntime) {
		causes = append(causes, correlatedCause{
			Priority:   74,
			Confidence: "环境侧主因",
			Summary:    "Kafka 数据目录、KRaft metadata 目录或 volume 挂载规划存在风险，这类问题容易引发启动异常、metadata 目录混写或数据持久化失真。",
			Actions:    mergeActionsWithFallback([]model.CheckResult{storageCapacity, storageLayout, storageMounts, storageTiered}, "优先核对磁盘容量、log.dirs、metadata.log.dir 与宿主机挂载规划。"),
		})
	}

	if isProblem(metricURP) || isProblem(metricMinISR) {
		causes = append(causes, correlatedCause{
			Priority:   73,
			Confidence: "中高置信度主因",
			Summary:    "JMX 已经观测到副本不足或 UnderMinISR 压力，说明问题不只是静态配置层，而是正在影响集群复制与写入安全边界。",
			Actions:    mergeActionsWithFallback([]model.CheckResult{metricURP, metricMinISR}, "优先核对 ISR、follower broker、复制链路与当前写入压力。"),
		})
	}

	if isProblem(metricOfflineLogDir) {
		causes = append(causes, correlatedCause{
			Priority:   72,
			Confidence: "高置信度主因",
			Summary:    "JMX 检测到 OfflineLogDirectoryCount，大概率已经出现目录、挂载、权限或磁盘层面的真实故障。",
			Actions:    firstActions(metricOfflineLogDir, "优先检查离线目录对应的挂载、权限、磁盘状态与 broker 日志。"),
		})
	}

	if isProblem(kafkaMetadataLatency) || isProblem(jvmNetworkIdle) || isProblem(jvmRequestIdle) || isProblem(jvmRequestPressure) || isProblem(jvmHeapGC) || isProblem(metricNetworkIdle) || isProblem(metricRequestIdle) || isProblem(quotaBackpressure) {
		causes = append(causes, correlatedCause{
			Priority:   69,
			Confidence: "运行态证据",
			Summary:    "控制面或 broker 线程池已出现运行态压力，metadata 变慢、请求堆积和复制波动可能互相放大。",
			Actions:    mergeActionsWithFallback([]model.CheckResult{kafkaMetadataLatency, jvmNetworkIdle, jvmRequestIdle, jvmRequestPressure, jvmHeapGC, quotaBackpressure}, "结合线程池 idle、请求延迟、GC 与 metadata 延迟一起定位 broker 压力。"),
		})
	}

	if isProblem(upgradeVersion) || isProblem(upgradeFeature) {
		causes = append(causes, correlatedCause{
			Priority:   60,
			Confidence: "环境背景证据",
			Summary:    "环境可能处于 rolling upgrade 或 feature/finalization 未收口状态，这会放大 metadata、controller 和协议兼容性问题。",
			Actions:    mergeActionsWithFallback([]model.CheckResult{upgradeVersion, upgradeFeature}, "优先确认升级窗口是否已完成，以及版本/feature 配置是否一致。"),
		})
	}

	if isProblem(quotaProduce) || isProblem(quotaFetch) || isProblem(quotaRequest) || isProblem(quotaBackpressure) || isProblem(producerThrottle) {
		causes = append(causes, correlatedCause{
			Priority:   67,
			Confidence: "运行态证据",
			Summary:    "当前问题可能已经被 quota 或 broker 背压机制影响，需要区分限流、线程池压力与真正的 broker 故障。",
			Actions:    mergeActionsWithFallback([]model.CheckResult{quotaProduce, quotaFetch, quotaRequest, quotaBackpressure, producerThrottle}, "优先核对 produce/fetch throttle、request percentage quota 与请求背压指标。"),
		})
	}

	if isProblem(jvmNetworkIdle) || isProblem(jvmRequestIdle) || isProblem(jvmRequestPressure) || isProblem(jvmHeapGC) || isProblem(metricNetworkIdle) || isProblem(metricRequestIdle) {
		causes = append(causes, correlatedCause{
			Priority:   68,
			Confidence: "运行态证据",
			Summary:    "JVM/JMX 指标显示 broker 线程池空闲率已经偏低，当前问题可能和实时负载、请求堆积或网络处理瓶颈有关。",
			Actions:    mergeActionsWithFallback([]model.CheckResult{jvmNetworkIdle, jvmRequestIdle, jvmRequestPressure, jvmHeapGC}, "结合请求延迟、purgatory、GC 和流量高峰继续排查 broker 压力。"),
			Limitations: []string{
				"线程空闲率更适合做运行态压力判断，仍需结合业务流量与请求延迟一起解释",
			},
		})
	}

	if isProblem(index["HOST-004"]) || isProblem(index["HOST-006"]) || isProblem(hostCapacity) || isProblem(hostFD) || isProblem(hostClock) || isProblem(hostListener) || isProblem(hostMemory) || isProblem(index["DKR-002"]) || isProblem(index["DKR-003"]) || isProblem(index["DKR-004"]) || isProblem(dockerMountRuntime) || isProblem(index["DKR-006"]) || isProblem(index["DKR-007"]) {
		causes = append(causes, correlatedCause{
			Priority:   70,
			Confidence: "环境侧主因",
			Summary:    "宿主机或 Docker 层存在异常，问题可能并非 Kafka 配置本身，而是资源、端口或挂载环境失真。",
			Actions:    mergeActionsWithFallback([]model.CheckResult{index["HOST-004"], index["HOST-006"], hostFD, hostClock, hostListener, hostMemory, index["DKR-002"], index["DKR-003"], index["DKR-004"], index["DKR-006"], index["DKR-007"]}, "优先排查宿主机资源、时钟、fd、端口监听与数据目录挂载。"),
		})
	}

	if isProblem(logs) || isProblem(index["LOG-005"]) || isProblem(index["LOG-006"]) || isProblem(index["LOG-007"]) {
		causes = append(causes, correlatedCause{
			Priority:   65,
			Confidence: "辅助证据",
			Summary:    "近期日志已经命中关键错误指纹，可把日志信号作为当前根因判断的加权证据。",
			Actions:    mergeActionsWithFallback([]model.CheckResult{logs, index["LOG-005"], index["LOG-006"], index["LOG-007"]}, "结合命中的日志指纹和解释结果，优先沿对应故障链路继续深挖。"),
			Limitations: []string{
				"日志指纹更适合做辅助归因，仍需结合实时探针、metadata 和副本状态交叉验证",
			},
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
