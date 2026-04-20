# Kdoctor 用户使用手册

## 1. 这份手册解决什么问题

这份手册面向真正要把 `Kdoctor` 用起来的人。

它重点回答这些问题：

- 只有一个 Kafka 地址时怎么查
- 有 `kdoctor.yaml`、`docker-compose.yml`、日志目录时怎么增强检查
- Windows 和 Linux 的命令怎么写
- 报告里的 `通过 / 告警 / 失败 / 严重 / 错误 / 跳过` 怎么理解
- 哪些是“真故障”，哪些只是“上下文提示”

## 2. Kdoctor 是什么

`Kdoctor` 是 Kafka 首轮巡检和排障收敛工具。

它不是：

- Kafka 管理平台
- 统一监控平台
- 自动修复系统

它更像值班现场的一把快刀：

- 先快速判断问题更像落在哪一层
- 再把零散症状收敛成 1 到 3 条主因判断
- 最后给出下一步最值得执行的动作

## 3. 使用前准备

### 3.1 二进制

Windows 使用：

```powershell
.\kdoctor.exe
```

Linux 使用：

```bash
./kdoctor
```

### 3.2 最低输入

最低只需要一个可访问的 Kafka `bootstrap` 地址，例如：

- `192.168.1.1:9192`

### 3.3 可选增强输入

如果有这些信息，报告会更完整：

- `kdoctor.yaml`
- `docker-compose.yml`
- Docker 运行时
- Kafka 日志目录

## 4. 快速开始

### 4.1 只有一个地址时

Windows:

```powershell
.\kdoctor.exe probe --bootstrap 192.168.1.1:9192
```

Linux:

```bash
./kdoctor probe --bootstrap 192.168.1.1:9192
```

这会执行：

- bootstrap TCP 检查
- metadata 拉取
- metadata 返回地址探测
- Topic / leader / ISR 检查
- `metadata -> produce -> consume -> commit -> e2e` 探针

### 4.2 使用配置文件

Windows:

```powershell
.\kdoctor.exe probe --config .\kdoctor.yaml
```

Linux:

```bash
./kdoctor probe --config ./kdoctor.yaml
```

### 4.3 再叠加 compose

Windows:

```powershell
.\kdoctor.exe probe --config .\kdoctor.yaml --compose .\docker-compose.yml
```

Linux:

```bash
./kdoctor probe --config ./kdoctor.yaml --compose ./docker-compose.yml
```

这会额外检查：

- `listeners`
- `advertised.listeners`
- `controller.quorum.voters`
- `node.id`
- `process.roles`
- `inter.broker.listener.name`
- Docker 挂载、持久化和数据目录规划

### 4.4 如果 Kafka 开启了 TLS / SSL / SASL

当前版本对加密 Kafka 的支持分成两层：

1. 支持“安全配置审计”
2. 暂不支持“带凭证直接连加密 listener 做 metadata / probe”

先说结论：

- 如果你有 `docker-compose.yml`，当前版本可以检查安全配置本身是否合理
- 如果你的 Kafka 只开放 `SSL / SASL_SSL / SASL_PLAINTEXT` listener，当前版本还不能直接拿用户名、密码、证书去跑真实客户端探针
- 如果你的集群另外保留了一个可达的 `PLAINTEXT` 运维 listener，那么可以把 `Kdoctor` 指到那个运维 listener 上，同时继续用 `compose` 审计安全配置

当前已经支持的安全相关检查有：

- `SEC-001`：`listener.security.protocol.map` 与 `profile.security_mode` 是否一致
- `SEC-002`：SASL 机制是否覆盖 `profile.sasl_mechanism`
- `SEC-003`：SSL / SASL_SSL listener 的证书握手、SAN、到期时间检查
- `SEC-004`：日志或探针里是否出现认证 / 授权拒绝证据
- `SEC-005`：`Authorizer` 是否配置、一致

当前还不支持的能力有：

- 在 `kdoctor.yaml` 里配置 Kafka TLS 客户端证书、私钥、CA
- 在 `kdoctor.yaml` 里配置 SASL 用户名、密码、JAAS
- 直接连接 `SSL / SASL_SSL / SASL_PLAINTEXT` listener 做 `metadata / produce / consume / commit / e2e`

也就是说，当前版本的“安全能力”主要是：

- 检查配置是否写对了
- 检查证书是否能握手、是否快过期
- 检查日志里是否已经出现认证 / 授权问题

而不是：

