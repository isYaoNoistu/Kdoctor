# Kdoctor 检查基线文档

本文件用于回答三个问题：

1. `Kdoctor` 当前到底会检查什么。
2. 每个检查项默认在什么条件下运行。
3. 阈值、告警线、失败线、判断依据分别是什么。

本文档以当前仓库实现为准，源码主入口在：

- `internal/runner/collect.go`：检查注册链路
- `internal/config/defaults.go`：默认阈值与默认值
- `internal/config/config.go`：可配置项定义
- `internal/checks/**`：单项检查实现

## 1. 使用说明

### 1.1 这份文档怎么读

- “默认注册”表示当前封版主链路里会注册，但最终是否真正执行，还取决于当前输入模式和证据是否可用。
- “条件注册”表示只有在给了对应输入后才会注册，比如 `compose`、Docker、日志、消费组目标、事务上下文。
- “代码存在但默认不注册”表示代码仓库里还有实现，但当前封版主链路默认不会注册到报告里。

### 1.2 状态语义

- `CRIT`：高危，通常表示强约束已破坏，或者直接影响可用性。
- `FAIL`：明确异常或高风险。
- `WARN`：提示、规划风险、证据质量不足、或上下文不理想。
- `ERROR`：执行检查时本身出错，无法完成判断。
- `PASS`：当前证据下未见问题。
- `SKIP`：本次运行没有纳入该检查，或缺少所需证据来源。

### 1.3 判定依据的优先级

`Kdoctor` 的检查依据来自以下几层，实际以能拿到的证据为准：

1. Kafka 元数据与 Admin/协议侧证据
2. `kdoctor.yaml` 里的期望基线
3. `docker-compose.yml` 静态配置
4. Docker 运行时
5. 宿主机信息
6. 日志来源
7. 探针链路 `metadata -> produce -> consume -> commit -> e2e`

## 2. 全局默认阈值

这些默认值来自 `internal/config/defaults.go`，可以被 `kdoctor.yaml` 覆盖。

### 2.1 日志

| 配置项 | 默认值 | 用途 |
| --- | --- | --- |
| `logs.tail_lines` | `500` | 每个日志来源最多拉取的行数 |
| `logs.lookback_minutes` | `30` | 日志时间窗口 |
| `logs.min_lines_per_source` | `20` | 单个来源样本充分的最低行数 |
| `logs.freshness_window` | `30m` | 判定“日志是否足够新鲜”的窗口 |
| `logs.max_files` | `12` | 最多处理的文件数 |
| `logs.max_bytes_per_source` | `1048576` | 单个来源最大字节数 |

### 2.2 探针

| 配置项 | 默认值 | 用途 |
| --- | --- | --- |
| `probe.topic` | `_kdoctor_probe` | 探针主题 |
| `probe.timeout` | `15s` | 探针阶段超时 |
| `probe.message_bytes` | `1024` | 探针消息大小 |
| `probe.produce_count` | `1` | 单次发送消息数 |
| `probe.acks` | `all` | 探针生产确认级别 |
| `probe.enable_idempotence` | `false` | 探针是否启用幂等 |
| `probe.cleanup_mode` | `disabled` | 探针后清理策略 |

### 2.3 执行超时

| 配置项 | 默认值 |
| --- | --- |
| `execution.timeout` | `30s` |
| `execution.metadata_timeout` | `5s` |
| `execution.tcp_timeout` | `3s` |
| `execution.admin_api_timeout` | `15s` |
| `execution.jmx_timeout` | `5s` |

### 2.4 宿主机与通用阈值

