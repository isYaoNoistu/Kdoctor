package localize

import (
	"path/filepath"
	"strings"

	"kdoctor/pkg/model"
)

var exactText = map[string]string{
	"no checks were executed":                                                   "本次未执行任何检查项",
	"probe snapshot missing":                                                    "探针快照缺失",
	"probe disabled for current mode":                                           "当前模式未启用探针",
	"probe topic is not configured":                                             "未配置探针主题",
	"no probe brokers are available":                                            "当前没有可用的探针 broker",
	"bootstrap endpoint reachable":                                              "bootstrap 地址可达",
	"no bootstrap endpoints reachable":                                          "所有 bootstrap 地址均不可达",
	"configured listeners cannot be evaluated":                                  "无法评估显式 listener 端点",
	"no explicit listener endpoints were provided":                              "当前输入未提供显式 listener 端点",
	"explicit listener endpoints are reachable from the current execution view": "从当前执行视角看，显式 listener 端点可达",
	"some internal controller listeners are not reachable from the current external execution view":           "从当前外部执行视角无法直连部分内部 controller listener",
	"some explicit listener endpoints are not reachable":                                                      "部分显式 listener 端点不可达",
	"metadata endpoints were not checked":                                                                     "未执行 metadata 返回端点探测",
	"metadata returned broker endpoints are reachable":                                                        "metadata 返回的 broker 端点可达",
	"metadata returned unreachable broker endpoints":                                                          "metadata 返回了不可达的 broker 端点",
	"DNS resolution cannot be evaluated":                                                                      "无法评估 DNS 解析情况",
	"no endpoint hosts are available for DNS resolution":                                                      "当前没有可用于 DNS 解析检查的主机名",
	"all configured endpoints use literal IP addresses":                                                       "所有配置端点均使用字面 IP，无需 DNS 解析",
	"all configured hostnames resolve successfully":                                                           "所有配置的主机名均解析成功",
	"some configured hostnames failed DNS resolution":                                                         "部分配置的主机名解析失败",
	"kafka metadata unavailable":                                                                              "Kafka metadata 不可用",
	"cluster metadata retrieved successfully":                                                                 "已成功获取集群 metadata",
	"broker registration cannot be evaluated":                                                                 "无法评估 broker 注册状态",
	"all expected brokers are registered":                                                                     "所有期望的 broker 都已注册",
	"broker registration count is below expectation":                                                          "已注册 broker 数量低于预期",
	"broker endpoint legality cannot be evaluated":                                                            "无法评估 broker 端点合法性",
	"no broker endpoints were returned by metadata":                                                           "metadata 未返回任何 broker 端点",
	"metadata returned malformed broker endpoint":                                                             "metadata 返回了格式非法的 broker 端点",
	"metadata returned duplicate broker endpoints":                                                            "metadata 返回了重复的 broker 端点",
	"broker endpoints are structurally valid":                                                                 "broker 端点结构合法",
	"metadata returned private broker endpoints for the current external client view":                         "对当前外部客户端视角，metadata 返回了私网 broker 端点",
	"internal topics cannot be evaluated":                                                                     "无法评估内部主题健康状态",
	"__consumer_offsets topic is missing":                                                                     "__consumer_offsets 主题缺失",
	"internal Kafka topics are unhealthy":                                                                     "Kafka 内部主题状态异常",
	"__transaction_state topic is not present yet":                                                            "__transaction_state 主题暂未出现",
	"internal Kafka topics are healthy":                                                                       "Kafka 内部主题状态健康",
	"controller quorum configuration is not available in the current input mode":                              "当前输入模式下没有可用的 controller quorum 配置",
	"controller quorum endpoints were provided explicitly":                                                    "已显式提供 controller quorum 端点",
	"controller node.id is missing or invalid in compose":                                                     "compose 中的 controller node.id 缺失或非法",
	"controller.quorum.voters is missing in a Kafka service":                                                  "某个 Kafka 服务缺少 controller.quorum.voters",
	"controller.quorum.voters format is invalid":                                                              "controller.quorum.voters 格式非法",
	"controller.quorum.voters differs across Kafka services":                                                  "不同 Kafka 服务的 controller.quorum.voters 不一致",
	"some controller node.id values are missing from controller.quorum.voters":                                "部分 controller node.id 未出现在 controller.quorum.voters 中",
	"explicit controller endpoints differ from compose quorum voters":                                         "显式 controller 端点与 compose 中的 quorum voters 不一致",
	"controller quorum configuration is consistent":                                                           "controller quorum 配置一致",
	"active controller cannot be evaluated":                                                                   "无法评估活动 controller 状态",
	"metadata did not report an active controller":                                                            "metadata 未报告活动 controller",
	"metadata reports an active controller":                                                                   "metadata 已报告活动 controller",
	"active controller is not present in the broker registration set":                                         "活动 controller 不在 broker 注册集合中",
	"metadata reports an active controller but its listener is not reachable from the current execution view": "metadata 已报告活动 controller，但从当前执行视角无法直连其 listener",
	"controller quorum endpoints are not available in the current input mode":                                 "当前输入模式下没有可用的 controller quorum 端点",
	"controller quorum has majority":                                                                          "controller quorum 具备多数派",
	"controller quorum cannot be directly verified from the current external probe view":                      "从当前外部探测视角无法直接验证 controller quorum",
	"controller quorum lost majority":                                                                         "controller quorum 已丢失多数派",
	"controller quorum still has majority but not all voters are reachable":                                   "controller quorum 仍有多数派，但并非所有 voter 可达",
	"topic leadership cannot be evaluated":                                                                    "无法评估 topic leader 状态",
	"all partitions have leaders":                                                                             "所有分区都有 leader",
	"partitions without leader detected":                                                                      "检测到无 leader 的分区",
	"ISR replica health cannot be evaluated":                                                                  "无法评估 ISR 副本健康状态",
	"all partitions have full ISR":                                                                            "所有分区的 ISR 都完整",
	"some partitions have empty ISR":                                                                          "部分分区的 ISR 为空",
	"some partitions are under replicated":                                                                    "部分分区处于副本不足状态",
	"ISR health cannot be evaluated":                                                                          "无法评估 ISR 健康状态",
	"ISR satisfies min.insync.replicas":                                                                       "ISR 满足 min.insync.replicas 要求",
	"some partitions are below min.insync.replicas and acks=all may fail":                                     "部分分区低于 min.insync.replicas，acks=all 可能失败",
	"some partitions are under replicated but still above min.insync.replicas":                                "部分分区副本不足，但仍高于 min.insync.replicas",
	"compose snapshot not available":                                                                          "没有可用的 compose 快照",
	"compose parsed successfully":                                                                             "compose 解析成功",
	"compose parsed but no Kafka services were detected":                                                      "compose 解析成功，但未识别出 Kafka 服务",
	"compose Kafka services not available":                                                                    "没有可用的 compose Kafka 服务",
	"node.id missing in Kafka service":                                                                        "Kafka 服务缺少 node.id",
	"duplicate node.id detected":                                                                              "检测到重复的 node.id",
	"node.id values are unique":                                                                               "node.id 唯一",
	"cluster.id missing in Kafka service":                                                                     "Kafka 服务缺少 cluster.id",
	"cluster.id is inconsistent across Kafka services":                                                        "不同 Kafka 服务的 cluster.id 不一致",
	"cluster.id is consistent across Kafka services":                                                          "各 Kafka 服务的 cluster.id 一致",
	"process.roles missing in Kafka service":                                                                  "Kafka 服务缺少 process.roles",
	"process.roles contains unsupported role":                                                                 "process.roles 包含不支持的角色",
	"process.roles is missing broker role":                                                                    "process.roles 缺少 broker 角色",
	"process.roles is missing controller role":                                                                "process.roles 缺少 controller 角色",
	"process.roles are legal for detected Kafka services":                                                     "识别到的 Kafka 服务 process.roles 合法",
	"controller.quorum.voters missing in Kafka service":                                                       "Kafka 服务缺少 controller.quorum.voters",
	"controller.quorum.voters is inconsistent across Kafka services":                                          "各 Kafka 服务的 controller.quorum.voters 不一致",
	"some node.id values are not represented in controller.quorum.voters":                                     "部分 node.id 未体现在 controller.quorum.voters 中",
	"controller.quorum.voters is consistent and matches node.id values":                                       "controller.quorum.voters 一致且与 node.id 对齐",
	"listeners format is invalid":                                                                             "listeners 格式非法",
	"advertised.listeners format is invalid":                                                                  "advertised.listeners 格式非法",
	"listener port conflict detected across Kafka services":                                                   "不同 Kafka 服务之间存在 listener 端口冲突",
	"advertised.listeners is missing a client-facing listener":                                                "advertised.listeners 缺少面向客户端的 listener",
	"advertised.listeners must not use 0.0.0.0":                                                               "advertised.listeners 不能使用 0.0.0.0",
	"listeners and advertised.listeners are structurally consistent":                                          "listeners 与 advertised.listeners 结构一致",
	"inter.broker.listener.name is missing":                                                                   "缺少 inter.broker.listener.name",
	"listeners format is invalid while validating inter.broker.listener.name":                                 "校验 inter.broker.listener.name 时，listeners 格式非法",
	"inter.broker.listener.name does not exist in listeners":                                                  "inter.broker.listener.name 未出现在 listeners 中",
	"inter.broker.listener.name points to EXTERNAL listener":                                                  "inter.broker.listener.name 指向了 EXTERNAL listener",
	"inter.broker.listener.name points to valid listener":                                                     "inter.broker.listener.name 指向合法 listener",
	"min.insync.replicas equals default.replication.factor and leaves no slack":                               "min.insync.replicas 等于 default.replication.factor，没有缓冲余量",
	"replication and ISR settings are structurally legal":                                                     "副本与 ISR 配置结构合法",
	"log collection is not enabled in the current input mode":                                                 "当前输入模式未启用日志采集",
	"no log sources were available from the current execution view":                                           "从当前执行视角没有可用的日志来源",
	"log sources were collected successfully":                                                                 "日志来源采集成功",
	"log fingerprints cannot be evaluated without collected log sources":                                      "缺少日志来源，无法评估日志指纹",
	"no known Kafka error fingerprints were found in recent logs":                                             "近期日志未命中已知 Kafka 错误指纹",
	"known Kafka error fingerprints were found in recent logs":                                                "近期日志命中了已知 Kafka 错误指纹",
	"log explanations cannot be generated without collected log sources":                                      "缺少日志来源，无法生成日志解释",
	"no matched log errors required explanation":                                                              "没有命中的日志错误需要额外解释",
	"matched log errors were explained and mapped to likely causes":                                           "已对命中的日志错误给出解释并映射到可能原因",
	"log aggregation cannot be evaluated without collected log sources":                                       "缺少日志来源，无法做日志聚合",
	"no repeated log fingerprints were observed":                                                              "未观察到重复日志指纹",
	"repeated log fingerprints were aggregated successfully":                                                  "重复日志指纹已成功聚合",
	"host disk usage is not available in the current input mode":                                              "当前输入模式下没有可用的宿主机磁盘使用信息",
	"host disk usage is within safe range":                                                                    "宿主机磁盘使用处于安全范围",
	"some Kafka disk paths are critically full":                                                               "部分 Kafka 磁盘路径已接近打满",
	"some Kafka disk paths are nearing capacity":                                                              "部分 Kafka 磁盘路径接近容量上限",
	"host listener ports are not available in the current input mode":                                         "当前输入模式下没有可用的宿主机 listener 端口信息",
	"expected Kafka listener ports are reachable from the host execution view":                                "从宿主机执行视角看，期望的 Kafka listener 端口可达",
	"some expected Kafka listener ports are not reachable from the host execution view":                       "从宿主机执行视角看，部分期望的 Kafka listener 端口不可达",
	"docker runtime is not enabled in the current input mode":                                                 "当前输入模式未启用 Docker 运行时采集",
	"docker runtime is not available on the current execution host":                                           "当前执行主机不可用 Docker 运行时",
	"all expected Kafka containers exist":                                                                     "所有期望的 Kafka 容器都存在",
	"some expected Kafka containers do not exist":                                                             "部分期望的 Kafka 容器不存在",
	"all expected Kafka containers are running":                                                               "所有期望的 Kafka 容器都在运行",
	"some expected Kafka containers are not running":                                                          "部分期望的 Kafka 容器未运行",
	"no Kafka container shows an OOMKilled runtime state":                                                     "没有 Kafka 容器出现 OOMKilled 运行时状态",
	"some Kafka containers were OOMKilled":                                                                    "部分 Kafka 容器发生过 OOMKilled",
	"compose Kafka services are not available for mount validation":                                           "没有可用于挂载校验的 compose Kafka 服务",
	"Kafka data and metadata paths are backed by docker mounts":                                               "Kafka 数据和 metadata 路径都由 Docker 挂载承载",
	"some Kafka data or metadata paths are not backed by docker mounts":                                       "部分 Kafka 数据或 metadata 路径没有 Docker 挂载承载",
	"bootstrap metadata probe succeeded":                                                                      "bootstrap 元数据探针成功",
	"bootstrap metadata probe failed":                                                                         "bootstrap 元数据探针失败",
	"producer probe succeeded":                                                                                "生产探针成功",
	"producer probe failed":                                                                                   "生产探针失败",
	"consumer probe succeeded":                                                                                "消费探针成功",
	"consumer probe failed":                                                                                   "消费探针失败",
	"consumer group commit probe succeeded":                                                                   "消费组提交位点探针成功",
	"consumer group commit probe failed":                                                                      "消费组提交位点探针失败",
	"end-to-end probe succeeded":                                                                              "端到端探针成功",
	"end-to-end probe failed":                                                                                 "端到端探针失败",
	"this may be acceptable if transactions are unused":                                                       "如果未使用事务，这种情况可能是可接受的",
	"verify probe topic exists":                                                                               "确认探针主题存在",
	"verify produce path and acks":                                                                            "检查生产链路与 acks 配置",
	"check ISR and leader health":                                                                             "检查 ISR 与 leader 健康状态",
	"verify topic leader and offsets":                                                                         "检查主题 leader 与位点状态",
	"verify fetch path from current client network":                                                           "检查从当前客户端网络到 fetch 路径的可达性",
	"check consumer side timeouts":                                                                            "检查消费端超时配置",
	"verify __consumer_offsets health":                                                                        "检查 __consumer_offsets 健康状态",
	"verify coordinator health":                                                                               "检查 coordinator 健康状态",
	"check controller and internal topic replicas":                                                            "检查 controller 与内部主题副本状态",
	"check the failing stage first":                                                                           "优先检查失败阶段本身",
	"verify probe topic and broker reachability":                                                              "检查探针主题与 broker 可达性",
	"correlate with network, ISR and controller checks":                                                       "结合网络、ISR 与 controller 检查一起判断",
	"verify bootstrap endpoints":                                                                              "检查 bootstrap 地址配置",
	"verify Kafka metadata requests are served":                                                               "确认 Kafka metadata 请求可正常响应",
	"check cluster and listener health":                                                                       "检查集群与 listener 健康状态",
	"verify bootstrap addresses":                                                                              "核对 bootstrap 地址",
	"verify firewall and security group":                                                                      "检查防火墙与安全组",
	"verify Kafka listener binding":                                                                           "检查 Kafka listener 绑定",
	"verify advertised.listeners":                                                                             "检查 advertised.listeners 配置",
	"verify returned broker ports are exposed":                                                                "确认返回的 broker 端口已暴露",
	"verify routing from current client network":                                                              "确认从当前客户端网络到 broker 的路由可达",
	"verify controller health":                                                                                "检查 controller 健康状态",
	"verify affected brokers are online":                                                                      "确认受影响 broker 在线",
	"check partition reassignment or recent broker failures":                                                  "检查分区迁移或近期 broker 故障",
	"verify broker replication pipeline":                                                                      "检查 broker 副本复制链路",
	"check affected broker health and disks":                                                                  "检查受影响 broker 的健康与磁盘状态",
	"check controller and partition movement":                                                                 "检查 controller 状态与分区迁移",
	"verify follower brokers are healthy":                                                                     "检查 follower broker 健康状态",
	"monitor ISR recovery before write pressure increases":                                                    "在写入压力升高前持续观察 ISR 恢复情况",
	"verify replica health":                                                                                   "检查副本健康状态",
	"reduce write pressure until ISR recovers":                                                                "在 ISR 恢复前降低写入压力",
	"verify broker processes are running":                                                                     "确认 broker 进程正在运行",
	"verify node.id and controller quorum settings":                                                           "检查 node.id 与 controller quorum 配置",
	"verify broker can register to cluster":                                                                   "确认 broker 可以注册到集群",
	"verify cluster metadata integrity":                                                                       "检查集群 metadata 完整性",
	"verify brokers can create and load internal topics":                                                      "确认 broker 可以创建并加载内部主题",
	"check controller and broker logs":                                                                        "检查 controller 与 broker 日志",
	"verify broker replication health":                                                                        "检查 broker 副本健康状态",
	"check internal topic leaders and ISR":                                                                    "检查内部主题的 leader 与 ISR",
	"verify transactional producers if they are expected":                                                     "如果预期存在事务生产者，请进一步核对事务链路",
	"run kdoctor from the Kafka internal network or host":                                                     "在 Kafka 内网或宿主机上执行 kdoctor",
	"verify controller listeners from the broker host":                                                        "从 broker 宿主机验证 controller listener",
	"use metadata and broker health as temporary reference":                                                   "暂时以 metadata 与 broker 健康状态作为参考",
	"verify controller listeners are reachable":                                                               "确认 controller listener 可达",
	"verify broker-controller processes are healthy":                                                          "确认 broker-controller 进程健康",
	"verify quorum voter configuration":                                                                       "检查 quorum voter 配置",
	"verify controller quorum majority":                                                                       "检查 controller quorum 多数派",
	"check controller listener reachability":                                                                  "检查 controller listener 可达性",
	"inspect broker logs for controller election errors":                                                      "检查 broker 日志中的 controller 选举错误",
	"verify broker registration":                                                                              "检查 broker 注册状态",
	"verify controller election completed":                                                                    "确认 controller 选举已完成",
	"inspect broker logs for registration failures":                                                           "检查 broker 日志中的注册失败信息",
	"verify controller listener binding and exposure":                                                         "检查 controller listener 绑定与暴露情况",
	"run kdoctor from the Kafka host or private network":                                                      "在 Kafka 宿主机或私网环境中执行 kdoctor",
	"check controller logs for election churn":                                                                "检查 controller 日志中是否存在频繁选举",
	"align profile controller_endpoints with compose quorum voters":                                           "让 profile 中的 controller_endpoints 与 compose quorum voters 保持一致",
	"verify the current execution view uses the intended controller listener addresses":                       "确认当前执行视角使用的是预期 controller listener 地址",
	"plan disk cleanup or capacity expansion":                                                                 "规划磁盘清理或容量扩容",
	"review retention and segment sizing":                                                                     "复核保留策略与 segment 大小配置",
	"monitor growth before the next traffic spike":                                                            "在下一次流量高峰前持续观察容量增长",
	"free disk space immediately":                                                                             "立即释放磁盘空间",
	"review retention and cleanup settings":                                                                   "复核保留与清理策略",
	"check whether a broker or metadata directory stopped writing":                                            "检查 broker 或 metadata 目录是否已停止写入",
	"verify broker processes are listening on the expected ports":                                             "确认 broker 进程监听了预期端口",
	"check docker host network or service binding":                                                            "检查 Docker host 网络或服务绑定状态",
	"compare compose listener settings with the active process state":                                         "对照 compose listener 配置与当前进程实际状态",
	"verify compose service names and container_names":                                                        "检查 compose 服务名与 container_name 配置",
	"start the missing containers":                                                                            "启动缺失的容器",
	"check whether the execution host is the intended Docker host":                                            "确认当前执行主机就是目标 Docker 宿主机",
	"restart the stopped containers":                                                                          "重启已停止的容器",
	"inspect docker logs for startup failures":                                                                "检查 Docker 日志中的启动失败信息",
	"verify host resources and port conflicts":                                                                "检查宿主机资源与端口冲突",
	"review container memory limits and heap size":                                                            "检查容器内存限制与 JVM 堆大小",
	"inspect broker logs for memory pressure":                                                                 "检查 broker 日志中的内存压力迹象",
	"reduce load or restart carefully after adding headroom":                                                  "增加资源余量后再谨慎降载或重启",
	"bind-mount Kafka data and metadata directories":                                                          "为 Kafka 数据与 metadata 目录配置 bind mount",
	"verify compose volume declarations":                                                                      "检查 compose volume 声明",
	"avoid storing Kafka state only in ephemeral container layers":                                            "避免仅把 Kafka 状态写入容器临时层",
	"prefer INTERNAL listener for broker-to-broker traffic":                                                   "broker 之间通信优先使用 INTERNAL listener",
	"verify listener security and routing for inter-broker communication":                                     "检查 broker 间通信链路的 listener 安全与路由",
	"verify this is intentional":                                                                              "确认这是有意为之的配置",
	"ensure producer ack expectations match the strict ISR policy":                                            "确认生产者 ack 预期与严格 ISR 策略相匹配",
}