- 作为一个完整的 TLS / SASL Kafka 客户端去登录集群

### 4.5 安全模式怎么写

`kdoctor.yaml` 里的 `profile.security_mode` 当前支持这些值：

- `plaintext`
- `ssl`
- `tls`
- `sasl`
- `sasl_plaintext`
- `sasl_ssl`

如果用了 SASL，还可以补：

- `profile.sasl_mechanism`

例如：

```yaml
profiles:
  secure-kafka:
    execution_view: "external"
    security_mode: "sasl_ssl"
    sasl_mechanism: "SCRAM-SHA-512"
    broker_count: 3
    expected_replication_factor: 3
    expected_min_isr: 2
```

这段配置的作用是：

- 告诉 `Kdoctor` 你期望当前执行视角走的是 `SASL_SSL`
- 告诉 `Kdoctor` 期望 broker 启用了 `SCRAM-SHA-512`

它会影响：

- `SEC-001`
- `SEC-002`
- `SEC-003`

但它不会让当前版本自动拿这些参数去登录 Kafka。

### 4.6 加密场景下现在怎么用

#### 场景 A：只有加密 listener，没有任何明文运维 listener

这种场景下，当前版本不适合直接做完整链路探针。

你可以把它当成“安全配置与证书审计工具”来辅助使用，但不能把 `probe` 结果当成完整可用性判断。

重点看：

- `SEC-001`
- `SEC-002`
- `SEC-003`
- `SEC-004`
- `SEC-005`

如果这类环境还需要 `Kdoctor` 去真正连 Kafka，那就属于后续要补的客户端安全连接能力，不是当前封版能力。

#### 场景 B：对外是加密 listener，但内网保留一个 PLAINTEXT 运维 listener

这种场景是当前版本最容易落地的方式。

做法是：

1. `bootstrap` 指向那个可达的明文运维 listener
2. `compose` 继续提供真实的安全配置
3. `profile.security_mode` 按你真正业务 listener 的模式填写

例如：

```yaml
profiles:
  prod:
    bootstrap_internal:
      - "192.168.1.1:9092"
    execution_view: "external"
    security_mode: "sasl_ssl"
    sasl_mechanism: "SCRAM-SHA-512"
```

然后运行：

Linux:

```bash
./kdoctor probe --bootstrap 192.168.1.1:9092 --config ./kdoctor.yaml --compose ./docker-compose.yml
```

Windows:

```powershell
.\kdoctor.exe probe --bootstrap 192.168.1.1:9092 --config .\kdoctor.yaml --compose .\docker-compose.yml
```

这种方式下：

- `CLI-*` 和 `KFK-*` 走明文运维 listener
- `SEC-*` 继续审计你真正暴露给业务的 `SSL / SASL` listener

这是当前版本在“加密 Kafka”场景下最现实的使用方式。

## 5. Windows / Linux 路径差异

### 5.1 配置文件路径

Windows:

```powershell
.\kdoctor.exe probe --config .\kdoctor.yaml
```

Linux:

```bash
./kdoctor probe --config ./kdoctor.yaml
```

不要混用：

- Windows 用 `.\file`
- Linux 用 `./file`

### 5.2 compose 路径

Windows:

```powershell
.\kdoctor.exe probe --compose .\docker-compose.yml
```

Linux:

```bash
./kdoctor probe --compose ./docker-compose.yml
```

## 6. 常用模式

### 6.1 `quick`

快速巡检，强调低成本和快速收敛。

```bash
./kdoctor quick --bootstrap 192.168.1.1:9192
```

### 6.2 `probe`

最常用模式。会执行真实客户端链路探针。

```bash
./kdoctor probe --bootstrap 192.168.1.1:9192
```

### 6.3 `lint`

偏静态配置审计。

```bash
./kdoctor lint --config ./kdoctor.yaml --compose ./docker-compose.yml
```

### 6.4 `incident`

更强调主因摘要和当前故障收敛。

```bash
./kdoctor incident --config ./kdoctor.yaml --compose ./docker-compose.yml
```

## 7. 输出格式

### 7.1 终端输出

默认终端输出特点：

- 中文
- 先看摘要
- 只展开 `CRIT / FAIL / WARN / ERROR`
- `PASS / SKIP` 默认折叠

适合值班现场。

### 7.2 JSON

```bash
./kdoctor probe --bootstrap 192.168.1.1:9192 --json
```

或：

```bash
./kdoctor probe --bootstrap 192.168.1.1:9192 --format json --output ./report.json
```

