# version2-2

## 阶段定位

这是第二阶段的第二轮推进，继续沿着 `version2.md` 里的 `P0` 往前补基础域。

这一轮重点不是再加 probe，而是把之前偏“Kafka 协议层”的诊断继续往运维现场推进两步：

- 补 `security` 基础域
- 补 `storage / logdir` 基础域

## 本轮完成内容

### 1. 安全域基础检查已接入主流程

本轮新增并接入：

- `SEC-001` listener 安全协议与当前执行视角一致性
- `SEC-002` SASL 机制一致性
- `SEC-005` Authorizer 配置基线

这几项检查目前主要依赖 `compose` 增强输入，但已经能直接识别以下高价值问题：

- `profile.security_mode` 与 `listener.security.protocol.map` 不一致
- 当前执行视角下的 client listener 缺少协议映射
- 启用了 `SASL_*` listener，但 broker 未声明 `KAFKA_CFG_SASL_ENABLED_MECHANISMS`
- `profile.sasl_mechanism` 与 broker 侧可用机制不一致
- KRaft 场景未使用 `StandardAuthorizer`，或各 broker 的 `authorizer.class.name` 不一致

### 2. 存储域基础检查已接入主流程

本轮新增并接入：

- `STG-003` metadata / data 目录规划检查
- `STG-005` 存储挂载规划检查

这两项检查目前已经能直接指出：

- `KAFKA_CFG_LOG_DIRS` 缺失
- controller/broker 节点没有显式设置 `KAFKA_CFG_METADATA_LOG_DIR`
- `metadata.log.dir` 与 `log.dirs` 共用或重叠
- Kafka 存储路径没有被 volume 承载
- Kafka 存储路径虽然挂载了，但仍依赖 named volume，宿主机可见性较弱

### 3. 第二阶段配置继续扩展

这一轮同时补了：

- `profiles.sasl_mechanism`

它用于把当前 profile 预期的 SASL 机制显式写进配置，后面做更深的安全探测时可以继续复用，不需要再回头拆配置结构。

### 4. 根因归并已吸收 security / storage

root cause 规则现在已经会吸收：

- `SEC-001 / SEC-002 / SEC-005`
- `STG-003 / STG-005`

这样输出不再只是多几条检查项，而是会把安全协议错配、SASL 机制错配、Authorizer 异常、目录规划和挂载风险提升到主因层。

## 验证情况

本轮已执行：

- `gofmt`
- `go test ./...`

结果通过。

## 当前阶段判断

第二阶段已经不再只是“消费组补盲区”了，`security` 和 `storage` 这两个基础域已经开始成形。

如果按 `version2.md` 的 `P0` 路线图看：

- `group / lag`：已落地
- `scheduler timeout / degrade`：已落地第一层
- `storage / logdir`：已进入可用版
- `security` 基础域：已进入可用版
- `JMX 指标采集`：仍是下一批最关键缺口

## 下一步建议

最合适的下一步是继续推进：

1. `JMX` 采集基础能力
2. `MET-* / TOP-006 / TOP-007 / STG-002`
3. 把 `security` 从静态配置一致性，推进到真实握手/认证层

这样第二阶段 `P0` 的主干就会真正闭合。