| 配置项 | 默认值 | 主要使用方 |
| --- | --- | --- |
| `host.fd_warn_pct` | `70` | `HOST-008` |
| `host.fd_crit_pct` | `85` | `HOST-008` |
| `host.clock_skew_warn_ms` | `500` | `HOST-009` |
| `thresholds.urp_warn` | `1` | `MET-001` |
| `thresholds.under_min_isr_crit` | `1` | `MET-002` |
| `thresholds.network_idle_warn` | `0.3` | `MET-005` / `JVM-001` / `QTA-004` |
| `thresholds.request_idle_warn` | `0.3` | `MET-006` / `JVM-002` |
| `thresholds.disk_warn_pct` | `75` | `HOST-004` / `HOST-007` / `STG-001` |
| `thresholds.disk_crit_pct` | `85` | `HOST-004` / `HOST-007` / `STG-001` |
| `thresholds.inode_warn_pct` | `80` | `HOST-007` / `STG-001` |
| `thresholds.replica_lag_warn` | `10000` | `TOP-010` / `MET-003` |
| `thresholds.leader_skew_warn_pct` | `30` | `TOP-009` |
| `thresholds.consumer_lag_warn` | `1000` | `CSM-001` |
| `thresholds.consumer_lag_crit` | `10000` | `CSM-001` |
| `thresholds.cert_expiry_warn_days` | `30` | `SEC-003` |
| `thresholds.produce_throttle_warn_ms` | `1` | `PRD-005` / `QTA-001` |
| `thresholds.fetch_throttle_warn_ms` | `1` | `QTA-002` |
| `thresholds.request_latency_warn_ms` | `100` | `JVM-003` / `QTA-004` |
| `thresholds.purgatory_warn_count` | `1` | `JVM-003` |
| `thresholds.heap_used_warn_pct` | `85` | `JVM-004` |
| `thresholds.gc_pause_warn_ms` | `200` | `JVM-004` |

### 2.5 输出与摘要

| 配置项 | 默认值 | 用途 |
| --- | --- | --- |
| `diagnosis.max_root_causes` | `3` | 根因摘要最多展示 3 条 |
| `output.max_evidence_items` | `8` | 单检查展示的证据上限 |
| `output.show_pass_checks` | `false` | 终端默认不展开 PASS |
| `output.show_skip_checks` | `false` | 终端默认不展开 SKIP |
| `output.verbose` | `false` | 终端默认非详细模式 |

## 3. 检查注册总览

### 3.1 默认主链路总是注册

- 网络：`NET-001~009`
- Kafka：`KFK-001~009`
- KRaft：`KRF-002~004`
- Topic：`TOP-003~011`
- Producer 审计：`PRD-001~004`、`PRD-006`
- 客户端探针：`CLI-001~005`

### 3.2 条件注册

| 条件 | 注册模块 |
| --- | --- |
| 提供 `compose` 或显式 `controller_endpoints` | `KRF-001` |
| 提供 `compose` | `KRF-005`、`CFG-001~014`、`SEC-001~003`、`SEC-005`、`STG-003`、`STG-005`、`STG-006` |
| 有 `compose` / 日志 / probe | `SEC-004` |
| 配置了 `group_probe_targets` | `CSM-001~006` |
| 开启事务上下文 | `TXN-001~005` |
| 宿主机证据可用 | `HOST-004`、`HOST-006~011`、`STG-001` |
| Docker 运行时证据可用 | `DKR-001~007` |
| 日志输入启用 | `LOG-001` |
| 日志来源实际可用 | `LOG-002~008` |

### 3.3 代码存在但当前封版默认不注册

这批检查在代码里仍然存在，但当前默认主链路不纳入报告：

- KRaft：`KRF-006`、`KRF-007`、`KRF-008`
- Producer：`PRD-005`
- Host：`HOST-009`
- Storage：`STG-002`、`STG-004`
- Upgrade：`UPG-001`、`UPG-002`、`UPG-003`
- Metrics / JVM / Quota：`MET-*`、`JVM-*`、`QTA-*`

## 4. 默认主链路检查清单

### 4.1 网络检查 `NET-*`

| 编号 | 检查内容 | 默认触发条件 | 触发阈值 / 判定规则 | 主要依据 |
| --- | --- | --- | --- | --- |
| `NET-001` | bootstrap 连通性 | 只要有 bootstrap | 任一 bootstrap 可达则 `PASS`；全部不可达则 `FAIL` | `bundle.Network.BootstrapChecks` |
| `NET-002` | 显式 listener / controller 端点连通性 | 提供显式 listener 或 controller 端点时 | 显式端点不可达则 `WARN/FAIL`；外部视角打不到内网 controller 时偏向 `WARN/SKIP` | profile / compose listener + TCP 探测 |
| `NET-003` | metadata 返回 broker 端点连通性 | Kafka metadata 可用 | metadata 返回端点中只要有不可达项就异常；已做端点去重 | `bundle.Network.MetadataChecks` |
| `NET-004` | DNS 解析检查 | 目标里包含主机名 | DNS 解析失败则 `FAIL`；纯 IP 场景 `SKIP` | DNS lookup 结果 |
| `NET-005` | 路由错配 / 可达路径不一致 | bootstrap 与 metadata 都有证据 | bootstrap 可达但 metadata 返回路由不可达，按影响范围 `WARN/FAIL` | bootstrap 路径 + metadata 路径 |
| `NET-006` | 外部视角下返回私网地址 | 外部执行视角 | metadata 返回私网 broker 地址则 `FAIL` | metadata broker endpoint |
| `NET-007` | bootstrap 负载均衡错觉 | 看到 bootstrap 与 metadata 路由差异时 | bootstrap 像单入口 LB，但后续 broker 路由不一致且不可达时 `WARN` | bootstrap 与 metadata 路由对比 |
| `NET-008` | DNS 漂移 | 存在主机名解析 | 主机名解析集与 metadata 视图漂移时 `WARN` | DNS A 记录与 metadata 路由 |
| `NET-009` | 协议不匹配 | TCP 可达但 Kafka 不通 | TCP 能连但拿不到 metadata，则 `FAIL` | TCP 成功 + metadata 失败 |