JSON 适合自动化系统消费，包含：

- `schema_version`
- `tool_version`
- `summary`
- `checks`
- `exit_code`

### 7.3 Markdown

```bash
./kdoctor probe --bootstrap 192.168.1.1:9192 --format markdown --output ./report.md
```

Markdown 适合留档、工单和审计。

## 8. 如何理解结果

### 8.1 状态等级

- `通过`：当前没有看到明显问题
- `跳过`：这次运行没有足够输入，不代表环境正常
- `告警`：存在风险或边界问题，未必立刻故障
- `失败`：已经出现明显异常，通常会影响链路
- `严重`：高危问题，应优先处理
- `错误`：工具这次没拿到足够结果或执行链路本身失败

### 8.2 证据覆盖

摘要里的“证据覆盖”不是“尝试采集过什么”，而是：

- `已获取证据`
- `无可用证据`
- `未纳入本次运行`
- `探针=已执行`

所以如果看到：

- `日志=无可用证据`

表示这次运行没有拿到可用于判断的日志证据，不是说日志一定正常或一定异常。

### 8.3 主因判断

主因判断最多收敛到几条最值得优先处理的问题，不是把所有检查项重复说一遍。

先看：

1. 严重 / 失败
2. 主因判断
3. 建议动作

## 9. 推荐使用顺序

### 9.1 首次接环境

1. 先跑 `bootstrap-only`
2. 再补 `kdoctor.yaml`
3. 再补 `docker-compose.yml`
4. 最后接日志 / Docker / 宿主机证据

### 9.2 线上故障现场

推荐先跑：

```bash
./kdoctor probe --config ./kdoctor.yaml --compose ./docker-compose.yml
```

如果当前机器不能访问 Docker 或日志，也没关系，工具会继续给出最小可用结论。

## 10. 配置文件说明

常见最小配置示例：

```yaml
profile: master-internal-kraft-prod
bootstrap: 192.168.1.1:9192

profiles:
  master-internal-kraft-prod:
    broker_count: 3
    expected_replication_factor: 3
    expected_min_isr: 2
    controller_endpoints:
      - 192.168.1.1:9193
      - 192.168.1.1:9195
      - 192.168.1.1:9197

probe:
  enabled: true
  topic: _kdoctor_probe
  message_bytes: 1024

output:
  format: terminal
  max_evidence_items: 8
  show_pass_checks: false
  show_skip_checks: false
```

## 11. 常见问题

### 11.1 为什么传了 `--config` 却像没生效

最常见原因是路径写错。

Linux 不要写：

```bash
./kdoctor --config .\kdoctor.yaml
```

Linux 要写：

```bash
./kdoctor --config ./kdoctor.yaml
```

### 11.2 为什么看到很多 `SKIP`

`SKIP` 不一定是坏事，通常表示：

- 这次没提供相关输入
- 当前执行机看不到那部分证据
- 该检查不属于当前模式

### 11.3 为什么探针成功，但还有失败项

这很常见。

例如：

- 链路本身可通
- 但 controller listener 不可达
- 或部分内部 topic / ISR 已经处在高风险边界

这种情况说明环境“还能用”，但不代表“健康”。

### 11.4 为什么我明明是加密 Kafka，`probe` 却失败了

如果你的集群只开放了：

- `SSL`
- `SASL_SSL`
- `SASL_PLAINTEXT`

那当前版本大概率不能直接完成 `probe`。

原因不是 Kafka 坏了，而是当前版本还没有把：

- TLS 客户端证书
- CA
- SASL 用户名 / 密码
- JAAS / SCRAM 登录参数

接进 Kafka 客户端传输层。

所以当前版本面对纯加密 listener 的能力边界是：

- 能审计配置和证书
- 不能直接当成完整安全客户端去登录 Kafka

## 12. 退出码

一般可以这样理解：

- `0`：没有发现会提升总体状态的问题
- 非 `0`：存在告警、失败、严重或运行错误

如果接自动化系统，建议直接读取 JSON 里的 `exit_code` 与 `summary.status`。

## 13. 版本与构建

查看版本：

Windows:

```powershell
.\kdoctor.exe --version
```

Linux:

```bash
./kdoctor --version
```

本地构建：

```powershell
.\scripts\build.ps1 -GOOS windows -GOARCH amd64
.\scripts\build.ps1 -GOOS linux -GOARCH amd64
```

构建产物统一输出到工作区根目录 `dist/`。
