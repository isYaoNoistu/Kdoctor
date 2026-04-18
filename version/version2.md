V2-高级 Kafka 运维工具详细设计标准与参数
内部使用 · 第二大版基线文档 · 适用于 Kdoctor V2 设计与实现
基于 Apache Kafka 官方文档、当前 Kdoctor 公开仓库代码与现有运行结果整理

文档使用方式： 第一份文档用于定义 V2 的目标能力、检查项体系、证据链与参数基线。
•	第二份文档用于审视当前仓库的真实能力边界，指出现有判定逻辑的不足，并给出具体代码级优化方向。
•	文档定位为内部运维设计基线，不面向外部发行平台。

 
1. 文档目标与定位
本文件的目标不是定义一个“发行平台”，而是定义一套内部 Kafka 高级排障工具的设计标准：在不牺牲执行速度和现场可用性的前提下，让工具能够发现并定位绝大多数常见问题，同时尽可能覆盖那些平时不常见、但一旦发生就排查成本极高的疑难问题。
V2 的定位仍然是“内部运维诊断器”，而不是统一运维门户。换句话说，最重要的目标不是炫技，而是：值班时拿起来就能用、输出对排障顺序有帮助、能把证据说清楚、能减少人肉试错。
V2 总体目标： 从“可用巡检器”升级为“高可信度 Kafka 分诊与定位器”。
•	从“看到表象”升级为“采集证据 → 规则关联 → 根因排序 → 动作顺序”。
•	从“固定少量检查项”升级为“可扩展的问题地图 + 参数基线 + 自定义规则库”。

2. 设计原则
•	证据优先：所有判断都必须能回答“依据是什么”，输出不能只有结论。
•	分层判断：必须区分入口连通、metadata 正常、leader 正常、真实生产消费链路正常，它们不是一回事。
•	误报控制优先：看不到就 `SKIP/WARN`，不要硬猜；但真正高危问题要果断升级严重级别。
•	模式驱动：至少保留 quick / probe / lint / full / incident 五种执行模式，并允许按模式切换默认参数。
•	视角感知：internal / external / host-network / docker-container / bastion 等执行视角会改变可达性与结论。
•	相关性而非堆叠：一个根因可能产生 4~6 个症状；V2 必须会做症状抑制和主因合并。
3. 当前 V1 基线（作为 V2 起点）
结合当前仓库代码与实际输出，现阶段工具已经具备以下值得保留的基础：bootstrap-only 模式、阶段化客户端 probe、自动准备 probe topic、内部主题的上下文化判断、基础 Docker/日志采集、Markdown/JSON/终端输出，以及初步的 root cause 规则关联。V2 应建立在这些能力之上，而不是推倒重来。
现有能力	现状	V2 保留/升级方向
bootstrap-only 运行	已可只给 bootstrap 执行快速巡检	继续保留，作为最小输入层
probe 阶段化	已能标出 failure stage 与 downstream skipped	升级为更强的 producer/consumer/transaction 诊断器
内部主题判断	`__consumer_offsets` 与 `__transaction_state` 已区分场景	继续结合 commit、事务和 group 状态细化
日志采集	支持 docker/file 日志并做指纹匹配	扩展为自定义规则库 + 时间线聚合
输出	terminal/json/markdown 已具备	增加证据置信度与面向值班的摘要模板

4. V2 总体架构
建议 V2 采用“多层输入 + 统一快照 + 规则引擎 + 诊断引擎 + 多格式输出”的结构。关键不是增加模块数量，而是让每个故障域都能拿到可靠证据，并让诊断层知道不同证据之间该如何关联。
层次	输入/模块	用途	V2 必须补充
输入层	bootstrap / profile / compose / docker / log-dir / host	延续 V1 最小输入与增强输入设计	新增 jmx / group describe / security / quotas / upgrade state
采集层	network / kafka / topic / logs / host / docker	形成原始证据快照	新增 metrics、consumer group、storage、security、version feature
规则层	单点检查项	把原始证据转换为 check result	新增阈值分级、视角感知、模式化参数
诊断层	root cause correlation	对重复症状做抑制和主因排序	新增 confidence、权重、图状关联、动作顺序
输出层	terminal / json / markdown	服务值班、工单、留档	新增 incident 摘要、值班摘要、证据索引