### 4.2 Kafka 元数据与 broker 检查 `KFK-*`

| 编号 | 检查内容 | 默认触发条件 | 触发阈值 / 判定规则 | 主要依据 |
| --- | --- | --- | --- | --- |
| `KFK-001` | metadata 可用性 | 总是注册 | 能成功拉到 metadata 则 `PASS`，否则 `FAIL` | Kafka metadata |
| `KFK-002` | broker 注册数量 | metadata 可用 | 实际 broker 数 `< profile.broker_count` 则 `FAIL` | metadata broker 列表 |
| `KFK-003` | broker 端点合法性 | metadata 可用 | 端点格式错误、重复、外部视角下是私网地址则异常 | metadata broker endpoint |
| `KFK-004` | 内部主题健康 | metadata 可用 | `__consumer_offsets` 缺失或 unhealthy 会 `WARN/FAIL`；事务未使用时 `__transaction_state` 缺失不直接判错 | topic metadata + probe 上下文 |
| `KFK-005` | metadata 返回路由可达性 | metadata 路由已采集 | 返回路由不可达则 `WARN/FAIL`；证据已去重 | metadata 路由探测 |
| `KFK-006` | broker 注册完整性 | metadata 可用 | 注册数低于期望则 `FAIL`；compose 期望 node.id 与 metadata 集合漂移时 `WARN` | metadata + compose |
| `KFK-007` | broker 身份唯一性 | metadata 可用 | broker 地址重复 / 身份冲突则 `FAIL`；compose 身份声明不完整时 `WARN` | metadata broker 集合 + compose |
| `KFK-008` | metadata 延迟 | metadata 采集到耗时 | `>=500ms` 告警，`>=2000ms` 失败 | metadata 延迟 |
| `KFK-009` | 拓扑一致性 | compose + metadata 同时可用时最有价值 | compose 拓扑与 metadata 拓扑漂移时 `WARN` | compose + metadata |

### 4.3 KRaft 检查 `KRF-*`

| 编号 | 检查内容 | 默认触发条件 | 触发阈值 / 判定规则 | 主要依据 |
| --- | --- | --- | --- | --- |
| `KRF-001` | quorum 配置一致性 | 提供 `compose` 或显式 `controller_endpoints` | `controller.quorum.voters` 在服务间不一致或与显式端点冲突时异常 | compose / profile |
| `KRF-002` | 活动 controller 状态 | metadata 可用 | 没有 active controller 或 controller 不在 broker 集合中则 `FAIL`；当前视角打不到 controller listener 可 `WARN` | metadata controller id + 网络探测 |
| `KRF-003` | quorum 可达多数派 | controller 端点可用时 | 可达 controller 数量 `< 多数派` 则 `CRIT`；只到达部分则 `WARN`；外部视角全是私网 controller 可 `SKIP` | controller endpoint reachability |
| `KRF-004` | controller 多数派证据 | controller 端点可用时 | 多数派不足则 `CRIT`，有多数派但不完整则 `WARN` | controller endpoint reachability |
| `KRF-005` | active controller 与 controller listener 配置一致性 | 需要 `compose` | 现在按 `controller id -> compose service -> CONTROLLER listener -> voters` 链路比对；配置映射异常才告警，不再把 broker listener 误判为 quorum 异常 | metadata controller id + compose listener/voter 配置 |

