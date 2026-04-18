# version2-4

## 阶段定位

这是第二阶段的一次“大规模覆盖推进”。

这轮不是只补单点能力，而是重新对照 `version2.md` 后，沿着还未完成的主干一次性推进了多条故障域：

- 高级网络
- Kafka / KRaft / Topic 深化规则
- producer / consumer / transaction 配置审计
- upgrade / version 场景
- logs / docker / storage / security 的第二层规则

## 本轮完成内容

### 1. 高级网络与返回路径判断已补齐一批

本轮新增并接入：

- `NET-005` metadata 返回路径错配
- `NET-006` 外部视角下返回私网地址
- `NET-007` 疑似 bootstrap-only LB
- `NET-009` 端口通但 Kafka 握手失败

这些检查会把“入口可达”和“后续 broker 路由正确”明确拆开，进一步减少把 listeners / NAT / LB 问题误判成“Kafka 完全挂了”的风险。

### 2. Kafka / KRaft / Topic 深化检查已接入

本轮新增并接入：

- `KFK-005` metadata 返回 broker 路径可达性
- `KFK-006` broker 注册完整性
- `KFK-007` broker 身份冲突
- `KFK-008` metadata latency
- `KFK-009` 运行态拓扑与 compose 偏离

- `KRF-005` 活动 controller 端点配置异常
- `KRF-007` unknown voter connections

- `TOP-006` UnderReplicatedPartitions
- `TOP-007` UnderMinISR / AtMinISR
- `TOP-008` OfflineReplica / 无 leader 分区
- `TOP-009` leader skew
- `TOP-010` replica lag
- `TOP-011` topic 分区 / RF 规划

这意味着第二阶段现在已经不只是“基础 leader / ISR 检查”，而是开始具备更细的拓扑、复制与分布判断能力。

### 3. producer / consumer / transaction 配置审计已进入主流程

这轮扩展了配置模型，新增：

- `profiles.producer.*`
- `profiles.consumer.*`

并接入了：

- `PRD-001` acks 与 minISR / durability
- `PRD-002` 幂等与重试乱序风险
- `PRD-003` delivery timeout sanity
- `PRD-006` transaction timeout sanity

- `CSM-003` max.poll.interval
- `CSM-004` heartbeat / session
- `CSM-005` auto.offset.reset 语义

- `TXN-001` 事务主题缺失上下文
- `TXN-002` 事务主题必需性
- `TXN-003` transaction timeout 上限
- `TXN-004` read_committed 前提

这样第二阶段已经不再只看运行结果，也开始能直接指出“配置组合本身就会踩雷”的问题。

### 4. upgrade / version 场景已补第一批

本轮新增并接入：

- `UPG-001` rolling upgrade 半完成
- `UPG-002` feature / finalization 版本偏离
- `UPG-003` tiered storage awareness

这批检查主要依赖 compose/镜像/env 静态信息，但对内部排障已经有实际价值，因为很多“奇怪的控制面问题”本质上是升级收口没做干净。

### 5. logs / docker / storage / security 再补一层

本轮新增并接入：

- `LOG-005` 命中上下文
- `LOG-006` 新鲜度与样本充分性
- `LOG-007` 重复指纹风暴
- `LOG-008` 自定义规则库可用性

- `DKR-006` 容器重启历史
- `DKR-007` 容器内存限制与 JVM 堆余量

- `STG-002` OfflineLogDirectory
- `STG-004` 部分 logdir 故障

- `SEC-004` 认证 / ACL 拒绝证据

同时还把日志指纹库补进了：

- 认证失败
- 授权拒绝
- 事务路径错误

## 验证情况

本轮已执行：

- `gofmt`
- `go test ./...`

结果通过。

## 重新对照 version2.md 后，当前仍未完全落地的点

这轮推进之后，`version2.md` 里剩余还没有完全落地、或仍属于“部分覆盖”的，主要还有：

- `NET-008` 真正的 DNS 漂移 / TTL / 多值可达性分析
- `KRF-004 / KRF-006 / KRF-008` 这类更依赖时间窗口与 JMX 历史的控制面判断
- `SEC-003` 证书过期 / SAN / CA 链
- `STG-001 / HOST-007` inode 与更完整的磁盘目录故障
- `QTA-001~004`
- `JVM-003 / JVM-004`
- `HOST-008 / HOST-009 / HOST-010 / HOST-011`
- 更完整的 `TXN-005`

换句话说：

- `P0` 主干已经非常接近成形
- 剩下的更多是深水区指标、证书、quota、inode、clock skew、GC 这类更高成本能力

## 当前阶段判断

第二阶段现在已经从“补盲区”推进到了“多故障域可联动判断”的阶段。

如果按 `version2.md` 的目标看：

- `P0`：大部分核心块已进入可用状态
- `P1/P2`：已经提前吃掉了其中一部分 static / upgrade / transaction awareness

## 下一步建议

如果继续沿着 `version2.md` 收口，下一批最值得做的是：

1. `QTA-* + JVM-003/004`
2. `SEC-003`
3. `HOST-008/009/010/011 + STG-001`

这样第二阶段会真正从“高级排障器”进一步走到“更完整的内部运维诊断器”。 