5. V2 故障地图与详细设计标准
下面这部分是本文件的核心。每个故障域都明确：为什么要看、必须采什么证据、怎么判断、应该命名成什么检查项、默认参数建议是什么。V2 实现时，应优先保证这些检查项能在 quick / full / incident 三种模式下以不同粒度运行。
5.1 接入与地址层（Bootstrap / DNS / NAT / LB / Listeners）
普通环境里最常见的 Kafka 问题，往往不是 broker 真挂了，而是 bootstrap、advertised.listeners、LB、DNS、NAT、双网卡、内外网混用导致“能连上，但后续 metadata 或 leader 路由不对”。V2 必须把“入口可达”“metadata 返回地址可达”“返回地址是否符合执行视角”拆开判断。
常见问题	必采证据	判断依据	V2 新检查 ID	建议参数
bootstrap 可达但 metadata 后续地址不可达	TCP 探测、metadata broker endpoints、执行视角标签（内网/公网）	bootstrap 成功 + metadata endpoints 局部不可达 => 非入口问题，而是 returned address 路由问题	NET-005 / KFK-005	tcp_timeout=3s；每地址至少 2 次探测
advertised.listeners 返回内网地址给公网客户端	metadata endpoints、profile.plaintext_external、profile.bootstrap_external	执行视角为 external 且 broker 仅返回 RFC1918 地址 => 高可信 listeners 暴露错误	NET-006 / CFG-009	视角标签 required；RFC1918/公网地址分类器
同一 broker 多 listener 混用、LB 只代理 bootstrap	listeners、advertised.listeners、客户端真实请求路径	bootstrap 走 LB 成功，但 leader/fetch 走 broker 直连失败 => 需要明确提示“LB 仅代理入口”	NET-007	bootstrap 与 metadata 地址分别取证
DNS 解析漂移、解析到旧 IP	DNS 解析结果、TCP 可达、证书 SAN（如启用 SSL）	解析结果多值且可达性不一致，或 A 记录与 metadata 返回地址不一致	NET-008	DNS 结果保留 TTL/多值
端口开放但协议错配	TCP 可达、Admin/Metadata 握手结果、SASL/SSL 预期	TCP success 不能等于 Kafka success；必须把 socket open 和协议握手分开	NET-009 / SEC-001	admin_api_timeout=10-30s

5.2 KRaft 控制面（Controller / Quorum / Epoch / KRaft 版本）
V1 已能看基础 quorum 可达与活动 controller，但 V2 必须覆盖 controller 选主、epoch、controller 角色配置、4.x 版本中的 dynamic controller membership、kraft.version 等问题，否则很多“偶发性控制面抖动”仍然看不到。
常见问题	必采证据	判断依据	V2 新检查 ID	建议参数
controller quorum 无多数派	controller endpoints、TCP 可达、JMX quorum metrics	controller 可达数 < 多数派阈值 => 直接高危	KRF-004	controller_probe_timeout=3s
活动 controller 存在，但 controller endpoint 集合配置异常	controller.listener.names、quorum voters/bootstrap servers、metadata active controller	active controller 地址不在配置集合内，或 broker/controller listener 混用 => 配置异常	KRF-005 / CFG-010	lint 必须解析 KRaft 配置
epoch 抖动或频繁切主	CurrentLeader、CurrentEpoch、日志指纹、时间窗口内变化次数	单位时间 controller epoch 频繁增长 => 控制面不稳定	KRF-006	jmx scrape 15s；窗口 5-15min
Unknown voter connections / follower controller 连接异常	JMX NumberOfUnknownVoterConnections、日志	unknown voter > 0 或 controller append/fetch 异常 => quorum 信任链有问题	KRF-007	unknown_voter_threshold=0
4.1 之后配置未完成迁移/Finalization 不完整	kraft.version、controller.quorum.bootstrap.servers、metadata.version	版本升级后仍停留在旧配置路径，或 feature 未 finalize	UPG-001 / KRF-008	必须采集 feature/finalization 状态