### 4.4 Topic / ISR / 规划检查 `TOP-*`

| 编号 | 检查内容 | 默认触发条件 | 触发阈值 / 判定规则 | 主要依据 |
| --- | --- | --- | --- | --- |
| `TOP-003` | leader 状态 | topic metadata 可用 | 任一分区无 leader 则 `FAIL` | topic metadata |
| `TOP-004` | 副本与 ISR 健康 | topic metadata 可用 | 空 ISR 直接 `FAIL`；副本未完全同步会 `WARN` | partition replicas / ISR |
| `TOP-005` | ISR 与 `min.insync.replicas` | topic metadata 可用 | `ISR < minISR` 则 `FAIL`；`ISR < replicas` 但仍 `>= minISR` 则 `WARN` | topic metadata + 期望 `minISR` |
| `TOP-006` | under-replicated partitions | topic metadata 可用 | 计数 `>=1` 即 `WARN` | URP 计数 |
| `TOP-007` | under-min-isr 风险 | topic metadata 可用 | `ISR < minISR` 直接 `FAIL`；`ISR == minISR` 且命中分区数 `>=1` 则 `WARN` | ISR 与 `minISR` |
| `TOP-008` | offline replicas | topic metadata 可用 | 存在 offline replica 或 leaderless 分区则 `FAIL` | topic metadata |
| `TOP-009` | leader 分布倾斜 | topic metadata 可用 | 最热 broker 的 leader 数 `> 平均值 * (1 + 30%)` 时 `WARN`；阈值来自 `thresholds.leader_skew_warn_pct` | partition leader 分布 |
| `TOP-010` | replica lag 指标 | 代码存在，但依赖 JMX 指标 | 默认阈值 `>=10000` `WARN`；当前封版默认不注册 | metrics snapshot |
| `TOP-011` | topic 规划 | topic metadata + broker 基线 | `RF > broker_count` 则 `FAIL`；`partition_count < broker_count` 则 `WARN`；只输出真正命中的 topic，证据最多 20 条 | topic 元数据 + broker 数基线 |

### 4.5 Producer 审计 `PRD-*`

| 编号 | 检查内容 | 默认触发条件 | 触发阈值 / 判定规则 | 主要依据 |
| --- | --- | --- | --- | --- |
| `PRD-001` | `acks` 与 durability | 提供 producer 基线时最有意义 | `expected_durability=strong` 但 `acks != all` 则 `FAIL`；`acks=all` 且 `minISR <= 1` 则 `WARN` | `profiles.*.producer` + Kafka/Topic 基线 |
| `PRD-002` | 幂等性风险 | 提供 producer 基线时最有意义 | `enable_idempotence=false` 且 `retries>0` 且 `max_in_flight>1` 则 `WARN` | `profiles.*.producer` |
| `PRD-003` | 生产超时关系 | 提供 producer 基线时最有意义 | `delivery.timeout.ms < request.timeout.ms + linger.ms` 则 `FAIL` | `profiles.*.producer` |
| `PRD-004` | 消息大小与 broker 上限 | 有消息大小基线或日志证据时 | 探针消息大小超过最小 `message.max.bytes` 或命中 `LOG-MESSAGE-TOO-LARGE` 则 `FAIL` | broker 配置 + 日志 |
| `PRD-005` | 生产 throttle | 代码存在，默认不注册 | `produce_throttle_ms >= 1` 则 `WARN` | 指标/JMX |
| `PRD-006` | 事务超时上限 | 事务上下文时更重要 | `transaction.timeout.ms > broker transaction.max.timeout.ms` 则 `FAIL` | producer 配置 + broker 配置 |

### 4.6 客户端探针 `CLI-*`

| 编号 | 检查内容 | 默认触发条件 | 触发阈值 / 判定规则 | 主要依据 |
| --- | --- | --- | --- | --- |
| `CLI-001` | metadata 探针 | `probe` 模式 | metadata 阶段成功则 `PASS`，失败则按探针错误 `FAIL` | probe snapshot |
| `CLI-002` | produce 探针 | `probe` 模式 | produce 阶段成功则 `PASS`，失败则 `FAIL` | probe snapshot |
| `CLI-003` | consume 探针 | `probe` 模式 | consume 阶段成功则 `PASS`，失败则 `FAIL` | probe snapshot |
| `CLI-004` | commit 探针 | `probe` 模式 | commit 阶段成功则 `PASS`，失败则 `FAIL` | probe snapshot |
| `CLI-005` | e2e 探针 | `probe` 模式 | 整条链路成功则 `PASS`，任一关键阶段失败则 `FAIL` | probe snapshot |

