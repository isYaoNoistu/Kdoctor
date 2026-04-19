package localize

import (
	"path/filepath"
	"strings"

	"kdoctor/pkg/model"
)

var exactText = map[string]string{
	"no checks were executed":                                                                                 "本次未执行任何检查项",
	"probe snapshot missing":                                                                                  "探针快照缺失",
	"probe disabled for current mode":                                                                         "当前模式未启用探针",
	"probe topic is not configured":                                                                           "未配置探针主题",
	"no probe brokers are available":                                                                          "当前没有可用的探针 broker",
	"probe topic already exists":                                                                              "探针主题已存在",
	"probe topic created for this run":                                                                        "探针主题已为本次检测自动创建",
	"probe topic became available during readiness check":                                                     "探针主题在就绪检查期间已可用",
	"probe topic could not be prepared":                                                                       "探针主题未准备就绪",
	"metadata stage failed; produce stage was not executed":                                                   "metadata 阶段失败，未执行生产阶段",
	"probe topic was not ready; produce stage was not executed":                                               "探针主题未就绪，未执行生产阶段",
	"execution context ended before produce stage":                                                            "执行上下文提前结束，未执行生产阶段",
	"metadata stage failed; consume stage was not executed":                                                   "metadata 阶段失败，未执行消费阶段",
	"probe topic was not ready; consume stage was not executed":                                               "探针主题未就绪，未执行消费阶段",
	"produce stage failed; consume stage was not executed":                                                    "生产阶段失败，未执行消费阶段",
	"execution context ended before consume stage":                                                            "执行上下文提前结束，未执行消费阶段",
	"metadata stage failed; commit stage was not executed":                                                    "metadata 阶段失败，未执行提交位点阶段",
	"probe topic was not ready; commit stage was not executed":                                                "探针主题未就绪，未执行提交位点阶段",
	"produce stage failed; commit stage was not executed":                                                     "生产阶段失败，未执行提交位点阶段",
	"consume stage failed; commit stage was not executed":                                                     "消费阶段失败，未执行提交位点阶段",
	"execution context ended before commit stage":                                                             "执行上下文提前结束，未执行提交位点阶段",
	"metadata stage failed; downstream probe stages were skipped":                                             "metadata 阶段失败，后续探针阶段已跳过",
	"probe topic was not ready; downstream probe stages were skipped":                                         "探针主题未就绪，后续探针阶段已跳过",
	"produce stage failed; downstream probe stages were skipped":                                              "生产阶段失败，后续探针阶段已跳过",
	"consume stage failed; downstream probe stages were skipped":                                              "消费阶段失败，后续探针阶段已跳过",
	"commit stage failed; end-to-end probe ended at commit stage":                                             "提交位点阶段失败，端到端探针在该阶段结束",
	"execution context ended before the probe finished":                                                       "执行上下文提前结束，探针未完成",
	"metadata stage was not executed":                                                                         "metadata 阶段未执行",
	"__consumer_offsets is missing after consumer group commit probe executed":                                "__consumer_offsets 缺失，且消费组提交位点探针已经执行",
	"__consumer_offsets is not present yet; cluster may still be fresh or commit path has not run":            "__consumer_offsets 尚未出现，当前更像是新集群或消费组提交链路尚未运行",
	"bootstrap endpoint reachable":                                                                            "bootstrap 地址可达",
	"no bootstrap endpoints reachable":                                                                        "所有 bootstrap 地址均不可达",
	"configured listeners cannot be evaluated":                                                                "无法评估显式 listener 端点",
	"no explicit listener endpoints were provided":                                                            "当前输入未提供显式 listener 端点",
	"explicit listener endpoints are reachable from the current execution view":                               "从当前执行视角看，显式 listener 端点可达",
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

var v2ExtraText = map[string]string{
	"verify bootstrap endpoints":                                                                                  "核对 bootstrap 地址",
	"verify Kafka metadata requests are served":                                                                   "确认 Kafka metadata 请求可被正常处理",
	"check cluster and listener health":                                                                           "检查集群与 listener 健康状态",
	"verify probe topic exists":                                                                                   "确认探针主题存在",
	"verify produce path and acks":                                                                                "检查生产链路与 acks 配置",
	"check ISR and leader health":                                                                                 "检查 ISR 与 leader 健康状态",
	"verify topic leader and offsets":                                                                             "检查主题 leader 与位点状态",
	"verify fetch path from current client network":                                                               "检查当前客户端网络到 fetch 链路的可达性",
	"check consumer side timeouts":                                                                                "检查消费端超时配置",
	"verify __consumer_offsets health":                                                                            "检查 __consumer_offsets 健康状态",
	"verify coordinator health":                                                                                   "检查 coordinator 健康状态",
	"check controller and internal topic replicas":                                                                "检查 controller 与内部主题副本状态",
	"check the failing stage first":                                                                               "优先检查失败阶段本身",
	"verify probe topic and broker reachability":                                                                  "检查探针主题与 broker 可达性",
	"correlate with network, ISR and controller checks":                                                           "结合网络、ISR 和 controller 检查一起判断",
	"verify listener binding addresses":                                                                           "核对 listener 绑定地址",
	"verify firewall and port exposure":                                                                           "检查防火墙与端口暴露",
	"compare the failing endpoints with compose and profile settings":                                             "对照 compose 和 profile 配置核查失败端点",
	"verify controller listener reachability between quorum voters":                                               "检查 quorum voter 之间的 controller listener 可达性",
	"check controller processes and recent election errors":                                                       "检查 controller 进程与最近的选举错误",
	"confirm controller.quorum.voters still reflects the active topology":                                         "确认 controller.quorum.voters 仍与当前拓扑一致",
	"stabilize the unreachable controller listeners before another failure happens":                               "在再次故障前先恢复不可达的 controller listener",
	"compare reachability from the current execution host and broker host":                                        "对比当前执行主机与 broker 宿主机的可达性差异",
	"inspect controller logs for intermittent network or append failures":                                         "检查 controller 日志中的间歇性网络或追加失败",
	"run kdoctor from the Kafka internal network or host":                                                         "从 Kafka 内网或宿主机侧执行 kdoctor",
	"verify controller listeners from the broker host":                                                            "在 broker 宿主机上验证 controller listener",
	"use metadata and broker health as temporary reference":                                                       "暂时以 metadata 和 broker 健康状态作为参考",
	"verify controller listeners are reachable":                                                                   "确认 controller listener 可达",
	"verify broker-controller processes are healthy":                                                              "确认 broker-controller 进程健康",
	"verify quorum voter configuration":                                                                           "检查 quorum voter 配置",
	"verify broker processes are listening on the expected ports":                                                 "确认 broker 进程监听了预期端口",
	"check docker host network or service binding":                                                                "检查 Docker host network 或服务绑定状态",
	"compare compose listener settings with the active process state":                                             "对照 compose listener 配置与当前进程实际状态",
	"verify broker replication pipeline":                                                                          "检查 broker 复制链路",
	"check affected broker health and disks":                                                                      "检查受影响 broker 的健康与磁盘状态",
	"check controller and partition movement":                                                                     "检查 controller 状态与分区迁移",
	"verify follower brokers are healthy":                                                                         "检查 follower broker 健康状态",
	"monitor ISR recovery before write pressure increases":                                                        "在写入压力升高前持续观察 ISR 恢复情况",
	"verify replica health":                                                                                       "检查副本健康状态",
	"check affected brokers and disks":                                                                            "检查受影响 broker 与磁盘",
	"reduce write pressure until ISR recovers":                                                                    "在 ISR 恢复前降低写入压力",
	"raise ulimit -n for Kafka and the execution environment":                                                     "提高 Kafka 与当前执行环境的 ulimit -n",
	"inspect current file descriptor growth and socket churn":                                                     "检查当前文件描述符增长与连接抖动",
	"verify recent connection spikes did not exhaust shared host limits":                                          "确认最近连接峰值没有耗尽宿主机共享 fd 上限",
	"review ulimit -n and current descriptor pressure before traffic increases":                                   "在流量升高前复核 ulimit -n 与当前描述符压力",
	"check whether connection churn or client retries are inflating descriptor usage":                             "检查连接抖动或客户端重试是否放大了 fd 使用量",
	"reserve more fd headroom for Kafka data and network workloads":                                               "为 Kafka 数据与网络负载预留更多 fd 余量",
	"raise ulimit -n for the Kafka service user":                                                                  "提高 Kafka 服务用户的 ulimit -n",
	"confirm the broker process inherits the intended soft and hard limits":                                       "确认 broker 进程继承了预期的软硬限制",
	"review listener and client connection fan-out before load grows":                                             "在负载增长前复核 listener 和客户端连接扇出",
	"compare ss or netstat output with Kafka listener configuration":                                              "对照 ss 或 netstat 输出与 Kafka listener 配置",
	"check whether the broker process bound a different port or address than expected":                            "检查 broker 进程是否绑定了非预期端口或地址",
	"confirm Docker host-network and listener exposure match the current runtime state":                           "确认 Docker host-network 与 listener 暴露状态匹配当前运行态",
	"verify compose service names and container_names":                                                            "检查 compose 服务名与 container_name 配置",
	"start the missing containers":                                                                                "启动缺失的容器",
	"check whether the execution host is the intended Docker host":                                                "确认当前执行主机就是目标 Docker 宿主机",
	"restart the stopped containers":                                                                              "重启已停止的容器",
	"inspect docker logs for startup failures":                                                                    "检查 docker 日志中的启动失败信息",
	"verify host resources and port conflicts":                                                                    "检查宿主机资源与端口冲突",
	"review container memory limits and heap size":                                                                "检查容器内存限制与 JVM 堆大小",
	"inspect broker logs for memory pressure":                                                                     "检查 broker 日志中的内存压力迹象",
	"reduce load or restart carefully after adding headroom":                                                      "补足资源余量后再谨慎降载或重启",
	"verify advertised.listeners":                                                                                 "检查 advertised.listeners 配置",
	"verify returned broker ports are exposed":                                                                    "确认返回的 broker 端口已暴露",
	"verify routing from current client network":                                                                  "确认当前客户端网络到 broker 的路由可达",
	"compose Kafka services are not available for runtime mount validation":                                       "当前没有可用的 compose Kafka 服务，无法校验运行时挂载预期",
	"docker inspect mounts match the expected Kafka storage paths":                                                "docker inspect 挂载结果与预期 Kafka 存储路径一致",
	"some Kafka storage paths are not mounted in the current docker runtime view":                                 "当前 docker 运行时视图下，部分 Kafka 存储路径没有正确挂载",
	"Kafka storage mounts exist but part of the runtime mount set is read-only or unusual":                        "Kafka 存储挂载已存在，但部分运行时挂载为只读或状态异常",
	"replica lag metrics are not available in the current JMX sources":                                            "当前 JMX 来源里没有可用的副本 lag 指标",
	"replica fetcher lag metrics look healthy in the current JMX window":                                          "当前 JMX 窗口内副本抓取 lag 指标正常",
	"replica fetcher lag is elevated and may soon erode ISR safety":                                               "副本抓取 lag 已升高，可能很快侵蚀 ISR 安全边界",
	"network idle metrics are not available in the current JMX sources":                                           "当前 JMX 来源里没有可用的网络 idle 指标",
	"network idle metrics still show healthy headroom":                                                            "网络 idle 指标仍显示有健康余量",
	"network idle metrics are critically low and already suggest broker-side saturation":                          "网络 idle 指标已降到危险水平，说明 broker 侧可能已接近饱和",
	"network idle metrics are getting low and broker network headroom is shrinking":                               "网络 idle 指标正在走低，broker 网络余量正在缩小",
	"request handler idle metrics are not available in the current JMX sources":                                   "当前 JMX 来源里没有可用的请求处理线程 idle 指标",
	"request handler idle metrics still show healthy processing headroom":                                         "请求处理线程 idle 指标仍显示有健康处理余量",
	"request handler idle metrics are critically low and broker request threads are close to saturation":          "请求处理线程 idle 指标已降到危险水平，broker 请求线程接近饱和",
	"request handler idle metrics are getting low and broker processing headroom is shrinking":                    "请求处理线程 idle 指标正在走低，broker 处理余量正在缩小",
	"host disk and inode evidence is not available in the current input mode":                                     "当前输入模式下没有可用的宿主机磁盘与 inode 证据",
	"host disk and inode headroom for Kafka paths looks acceptable":                                               "Kafka 宿主机路径的磁盘与 inode 余量看起来仍然可接受",
	"some Kafka host paths are critically close to disk exhaustion":                                               "部分 Kafka 宿主机路径已非常接近磁盘耗尽",
	"host disk or inode headroom for Kafka paths is getting tight":                                                "Kafka 宿主机路径的磁盘或 inode 余量开始变紧",
	"compare docker inspect mounts with compose volume declarations":                                              "对照 docker inspect 挂载结果与 compose volume 声明",
	"bind-mount Kafka data and metadata directories explicitly":                                                   "为 Kafka 数据目录和 metadata 目录显式配置 bind mount",
	"avoid leaving Kafka state only inside container layers":                                                      "避免只把 Kafka 状态保存在容器层内",
	"confirm Kafka data and metadata directories are mounted read-write":                                          "确认 Kafka 数据目录和 metadata 目录都是读写挂载",
	"check whether the current container mount policy matches the intended persistence design":                    "确认当前容器挂载策略与预期持久化设计一致",
	"review recent container recreation or host-path changes":                                                     "检查最近是否发生过容器重建或宿主机路径变更",
	"check follower broker disk and network pressure":                                                             "检查 follower broker 的磁盘与网络压力",
	"correlate replica lag with ISR and under-replicated partition signals":                                       "结合 ISR 与 UnderReplicatedPartitions 信号一起判断副本 lag",
	"review whether the current write burst is outrunning replica catch-up capacity":                              "检查当前写入突发是否已经超过副本追赶能力",
	"check connection churn and listener traffic concentration":                                                   "检查连接抖动和 listener 流量是否过度集中",
	"correlate network idle with request latency and quota/backpressure signals":                                  "结合网络 idle、请求延迟和 quota/backpressure 信号一起判断",
	"verify the current route design is not funneling all traffic through a hot broker":                           "确认当前路由设计没有把流量集中打到单个热点 broker",
	"watch network idle over a longer window":                                                                     "在更长时间窗口内持续观察 network idle",
	"check whether the current traffic spike or client fan-out is sustainable":                                    "检查当前流量高峰或客户端扇出是否可持续",
	"review listener routing and load distribution before pressure worsens":                                       "在压力进一步恶化前复核 listener 路由和负载分布",
	"correlate request idle with queue time, purgatory, and replica pressure":                                     "结合队列时间、purgatory 和副本压力一起判断 request idle",
	"check whether recent traffic or rebalance storms are saturating broker handlers":                             "检查最近的流量高峰或 rebalance 风暴是否压满了 broker 处理线程",
	"review disk, ISR, and GC pressure before scaling or rerouting traffic":                                       "在扩容或重分流前先复核磁盘、ISR 和 GC 压力",
	"watch request idle over the next peak window":                                                                "在下一个高峰窗口继续观察 request idle",
	"correlate handler idle with request latency and quota pressure":                                              "结合处理线程 idle、请求延迟和 quota 压力一起判断",
	"review whether hot partitions or hot brokers are concentrating request load":                                 "检查是否有热点分区或热点 broker 正在集中承载请求",
	"free disk space immediately on the affected host path":                                                       "立即释放受影响宿主机路径的磁盘空间",
	"review inode and retention growth before the next write peak":                                                "在下一次写入高峰前复核 inode 与保留策略增长情况",
	"check whether the host-level mount or filesystem is already impairing broker writes":                         "检查宿主机挂载点或文件系统是否已经影响 broker 写入",
	"plan host-level cleanup or capacity expansion":                                                               "规划宿主机层面的清理或容量扩展",
	"review inode usage on Kafka data and metadata directories":                                                   "复核 Kafka 数据目录与 metadata 目录的 inode 使用情况",
	"track which broker path is growing fastest before it becomes a write outage":                                 "找出增长最快的 broker 路径，提前处理避免变成写入故障",
	"network snapshot is not available":                                                                           "当前没有可用的网络快照",
	"no hostname-based endpoints are available for DNS drift analysis":                                            "当前没有基于主机名的端点，无法评估 DNS 漂移",
	"hostname resolution is broadly consistent with the current Kafka route view":                                 "主机名解析结果与当前 Kafka 路由视图基本一致",
	"DNS resolution differs from the current metadata route view and may indicate stale records or split routing": "DNS 解析结果与当前 metadata 路由视图不一致，可能存在旧记录或分流路由问题",
	"some Kafka hostnames resolve to multiple addresses; verify that all returned routes are intentional":         "部分 Kafka 主机名解析到多个地址，请确认这些返回路由都是有意设计的",
	"controller quorum majority evidence is healthy":                                                              "controller quorum 的多数派证据正常",
	"controller quorum does not currently have majority evidence":                                                 "controller quorum 当前缺少多数派证据",
	"different JMX endpoints report different controller epoch or leader values":                                  "不同 JMX 端点报告了不一致的 controller epoch 或 leader 值",
	"controller epoch and leader view are stable across the current JMX endpoints":                                "当前各个 JMX 端点的 controller epoch 与 leader 视图一致",
	"some SSL listeners failed certificate validation, SAN matching, or expiry checks":                            "部分 SSL listener 的证书校验、SAN 匹配或到期检查失败",
	"TLS certificate chain and expiry look healthy for the current listener set":                                  "当前 listener 集合的 TLS 证书链与到期时间正常",
	"some SSL listener certificates are approaching expiry":                                                       "部分 SSL listener 证书即将过期",
	"producer throttle time is above zero and may already be affecting write latency":                             "producer throttle time 已大于 0，可能已经影响写入延迟",
	"produce throttle time is above zero and may already be limiting write throughput":                            "produce throttle time 已大于 0，可能已经限制写入吞吐",
	"fetch throttle time is above zero and may already be slowing consumers":                                      "fetch throttle time 已大于 0，可能已经拖慢消费端",
	"request percentage quota is saturated and may already be throttling client requests":                         "request percentage quota 已接近或达到饱和，可能已经在限流客户端请求",
	"request percentage quota usage is close to saturation":                                                       "request percentage quota 使用率接近饱和",
	"broker idle headroom or request latency already suggests backpressure":                                       "broker 空闲余量或请求延迟已经显示出背压迹象",
	"request latency or purgatory backlog is elevated and may already be contributing to broker pressure":         "请求延迟或 purgatory 堆积升高，可能已经造成 broker 压力",
	"heap usage or GC pause metrics indicate rising JVM pressure":                                                 "heap 使用率或 GC pause 指标显示 JVM 压力正在上升",
	"host file descriptor headroom is critically low":                                                             "宿主机文件描述符余量已经非常紧张",
	"host file descriptor headroom is getting tight":                                                              "宿主机文件描述符余量开始变紧",
	"clock skew between hosts is larger than expected and can affect SSL, logs, and transaction timing":           "主机之间的时钟偏移大于预期，可能影响 SSL、日志时序与事务时间",
	"some expected Kafka listener ports are missing from the host listening table":                                "部分预期 Kafka listener 端口未出现在宿主机监听表中",
	"host memory pressure is high and may amplify JVM or container instability":                                   "宿主机内存压力较高，可能放大 JVM 或容器不稳定问题",
	"at least one Kafka storage path is critically close to disk exhaustion":                                      "至少有一个 Kafka 存储路径已经非常接近磁盘耗尽",
	"Kafka storage headroom is getting tight on disk space or inodes":                                             "Kafka 存储在磁盘空间或 inode 上的余量开始变紧",
	"transaction logs already show commit, abort, or coordinator-side transaction errors":                         "当前日志已经出现事务提交、中止或协调器侧事务错误",
}

var fragmentReplacer = strings.NewReplacer(
	"failure_stage=metadata", "失败阶段=metadata",
	"failure_stage=produce", "失败阶段=生产",
	"failure_stage=consume", "失败阶段=消费",
	"failure_stage=commit", "失败阶段=提交位点",
	"failure_stage=topic_ready", "失败阶段=探针主题就绪检查",
	"failure_stage=context", "失败阶段=上下文取消",
	"executed_stage=bootstrap", "已执行到=bootstrap",
	"executed_stage=metadata", "已执行到=metadata",
	"executed_stage=topic_ready", "已执行到=探针主题就绪检查",
	"executed_stage=produce", "已执行到=生产",
	"executed_stage=consume", "已执行到=消费",
	"executed_stage=commit", "已执行到=提交位点",
	"executed_stage=complete", "已执行到=完成",
	"bootstrap=", "bootstrap=",
	"topic=", "主题=",
	"topic_ready_reason=", "主题就绪说明=",
	"topic_created=", "已创建主题=",
	"cleanup_attempted=", "已尝试清理=",
	"cleanup_ok=", "清理成功=",
	"cleanup_error=", "清理错误=",
	"produce_count=", "已生产消息数=",
	"commit_executed=", "提交阶段已执行=",
	"commit_ok=", "提交阶段成功=",
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
	" unreachable 错误=", " 不可达 错误=",
	"create kafka client:", "创建 Kafka 客户端失败：",
	"kafka: client has run out of available brokers to talk to:", "Kafka 客户端已耗尽可用 broker：",
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
	if translated, ok := v2ExtraText[input]; ok {
		return translated
	}

	output := fragmentReplacer.Replace(input)
	for oldValue, newValue := range exactText {
		if strings.Contains(output, oldValue) {
			output = strings.ReplaceAll(output, oldValue, newValue)
		}
	}
	for oldValue, newValue := range v2ExtraText {
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
	if module == "quota" {
		return "閰嶉"
	}
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
	case "consumer":
		return "消费组"
	case "security":
		return "安全"
	case "storage":
		return "存储"
	case "metrics":
		return "指标"
	case "jvm":
		return "JVM"
	case "producer":
		return "生产者"
	case "transaction":
		return "事务"
	case "upgrade":
		return "升级"
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