5.3 Broker 注册、Metadata 与集群拓扑
Broker 注册看似简单，但很多线上问题发生在“broker 进程在、端口在、metadata 也能拉到”，实际却是注册不完整、拓扑漂移、controller 还没收敛，或者 broker ID/地址复用。
常见问题	必采证据	判断依据	V2 新检查 ID	建议参数
broker 存活但注册不完整	metadata brokers、expected broker count、controller 视图	期望=3 实际=2 或注册地址异常 => broker registration 问题	KFK-006	broker_count 来自 profile/compose
broker ID 冲突或地址重复	metadata broker list、node.id、listeners	多个 broker 地址重复或 ID/地址错位 => 高危配置错误	KFK-007 / CFG-011	必须对 ID/地址做去重校验
metadata 拉取正常但响应慢/反复超时	metadata RTT、Admin API timeout 分布、日志	可用但慢说明控制面或 broker 线程池压力，不应仅报通过	KFK-008	metadata_warn_ms=500；crit_ms=2000
集群拓扑与 compose/profile 描述不一致	compose services、profile、metadata	profile/compose/broker metadata 三者不一致 => 说明环境描述或实际拓扑变更	CFG-012 / KFK-009	strict_mode 可配置

5.4 Topic / Partition / Leader / ISR / 副本与均衡
这是 Kafka 业务稳定性的核心。V1 只覆盖了 leader 存在、ISR 完整和 min.insync.replicas 三项，远远不够。V2 要补齐 URP、UnderMinISR、AtMinISR、OfflineReplica、leader skew、partition skew、replica lag 等。
常见问题	必采证据	判断依据	V2 新检查 ID	建议参数
UnderReplicatedPartitions > 0	JMX URP、topic describe、ISR 列表	URP 不是普通告警，应结合 broker/磁盘/网络进一步关联主因	TOP-006 / MET-001	urp_warn=1；crit=1
UnderMinISR / AtMinISR	JMX UnderMinIsrPartitionCount、AtMinIsrPartitionCount、broker/topic min.insync.replicas	acks=all 场景下直接影响生产可用性	TOP-007 / MET-002	under_min_isr_crit=1；at_min_isr_warn=1
OfflineReplica / OfflinePartitions	JMX OfflineReplicaCount、topic describe	副本离线说明硬故障或目录故障，不应与普通 ISR 缺失等价	TOP-008	offline_replica_crit=1
leader 分布极不均衡	每 broker LeaderCount、PartitionCount	单 broker leader 过载会引发热点和延迟抖动	TOP-009	leader_skew_warn=30%
副本同步 lag 过大	ReplicaFetcherManager.MaxLag、per follower lag	ISR 仍完整但 lag 已很高时，需要提前预警	TOP-010 / MET-003	replica_lag_warn=10000 messages 或按 bytes/time
分区/副本数量规划异常	topic describe、默认分区数、RF、broker 数	单机三 broker 场景常见参数错配：RF>broker_count、partitions 过少/过多	TOP-011 / CFG-013	per_topic_rules 可配置

