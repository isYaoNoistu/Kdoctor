我这次按**公开仓库 main 分支**重新评估后的结论是：

**这个版本更接近 v1-rc，不建议现在就把 GitHub 上这个 main 直接定成“正式 v1”。**
原因不是检查项不够，而是还有几个**闭环级**问题没收口；另外，你贴出来的本地输出，和公开仓库当前 `diagnose` 实现存在明显不一致，我没法把“你本地跑出来的效果”直接等同于“仓库当前可复现状态”。仓库 README/设计文档已经把 V1 范围定义得很清楚，但公开 main 现在仍然只有 2 次提交、没有 release，而且 `rootcause.go` 还是非常薄的实现，`incident.go` 仍是占位。([GitHub][1])

先说你这份样例输出本身暴露了什么。
从你贴的结果看，**网络、bootstrap、metadata、broker 注册、controller、leader、ISR 都是通的**，所以这不是“集群整体挂了”，更像是**探针主题未就绪 / 内部 offset 主题未就绪**，导致真实链路探针在“生产阶段”就断掉了。这个判断是对的，至少方向没有跑偏。真正的问题在于：工具现在把一个上游失败，展开成了 4 个客户端 FAIL，噪声还是偏大。

我建议你把正式 v1 前的修复，分成下面这几项。

---

## 必修 P0：不改不建议定版

### 1）把 probe 的“级联失败”改成“阶段感知”

你当前 `probe.Run()` 是**在 metadata / produce / consume / commit 任一阶段失败后直接 return**；但 `CLI-003/004/005` 这些 checker 又只是简单看 `ConsumeOK/CommitOK/...` 是否为 false，于是当 **produce 失败** 时，`consumer`、`commit`、`e2e` 也会一起报 FAIL。你样例里那 4 个客户端失败，本质上就是这个逻辑导致的。([GitHub][2])

这个要改成：

* `metadata` 失败 → `CLI-002/003/004/005` 应该大多 `SKIP`，并写明“上游 metadata 已失败，后续未执行”
* `produce` 失败 → `CLI-003/004/005` 不该继续算独立 FAIL，而应 `SKIP` 或“继发失败”
* `consume` 失败 → `CLI-004/005` 同理
* `commit` 失败 → 只保留 `CLI-004/005` 失败

这样报告会立刻干净很多，主因也更稳定。

### 2）补 probe 主题生命周期，不然 fresh cluster 会被误判

默认配置里 probe 主题就是 `_kdoctor_probe`，而当前探针路径只是直接往这个 topic 发消息；代码里没有“预检查 topic 是否存在”、没有“必要时创建”、也没有 cleanup 分支。更关键的是，`ProbeConfig` 里已经暴露了 `cleanup` 和 `produce_count` 字段，但公开 probe 执行路径里没有真正消费这些配置。([GitHub][3])

正式 v1 这里至少二选一：

* **方案 A：实现 `ensureProbeTopic`**

  * topic 不存在时，先做 admin create
  * 支持可配置 partition / replication-factor
  * `cleanup=true` 时允许删除或清理探针 topic
* **方案 B：明确把 probe topic 变成前置约束**

  * 在 probe 开始前专门做 `PRE-001 probe topic readiness`
  * topic 不存在时，直接输出“环境未准备好”，不要把它包装成 Kafka 业务链路故障

站在工具可用性上，我更建议 A；站在副作用控制上，至少要把 B 做完整。

### 3）`KFK-004` 不能无脑把 `__consumer_offsets` 缺失判成硬 FAIL

现在 `internal_topics.go` 里只要找不到 `__consumer_offsets`，就直接 `FAIL`。这对**长期运行的生产集群**通常是合理的，但对**新建/几乎没使用过的集群**，这个判断太硬。Kafka 的 consumer group / offsets 状态本来就和“首次消费组活动、首次 offset 提交”有关，历史上也有客户端侧观察到首次 consumer-group 管理请求会触发 `__consumer_offsets` 的创建。你这个样例就是“本地刚起、没怎么用过的 Kafka”，这里更像是“未初始化”或“未触发”，不应直接当成高优先级集群故障。([GitHub][4])

