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