5.5 Producer 链路与写入一致性
很多业务层“Kafka 偶发失败”实际发生在 producer 侧参数、ack 语义、幂等与超时组合不当。V2 不能只做一次消息发送，需要把生产参数与 broker 约束一并判断。
常见问题	必采证据	判断依据	V2 新检查 ID	建议参数
acks 与 min.insync.replicas 组合不安全	producer 配置、broker/topic min.insync.replicas	acks=1 与业务宣称强一致不一致；acks=all 但 minISR 过低也不稳	PRD-001	producer_acks、expected_durability_required
未启用幂等且启用重试，可能乱序/重复	producer config	enable.idempotence=false 且 retries>0、max.in.flight>1 => 中高风险	PRD-002	lint on producer profile
delivery.timeout/request.timeout/linger 组合错误	producer config	delivery.timeout.ms < request.timeout.ms + linger.ms => 配置不合理	PRD-003	delivery_timeout_sanity=true
消息大小超限	message.max.bytes、max.request.size、日志指纹	服务端或客户端 RecordTooLarge/MessageTooLarge 必须直接提示链路点位	PRD-004 / LOG-006	message_bytes_probe 可变
写入被 quota/throttle	produce throttle metrics、request latency	业务误以为 Kafka 卡顿，实际是限流	PRD-005 / QTA-001	produce_throttle_warn_ms>0 持续 1-5min
事务生产者超时	transaction.timeout.ms、broker transaction.max.timeout.ms	超出 broker 上限会抛 InvalidTxnTimeoutException	PRD-006 / TXN-003	tx_timeout_ms lint

5.6 Consumer / Group / Offset / Lag / Rebalance
这是当前仓库最大的盲区之一。仅靠 probe 无法代表真实消费组行为。V2 要引入 group describe、coordinator、lag、rebalance、session/max.poll 参数审计。
常见问题	必采证据	判断依据	V2 新检查 ID	建议参数
消费组堆积（lag 高）	listConsumerGroupOffsets、end offsets、topic lag	真正生产问题里最常见，但当前工具完全看不到	CSM-001	lag_warn/crit 可按 group/topic 配
频繁 rebalance	consumer logs、group state、join/sync 频次、session/max.poll 参数	rebalance 风暴常被误判成 broker 问题	CSM-002 / LOG-007	rebalance_window=5min
max.poll.interval 过短导致踢出组	consumer config / profile	业务处理时间 > poll 间隔 => 典型配置问题	CSM-003	poll_interval_profile
heartbeat/session 超时配置不合理	consumer config + broker group settings	classic 模式 heartbeat 通常应显著小于 session timeout	CSM-004	heartbeat<=session/3 校验
offset reset 行为不符合预期	consumer config、group state	auto.offset.reset=latest/earliest/by_duration 错配会导致“消费不到旧数据”	CSM-005	显式输出当前语义
coordinator 异常 / offset commit 异常	group coordinator、__consumer_offsets、probe commit	需要将 commit 失败和 internal topic/coordinator 关联	CSM-006	commit_probe_enabled=true

5.7 事务 / Exactly-once / read_committed
当前仓库对 `__transaction_state` 只有存在性告警，还没有真正的事务链路探针。V2 要能区分“未使用事务所以不存在”“正在使用事务但状态主题异常”“事务超时/未提交导致 read_committed 看不到数据”。
常见问题	必采证据	判断依据	V2 新检查 ID	建议参数
未使用事务但看到 __transaction_state 缺失	internal topic state、是否启用 transactional.id	只应 WARN，不应误判为故障	TXN-001	transaction_probe=false 默认
事务主题缺失且存在事务生产者	transactional.id 配置、事务 probe、internal topic state	真正故障应升级为 FAIL	TXN-002	transaction_probe=true in full/incident
read_committed 读不到消息	consumer isolation.level、事务提交状态、未完成事务	不能只看 topic 是否有数据，必须看事务状态	TXN-004	probe_isolation_levels=both
事务提交/中止异常	日志指纹、事务 probe、broker metrics	需要与 transaction coordinator、topic state 联动判断	TXN-005	tx_probe_timeout=20s