这个检查建议改成三段式：

* **commit probe 已执行且失败，同时 `__consumer_offsets` 缺失** → `FAIL`
* **集群是 fresh / low-usage / 首次探测场景，且还没形成有效 consumer-group 活动** → `WARN`
* **明确要求 consumer group 可用，但 offsets 仍缺失** → `CRIT/FAIL`

也就是把 `KFK-004` 从“绝对规则”改成“上下文规则”。

### 4）把 Markdown 输出真正接到 CLI，不然你自己的 V1 标准还没达成

你的设计文档明确写了 V1 支持 **终端 / JSON / Markdown** 三种输出；仓库里也确实已经有 `internal/output/markdown/renderer.go`，而且目录里还有 `renderer_test.go`。但当前 CLI 只有 `--json`，`app.render()` 也只在 **JSON 和 terminal** 之间二选一，Markdown 根本没有正式出口。([GitHub][5])

这项是标准意义上的 **“文档承诺未兑现”**。
正式 v1 建议直接改成：

* `--format terminal|json|markdown`
* `--output report.md`
* 默认 terminal

不要再用单独 `--json` 这种会把格式扩展卡死的参数。

### 5）公开仓库里的 diagnose 层，仍然不够格叫“正式 v1”

你自己的文档对 V1 写得很明确：**不能只列问题，至少要做基本关联归因**。但当前公开 main 的 `rootcause.go` 仍然只是：

* Overview = `"highest status is ..."`
* RootCauses = 取前 3 个 problem summary
* Actions = 抽前几个 next actions

而 `incident.go` 还是一句占位注释。([GitHub][5])

这也是为什么我说：**你贴出来的本地输出，和 GitHub 当前代码不一致。**
因为你样例里的“高优先级主因 / 业务链路主因 / 动作顺序”明显不是这份公开 `rootcause.go` 能生成出来的。

所以正式 v1 前你必须做一件事：
**确认“正式服务版本”到底是本地未推送代码，还是 GitHub main。**

如果 GitHub main 是准备用来发版的源码，那就必须把 diagnose 补到至少这种程度：

* `CLI-002` produce fail + `KFK-004` offsets missing → 主因优先落到“probe topic/internal topic readiness”
* `NET-003` fail + `KFK-003` endpoint anomaly → 主因归到 `advertised.listeners`
* `KRF-*` 异常优先压过下游 topic/probe 症状
* 同源问题只给一个主因，不要 4 个重复症状

### 6）调度器还太薄，不符合你架构文档里对 degrade/timeout 的要求

你架构文档要求 runner 要支持“数据源降级”和整体 timeout 控制；但当前 `scheduler.go` 只是一个 `WaitGroup` 并发器，没有 task 级 timeout、没有 stage 级 cancel、没有 fail-open / fail-soft 分类。([GitHub][6])

这个问题现在未必立刻炸，但它会直接影响两个场景：

* Docker / logs / host 某一段采集变慢时，整体体验变差
* 你后面加更多 collector/checker 后，故障现场输出不稳定

正式 v1 不一定要做复杂调度框架，但至少要补：

* 每个 collector/checker 的超时上限
* 超时后的 `SKIP/WARN/TIMEOUT` 归类
* 报告里标明“哪些阶段没完成，为什么没完成”

### 7）要做“真正的发版动作”，不能只停留在 main

当前仓库公开页显示只有 **2 commits**，且 **No releases published**。如果你要把它当正式 v1 服务版本，这一步不能省。([GitHub][1])

至少要补这几样：

* Git tag：`v1.0.0`
* Release note：范围、已知限制、输入模式、退出码
* 4 份样例报告：

  * healthy
  * probe topic missing
  * `__consumer_offsets` missing
  * advertised.listeners mismatch

---

## P1：可以留到 v1.1，但最好尽快做