var fragmentReplacer = strings.NewReplacer(
	"failure_stage=metadata", "失败阶段=metadata",
	"failure_stage=produce", "失败阶段=生产",
	"failure_stage=consume", "失败阶段=消费",
	"failure_stage=commit", "失败阶段=提交位点",
	"failure_stage=context", "失败阶段=上下文取消",
	"bootstrap=", "bootstrap=",
	"topic=", "主题=",
	"partition=", "分区=",
	"offset=", "位点=",
	"group_id=", "消费组=",
	"message_id=", "消息 ID=",
	"error=", "错误=",
	"expected=", "期望=",
	"actual=", "实际=",
	"service=", "服务=",
	"source=", "来源=",
	"controller endpoint=", "controller 端点=",
	"controller id=", "controller ID=",
	"controller_id=", "controller ID=",
	"cluster_id=", "集群 ID=",
	"brokers=", "broker 数=",
	"broker_id=", "broker ID=",
	"broker_ids=", "broker ID 列表=",
	"address=", "地址=",
	"baseline=", "基线=",
	"compose voters=", "compose 投票节点=",
	"runtime endpoints=", "运行时端点=",
	"private_endpoints=", "私网端点=",
	"missing node.id=", "缺少 node.id=",
	"missing controller node.id=", "缺少 controller node.id=",
	"roles=", "角色=",
	"voters=", "投票节点=",
	"count=", "次数=",
	"sources=", "来源数=",
	"severity=", "严重级别=",
	"meaning=", "含义=",
	"collector warning=", "采集告警=",
	"metadata_duration_ms=", "metadata 耗时(ms)=",
	"produce_duration_ms=", "生产耗时(ms)=",
	"consume_duration_ms=", "消费耗时(ms)=",
	"commit_duration_ms=", "提交位点耗时(ms)=",
	"reachable=", "可达数量=",
	"majority=", "多数派阈值=",
	"reachable in ", "可达，耗时 ",
	" unreachable: ", " 不可达：",
	"message mismatch: expected ", "消息不匹配：期望 ",
	" got ", "，实际 ",
	"literal-ip", "字面 IP",
	"produce probe: ", "生产探针：",
	"commit probe offset: ", "提交位点探针：",
	"consume partition: ", "消费分区：",
	"send probe message: ", "发送探针消息：",
	"kafka server: ", "Kafka 服务端：",
	"Request was for a topic or partition that does not exist on this broker", "请求的主题或分区在该 broker 上不存在",
	"requested topic or partition does not exist on the cluster", "请求的主题或分区在集群中不存在",
	"OOMKilled=true", "OOMKilled=true",
	"restart_count=", "重启次数=",
	"controller listener reachable in ", "controller listener 可达，耗时 ",
	"controller listener was not directly probed in the current input mode", "当前输入模式未直接探测 controller listener",
	"controller listener is private and was not directly reachable from the current external view", "controller listener 为私网地址，当前外部视角无法直接连通",
	"active controller address was not part of the explicit controller endpoint set", "活动 controller 地址不在显式 controller 端点集合中",
)