5.8 安全（SSL / SASL / ACL / Authorizer）
V2 必须新增安全域；目前仓库几乎没有覆盖。否则一大类最常见的连接失败、认证失败、ACL 拒绝、证书过期/域名不匹配都无法定位。
常见问题	必采证据	判断依据	V2 新检查 ID	建议参数
监听器协议与 client 配置错配	listener.security.protocol.map、client security.protocol	TCP 通但握手失败；必须分离网络与安全故障	SEC-001	security_probe_enabled=true
SASL 机制不一致（PLAIN/SCRAM/OAUTH等）	listener SASL config、client mechanism	机制或 JAAS 配置不匹配 => 认证失败	SEC-002	sasl_mechanism_required
证书过期 / SAN 不匹配 / CA 链错误	broker/client cert、主机名、metadata endpoints	这是 SSL 常见故障源	SEC-003	cert_expiry_warn_days=30
ACL/Authorizer 导致 Describe/Produce/Consume 拒绝	StandardAuthorizer、ACL describe、错误码	需要将 authz 拒绝与普通网络故障区分	SEC-004	acl_probe_subjects configurable
KRaft 未使用标准 Authorizer 或迁移不完整	authorizer.class.name、ACL 状态	KRaft 下标准 Authorizer 是关键路径	SEC-005	lint authorizer config

5.9 存储 / LogDir / MetadataLogDir / 磁盘与目录故障
Kafka 很多硬故障最终都落到磁盘、目录、权限、inode、JBOD、离线 log dir。V2 需要把数据目录、元数据目录、可用空间、目录失败计数、日志目录离线指标纳入判定。
常见问题	必采证据	判断依据	V2 新检查 ID	建议参数
磁盘空间/ inode 紧张	df、inode、host thresholds	Kafka 很多毛病最终都是磁盘占满或 inode 枯竭	STG-001 / HOST-007	disk_warn=75%；crit=85%；inode_warn=80%
log dir 离线	OfflineLogDirectoryCount、日志、目录权限	这是高危硬故障，不应埋在日志里	STG-002 / MET-004	offline_log_dir_crit=1
metadata.log.dir 与 data dir 规划异常	broker config、目录存在性/权限	KRaft 控制面目录异常会导致启动/控制面问题	STG-003 / CFG-014	检查权限与挂载
JBOD 某一目录失败但 broker 仍在	log dir failure timeout、日志、metrics	需要能识别“部分目录故障”而不是只看 broker up/down	STG-004	log_dir_failure_timeout aware
容器挂载不符合预期	docker inspect mounts、compose volumes	看似 Kafka 故障，实则挂载/路径问题	DKR-005 / STG-005	inspect_mounts=true

5.10 Quota / Throttle / 背压
生产现场很容易把“Kafka 慢”“接口超时”误判成网络或 broker 崩溃，实际上是 quota 或 broker 背压。V2 应能识别 throttle-time、request percentage、produce/fetch 限速。
常见问题	必采证据	判断依据	V2 新检查 ID	建议参数
Produce throttle	Produce/Request throttle-time metrics、quota config	有时只是限流，不是 broker 故障	QTA-001	throttle_warn_ms>0 持续 1-5min
Fetch throttle	Fetch throttle metrics、consumer complaints	消费慢不一定是 lag 本身，可能是限流	QTA-002	fetch_throttle_warn_ms
Request percentage quota 触发	quota config、request metrics	常见于多租户环境	QTA-003	collect quota entities
连接数/请求数过大导致背压	request latency、network idle、handler idle	需要把背压与 quota/线程池一起关联	QTA-004 / JVM-003	idle thresholds

5.11 JVM / 线程池 / 请求队列 / GC
这是高级运维真正看重的层面。当前仓库完全缺。V2 要能通过 JMX 看线程闲置率、请求排队、GC、堆使用、网络线程/请求线程是否枯竭。
常见问题	必采证据	判断依据	V2 新检查 ID	建议参数
网络线程空闲率过低	NetworkProcessorAvgIdlePercent	官方建议一般保持明显余量，持续很低表示网络线程饱和	JVM-001 / MET-005	warn<0.3；crit<0.1
请求处理线程空闲率过低	RequestHandlerAvgIdlePercent	持续低空闲是 broker 压力核心信号	JVM-002 / MET-006	warn<0.3；crit<0.1
请求延迟高 / purgatory 堆积	RequestQueueTime、Local/RemoteTime、PurgatorySize	能定位为网络、磁盘、复制等待还是处理线程问题	JVM-003	latency buckets + purgatory
堆/GC 异常	JVM memory、GC pause、OOM 日志	当前只看 OOMKilled 远远不够	JVM-004	gc_pause_warn_ms、heap_used_warn_pct

