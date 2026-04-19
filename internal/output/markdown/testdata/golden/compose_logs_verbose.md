# Kdoctor 检查报告

## 摘要

| 项目 | 值 |
| --- | --- |
| 模式 | `incident` |
| 配置模板 | `single-host-3broker-kraft-prod` |
| 总体状态 | `失败` |
| 检查时间 | `2026-04-19 21:00:00+08:00` |
| 耗时 | `1234ms` |
| Broker 存活 | `2/3` |
| 检查统计 | 严重 0 / 失败 1 / 告警 1 / 错误 0 / 通过 1 / 跳过 0 |

## 概览

本次共执行 3 项检查，最高状态为 失败。已识别 1 个优先级较高的主因，请优先按建议动作顺序处理。

## 证据覆盖

| 来源 | 状态 |
| --- | --- |
| 网络 | 已启用，已获取证据 |
| Compose | 未纳入本次运行 |
| 日志 | 已启用，未获取证据 |

## 主因判断

- 最可能主因：metadata 返回地址与当前客户端视角不匹配。

## 建议动作

- 优先核对 advertised.listeners 与当前客户端网络路径。

## 重点问题

| 状态 | 编号 | 模块 | 摘要 |
| --- | --- | --- | --- |
| 失败 | DKR-003 | Docker | 部分 Kafka 容器发生过 OOMKilled |
| 告警 | LOG-001 | 日志 | 日志来源已获取，但部分样本不足或不够新鲜，后续日志判断需要谨慎解释 |

## 重点问题详情

### DKR-003 Docker

- 状态：`失败`
- 摘要：部分 Kafka 容器发生过 OOMKilled
- 核心证据：
  - container=kafka2 oom_killed=true
- 下一步：
  - 检查容器内存限制与 JVM 堆大小

### LOG-001 日志

- 状态：`告警`
- 摘要：日志来源已获取，但部分样本不足或不够新鲜，后续日志判断需要谨慎解释
- 核心证据：
  - source=file:/data/kafka/server.log line_count=80 byte_count=4096 latest_timestamp=2026-04-19T20:58:00+08:00 freshness=2m0s sample_sufficient=false empty=false

## 完整附录

<details>
<summary>展开 PASS / SKIP 明细</summary>

### CFG-006 配置

- 状态：`通过`
- 摘要：listeners 与 advertised.listeners 结构一致

</details>