## 5. 条件注册检查清单

### 5.1 Compose / 配置静态审计 `CFG-*`

只有提供 `docker-compose.yml` 或显式 compose 输入时才注册。

| 编号 | 检查内容 | 触发阈值 / 判定规则 | 主要依据 |
| --- | --- | --- | --- |
| `CFG-001` | compose 解析 | compose 无法解析则失败 | compose 文件 |
| `CFG-002` | `node.id` 唯一性 | 有重复 `node.id` 则 `FAIL` | compose 环境变量 |
| `CFG-003` | `cluster.id` 一致性 | broker 间 `cluster.id` 不一致则 `FAIL` | compose 环境变量 |
| `CFG-004` | `process.roles` 合法性 | 非法角色或角色缺失则 `FAIL` | compose 环境变量 |
| `CFG-005` | `controller.quorum.voters` 一致性 | 服务间 voters 不一致或格式非法则 `FAIL` | compose 环境变量 |
| `CFG-006` | `listeners / advertised.listeners` | 端口冲突、缺失 client-facing advertised listener、`advertised.listeners` 使用 `0.0.0.0` 均为 `FAIL` | compose listeners |
| `CFG-007` | `inter.broker.listener.name` | 缺失、未出现在 listeners 中则 `FAIL`；若为 `EXTERNAL` 则 `WARN` | compose 环境变量 |
| `CFG-008` | 副本与 ISR 合法性 | `RF > broker_count`、`txn.min.isr > txn RF`、`minISR > default RF` 为 `FAIL`；`minISR == default RF` 为 `WARN` | compose 环境变量 + broker 数 |
| `CFG-009` | advertised listener 与执行视角匹配 | 内外网视角与 `advertised.listeners` 不匹配时 `WARN/FAIL` | compose listeners + execution view |
| `CFG-010` | controller listener 映射 | `controller.listener.names` 与实际 listener 不匹配时异常 | compose listeners |
| `CFG-011` | broker 身份唯一性 | broker 身份声明冲突时异常 | compose |
| `CFG-012` | profile 与 compose 拓扑一致性 | 期望 broker/controller 数与 compose 实际不一致时 `WARN` | `kdoctor.yaml` + compose |
| `CFG-013` | 默认 topic 规划 | `default RF > broker_count` 或 `minISR > default RF` 为 `FAIL`；默认 partitions `< broker_count` 为 `WARN` | compose 环境变量 |
| `CFG-014` | `metadata.log.dir` 规划 | metadata 目录无清晰 volume 承载则 `FAIL/WARN` | compose volume + metadata dir |

### 5.2 消费组检查 `CSM-*`

只有配置了 `profiles.<name>.group_probe_targets` 才注册。

| 编号 | 检查内容 | 默认阈值 / 判定规则 | 主要依据 |
| --- | --- | --- | --- |
| `CSM-001` | 消费组 lag | 默认 `warn=1000`、`crit=10000`；也可在目标级别用 `lag_warn`、`lag_crit` 覆盖 | 消费组采集结果 |
| `CSM-002` | rebalance / group state | 状态为 `Dead` 则 `FAIL`；状态包含 `rebalance` 则 `WARN` | group state |
| `CSM-003` | `max.poll.interval.ms` | `<60000` 则 `WARN` | consumer 配置 |
| `CSM-004` | heartbeat 与 session timeout | `heartbeat.interval.ms > session.timeout.ms / 3` 则 `WARN` | consumer 配置 |
| `CSM-005` | `auto.offset.reset` | `latest` 或 `none` 视为 `WARN`；其他显式值 `PASS` | consumer 配置 |
| `CSM-006` | coordinator / 位点完整性 | coordinator 缺失则 `FAIL`；`missing_offsets > 0` 则 `WARN` | 消费组采集结果 |

### 5.3 事务上下文检查 `TXN-*`

只有存在事务上下文时才注册，触发方式包括：

- `probe.tx_probe_enabled=true`
- `profiles.<name>.producer.transactional_id` 非空
- `profiles.<name>.consumer.isolation_level=read_committed`