5.12 宿主机 / Docker / 时钟 / 网络栈
当前宿主机与 Docker 检查是基础版，V2 需要把 fd、inode、磁盘水位、时钟偏移、容器重启、oom、sysctl、连接数、网卡异常补上。
常见问题	必采证据	判断依据	V2 新检查 ID	建议参数
容器在 / broker 在，但不断重启	docker restart count、status、recent logs	单次状态为 Up 不能说明过去稳定	DKR-006	restart_count_window=24h
fd 不足 / ulimit 过低	ulimit -n、进程 fd 使用量、日志	会导致各种隐蔽性故障	HOST-008	fd_warn=70%；crit=85%
时钟偏移	controller/broker/host time、NTP 状态	日志排序、SSL 验证、事务时间都受影响	HOST-009	clock_skew_warn_ms=500
端口被占用或监听漂移	ss/netstat、listener config	配置没问题但实际没按预期监听	HOST-010	listener_port_probe=true
OOMKilled / 内存逼近	docker inspect、cgroup 内存、JVM 堆配置	容器未 OOMKilled 也可能已处于高压	DKR-007 / HOST-011	mem_warn=85%

5.13 日志 / 指纹 / 证据引擎
V1 日志模块是真实采集，但能力仍偏弱：只扫最近窗口 + 固定少量正则。V2 要扩为可配置模式库、时间线聚合、source freshness、同类异常计数与主因加权。
常见问题	必采证据	判断依据	V2 新检查 ID	建议参数
已知错误指纹命中	日志时间、源、匹配规则、连续次数	不能只说“命中”，要说命中在哪、多久、几次、影响哪个 broker	LOG-005	custom_patterns 支持
日志有来源但实际上为空/太旧	line_count、byte_count、last_log_ts	采集成功不等于日志有价值，必须显式给出 freshness	LOG-006	min_lines_per_source / freshness_window
重复指纹风暴	fingerprint count、受影响 sources、时间窗口	需要区分单次异常与持续性雪崩	LOG-007	repeat_window=5-15min
业务异常未覆盖在内置指纹库	自定义指纹库、排障经验固化	内部工具必须支持自定义模式库而非写死正则	LOG-008	patterns.d/*.yaml

5.14 升级 / 版本 / Feature Finalization / Tiered Storage
少见但价值极高的故障域。V2 应覆盖 rolling upgrade 半完成、metadata.version 未 finalize、kraft.version 不一致、tiered storage 参数错配等问题。
常见问题	必采证据	判断依据	V2 新检查 ID	建议参数
rolling upgrade 半完成	broker versions、feature state、metadata.version	常见于版本升级后长期未 finalize	UPG-001	collect broker/version/features
kraft.version / 元数据特性不一致	feature states、controller state	控制面兼容问题往往隐藏较深	UPG-002	warn_on_unfinalized=true
tiered storage 配置打开但客户端/拉取参数未适配	remote storage config、fetch bytes、相关日志	少见，但一旦踩中排查成本极高	UPG-003 / STG-006	remote_storage_enabled awareness

6. V2 参数体系设计
当前仓库的配置已经有一个良好起点，但远远不够支撑高级运维判断。V2 应按“执行参数、探针参数、日志参数、JMX/metrics、安全、阈值、诊断策略”分段设计配置。下面给出建议字段与默认值基线。
参数	当前基线	V2 建议	说明
execution.timeout	30s	quick=30s / full=60s / incident=120s	总执行超时
execution.metadata_timeout	5s	5s	metadata/API 级超时
execution.admin_api_timeout	无	10s~30s	新增，专供 AdminClient / group / describe 调用
execution.jmx_timeout	无	3s~5s	新增，JMX 抓取超时
probe.enabled	true	mode preset	按模式决定是否运行真实链路探针
probe.topic	_kdoctor_probe	_kdoctor_probe	探针主题名
probe.produce_count	1	quick=1 / incident=3	多次发送便于判定偶发性
probe.cleanup	false	false 或 when_created	不建议默认强清理生产环境资源
probe.acks	无	all	V2 建议显式参数化
probe.enable_idempotence	无	true（incident 可开）	用于校验高可靠写入链路
probe.tx_probe_enabled	无	false（full/incident 可开）	事务链路探针
logs.tail_lines	300	quick=300 / incident=1000	日志样本窗口
logs.lookback_minutes	15	quick=15 / full=30 / incident=60	日志时间窗口
logs.max_files	固定 12	可配置 12~24	文件日志源数量上限
logs.max_bytes_per_source	固定 512KiB	1~2MiB	避免样本过窄
logs.custom_patterns	无	支持目录/文件加载	内部经验规则库
jmx.enabled	无	false by default	建议新增
jmx.metric_sets	无	kraft / broker / replica / request / quota / jvm	可按模式裁剪
thresholds.network_idle_warn	无	0.3	官方经验基线
thresholds.request_idle_warn	无	0.3	官方经验基线
thresholds.disk_warn	无	75%	宿主机/挂载水位告警
thresholds.disk_crit	无	85%	宿主机/挂载高危阈值
thresholds.clock_skew_warn_ms	无	500ms	时钟偏移
diagnosis.enable_confidence	无	true	启用主因置信度
diagnosis.suppress_downstream_symptoms	无	true	抑制继发症状刷屏

7. 建议的 V2 YAML 结构
version: "2"
default_profile: prod-internal-kraft

profiles:
  prod-internal-kraft:
    bootstrap_internal:
      - 192.168.119.7:9192
      - 192.168.119.7:9194
      - 192.168.119.7:9196
    controller_endpoints:
      - 192.168.119.7:9193
      - 192.168.119.7:9195
      - 192.168.119.7:9197
    broker_count: 3
    expected_min_isr: 2
    expected_replication_factor: 3
    execution_view: internal
    security_mode: plaintext
    group_probe_targets:
      - name: critical-group-a
        topic: business-topic-a

docker:
  enabled: true
  compose_file: ./docker-compose.yml
  container_names: [kafka1, kafka2, kafka3]
  inspect_mounts: true

logs:
  enabled: true
  lookback_minutes: 30
  tail_lines: 800
  max_files: 20
  max_bytes_per_source: 2097152
  custom_patterns_dir: ./rules/log-patterns.d

probe:
  enabled: true
  topic: _kdoctor_probe
  group_prefix: kdoctor-probe
  timeout: 20s
  produce_count: 3
  message_bytes: 1024
  acks: all
  enable_idempotence: true
  cleanup_mode: when_created
  tx_probe_enabled: false

jmx:
  enabled: true
  scrape_timeout: 5s
  metric_sets: [kraft, broker, replica, request, quota, jvm]

host:
  enabled: true
  disk_paths: [/data/kafka, /bitnami/kafka]
  check_ports: [9192, 9193, 9194, 9195, 9196, 9197]
  fd_warn_pct: 70
  fd_crit_pct: 85
  clock_skew_warn_ms: 500

thresholds:
  urp_warn: 1
  under_min_isr_crit: 1
  network_idle_warn: 0.3
  request_idle_warn: 0.3
  disk_warn_pct: 75
  disk_crit_pct: 85
  replica_lag_warn: 10000
  leader_skew_warn_pct: 30

diagnosis:
  enable_confidence: true
  suppress_downstream_symptoms: true
  max_root_causes: 3
  rule_packs: [builtin, internal-kafka-prod]

8. 诊断模型：从检查结果到主因排序
V2 不应继续把所有检查结果平铺出来再简单摘前几条。建议引入“证据权重 + 置信度 + 继发症状抑制”模型。一个建议的思路是：
•	每个 check result 除 status 外，还应有 `evidence_strength`、`scope`、`source_freshness`、`confidence` 四个维度。
•	诊断层维护 root cause rule pack，例如：`NET-006 advertised.listeners mismatch` 可吸收 `NET-003 metadata endpoint 不可达`、`KFK-003 endpoint structure abnormal`、`LOG-005 Connection to node ... could not be established` 等症状。
•	对“上游失败导致下游未执行”的情况，统一标记 `downstream_skipped`，而不是重复 FAIL。
•	输出时同时给出：主因、主要证据、反证/局限、下一步动作顺序。
建议输出字段： root_causes[].id / summary / confidence / evidence_refs / next_actions
•	checks[].status / severity / evidence / source / freshness / skipped_reason
•	summary.overview / mode / runtime / data_source_coverage

9. 实施优先级建议
阶段	目标	必须完成项
P0（先补盲区）	让工具真正覆盖大头问题	JMX 指标采集；group/lag；storage/logdir；安全域基础；scheduler timeout/degrade
P1（增强判定）	把“能看见”升级成“能定位”	confidence root cause；自定义日志规则；leader/partition skew；quota/throttle；upgrade checks
P2（高级场景）	覆盖少见但昂贵的问题	事务 probe；tiered storage；version finalization；复杂安全与多视角

10. 对内部使用场景的落地建议
因为这个工具定位为内部排障工具，所以 V2 的成功标准不是“是否覆盖所有 Kafka 功能”，而是：面对你们常见的一机三 broker / KRaft / Docker / host network / 内外网混合接入的真实环境，能否快速把问题指向正确层级。建议优先把你们历史故障、值班手册、常见日志关键字都沉淀成自定义规则包，与 V2 的标准规则一起维护。
另一个重要建议是把模式预设做成“快检/战时”两套：平时以 quick/lint 低成本跑，出问题时一键切到 incident，自动打开更长日志窗口、更高 tail lines、更强 probe 和更多 metrics。这比一上来把所有检查都堆到默认模式更实用。
附录：主要依据（官方与仓库）
以下资料用于定义 V2 的检查边界、判断依据和参数基线。内部设计时，建议优先以 Apache Kafka 官方文档和当前 Kdoctor 仓库代码为准，其次再参考现场经验。
[R1] Apache Kafka 官方文档总入口：https://kafka.apache.org/documentation/
[R2] Kafka KRaft 配置与 controller / listener 官方说明：https://kafka.apache.org/41/documentation/
[R3] Kafka broker 配置参考（listeners / advertised.listeners / inter.broker.listener.name / controller.listener.names 等）：https://kafka.apache.org/40/generated/kafka_config.html
[R4] Kafka producer 配置参考（acks / enable.idempotence / delivery.timeout.ms / transaction.timeout.ms）：https://kafka.apache.org/40/generated/producer_config.html
[R5] Kafka consumer 配置参考（heartbeat / session / max.poll.interval / auto.offset.reset / isolation.level）：https://kafka.apache.org/40/generated/consumer_config.html
[R6] Kafka 监控文档（JMX 指标：ActiveControllerCount、UnderReplicatedPartitions、OfflineLogDirectoryCount、IdlePercent 等）：https://kafka.apache.org/40/documentation/#monitoring
[R7] Kafka Quotas 官方说明：https://kafka.apache.org/40/documentation/#design_quotas
[R8] Kafka 安全文档（SSL / SASL / ACL / StandardAuthorizer）：https://kafka.apache.org/40/documentation/#security
[R9] Kdoctor 公开仓库 README：https://github.com/isYaoNoistu/Kdoctor
[R10] Kdoctor 架构文档 architecture.md：https://github.com/isYaoNoistu/Kdoctor/blob/main/architecture.md
[R11] Kdoctor V1 设计说明 doc.md：https://github.com/isYaoNoistu/Kdoctor/blob/main/doc.md