### 1）把 capacity 目录真正接进主流程

`internal/checks` 里已经有 `capacity` 目录，但当前 `runChecks()` 列表里没有把 capacity checker 接进来。说明你已经开始考虑容量/资源面，但还没进主干。([GitHub][7])

这不挡 v1 发版，但很适合放到 v1.1：

* 磁盘水位
* FD / inode
* OOM / restart 关联
* page cache / IO wait 风险提示

### 2）把“无 profile 时的文案”做得更像人话

你样例里 `KFK-002` 的证据是“期望=0 实际=3”。从逻辑上没错，因为 `registration.go` 就是这么输出的；但从产品表达上，这不够成熟。([GitHub][8])

更好的写法应该是：

* 未配置期望 broker 数，当前已发现 3 个 broker

而不是把 `0` 当成一个真的期望值打印出来。

### 3）补 diagnose / runner / terminal 的测试空白

目前仓库里能看到一些测试资产，比如：

* `config_test.go`
* `exitcode_test.go`
* `markdown/renderer_test.go`
* `kafka/endpoint_test.go`
* `kafka/internal_topics_test.go`

但 `diagnose` 目录没有测试文件，`runner` 目录也没有测试文件，`terminal` 目录同样看不到对应测试。([GitHub][9])

这块不一定挡 v1，但会直接影响你以后改归因逻辑和输出模板时的稳定性。

---

## 我对这版的最终定性

如果**以公开 GitHub main 为准**，我的定性是：

**现在是 v1-rc，离正式 v1 还差 5 个核心修复：**

1. probe 失败阶段要做下游 `SKIP` 抑制
2. probe topic 生命周期要补齐
3. `KFK-004` 要从绝对 FAIL 改成上下文判断
4. Markdown 输出要真正接入 CLI
5. diagnose / incident 层要和你文档承诺一致

这 5 个修完，再打 `v1.0.0` tag，我认为就可以作为**正式第一版**投入使用。
如果不修，我不建议你把当前公开仓库直接当“正式服务版本”。

---

## 你这份本地样例，我给出的直接判断

针对你贴的这次本地探测，我会给这样的运维结论：

**这台 Kafka 不是“基础通信故障”，而是“探针环境未准备好”。”**
优先顺序应该是：

1. 先确认 `_kdoctor_probe` 是否存在，或是否允许自动创建
2. 再确认 `__consumer_offsets` 是否会在首次 group/commit 后自动建立
3. 如果这两个补齐后 probe 仍失败，再去看 controller / internal topic / coordinator

也就是说，这份样例更像是在帮你暴露 **kdoctor 目前的产品边界**，而不是在证明 Kafka 集群已经严重异常。

你要是愿意，我下一条可以直接给你一份**“v1.0.0 最终修复清单”**，按 GitHub Issue 形式写成：标题、修改文件、修改目标、验收标准。

[1]: https://github.com/isYaoNoistu/kdoctor "GitHub - isYaoNoistu/kdoctor · GitHub"
[2]: https://github.com/isYaoNoistu/Kdoctor/raw/refs/heads/main/internal/probe/e2e.go "raw.githubusercontent.com"
[3]: https://github.com/isYaoNoistu/kdoctor/raw/refs/heads/main/internal/config/defaults.go "raw.githubusercontent.com"
[4]: https://github.com/isYaoNoistu/kdoctor/raw/refs/heads/main/internal/checks/kafka/internal_topics.go "raw.githubusercontent.com"
[5]: https://raw.githubusercontent.com/isYaoNoistu/kdoctor/main/doc.md "raw.githubusercontent.com"
[6]: https://raw.githubusercontent.com/isYaoNoistu/kdoctor/main/architecture.md "raw.githubusercontent.com"
[7]: https://github.com/isYaoNoistu/kdoctor/tree/main/internal/checks "kdoctor/internal/checks at main · isYaoNoistu/kdoctor · GitHub"
[8]: https://github.com/isYaoNoistu/kdoctor/blob/main/internal/checks/kafka/registration.go "Kdoctor/internal/checks/kafka/registration.go at main · isYaoNoistu/Kdoctor · GitHub"
[9]: https://github.com/isYaoNoistu/kdoctor/tree/main/internal/config "kdoctor/internal/config at main · isYaoNoistu/kdoctor · GitHub"