| 编号 | 检查内容 | 触发阈值 / 判定规则 | 主要依据 |
| --- | --- | --- | --- |
| `TXN-001` | 事务主题缺失提示 | `__transaction_state` 缺失但当前没有事务使用证据时做上下文提示，不直接判故障 | topic metadata + 事务上下文 |
| `TXN-002` | 事务主题必须存在 | 明确启用事务但 `__transaction_state` 缺失则 `FAIL` | topic metadata + 事务上下文 |
| `TXN-003` | 事务超时是否合法 | `transaction.timeout.ms > broker max timeout` 则 `FAIL` | producer 配置 + broker 配置 |
| `TXN-004` | 隔离级别与事务能力一致性 | `read_committed` 但 tx probe 未启用则 `WARN`；`read_committed` 且事务主题缺失则 `FAIL` | consumer 配置 + topic metadata |
| `TXN-005` | 事务结果证据 | 命中事务故障日志指纹 `LOG-TRANSACTION` 则 `FAIL`；无证据时 `PASS/SKIP` | 日志与事务上下文 |

### 5.4 安全检查 `SEC-*`

| 编号 | 检查内容 | 注册条件 | 默认阈值 / 判定规则 | 主要依据 |
| --- | --- | --- | --- | --- |
| `SEC-001` | listener 安全协议映射 | compose | `security_mode` 与 `listener.security.protocol.map` 不一致则 `FAIL` | compose + profile |
| `SEC-002` | SASL 机制一致性 | compose，且使用 SASL | 未声明 enabled mechanisms 或不包含 `profiles.*.sasl_mechanism` 则 `FAIL` | compose + profile |
| `SEC-003` | TLS 证书健康 | compose，且存在 `SSL/SASL_SSL` listener | 握手失败、证书链/SAN/过期问题 `FAIL`；距离过期 `<30` 天 `WARN` | TLS handshake + 证书 |
| `SEC-004` | 认证/授权拒绝证据 | compose、日志或 probe 任一可用 | 日志命中 `LOG-AUTHORIZATION` / `LOG-AUTHENTICATION`，或 probe 错误包含 authorization，则 `FAIL` | 日志 + probe |
| `SEC-005` | Authorizer 一致性 | compose | 未配置任何 authorizer 为 `WARN`；非 `StandardAuthorizer` 或服务间不一致为 `FAIL` | compose 环境变量 |

### 5.5 存储检查 `STG-*`

| 编号 | 检查内容 | 注册条件 | 默认阈值 / 判定规则 | 主要依据 |
| --- | --- | --- | --- | --- |
| `STG-001` | 磁盘与 inode 容量 | 宿主机证据可用 | 磁盘 `>=85%` `FAIL`，`>=75%` `WARN`；inode `>=80%` `WARN` | host disk usage |
| `STG-003` | `log.dirs / metadata.log.dir` 布局 | compose | 缺失关键目录、布局明显不合理时异常 | compose 环境变量 |
| `STG-005` | 存储挂载规划 | compose | 存储路径无 volume 承载 `FAIL`；使用 named volume `WARN` | compose volume |
| `STG-006` | tiered storage 提示 | compose | 检测到 tiered storage 配置时 `WARN` | compose 环境变量 |

### 5.6 宿主机检查 `HOST-*`

| 编号 | 检查内容 | 注册条件 | 默认阈值 / 判定规则 | 主要依据 |
| --- | --- | --- | --- | --- |
| `HOST-004` | 宿主机磁盘使用率 | host 证据 | 磁盘 `>=85%` `FAIL`，`>=75%` `WARN` | host disk usage |
| `HOST-006` | 宿主机端口可达性 | host 证据 | 期望 listener 端口不通则 `FAIL` | host port checks |
| `HOST-007` | Kafka 相关路径容量 | host 证据 | 磁盘 `>=85%` `FAIL`，`>=75%` `WARN`；inode `>=80%` `WARN` | host disk usage |
| `HOST-008` | 文件描述符余量 | host 或 Docker 证据 | Docker 场景优先取 Kafka 容器 `/proc/1/limits`；容器 `soft_limit <32768` `FAIL`，`soft_limit <65536` `WARN`；若只能看到系统使用率，则 `>=85%` `FAIL`，`>=70%` `WARN` | container fd limit / host fd stats |
| `HOST-010` | listener 漂移 | host 证据 | 期望端口不在宿主机监听表中则 `FAIL` | host listening sockets |
| `HOST-011` | 宿主机内存压力 | host 证据 | `used_pct >=85%` `WARN` | host memory |

