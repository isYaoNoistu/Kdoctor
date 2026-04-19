封版前最终优化文档
本轮审查基于 GitHub 上 Kdoctor 仓库的 README、设计/架构文档、关键源码，以及你提供的两份真实运行输出完成。总体判断是：Kdoctor 已经具备内部 V2 封版条件，但上线前仍有几类“可信度与一致性”问题必须收口，否则工具会在默认输出层面继续制造噪声，削弱值班人员对结果的信任。
 
 

从代码结构看，当前版本的优点已经比较明确：探针链路已经做到了分阶段执行、上游失败时给下游留出跳过提示；调度层已有任务级 timeout/soft degrade；宿主机路径解析对 volume 根目录做了安全收敛；日志采集也已经限制了文件数量和读取窗口。这些都说明它不是 demo，而是可用于生产现场首轮排障的工具。

但当前阻断封版的核心并不是“功能还不够多”，而是这四类问题：采集覆盖状态不够真实、JMX 路径没有彻底退出默认报告、输出层与 README 的默认行为不一致、部分检查项的证据组织方式还会误导人。你给出的最新 probe 输出里，摘要写着“宿主机=已采集、日志=已采集”，但明细里 HOST/LOG 大量 SKIP；同时默认终端输出仍展开了大量 PASS/SKIP，并且还出现了 JMX/Quota 相关的 SKIP 与乱码模块名，这与 README 中“封版已移除 JMX、默认折叠 PASS/SKIP”的表述不一致。
 

本轮我建议的执行口径很简单：只修现有行为，不扩新功能；P0 修完即封版；P1 补可信度；P2 只做维护性收尾。 下面这份文档可以直接拆成 GitHub Issues 和 PR。

本轮我将执行的具体验证步骤如下：

静态审查：核对 README、doc.md、architecture.md 与 internal/ 关键实现是否一致。
单元测试：执行 go test ./...，优先关注 checks、config、renderer、rootcause 相关测试。README 也明确把 go test ./... 作为构建前提。
集成/端到端 probe 运行：验证 metadata → topic-ready → produce → consume → commit → e2e 的阶段收口与超时行为。
日志采集验证：分别验证“无日志源”“Docker 日志”“文件日志目录”三类场景，检查 LOG-001~008 与采集覆盖摘要是否一致。
输出比对：分别校验 terminal / json / markdown 三种输出，与 README 声明的默认行为和 golden 报告保持一致。
按架构文档建议，本轮发布前验证流应保持下面这个顺序。

静态审查

go test ./...

quick / probe 集成验证

日志采集验证

terminal/json/markdown 比对

生成 golden 报告

打包与打 tag



显示代码
验证步骤与检查范围
你要求覆盖的模块与维度，结合仓库当前目录结构，可以按下面这张表执行。目录与文件路径来自仓库目录树和架构说明；其中少数目录树抓取结果不稳定，但模块边界本身是清晰的。