# 修复方向：
下面这份就是我建议你直接落到仓库里的 **`v1.0.0 最终修复清单`**。
基线判断是：仓库已经把 V1 能力范围、输出要求和验收标准写出来了，但当前公开 main 仍只有 2 次提交、还没有 release；同时 probe、diagnose、output、runner 这四层还有几个关键闭环没收口，所以更像 **v1.0.0-rc**，不是我会直接盖章的正式版。([GitHub][1])

## 1. 修复 probe 级联失败噪声

**标题**：`probe: make downstream checks stage-aware and avoid cascade FAILs`

**目标**：
当 `metadata / produce / consume / commit` 某一阶段失败时，后续未执行阶段不要继续报独立 FAIL，而应标记为 `SKIP` 或“继发失败”。当前 `probe.Run()` 在 metadata、produce、consume、commit 任一阶段报错时都会直接 `return`，这会让下游检查拿到一堆布尔 false，从而产生噪声。([GitHub][2])

**建议修改文件**：

* `internal/probe/e2e.go`
* `internal/checks/client/*.go`（如果 `CLI-002~005` 分散在多个 checker 里）

**必须做到**：

* `metadata` 失败：`produce/consume/commit/e2e` 不应继续算独立 FAIL
* `produce` 失败：`consume/commit/e2e` 统一降为 `SKIP` 或“未执行”
* 报告里增加字段：`executed_stage`、`failure_stage`、`downstream_skipped_reason`

**验收标准**：

* topic 不存在时，报告中最多 1~2 个核心 FAIL，不再出现 4 个客户端重复失败
* terminal / JSON / markdown 三种输出都能看出失败发生在哪个阶段

---

## 2. 补 probe topic 就绪与生命周期管理

**标题**：`probe: add topic readiness / optional create / cleanup flow`

**目标**：
当前默认 probe topic 是 `_kdoctor_probe`，配置里也有 `ProduceCount` 和 `Cleanup`，但公开实现里 `probe.Run()` 只直接用 `env.ProbeTopic` 执行 produce/consume/commit，代码中看不到对 `ProduceCount` 或 `Cleanup` 的实际消费，也没有 topic 预检查、创建或清理逻辑。这个会让“新集群、低使用集群、自动建 topic 未开启”的场景被放大成链路故障。([GitHub][3])

**建议修改文件**：

* `internal/probe/e2e.go`
* `internal/transport/kafka/*`
* `internal/config/*`

**二选一方案**：

* **推荐方案 A**：实现 `ensureProbeTopic()`

  * topic 不存在时可选自动创建
  * 支持 partitions / replication-factor / retention 的最小配置
  * `cleanup=true` 时在探测后删除或清理 topic
* **保守方案 B**：不自动创建，但在 probe 前增加 `PRE-001 probe topic readiness`

  * topic 不存在时直接给出“环境未准备好”，不要进入 produce fail

**验收标准**：

* fresh Kafka 上，至少能明确区分“探针环境未准备好”和“真实链路故障”
* `ProduceCount`、`Cleanup` 不再是死配置
* 自动建 topic 关闭时，报告文案能明确指出前置条件不足

---

## 3. 把 `KFK-004` 从绝对 FAIL 改成上下文判断

**标题**：`kafka: downgrade __consumer_offsets missing from absolute FAIL to context-aware verdict`

**目标**：
当前 `InternalTopicsChecker` 只要找不到 `__consumer_offsets`，就直接返回 `FAIL` 并附带下一步动作。这个规则在成熟生产集群上通常成立，但对“刚启动、几乎没使用、还没形成有效 consumer-group/commit 行为”的 Kafka 太硬。你自己的 V1 文档也强调了：视角不足时要优先 `SKIP`，避免明显误报。([GitHub][4])