### 5.7 Docker 检查 `DKR-*`

| 编号 | 检查内容 | 注册条件 | 默认阈值 / 判定规则 | 主要依据 |
| --- | --- | --- | --- | --- |
| `DKR-001` | 预期容器是否存在 | Docker 可用 | 缺少预期 Kafka 容器则 `FAIL` | Docker inspect / ps |
| `DKR-002` | 容器运行状态 | Docker 可用 | 预期 Kafka 容器未运行则 `FAIL` | Docker inspect / ps |
| `DKR-003` | OOMKilled | Docker 可用 | 任一 Kafka 容器 `OOMKilled=true` 则 `FAIL` | Docker runtime state |
| `DKR-004` | 数据与 metadata 路径是否由 mount 承载 | Docker + compose | Kafka 数据或 metadata 路径没有 docker mount 则 `FAIL` | docker inspect mounts + compose |
| `DKR-005` | 运行时 mount 与预期是否一致 | Docker + compose | 预期存储路径未挂载则 `FAIL`；挂载存在但只读/异常则 `WARN` | docker inspect mounts |
| `DKR-006` | 容器重启历史 | Docker 可用 | 任一 Kafka 容器 `restart_count > 0` 则 `WARN` | Docker restart count |
| `DKR-007` | 容器内存与 JVM 堆规划 | compose | `heap_to_limit_ratio >=0.9` `FAIL`；`>=0.8` `WARN` | `mem_limit` + `KAFKA_HEAP_OPTS` |

### 5.8 日志检查 `LOG-*`

| 编号 | 检查内容 | 注册条件 | 默认阈值 / 判定规则 | 主要依据 |
| --- | --- | --- | --- | --- |
| `LOG-001` | 日志来源与样本质量 | 日志采集启用 | 来源为空则 `SKIP`；有来源但存在 stale / sparse / empty / source warning 则 `WARN` | logs snapshot |
| `LOG-002` | 已知错误指纹 | 日志来源可用 | 未命中指纹 `PASS`；命中则按最高严重级别输出 | builtin/custom pattern matches |
| `LOG-003` | 日志解释 | 日志来源可用 | 命中指纹时输出解释；无命中 `PASS` | fingerprints + meaning |
| `LOG-004` | 重复错误聚合 | 日志来源可用 | 有命中指纹则 `WARN`，用于聚合相同问题 | fingerprints aggregation |
| `LOG-005` | 命中上下文 | 日志来源可用 | 有命中则按最高严重级别输出来源、次数、含义 | fingerprints context |
| `LOG-006` | 日志新鲜度与样本充分性 | 日志来源可用 | 来源 stale / sparse / empty 任一成立则 `WARN` | `lookback_minutes`、`freshness_window`、`min_lines_per_source` |
| `LOG-007` | 指纹风暴 | 日志来源可用 | 默认 `repeat_threshold=5`；某指纹次数 `>=5` 或影响多个来源则 `WARN` | fingerprints repeat count |
| `LOG-008` | 自定义规则库 | 日志采集启用 | 加载到自定义规则则 `PASS`；配置了目录但未成功加载任何规则则 `WARN`；未配置则 `SKIP` | custom pattern library |

## 6. 代码存在但默认不注册的检查

这部分也列出来，方便你们知道代码里“还有什么”，但要注意：它们当前不是封版主链路的一部分。

### 6.1 KRaft 深化 `KRF-006~008`

| 编号 | 检查内容 | 默认阈值 / 依据 |
| --- | --- | --- |
| `KRF-006` | epoch / controller 代际异常 | 依赖更深的控制面证据，当前默认不注册 |
| `KRF-007` | unknown voter | 依赖控制面证据，当前默认不注册 |
| `KRF-008` | finalization / 元数据终结状态 | 依赖控制面证据，当前默认不注册 |

### 6.2 Host / Storage / Upgrade

| 编号 | 检查内容 | 默认阈值 / 依据 |
| --- | --- | --- |
| `HOST-009` | 时钟偏移 | `clock_skew_warn_ms=500`，依赖 JMX/时间证据 |
| `STG-002` | offline log dir | 依赖指标/JMX |
| `STG-004` | partial storage failure | 依赖指标/JMX |
| `UPG-001` | 滚动升级版本混用 | 多 image version 时 `WARN` |
| `UPG-002` | feature / metadata 版本一致性 | 版本不一致或残留旧配置时 `WARN` |
| `UPG-003` | tiered storage 升级提示 | 检测到 tiered storage 配置时 `WARN` |