模块/维度	关键路径	本轮关注点
app	internal/app/app.go	运行时装配、格式选择、输出分发、总体 timeout
collector logs	internal/collector/logs/collector.go	日志源判定、采集窗口、指纹匹配、可信度
collector docker	internal/collector/docker/collector.go	容器存在/运行/OOMKilled/mount 证据
collector host	internal/collector/host/collector.go	磁盘、listener 端口、路径映射与 root 安全
probe e2e	internal/probe/e2e.go	topic-ready、阶段收口、cleanup、超时
checks kafka	internal/checks/kafka/cluster.go endpoint.go internal_topics.go registration.go	metadata、broker 注册、内部主题
checks kraft	internal/checks/kraft/config.go controller.go quorum.go	controller、quorum、多数派、文案
checks topic	internal/checks/topic/leader.go isr.go replica.go planning.go	leader、ISR、AtMinISR、Topic 规划
checks client	internal/checks/client/metadata.go producer.go consumer.go commit.go e2e.go	端到端链路与阶段性失败
checks logs / capacity	internal/checks/logs/* internal/checks/capacity/*	日志解释、上下文、容量边界
diagnose	internal/diagnose/rootcause.go internal/diagnose/incident.go	主因归并、覆盖摘要、建议动作
runner/scheduler	internal/runner/scheduler.go 及 runner 组装文件	任务 timeout、soft degrade、顺序稳定性
output	internal/output/terminal/renderer.go internal/output/json/renderer.go internal/output/markdown/renderer.go	默认折叠、字段一致性、证据截断
config/defaults	internal/config/config.go defaults.go runtime.go validate.go	默认值、merge 语义、配置漂移
tests	internal/checks/kafka/*_test.go internal/checks/kraft/*_test.go internal/checks/topic/replica_test.go internal/checks/client/stage_test.go internal/output/markdown/renderer_test.go internal/config/config_test.go	现有覆盖、回归基线、golden 输出

必做项
下面这几项属于封版阻断项。不修，工具仍然能跑；但拿去做内部正式版本时，值班人员会继续遇到“摘要说采到了，明细却全跳过”“README 说默认折叠，实际全展开”“README 说去掉 JMX，报告里还全是 JMX 噪声”这类问题。
 

优先级	问题描述	修改文件/函数	修改要点	验收标准	回归命令	依据
P0	采集覆盖必须按“有可用证据”展示，不能按“尝试过采集”展示	internal/collector/logs/collector.go::Collect；internal/collector/host/collector.go::Collect；internal/diagnose/incident* 覆盖摘要生成处	Collected 与 Available 语义分开；覆盖摘要只根据“有实际可用证据”显示“已采集”	无 --log-dir/--compose 时，摘要不得再显示“日志=已采集 / 宿主机=已采集”；LOG/HOST 的 SKIP 与摘要一致	./kdoctor probe --bootstrap 192.168.100.78:9192 > out.txt	日志/宿主机 collector 一开始就把 Collected 置为 true，但没有证据时直接返回；你给出的 probe 输出也确实出现了“摘要已采集，明细全跳过”。
 
P0	默认输出必须真正折叠 PASS/SKIP，并对证据去重/截断	internal/output/terminal/renderer.go::Render；internal/output/markdown/renderer.go::Render；必要时 internal/app/app.go::render	默认只展开 WARN/FAIL/CRIT/ERROR；--verbose 再展开全部；证据统一 dedupe + max_evidence_items	默认 terminal 不再出现 - [通过] / - [跳过] 明细；--verbose 时才展开；重复 endpoint 证据消失	./kdoctor probe --bootstrap 192.168.100.78:9192 > out.txt；./kdoctor probe --bootstrap 192.168.100.78:9192 --verbose > out.verbose.txt	README 明确写了默认折叠 PASS/SKIP，但渲染器公开实现仍是“循环打印全部 checks”；你给出的 probe 输出也的确全展开了 PASS/SKIP。
 
P0	JMX / Metrics / JVM / Quota 路径必须彻底退出默认运行与默认报告	internal/runner/* 检查注册处；README.md；doc.md；architecture.md；internal/localize/*	不再把 JMX 依赖检查注册进默认 quick/probe；移除遗留 quota/JVM 文案与乱码模块名	默认 quick/probe 报告中不再出现 JVM-* / MET-* / QTA-* / JMX 依赖型 SKIP；不再出现“閰嶉”乱码	`./kdoctor probe --bootstrap 192.168.100.78:9192 > out.txt && ! grep -Eq 'JVM-	MET-
P0	TOP-011 只能输出真正命中的 topic，不得把正常 topic 一起塞进告警证据	internal/checks/topic/planning.go::Run	把 failEvidence 与 warnEvidence 分开；仅输出触发条件的 topic；超过上限时输出“前 N 项 + 其余数量”	TOP-011 告警证据只保留真正命中项；正常的 3/12/50 分区 topic 不再出现在告警里的“证据”中	./kdoctor probe --bootstrap 192.168.100.78:9192 > out.txt	当前 planning.go 在循环里对所有 topic append(evidence, ...)，而不是只对 fail/warn topic 收集证据；你给出的 probe 输出也确实把大量正常 topic 混进了 TOP-011 告警证据。
 
P0	日志 collector 的性能与对外承诺要一致	internal/collector/logs/collector.go::aggregateMatches / fingerprints；internal/config/config.go；internal/config/defaults.go；README.md	指纹正则只编译一次；如果本版不做 lines/bytes/latest_ts/custom_patterns_dir，就把 README 与 LOG-006/008 的承诺删掉；不要“说有但代码没有”	go test ./... 稳定通过；probe/quick 日志检查不再因为 regex 重复编译出现不必要开销；文档/输出与实现一致	go test ./...；./kdoctor --config ./kdoctor.yaml > out.txt	aggregateMatches 在逐行扫描时反复调用 fingerprints()，而 fingerprints() 每次都 regexp.MustCompile(...)；同时 README 宣称支持每源统计与 logs.custom_patterns_dir，但 LogConfig 当前只有 enabled/log_dir/tail_lines/lookback_minutes 四个字段。

采集覆盖必须按真实可用证据展示
这项是最该优先修的，因为它直接影响摘要可信度。当前 logs collector 在函数一开始就执行 out := &snapshot.LogSnapshot{Collected: true}，host collector 也在创建快照时立刻置 Collected: true；如果后续发现没有任何可用源，它们会返回一个“已采集但不可用”的快照。只要摘要层拿 Collected 来渲染，就会出现你现在 probe 输出里的“日志=已采集、宿主机=已采集”，但明细全是 SKIP。
 

建议直接改成下面这种最小补丁思路：

diff
复制
diff --git a/internal/collector/logs/collector.go b/internal/collector/logs/collector.go
@@
- out := &snapshot.LogSnapshot{Collected: true}
+ out := &snapshot.LogSnapshot{}
@@
- if len(sourceContents) == 0 {
-     return out
- }
- out.Available = true
+ if len(sourceContents) == 0 {
+     out.Collected = len(out.Errors) > 0
+     out.Available = false
+     return out
+ }
+ out.Collected = true
+ out.Available = true
diff
复制
diff --git a/internal/collector/host/collector.go b/internal/collector/host/collector.go
@@
- out := &snapshot.HostSnapshot{Collected: true, Raw: map[string]string{}}
+ out := &snapshot.HostSnapshot{Raw: map[string]string{}}
@@
- if len(diskTargets) == 0 && !hostContext {
+ if len(diskTargets) == 0 && !hostContext {
+     out.Collected = false
      out.Errors = append(out.Errors, "host-level evidence is not available from the current input mode")
      return out
  }
@@
- out.Available = len(out.DiskUsages) > 0 || len(out.PortChecks) > 0
+ out.Collected = true
+ out.Available = len(out.DiskUsages) > 0 || len(out.PortChecks) > 0
摘要层的补丁思路也很直接：“已采集”只看 Available，不要看 Collected；Collected && !Available 应显示为“已尝试但无有效证据”或直接“无可用证据”。这不会改 CLI 语义，只是把摘要从“乐观说法”改成“证据说法”。

默认输出必须真正折叠 PASS/SKIP，并对证据截断去重
README 已经把封版默认行为写得很明确：终端默认只展开重点问题，PASS/SKIP 默认折叠，--verbose 才展开；还给出了 output.max_evidence_items=8 等默认值。
 但当前公开抓取得到的 terminal 渲染器与 markdown 渲染器都仍是简单地“遍历 report.Checks 全量输出”，没有实现默认折叠、证据上限、证据去重。

建议最小补丁如下：

diff
复制
diff --git a/internal/output/terminal/renderer.go b/internal/output/terminal/renderer.go
@@
-type Renderer struct{}
+type Renderer struct {
+    MaxEvidenceItems int
+    ShowPassChecks   bool
+    ShowSkipChecks   bool
+    Verbose          bool
+}
@@
- for _, check := range report.Checks {
+ for _, check := range report.Checks {
+     if !r.visible(check.Status) {
+         continue
+     }
      fmt.Fprintf(&buf, "- [%s] %s %s：%s\n", ...)
-     if len(check.Evidence) > 0 {
-         fmt.Fprintf(&buf, " 证据：%s\n", strings.Join(check.Evidence, " | "))
+     evidence := capAndDedupe(check.Evidence, r.MaxEvidenceItems)
+     if len(evidence) > 0 {
+         fmt.Fprintf(&buf, " 证据：%s\n", strings.Join(evidence, " | "))
      }
  }
go
复制
func (r Renderer) visible(status string) bool {
    if r.Verbose {
        return true
    }
    switch strings.ToLower(strings.TrimSpace(status)) {
    case "pass":
        return r.ShowPassChecks
    case "skip":
        return r.ShowSkipChecks
    default:
        return true
    }
}
同样的逻辑要在 Markdown 渲染器里做一次，否则 terminal/markdown 会继续不一致。渲染层做一层“通用 dedupe”也很值，因为你给出的 probe 输出已经出现了 KFK-005 / NET-003 / NET-005 的地址证据重复。

JMX / Metrics / JVM / Quota 路径必须彻底退出默认报告
如果你的封版目标已经明确为**“不启用 JMX、移除 JMX 相关检查”**，那就不要再让默认 report 出现任何 JVM-*、MET-*、QTA-*、JMX 依赖型 HOST-009、KRF-006/007 的 SKIP。否则一线看到的仍然是“我没打算启 JMX，但报告里一堆 JMX 相关噪声”。你给出的最新 probe 输出就是这个状态。

这里不建议做复杂的“隐藏规则”，而是直接在 runner 的检查注册处把这些检查从默认 quick/probe/full registry 中移除，只保留真正不依赖 JMX 的路径。这样不改 CLI 参数语义，也最稳。同步还要修两处文档漂移：一处是 README / doc.md / architecture.md；另一处是本地化映射，避免 quota 模块继续出现 閰嶉 这类乱码。README 原始版本已经在写“JMX 已从封版默认能力中移除”，所以现在真正要做的是把实现收敛到文档，而不是再去扩 JMX。

伪代码可以是这样：

go
复制
// internal/runner/<registry>.go
checks = []Checker{
    // keep
    netChecks...,
    kafkaChecks...,
    kraftStaticChecks...,
    topicChecks...,
    clientChecks...,
    dockerChecks...,
    hostNonJMXChecks...,
    logsChecks...,
    // remove from default release profile:
    // metricsChecks...,
    // jvmChecks...,
    // quotaChecks...,
    // hostJMXChecks...,
    // kraftJMXChecks...,
}
TOP-011 证据必须只列命中项
这项完全是“证据组织 bug”，而不是“规则方向 bug”。TOP-011 当前的规则方向没有大问题：副本因子超过 broker 数应 FAIL，分区数低于 broker 数是规划性 WARN。问题在于你现在的实现把所有 topic 都 append 进证据，然后只用 warnings > 0 / failures > 0 决定最终状态。结果就是：用户看到的是一大串 topic 名称，以为它们全部有问题，但其实大多数只是被顺手打印出来了。 代码本身已经能直接证明这一点。

建议直接改成：

diff
复制
diff --git a/internal/checks/topic/planning.go b/internal/checks/topic/planning.go
@@
- evidence := []string{}
+ failEvidence := []string{}
+ warnEvidence := []string{}
@@
- evidence = append(evidence, fmt.Sprintf("topic=%s partitions=%d rf=%d", ...))
- if replicationFactor > c.ExpectedBrokerCount { failures++ }
- else if partitionCount < c.ExpectedBrokerCount { warnings++ }
+ item := fmt.Sprintf("topic=%s partitions=%d rf=%d", topic.Name, partitionCount, replicationFactor)
+ if replicationFactor > c.ExpectedBrokerCount {
+     failures++
+     failEvidence = append(failEvidence, item)
+ } else if partitionCount < c.ExpectedBrokerCount {
+     warnings++
+     warnEvidence = append(warnEvidence, item)
+ }
@@
- result.Evidence = evidence
+ result.Evidence = failEvidence
@@
- result.Evidence = evidence
+ result.Evidence = warnEvidence
如果你想进一步优化值班可读性，再加一层摘要化即可，例如："共 17 个 topic 命中，展示前 8 个"。这样 TOP-011 仍然能保留，但不会再污染可读性。

日志 collector 的性能与对外承诺必须对齐
这项与工具“可信度”关系很大。当前日志 collector 已经做了两件好事：读取文件数量有限制，并且 readTail 会把单文件读取窗口限制在 512 KiB；这说明资源边界意识已经有了。
 但它还有两个明显问题。

第一，aggregateMatches 在逐行扫描时反复调用 fingerprints()，而 fingerprints() 里每次都重新 regexp.MustCompile(...)。这会把 regex 编译放到热路径里，属于非常典型、非常不值当的 CPU 浪费。
 第二，README 公开承诺了“每个日志源的行数、字节数、最新时间、新鲜度、样本充分性”和 logs.custom_patterns_dir，但当前 LogConfig 里并没有这个字段，collector 也没有计算这些统计。

如果你坚持“封版前不扩功能”，那这一项的最小正确做法不是去赶一个半成品的外部规则库，而是：

diff
复制
diff --git a/internal/collector/logs/collector.go b/internal/collector/logs/collector.go
@@
-func aggregateMatches(sourceContents map[string]string) []snapshot.LogPatternMatch {
+var builtinFingerprints = []fingerprint{
+    newFingerprint(...),
+    ...
+}
+
+func aggregateMatches(sourceContents map[string]string) []snapshot.LogPatternMatch {
@@
-    for _, fp := range fingerprints() {
+    for _, fp := range builtinFingerprints {
         ...
     }
 }
然后二选一：

方案 A：这版就不再声称支持 lines/bytes/latest_ts/custom_patterns_dir，同步删 README/LOG-006/LOG-008 的承诺。
方案 B：做最小实现，只增加每源 lines、bytes、latest_source_ts，暂时不做复杂日志行时间解析；custom_patterns_dir 如果要保留，就先把配置字段补进 LogConfig 并配两条单测。
考虑你当前的封版目标，我更推荐 方案 A + regex 预编译，把“说了但没实现”的部分删干净，反而更稳。

建议项
下面这些项不一定阻断封版，但它们会明显提升“误报控制”和“报告专业感”。其中几项甚至只改几行代码，却能显著提高你的内部口碑。

优先级	问题描述	修改文件/函数	修改要点	验收标准	回归命令	依据
P1	__transaction_state 缺失在“未使用事务”场景下不应形成重复告警	internal/checks/kafka/internal_topics.go::Run；事务检查器；internal/diagnose/rootcause.go	无 transactional.id、无 tx probe、无事务主题使用证据时，把 KFK-004 降为上下文或让 TXN-001 独占提示	非事务集群中，__transaction_state 缺失不再造成双重告警	./kdoctor --config ./kdoctor.yaml > out.txt	InternalTopicsChecker 当前只要没看到 __transaction_state 就 WARN；而 Kafka 官方文档明确写了 transactional.id 默认未配置、默认不能使用事务，事务内部主题的创建也受 transaction.state.log.replication.factor 约束。
 
 
P1	配置 merge 语义里布尔值无法显式关闭	internal/config/defaults.go	把 Enabled bool 改成 *bool 或引入显式 merge 语义，避免 `result.Enabled = result.Enabled		override.Enabled`	enabled: false 能真正禁用 logs/docker/probe
P1	调度错误顺序目前是非确定性的，不利于 golden 测试	internal/runner/scheduler.go::runTasks	在返回前对 errs 做稳定排序，或按 task 声明顺序收集错误	多次运行同一失败场景，错误顺序一致	for i in {1..5}; do ./kdoctor ...; done	runTasks 采用 goroutine 并发执行后再从 channel 汇总错误，天然受完成顺序影响。
P1	probe 主题自动创建后应做短暂 metadata/leader 就绪等待	internal/probe/e2e.go::ensureTopicReady	topic 新建成功后增加短轮询，确认 leader 可用再进入 produce	fresh cluster 首次 probe 不再偶发 “topic created but produce still not ready”	./kdoctor probe --bootstrap <fresh-cluster> 连续跑多次	现在 ensureTopicReady 成功后立即返回；它没有在新建 topic 后重新拉 metadata 或等待 leader 收敛。
P1	KRaft controller 文案需要更精确，避免把“broker listener 地址”说成 controller 端点不一致	internal/checks/kraft/controller.go 或对应本地化文本	将 KRF-002 证据文案改成“metadata 返回 controller 所属 broker 地址；不要求等于 controller.quorum.voters 的 CONTROLLER 端点”	类似你给出的 compose 场景下，值班人员不再被“地址不在集合中”误导	./kdoctor --config ./kdoctor.yaml > out.txt	你给出的运行输出里，KRF-002 为 PASS，但证据写着“活动 controller 地址不在显式 controller 端点集合中”；结合 compose 配置可知这是 combined mode 下 broker listener 与 controller listener 不同导致的正常现象。
P1	README / architecture / 实际输出仍有明显漂移，发布前必须把仓库与二进制 commit 对齐	README.md；architecture.md；版本信息注入处；golden 样例	生成二进制时嵌入 commit/version；release note 标明冻结 commit；golden 报告与该 commit 对应	二进制 --version、tag、release note、golden 报告四者一致	git rev-parse HEAD；./kdoctor --version	README raw 已写“封版已移除 JMX、默认折叠 PASS/SKIP”，但你给出的最新运行输出还不是这个行为；同时架构文档仍写“V1 已交付、开始稳定化优化”。
 

这里我特别强调一下 KFK-004。从 Kafka 官方文档看，只有配置了 transactional.id 才会启用事务，而且默认交易链路并不启用；transaction.state.log.replication.factor 不满足时，事务内部主题甚至不会创建成功。也就是说，在“没有事务使用证据”的集群里，单独把 __transaction_state missing 抬成 Kafka 层 WARN，很容易制造重复上下文噪声。当前 internal_topics.go 的实现就是这样。

同样，TOP-007 一类与 min.insync.replicas 相关的提示是有官方依据的：Kafka 官方明确说明，当 producer 使用 acks=all 且 ISR 成员数低于 min.insync.replicas 时，写入会失败；因此 AtMinISR 风险提示并不是误报方向，而是需要更好地压缩证据，避免 50 个 partition 一行行刷屏。
 

可选项
这些项不影响你现在封版，但会让后续维护舒服很多。如果你这次想做到“封版后基本不再动”，可以做；如果你只想一周内稳定上线，可以先不做。

优先级	问题描述	修改文件/函数	修改要点	验收标准	回归命令	依据
P2	给 JSON 报告加 schema_version 与 tool_version	internal/output/json/renderer.go；pkg/model	只新增字段，不删旧字段，向后兼容	jq -e '.schema_version and .tool_version' out.json	./kdoctor probe --bootstrap ... --json --output out.json	设计文档要求 JSON 结构稳定、适合自动化，但当前公开文档未给出明确 schema 版本字段。
P2	把架构文档中的 V1 文案全部收敛到 V2 封版状态	architecture.md doc.md README.md	统一为“V2 封版、JMX 已退出默认能力、以可信度优化为主”	三份文档没有 V1/V2 混用	文档 diff 审查	architecture.md 仍写“V1 已交付，开始稳定化优化”，与当前封版目标不一致。
P2	如要保留外部日志规则库承诺，就做成正式配置与测试	internal/config/config.go；internal/collector/logs/collector.go；对应 tests	把 logs.custom_patterns_dir 作为正式配置落地；否则删 README 承诺	配置文件可解析、自定义规则可命中、有测试	go test ./... + 自定义规则最小样例	README 目前宣称支持 logs.custom_patterns_dir，但 LogConfig 中并没有这个字段。
P2	补一套固定的 golden 输出仓	testdata/golden/*；renderer tests	terminal/json/markdown 各保留一份 golden	go test ./... 时自动比对	go test ./...	架构文档已要求契约测试与真实环境验证，但当前 golden 输出体系未在文档层明确。

发布与回归
发布步骤
README 已经给出了最小构建路径：仓库根目录先执行 go test ./...，再执行构建脚本；如需 Linux 交叉编译，用 PowerShell 构建脚本传 GOOS linux -GOARCH amd64。

但考虑你这次是内部 Linux 二进制正式封版，我建议把发布步骤固定成下面这一版：

bash
复制
# 进入冻结 commit
git checkout main
git pull --ff-only

# 单元与契约测试
go test ./...

# 构建 Linux amd64 静态二进制
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
go build -trimpath \
  -ldflags="-s -w -X main.version=v2.0.0 -X main.commit=$(git rev-parse --short HEAD)" \
  -o dist/kdoctor-linux-amd64 \
  ./cmd/kdoctor

# 打包
cp README.md USER_GUIDE.md kdoctor.example.yaml dist/
tar -C dist -czf dist/kdoctor-v2.0.0-linux-amd64.tar.gz \
  kdoctor-linux-amd64 README.md USER_GUIDE.md kdoctor.example.yaml

# 校验和
sha256sum dist/kdoctor-v2.0.0-linux-amd64.tar.gz > dist/kdoctor-v2.0.0-linux-amd64.tar.gz.sha256

# 打 tag
git tag -a v2.0.0 -m "Kdoctor internal V2 release"
git push origin main --tags
如果你坚持走仓库已有脚本，可以保留一条 PowerShell 路径作为备份，但正式 release note 仍建议以上面的 Linux go build 命令为准，因为这是你最终内部使用的目标产物。README 已明确仓库有 scripts/build.ps1 和 dist/ 产物目录。

建议 release note 固定成四段：

范围：只修 bug、优化输出、移除 JMX，不新增功能。
已修复：列出本文 P0 全部项目。
已知限制：事务未使用场景、compose 缺失场景、consumer group 未配置场景。
回归通过：列出下面的 6 个场景全部通过。
回归测试矩阵
下表是我建议你冻结前必须跑过的最小矩阵。它既满足架构文档对 bootstrap-only / compose / JSON / Markdown / 真实环境验证 的要求，也和你当前工具的真实使用方式一致。

场景	命令	预期断言	golden 文件建议
健康 quick 场景	./kdoctor --config ./kdoctor.yaml > out.txt	有 模式：快速巡检；无 JVM-/MET-/QTA-；默认不展开 PASS/SKIP	testdata/golden/quick_healthy.txt
健康 probe 场景	./kdoctor probe --bootstrap 192.168.100.78:9192 > out.txt	CLI-001~005 全通过；若无日志/compose，不再显示“日志=已采集”	testdata/golden/probe_healthy.txt
verbose 输出场景	./kdoctor probe --bootstrap 192.168.100.78:9192 --verbose > out.txt	明确展开 - [通过] 与 - [跳过]	testdata/golden/probe_verbose.txt
JSON 契约场景	./kdoctor probe --bootstrap 192.168.100.78:9192 --json --output out.json	jq -e '.summary.status and .checks' out.json 通过；如加 schema 版号则断言该字段存在	testdata/golden/probe_healthy.json
Markdown 留档场景	./kdoctor probe --bootstrap 192.168.100.78:9192 --format markdown --output report.md	存在 # kdoctor 检查报告、## 主因判断、## 建议动作、## 检查项	testdata/golden/probe_healthy.md
compose + docker + logs 场景	./kdoctor --config ./kdoctor.yaml > out.txt	DKR-001~003 为通过；LOG-001 通过；若你实现了日志源统计，则证据包含 source/lines/bytes/latest_ts	testdata/golden/compose_docker_logs.txt
fresh cluster 首次 probe 场景	./kdoctor probe --bootstrap <fresh-bootstrap> > out.txt	首次运行可自动创建 probe topic 并成功或给出明确 topic-ready 阶段错误；不应继续扩散成 3 条重复 fail	testdata/golden/probe_fresh_cluster.txt
非事务集群场景	./kdoctor --config ./kdoctor.yaml > out.txt	没有事务使用证据时，__transaction_state missing 不再形成重复告警	testdata/golden/quick_non_txn.txt

golden 报告样例
封版后建议直接把以下风格冻结为 golden 标准。字段名来自设计文档要求；内容上只做脱敏，不做简写。

terminal 预期骨架

text
复制
模式：链路探针
配置模板：generic-bootstrap
总体状态：告警
检查时间：2026-04-19 20:00:00+08:00
耗时：1686ms
Broker 存活：3/3
概览：...
采集覆盖：
- 网络=已采集
- Compose=未提供
- Kafka=已采集
- Docker=未提供
- 宿主机=无可用证据
- 日志=无可用证据
- 探针=已执行
主因判断：
- ...
建议动作：
- ...

检查结果：
- [告警] TOP-007 主题：...
  证据：...
JSON 预期骨架

json
复制
{
  "schema_version": "kdoctor.report.v2",
  "tool_version": "v2.0.0",
  "mode": "probe",
  "profile": "generic-bootstrap",
  "checked_at": "2026-04-19T20:00:00+08:00",
  "elapsed_ms": 1686,
  "exit_code": 1,
  "summary": {
    "status": "warn",
    "broker_total": 3,
    "broker_alive": 3,
    "overview": "...",
    "root_causes": ["..."],
    "recommended_actions": ["..."]
  },
  "coverage": {
    "network": "available",
    "compose": "missing",
    "kafka": "available",
    "docker": "missing",
    "host": "no_evidence",
    "logs": "no_evidence",
    "probe": "executed"
  },
  "checks": [
    {
      "id": "TOP-007",
      "module": "topic",
      "status": "warn",
      "summary": "...",
      "evidence": ["..."],
      "possible_causes": [],
      "next_actions": ["..."]
    }
  ]
}
输出规范与审计要点
输出格式规范
设计文档和架构文档对报告模型已经给出了最小要求：报告层至少要有 模式/配置模板/检查时间/耗时/总体状态/broker 总数与存活数/主因判断/建议动作/逐项检查结果；检查项至少要有 id/module/status/summary/evidence/possible_causes/next_actions。

封版建议把三种输出统一到下面这个规则：

输出	必须字段	默认行为	备注
terminal	mode/profile/status/checked_at/elapsed_ms/broker_alive/overview/coverage/root_causes/recommended_actions/checks	默认折叠 PASS/SKIP；证据去重并截断	面向值班排障
json	同 terminal + schema_version/tool_version/exit_code	结构稳定，字段英文，值可中文	面向自动化
markdown	与 terminal 同语义	章节固定：概览 / 主因判断 / 建议动作 / 检查项	面向留档/工单

代码审计要点
从这轮静态审查来看，Kdoctor 当前代码层面最值得肯定的地方是：probe 模块有明确的阶段边界与 cleanup 策略，scheduler 已有任务级 timeout 与 soft degrade，host 路径解析有 root 约束，logs 读取窗口也有限制。

但封版前仍建议你用下面这张审计表统一扫一遍：

维度	当前观察	风险	建议
安全	host collector 的路径映射会在 mount root 内向上回退，避免越界；日志读取限制为最多 12 个文件、单文件最多 512 KiB	风险可控	保持现有 root 约束与读取上限，不要放开
并发	scheduler 已按 task 建立 goroutine 与 timeout，但错误汇总顺序不稳定	golden 输出可能抖动	返回前排序错误，或按 task 顺序收集
timeout	runner 已支持 task 级 timeout；probe 也有独立 timeout	基本合理	auto-create topic 后再加一次短轮询，不要把 topic ready 与 produce 紧贴
资源泄露	readTail 会 defer file.Close()；probe cleanup 只在 TopicCreated && Cleanup 时做	基本合理	保持 cleanup 条件，不要对已有业务 topic 误删
错误处理	Collected/Available 语义混淆	摘要误导	用 Available 驱动采集覆盖摘要
性能	日志指纹正则在热路径重复编译	无谓 CPU 开销	预编译为包级变量，按严重级排序仍保留
本地化	架构要求 UTF-8 与中文输出，但最新 probe 输出仍有乱码模块名	影响专业观感	修正 localize/字体编码/模块映射，并加入输出回归断言

日志采集可信度提升建议
如果你只想做“封版前最小必要修复”，那我建议你把日志可信度提升分成两层处理。

第一层，封版必做：
LOG-001 的“采集成功”只表示“窗口内确实拿到了可用源”，不要再暗示“样本已经足够”；LOG-006/008 如果当前没有实现“新鲜度/样本充分性/自定义规则库”，就不要在 README 或默认输出里承诺这些能力。这样做最稳。

第二层，若你愿意再补一刀：
给每个 source 增加下面五个证据，哪怕先只做到文件级/容器级统计，也足够让 LOG 模块从“能采到”变成“值得信”：

证据项	最小实现	建议默认值
source	docker:kafka1 / file:/path/server.log	必填
lines	strings.Count(content, "\n")+1	仅窗口内统计
bytes	len([]byte(content))	仅窗口内统计
latest_source_ts	文件 modTime 或 Docker 日志返回时间；先不强求逐行解析	RFC3339
sample_enough	lines >= 20 或 bytes >= 8KB	建议先用简单阈值

指纹匹配窗口目前已有两个合理默认：tail_lines=300、lookback_minutes=15。这两个值就在默认配置里，完全可以直接冻结为封版基线。

至于“指纹库位置与可扩展方式”，考虑你这次明确要求不扩新功能，我给你的封版建议是：

本版正式基线：内置指纹库仍放在 internal/collector/logs/collector.go，作为代码内规则表维护。
对外文档口径：先删除 logs.custom_patterns_dir 的公开承诺，避免“文档有、配置没有、实现也没有”。
如果未来真的要扩：再把 custom_patterns_dir 做成正式配置项，并补 2 条测试——一条配置解析测试，一条规则命中测试。那时它才算“功能”，而不是“注释里的愿望”。
最终上线判断
综合这次静态审查与两份真实输出，我的最终判断是：Kdoctor 可以作为内部 V2 正式版本上线，但前提是本文列出的 P0 必做项全部完成并回归通过。 这些 P0 不涉及新功能扩展，都是把你已经设计好的行为做实，把 README、默认输出和真实代码拉回同一条线上。
 
 

如果按封版优先级排序，我建议你的最后一轮实施顺序是：

修采集覆盖真实度
修默认输出折叠/截断/去重
彻底移除 JMX 默认路径
修 TOP-011 证据污染
修日志 collector 指纹预编译与 README 承诺对齐
跑回归矩阵，打包、打 tag、冻结 golden
做到这一步，这个项目就可以结束“继续加功能”的阶段，进入“正式内部使用 + 只修已知缺陷”的状态。