**建议修改文件**：

* `internal/checks/kafka/internal_topics.go`
* 关联 `CLI-004 / CLI-005` 归因逻辑

**建议判定规则**：

* `commit probe` 已经执行且失败，同时 `__consumer_offsets` 缺失 → `FAIL`
* 首次探测 / fresh cluster / 低使用上下文，且未形成真实 commit 行为 → `WARN`
* profile 或命令显式声明“要求 group/commit 正常” → `FAIL`

**验收标准**：

* fresh cluster 不再被直接打成高优先级 Kafka 内部主题故障
* 只有在真实消费组位点链路受影响时，`KFK-004` 才进入高优先级主因

---

## 4. 真正接通 Markdown 输出

**标题**：`output: add --format and wire markdown renderer into CLI`

**目标**：
设计文档明确写了 V1 支持三种输出：终端文本、JSON、Markdown；仓库里也已经存在 `internal/output/markdown/renderer.go`。但 `app.render()` 目前只在 `JSONOutput=true` 时走 JSON，否则一律走 terminal，Markdown 渲染器并没有接入 CLI。这个属于“文档承诺已写，产品入口未落地”。([GitHub][5])

**建议修改文件**：

* `cmd/kdoctor/*`
* `internal/app/app.go`

**建议改法**：

* 废弃单独 `--json`
* 改为统一参数：`--format terminal|json|markdown`
* 保留 `--output <path>` 输出到文件
* terminal 仍为默认值

**验收标准**：

* `kdoctor probe --format markdown --output report.md` 可直接生成报告
* JSON / terminal / markdown 输出字段语义一致
* README 与 doc.md 的用法同步更新

---

## 5. 把 diagnose 从“摘前三条”升级成最小可用归因层

**标题**：`diagnose: implement correlation-based root cause ranking`

**目标**：
你的设计文档对 V1 的要求很明确：不能只停留在“列出问题”，要做基本关联、给出 1~3 个最高优先级主因、并输出动作顺序。可当前 `RootCause.Diagnose()` 只是把 `Overview` 写成 “highest status is …”，然后遍历问题，把前三个 problem summary 和前三个 next action 塞进 summary。这个实现明显低于你在文档里给自己的标准。([GitHub][5])

**建议修改文件**：

* `internal/diagnose/rootcause.go`
* `internal/diagnose/incident.go`（如果你准备把 incident 也收口）

**最低限度要实现的关联**：

* `NET-003 + KFK-003` → 优先归因 `advertised.listeners / endpoint 暴露问题`
* `KFK-004 + CLI-004/005` → 优先归因内部主题 / coordinator
* `KRF-*` 异常 → 压过下游 topic/probe 症状
* `produce fail` 时，不再把 `consume/commit/e2e` 当四个独立主因

**验收标准**：

* 同一主因不会在 summary 里重复出现 3 次不同表述
* 输出里真正体现“主因判断”和“建议动作顺序”
* 你的本地样例结果，能由公开仓库当前源码稳定复现

---

## 6. 给 runner 增加最小超时与降级语义

**标题**：`runner: add per-stage timeout / timeout reason / degrade semantics`

**目标**：
架构文档写的是要支持数据源降级和现场可用性，但当前 `internal/runner/scheduler.go` 只是一个 `parallel(tasks ...func())` 的 `WaitGroup` 包装，没有 task 级 timeout、没有 fail-soft 分类、没有明确的 timeout reason。这个实现对现在的小规模检查还能用，但对正式版不够稳。([GitHub][6])

**建议修改文件**：

* `internal/runner/scheduler.go`
* `internal/runner/*`
* 必要时补 `pkg/model` 里的状态字段

**建议改法**：

* 每个 collector/checker 都有自己的 timeout
* 超时后统一落成 `TIMEOUT` / `SKIP` / `ERROR` 中之一
* 报告里显示：哪个阶段没完成、为什么没完成、对结论影响多大