### 6.3 Metrics / JVM / Quota

| 编号 | 检查内容 | 默认阈值 |
| --- | --- | --- |
| `MET-001` | under-replicated partitions 指标 | `urp_warn=1` |
| `MET-002` | under-min-isr / at-min-isr 指标 | `under_min_isr_crit=1`，at-min-isr `>=1` 告警 |
| `MET-003` | replica lag 指标 | `replica_lag_warn=10000` |
| `MET-004` | offline log dir 指标 | `>=1` 失败 |
| `MET-005` | network idle 指标 | `<=0.3` 告警，`<=0.1` 失败 |
| `MET-006` | request idle 指标 | `<=0.3` 告警，`<=0.1` 失败 |
| `JVM-001` | 网络线程空闲度 | 同 `MET-005` |
| `JVM-002` | 请求处理线程空闲度 | 同 `MET-006` |
| `JVM-003` | 请求压力 / purgatory | `request_latency >=100ms` 或 `purgatory >=1` 告警 |
| `JVM-004` | heap / GC 压力 | `heap_used >=85%` 或 `gc_pause >=200ms` 告警 |
| `QTA-001` | produce throttle | `>=1ms` 告警 |
| `QTA-002` | fetch throttle | `>=1ms` 告警 |
| `QTA-003` | request quota 占用 | `>=0.8` 告警，`>=1` 视为饱和 |
| `QTA-004` | backpressure | `network_idle <0.2` 或 `request_latency >=100ms` 告警 |

## 7. 配置覆盖关系

### 7.1 最重要的可调项

- broker / ISR / 副本基线：
  - `profiles.<name>.broker_count`
  - `profiles.<name>.expected_min_isr`
  - `profiles.<name>.expected_replication_factor`
- Producer / Consumer / Transaction：
  - `profiles.<name>.producer.*`
  - `profiles.<name>.consumer.*`
- 消费组 lag：
  - 全局：`thresholds.consumer_lag_warn`、`thresholds.consumer_lag_crit`
  - 单目标覆盖：`profiles.<name>.group_probe_targets[].lag_warn / lag_crit`
- 宿主机：
  - `host.fd_warn_pct`
  - `host.fd_crit_pct`
  - `thresholds.disk_warn_pct`
  - `thresholds.disk_crit_pct`
  - `thresholds.inode_warn_pct`
- 日志：
  - `logs.lookback_minutes`
  - `logs.tail_lines`
  - `logs.min_lines_per_source`
  - `logs.freshness_window`
- 输出：
  - `output.max_evidence_items`
  - `output.show_pass_checks`
  - `output.show_skip_checks`
  - `output.verbose`

### 7.2 运行结果为什么会和这份清单“数量不同”

因为 `Kdoctor` 不是每次都把所有检查都跑出来，而是按证据来源动态注册和动态判定：

- 只给 `bootstrap`：看不到 `CFG-*`、多数 `DKR-*`、多数 `LOG-*`
- 给了 `compose`：会多出配置、安全、存储、部分 KRaft 检查
- 当前主机能访问 Docker：才会看到 `DKR-*`
- 当前主机能拿到宿主机证据：才会看到 `HOST-*`
- 开了日志采集：才会看到 `LOG-*`
- 配了消费组目标：才会看到 `CSM-*`
- 开了事务上下文：才会看到 `TXN-*`

## 8. 结论

如果只想用一句话概括这份基线文档，可以这样理解：

- `Kdoctor` 当前默认主链路覆盖的是：网络、metadata、KRaft、Topic/ISR、probe、compose、Docker、Host、Logs、Consumer、Producer、Transaction。
- 阈值主要由 `kdoctor.yaml` 里的 `profiles.*`、`thresholds.*`、`logs.*`、`host.*` 决定。
- JMX / Metrics / JVM / Quota 相关代码还在，但当前封版默认不进主链路报告。

后续如果你们要做内部审计或运维基线评审，建议把这份文档和 `kdoctor.yaml` 放在一起维护：  
`CHECK_BASELINE.md` 负责说明“工具会怎么判”，`kdoctor.yaml` 负责说明“你们希望它按什么标准判”。