func ApplyChinese(report *model.Report) {
	if report == nil {
		return
	}

	report.Mode = TranslateMode(report.Mode)
	report.Summary.Overview = TranslateText(report.Summary.Overview)
	report.Summary.RootCauses = translateSlice(report.Summary.RootCauses)
	report.Summary.RecommendedActions = translateSlice(report.Summary.RecommendedActions)
	report.Errors = translateSlice(report.Errors)

	for i := range report.Checks {
		report.Checks[i].Module = TranslateModule(report.Checks[i].Module)
		report.Checks[i].Summary = TranslateText(report.Checks[i].Summary)
		report.Checks[i].Evidence = translateSlice(report.Checks[i].Evidence)
		report.Checks[i].Impact = TranslateText(report.Checks[i].Impact)
		report.Checks[i].PossibleCauses = translateSlice(report.Checks[i].PossibleCauses)
		report.Checks[i].NextActions = translateSlice(report.Checks[i].NextActions)
		report.Checks[i].ErrorMessage = TranslateText(report.Checks[i].ErrorMessage)
	}
}

func TranslateText(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return input
	}
	if translated, ok := exactText[input]; ok {
		return translated
	}

	output := fragmentReplacer.Replace(input)
	for oldValue, newValue := range exactText {
		if strings.Contains(output, oldValue) {
			output = strings.ReplaceAll(output, oldValue, newValue)
		}
	}
	return output
}