**验收标准**：

* 某个增强采集源变慢时，整体报告仍能稳定输出
* 用户能区分“真的失败”和“因为上下文/超时未完成”

---

## 7. 增加金标样例和回归测试

**标题**：`tests: add golden reports and scenario fixtures for v1`

**目标**：
V1 文档要求“基本可用可信”，这意味着你不能只靠人工跑几个现场命令。现在最需要的是把几类经典场景固定成夹具和金标报告，这样你后续改 diagnose、probe、output 时不会把旧行为改坏。这个是我建议你在正式发版前一定补的工程化动作。依据来自你文档里对 probe 主流程、误报控制、JSON/Markdown 输出和基本可信度的要求。([GitHub][5])

**建议测试场景**：

* healthy cluster
* probe topic missing
* `__consumer_offsets` missing on fresh cluster
* `advertised.listeners` mismatch
* controller / quorum 异常
* metadata 正常但 consume/commit 失败

**验收标准**：

* `go test ./...` 覆盖这些典型场景
* terminal / json / markdown 都有金标样例
* summary 主因和 recommended actions 有回归校验

---

## 8. 做真正的 `v1.0.0` 发版动作

**标题**：`release: cut v1.0.0 tag and publish release notes`

**目标**：
当前公开仓库显示 main 只有 2 次提交，而且还没有任何 release。既然你要把它定义成正式版本，就应该让“正式”体现在仓库发布动作上，而不是只停留在口头上。([GitHub][1])

**建议补齐**：

* `v1.0.0` git tag
* GitHub Release
* release note 至少包含：

  * 支持的模式
  * 输入边界
  * 已知限制
  * 退出码说明
  * 三份示例报告链接

**验收标准**：

* 别人进入仓库，不看聊天上下文，也知道哪个版本是正式版
* 版本说明能和 README / doc.md 对上

---

## 最终执行顺序

我建议你按这个顺序推进：

1. **Issue 1：probe 级联失败抑制**
2. **Issue 2：probe topic 就绪与生命周期**
3. **Issue 3：`KFK-004` 上下文化**
4. **Issue 5：diagnose 归因层升级**
5. **Issue 4：Markdown 输出接 CLI**
6. **Issue 6：runner 超时/降级语义**
7. **Issue 7：金标场景测试**
8. **Issue 8：打 tag 发 release**

前 1~6 修完，我认为就可以叫 **代码层面的 v1.0.0**。
7~8 做完，才是 **仓库层面的正式 v1.0.0**。这个顺序最符合你当前仓库状态和文档承诺。([GitHub][5])

你要的话，我下一条直接把这 8 条改写成 **GitHub Issue 模板格式**，每条带：背景、现状、改动点、验收标准、优先级。

[1]: https://github.com/isYaoNoistu/kdoctor "GitHub - isYaoNoistu/kdoctor · GitHub"
[2]: https://github.com/isYaoNoistu/kdoctor/blob/main/internal/probe/e2e.go "Kdoctor/internal/probe/e2e.go at main · isYaoNoistu/Kdoctor · GitHub"
[3]: https://github.com/isYaoNoistu/kdoctor/blob/main/internal/config/defaults.go "Kdoctor/internal/config/defaults.go at main · isYaoNoistu/Kdoctor · GitHub"
[4]: https://github.com/isYaoNoistu/kdoctor/blob/main/internal/checks/kafka/internal_topics.go "Kdoctor/internal/checks/kafka/internal_topics.go at main · isYaoNoistu/Kdoctor · GitHub"
[5]: https://raw.githubusercontent.com/isYaoNoistu/kdoctor/main/doc.md "raw.githubusercontent.com"
[6]: https://github.com/isYaoNoistu/kdoctor/blob/main/internal/runner/scheduler.go "kdoctor/internal/runner/scheduler.go at main · isYaoNoistu/kdoctor · GitHub"