func TranslateMode(mode string) string {
	switch mode {
	case model.ModeQuick:
		return "快速巡检"
	case model.ModeFull:
		return "全量体检"
	case model.ModeProbe:
		return "链路探针"
	case model.ModeIncident:
		return "战时排障"
	case model.ModeLint:
		return "配置审计"
	default:
		return mode
	}
}

func TranslateModule(module string) string {
	switch module {
	case "network":
		return "网络"
	case "kafka":
		return "Kafka"
	case "kraft":
		return "KRaft"
	case "topic":
		return "主题"
	case "client":
		return "客户端"
	case "lint":
		return "配置"
	case "logs":
		return "日志"
	case "host":
		return "宿主机"
	case "docker":
		return "Docker"
	case "capacity":
		return "容量"
	default:
		return module
	}
}

func TranslateStatus(status model.CheckStatus) string {
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
	case model.StatusSkip:
		return "跳过"
	case model.StatusTimeout:
		return "超时"
	default:
		return string(status)
	}
}

func GuessFormat(format string, outputPath string, jsonFlag bool) string {
	format = strings.TrimSpace(strings.ToLower(format))
	if jsonFlag {
		return "json"
	}
	if format != "" {
		return format
	}
	switch strings.ToLower(filepath.Ext(strings.TrimSpace(outputPath))) {
	case ".json":
		return "json"
	case ".md", ".markdown":
		return "markdown"
	default:
		return "terminal"
	}
}

func translateSlice(values []string) []string {
	if len(values) == 0 {
		return values
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, TranslateText(value))
	}
	return out
}
