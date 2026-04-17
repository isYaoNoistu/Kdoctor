# Kdoctor 用户使用手册

## 1. 这份手册解决什么问题

这份手册不是介绍代码结构，而是告诉你：

- 这个工具适合什么时候用
- 你手里只有一个 Kafka 地址时该怎么跑
- 你有 `compose`、配置文件、日志目录时又该怎么增强检查
- 报告里每类结果大概代表什么
- 遇到 `SKIP`、`WARN`、`FAIL` 时应该怎么理解

如果你是第一次真正使用 `Kdoctor`，建议直接按下面顺序阅读：

1. 先看“快速上手”
2. 再看“最常用命令”
3. 然后看“结果怎么解读”
4. 最后按需要查“配置文件”和“常见问题”

## 2. Kdoctor 是什么

`Kdoctor` 是一个面向 Kafka 运维排障的命令行诊断工具。

它不是 Kafka 管理平台，也不是监控系统。它更像是：

- 现场排障工具
- 集群健康快检工具
- 配置 / 网络 / Topic / KRaft / 客户端链路联合诊断工具

它特别适合这些场景：

- 手里只有一个 `bootstrap` 地址，想先判断集群是不是“基本可用”
- Kafka 连不上，但你不确定是网络、`advertised.listeners`、Topic、ISR，还是客户端探针链路的问题
- 你有 `docker-compose.yml`，想让工具顺便做静态配置 lint
- 你有日志目录、容器环境，想让工具补充部署侧和日志侧诊断

## 3. 使用前你需要知道的最重要概念

### 3.1 最小输入模式

`Kdoctor` 最重要的设计目标之一，就是**不依赖 compose 才能运行**。

也就是说，只给一个 Kafka 地址也能执行：

- 网络检查
- metadata 检查
- broker endpoint 检查
- Topic / ISR 检查
- 真实 client probe

最小命令就是：

```powershell
kdoctor.exe probe --bootstrap 192.168.1.1:9192
```

### 3.2 增强输入模式

如果你还能提供更多信息，`Kdoctor` 会做更多检查：

- `--profile`：补充环境预期，例如 broker 数、controller 端点、复制因子
- `--compose`：做配置 lint 和部署对照
- `--config`：统一加载 YAML 配置
- `--log-dir`：增加日志侧检查

理解方式很简单：

- `bootstrap` 解决“能不能开始看”
- `profile` 解决“我预期它应该是什么样”
- `compose` 解决“部署配置本身写得对不对”
- `log-dir` 解决“近期日志里有没有典型错误”

### 3.3 `probe` 是真实探针，不是占位检查

`probe` 模式下会真实尝试跑一条客户端链路：

1. 拉 metadata
2. 检查 / 准备 `_kdoctor_probe`
3. produce
4. consume
5. commit

当前版本里：

- `_kdoctor_probe` 不存在时，会尝试自动创建
- 上游阶段失败后，下游未执行阶段会显示为 `SKIP`
- 不会再把一处失败扩散成多条重复 `FAIL`

### 3.4 Windows 和 Linux 的路径写法不同

这是最容易踩坑的一点。

Windows 常见写法：

```powershell
.\kdoctor.exe
.\kdoctor.yaml
.\docker-compose.yml
```

Linux 常见写法：

```bash
./kdoctor
./kdoctor.yaml
./docker-compose.yml
```

请特别注意：

- Windows 下常写 `.\kdoctor.yaml`
- Linux 下必须优先写 `./kdoctor.yaml`

如果你在 Linux 下把路径写成 `.\kdoctor.yaml`，虽然当前版本已经做了兼容，但仍然建议你按 Linux 习惯写 `./kdoctor.yaml`，这样最稳定、最不容易误解。

## 4. 快速上手

### 4.1 方式一：直接使用打好的二进制

Windows：

```powershell
.\kdoctor.exe probe --bootstrap 192.168.1.1:9192
```

Linux：

```bash
./kdoctor probe --bootstrap 192.168.1.1:9192
```

这是最推荐的使用方式。

### 4.2 方式二：直接用源码运行

在仓库根目录执行：

Windows：

```powershell
go run ./cmd/kdoctor probe --bootstrap 192.168.1.1:9192
```

Linux：

```bash
go run ./cmd/kdoctor probe --bootstrap 192.168.1.1:9192
```

适合开发阶段或你还没打包的时候。

### 4.3 第一次实际使用时，建议你先跑这条命令

Windows：

```powershell
.\kdoctor.exe probe --bootstrap 192.168.1.1:9192
```

Linux：

```bash
./kdoctor probe --bootstrap 192.168.1.1:9192
```

原因很简单：

- `probe` 是当前最有价值、最接近真实业务链路的模式
- 只需要一个地址
- 输出最容易看出问题到底卡在哪一层

## 5. 最常用命令

### 5.1 最小可用检查

Windows：

```powershell
.\kdoctor.exe probe --bootstrap 192.168.1.1:9192
```

Linux：

```bash
./kdoctor probe --bootstrap 192.168.1.1:9192
```

适合：

- 先看 Kafka 基本是不是通的
- 不确定集群有没有明显元数据 / 网络 / Topic 问题

### 5.2 多个 bootstrap 地址一起传入

多个地址用英文逗号分隔：

Windows：

```powershell
.\kdoctor.exe probe --bootstrap 192.168.1.1:9192,192.168.1.1:9194,192.168.1.1:9196
```

Linux：

```bash
./kdoctor probe --bootstrap 192.168.1.1:9192,192.168.1.1:9194,192.168.1.1:9196
```

适合：

- 你不确定某一个地址是否稳定
- 希望工具有更多可回退的 bootstrap 节点

### 5.3 使用 profile 增强诊断

Windows：

```powershell
.\kdoctor.exe probe --bootstrap 192.168.1.1:9192 --profile generic-bootstrap
```

Linux：

```bash
./kdoctor probe --bootstrap 192.168.1.1:9192 --profile generic-bootstrap
```

如果你已经在 `kdoctor.yaml` 里定义了自己的 profile：

Windows：

```powershell
.\kdoctor.exe probe --profile my-prod
```

Linux：

```bash
./kdoctor probe --profile my-prod
```

适合：

- 你已经知道这个环境应该有几个 broker
- 你知道预期的 replication factor 和 min ISR
- 你希望工具不仅看“有没有问题”，还看“是否低于预期”

### 5.4 使用 compose 增强检查

Windows：

```powershell
.\kdoctor.exe probe --bootstrap 192.168.1.1:9192 --compose .\docker-compose.yml
```

Linux：

```bash
./kdoctor probe --bootstrap 192.168.1.1:9192 --compose ./docker-compose.yml
```

适合：

- 你有 `docker-compose.yml`
- 你想顺手检查 `listeners`、`advertised.listeners`、`controller.quorum.voters`、`node.id` 等配置

### 5.5 输出 JSON

Windows：

```powershell
.\kdoctor.exe probe --bootstrap 192.168.1.1:9192 --json
```

Linux：

```bash
./kdoctor probe --bootstrap 192.168.1.1:9192 --json
```

适合：

- 后续要接脚本处理
- 想把结果喂给其他自动化工具

### 5.6 输出 Markdown 报告

Windows：

```powershell
.\kdoctor.exe probe --bootstrap 192.168.1.1:9192 --format markdown --output .\report.md
```

Linux：

```bash
./kdoctor probe --bootstrap 192.168.1.1:9192 --format markdown --output ./report.md
```

适合：

- 发群
- 提工单
- 做排障留档

### 5.7 只做快速巡检

Windows：

```powershell
.\kdoctor.exe quick --bootstrap 192.168.1.1:9192
```

Linux：

```bash
./kdoctor quick --bootstrap 192.168.1.1:9192
```

适合：

- 只想快速看整体健康情况
- 不一定需要真实 client probe

### 5.8 只做配置审计

Windows：

```powershell
.\kdoctor.exe lint --compose .\docker-compose.yml
```

Linux：

```bash
./kdoctor lint --compose ./docker-compose.yml
```

适合：

- 集群还没正式启动
- 你想先看配置有没有明显结构问题

### 5.9 故障现场模式

Windows：

```powershell
.\kdoctor.exe incident --bootstrap 192.168.1.1:9192
```

Linux：

```bash
./kdoctor incident --bootstrap 192.168.1.1:9192
```

适合：

- 现场更关注“先看主因和动作建议”
- 不想先读完整检查清单

## 6. 所有运行模式怎么选

### 6.1 `probe`

最常用模式。

特点：

- 做真实链路探针
- 最适合判断“现在这个 Kafka 从客户端角度到底能不能用”

建议：

- 你平时优先用这个

### 6.2 `quick`

快速巡检模式。

特点：

- 优先给出整体健康信号
- 通常噪声更少

建议：

- 适合日常巡检、初筛

### 6.3 `full`

尽量多做检查。

特点：

- 适合已经有较完整上下文的时候

建议：

- 适合排障过程中期，不建议第一次就无脑上

### 6.4 `lint`

配置审计模式。

特点：

- 偏配置和结构检查
- 更适合 `compose` / 配置文件场景

### 6.5 `incident`

故障现场摘要模式。

特点：

- 概览更聚焦
- 适合快速给运维结论

## 7. 命令行参数说明

当前最重要的参数如下。

### 7.1 `--bootstrap`

最常用参数。

作用：

- 指定一个或多个 Kafka bootstrap 地址

示例：

```powershell
--bootstrap 192.168.1.1:9192
--bootstrap 192.168.1.1:9192,192.168.1.1:9194
```

### 7.2 `--bootstrap-internal`

作用：

- 显式指定内网 bootstrap 地址

适合：

- 宿主机 / 内网探测视角
- 你希望区分外部入口和内部入口

### 7.3 `--bootstrap-external`

作用：

- 显式指定外部 bootstrap 地址

适合：

- 外网 / 客户端视角探测

### 7.4 `--profile`

作用：

- 选择运行 profile

当前内置 profile：

- `generic-bootstrap`
- `single-host-3broker-kraft-prod`
- `single-host-3broker-kraft-uat`

### 7.5 `--config`

作用：

- 指定配置文件路径

默认值：

```text
kdoctor.yaml
```

说明：

- 不显式传 `--config` 时，工具会尝试读取当前目录下的 `kdoctor.yaml`
- 如果这个默认文件不存在，工具会继续使用默认配置
- 如果你显式传了 `--config`，但路径不存在，工具现在会直接报错，不再静默退回默认配置

示例：

Windows：

```powershell
.\kdoctor.exe probe --config .\kdoctor.yaml
```

Linux：

```bash
./kdoctor probe --config ./kdoctor.yaml
```

### 7.6 `--compose`

作用：

- 指定 `docker-compose.yml` 路径

适合：

- 做配置 lint
- 对照部署配置和运行结果

### 7.7 `--log-dir`

作用：

- 指定 Kafka 日志目录

适合：

- 想让工具补做日志侧检查时使用

### 7.8 `--format`

作用：

- 指定输出格式

可选值：

- `terminal`
- `json`
- `markdown`

### 7.9 `--json`

作用：

- 直接输出 JSON

说明：

- 这是一个快捷参数
- 如果同时传了 `--json` 和 `--format markdown`，最终会按 JSON 输出

### 7.10 `--output`

作用：

- 把输出写到文件

示例：

Windows：

```powershell
--output .\report.md
--output .\report.json
```

Linux：

```bash
--output ./report.md
--output ./report.json
```

### 7.11 `--timeout`

作用：

- 指定本次整体执行超时

示例：

```powershell
--timeout 30s
--timeout 1m
```

### 7.12 `--severity`

作用：

- 预留的最小输出严重级别参数

当前建议：

- 暂时不要把它当成主功能依赖
- 当前版本的核心使用方式还是直接看完整输出

## 8. 参数优先级

这是最容易搞混的地方。

当前大致优先级是：

1. 命令行参数
2. `kdoctor.yaml`
3. 内置默认值 / 内置 profile

更具体一点：

- `--profile` 会决定选哪个 profile
- `--bootstrap` / `--bootstrap-external` / `--bootstrap-internal` 会覆盖 profile 里的对应地址
- `--compose` 会覆盖配置文件里的 `docker.compose_file`
- `--log-dir` 会覆盖配置文件里的 `logs.log_dir`
- `--timeout` 会覆盖配置文件里的 `execution.timeout`

如果你记不住，按这个理解就够了：

**命令行永远是你这次执行的最终裁定。**

## 9. 配置文件怎么写

默认配置文件名是：

```text
kdoctor.yaml
```

你可以直接从示例文件开始改：

- [kdoctor.example.yaml](/d:/project/project/Kdoctor/kdoctor.example.yaml:1)

### 9.1 一个最小示例

```yaml
version: 1
default_profile: generic-bootstrap

profiles:
  generic-bootstrap:
    bootstrap_external:
      - "192.168.1.1:9192"

probe:
  enabled: true
  topic: "_kdoctor_probe"
  timeout: "15s"

execution:
  timeout: "30s"
```

### 9.2 常用字段解释

`default_profile`

- 默认使用哪个 profile

`profiles.<name>.bootstrap_external`

- 外部客户端视角的 Kafka 地址

`profiles.<name>.bootstrap_internal`

- 内部网络视角的 Kafka 地址

`profiles.<name>.controller_endpoints`

- KRaft controller 端点

`profiles.<name>.broker_count`

- 预期 broker 数量

`profiles.<name>.expected_min_isr`

- 预期最小 ISR

`profiles.<name>.expected_replication_factor`

- 预期复制因子

`docker.compose_file`

- 默认 compose 文件路径

`logs.log_dir`

- 默认日志目录

`logs.min_lines_per_source`

- 单个日志来源至少需要多少行样本才算“样本充足”

`logs.freshness_window`

- 日志新鲜度窗口，超过这个时间的来源会被标记为“不够新鲜”

`logs.max_files`

- 本次日志采样最多处理多少个文件

`logs.max_bytes_per_source`

- 单个日志来源最多读取多少字节，用于控制现场读取开销

`logs.custom_patterns_dir`

- 自定义日志指纹规则目录，用于补充团队自己的错误模式

`probe.topic`

- 探针主题名

`probe.group_prefix`

- 探针消费组前缀

`probe.timeout`

- 探针阶段超时

`probe.message_bytes`

- 探针消息大小

`probe.produce_count`

- 探针每次生产的消息数

`probe.cleanup`

- 是否在本轮结束后尝试清理由工具自动创建的 probe topic

`execution.timeout`

- 整体执行超时

`execution.metadata_timeout`

- metadata 超时

`execution.tcp_timeout`

- TCP 探测超时

`execution.admin_api_timeout`

- Admin API 相关动作的独立超时

`execution.jmx_timeout`

- JMX 相关动作的独立超时

`diagnosis.max_root_causes`

- 报告摘要最多输出多少个主因

`diagnosis.enable_confidence`

- 是否在根因摘要里显示“高置信度”“中高置信度”一类提示

## 10. 输出怎么理解

### 10.1 顶部概览

你通常会先看到这些字段：

- 模式
- 配置模板
- 总体状态
- 检查时间
- 耗时
- Broker 存活
- 概览
- 主因判断
- 建议动作

可以把它理解成：

- 第一层：先告诉你“严重不严重”
- 第二层：再告诉你“最像是什么原因”
- 第三层：最后告诉你“先做什么”

### 10.2 状态含义

`PASS`

- 通过，没有发现该项问题

`WARN`

- 有风险或有上下文不足，但不一定已经是故障

`FAIL`

- 该项检查明确失败

`CRIT`

- 严重失败，通常意味着高优先级问题

`SKIP`

- 不是错误，表示这项没有执行或当前上下文不足

`ERROR`

- 工具在执行该项时出现内部执行错误

`TIMEOUT`

- 该项超时

### 10.3 `SKIP` 不是坏事

很多使用者第一次会误会 `SKIP`。

其实 `SKIP` 很常见，而且很多时候是合理的：

- 你没传 `compose`，配置 lint 相关检查就会 `SKIP`
- 你没提供日志目录，日志相关检查就会 `SKIP`
- 你在外网视角执行，而 controller listener 是内网地址，部分 KRaft 检查就会 `SKIP`
- probe 上游阶段失败，下游未执行阶段也会 `SKIP`

所以判断标准不是“有没有 `SKIP`”，而是：

- 这些 `SKIP` 是否符合你当前输入条件

### 10.4 `probe` 相关检查怎么读

`CLI-001`

- `bootstrap -> metadata`

`CLI-002`

- produce

`CLI-003`

- consume

`CLI-004`

- commit

`CLI-005`

- 端到端综合结果

当前版本里：

- 如果 `CLI-002` 失败，`CLI-003/004` 可能会 `SKIP`
- 这表示工具已经知道“后面没必要继续执行了”

这比把所有后续阶段都打成 `FAIL` 更可信。

## 11. 常见场景怎么用

### 11.1 我只有一个 Kafka 地址

直接这样：

Windows：

```powershell
.\kdoctor.exe probe --bootstrap 192.168.1.1:9192
```

Linux：

```bash
./kdoctor probe --bootstrap 192.168.1.1:9192
```

这是默认推荐姿势。

### 11.2 我有三个 broker 地址

Windows：

```powershell
.\kdoctor.exe probe --bootstrap 192.168.1.1:9192,192.168.1.1:9194,192.168.1.1:9196
```

Linux：

```bash
./kdoctor probe --bootstrap 192.168.1.1:9192,192.168.1.1:9194,192.168.1.1:9196
```

### 11.3 我有 compose 文件

Windows：

```powershell
.\kdoctor.exe probe --bootstrap 192.168.1.1:9192 --compose .\docker-compose.yml
```

Linux：

```bash
./kdoctor probe --bootstrap 192.168.1.1:9192 --compose ./docker-compose.yml
```

### 11.4 我想让别人也看懂结果

Windows：

```powershell
.\kdoctor.exe probe --bootstrap 192.168.1.1:9192 --format markdown --output .\report.md
```

Linux：

```bash
./kdoctor probe --bootstrap 192.168.1.1:9192 --format markdown --output ./report.md
```

### 11.5 我想给脚本处理

Windows：

```powershell
.\kdoctor.exe probe --bootstrap 192.168.1.1:9192 --json
```

Linux：

```bash
./kdoctor probe --bootstrap 192.168.1.1:9192 --json
```

### 11.6 我想把常用环境固化下来

做一个 `kdoctor.yaml`：

```yaml
version: 1
default_profile: my-prod

profiles:
  my-prod:
    bootstrap_external:
      - "192.168.1.1:9192"
      - "192.168.1.1:9194"
      - "192.168.1.1:9196"
    broker_count: 3
    expected_min_isr: 2
    expected_replication_factor: 3
```

然后直接跑：

Windows：

```powershell
.\kdoctor.exe probe --config .\kdoctor.yaml
```

Linux：

```bash
./kdoctor probe --config ./kdoctor.yaml
```

或者：

Windows：

```powershell
.\kdoctor.exe probe --config .\kdoctor.yaml --profile my-prod
```

Linux：

```bash
./kdoctor probe --config ./kdoctor.yaml --profile my-prod
```

## 12. 常见结果解释

### 12.1 `NET-003` 失败

通常表示：

- metadata 返回的 broker 地址对当前客户端不可达

优先怀疑：

- `advertised.listeners`
- 端口暴露
- 防火墙
- 路由

### 12.2 `CLI-001` 通过，但 `CLI-002~005` 失败

通常表示：

- bootstrap 和 metadata 基本通
- 但真实 produce / consume / commit 链路有问题

优先看：

- `CLI-*` 的 `failure_stage`
- `KFK-004`
- `TOP-*`
- `NET-003`

### 12.3 `KFK-004` 提示 `__transaction_state` 缺失

当前版本里，这通常是 `WARN`。

解释：

- 如果你没用事务生产者，这通常不算阻塞问题
- 如果你明确依赖事务，再继续深挖

### 12.4 `KRF-*` 大量 `SKIP`

通常表示：

- 你当前提供的信息不足
- 或你是在外部视角探测，而 controller listener 是内网地址

这不一定是故障。

### 12.5 大量 `CFG-*` 都是 `SKIP`

通常表示：

- 你没有提供 `compose`

这也是正常现象。

## 13. 退出码怎么用

退出码与报告最高状态对应：

- `0`：最高状态是 `PASS`
- `1`：最高状态是 `WARN`
- `2`：最高状态是 `FAIL`
- `3`：最高状态是 `CRIT`
- `5`：最高状态是 `ERROR`
- `6`：最高状态是 `TIMEOUT`

最典型的用法是脚本里这样判断：

```powershell
.\kdoctor.exe probe --bootstrap 192.168.1.1:9192
if ($LASTEXITCODE -ge 2) {
    Write-Host "存在明确失败项，需要人工排查"
}
```

## 14. 从源码构建

在仓库根目录执行：

```powershell
go test ./...
.\scripts\build.ps1
```

默认输出：

```text
dist/kdoctor-windows-amd64.exe
```

交叉编译 Linux：

```powershell
.\scripts\build.ps1 -GOOS linux -GOARCH amd64
```

## 15. 常见问题

### 15.1 没有 `kdoctor.yaml` 能不能跑

可以。

没有配置文件时，工具会使用默认配置。

### 15.2 没有 compose 能不能跑

可以。

这是工具设计上的基本要求。

### 15.3 为什么报告里很多 `SKIP`

先看你有没有提供：

- `compose`
- `log-dir`
- 内网视角的 controller 地址

如果没有，这些 `SKIP` 往往是合理的。

### 15.4 Windows 和 Linux 的路径到底怎么写

Windows：

```powershell
.\kdoctor.exe probe --config .\kdoctor.yaml --compose .\docker-compose.yml
```

Linux：

```bash
./kdoctor probe --config ./kdoctor.yaml --compose ./docker-compose.yml
```

不要混写：

- 不要在 Linux 里优先照抄 `.\kdoctor.yaml`
- 不要在 Windows 里把所有示例都改成 `./kdoctor.yaml` 之后又拿 PowerShell 去跑
### 15.5 为什么我传了 `--json` 还不是 Markdown

因为 `--json` 优先级高于 `--format`。

### 15.6 为什么我明明传了 `--config`，结果还是走默认 profile

优先检查两件事：

- 配置文件路径是不是真的存在
- 路径写法是不是符合当前系统

典型错误：

- Linux 下写成 `--config .\kdoctor.yaml`
- `kdoctor.yaml` 文件根本不在当前目录

当前版本里，如果你显式传了 `--config` 但路径不存在，工具会直接报错。

### 15.7 为什么 `probe` 还会自动创建 topic

因为当前版本希望减少 fresh cluster 的误判。

它会优先尝试准备 `_kdoctor_probe`，而不是因为 topic 不存在就直接把整条链路判成业务故障。

### 15.8 如果我不希望 topic 留在环境里怎么办

在配置文件里把：

```yaml
probe:
  cleanup: true
```

当前行为是：

- 仅当本轮 topic 是由 `Kdoctor` 自动创建时，才会尝试清理

## 16. 推荐的实际使用姿势

如果你是运维或开发同学，我建议你这样用：

### 第一步：先做最小探测

Windows：

```powershell
.\kdoctor.exe probe --bootstrap 192.168.1.1:9192
```

Linux：

```bash
./kdoctor probe --bootstrap 192.168.1.1:9192
```

### 第二步：如果你有 compose，再补一轮

Windows：

```powershell
.\kdoctor.exe probe --bootstrap 192.168.1.1:9192 --compose .\docker-compose.yml
```

Linux：

```bash
./kdoctor probe --bootstrap 192.168.1.1:9192 --compose ./docker-compose.yml
```

### 第三步：如果你要发给别人看，输出 Markdown

Windows：

```powershell
.\kdoctor.exe probe --bootstrap 192.168.1.1:9192 --format markdown --output .\report.md
```

Linux：

```bash
./kdoctor probe --bootstrap 192.168.1.1:9192 --format markdown --output ./report.md
```

### 第四步：如果这个环境以后会反复检查，固化到 `kdoctor.yaml`

这样你后面只要跑：

Windows：

```powershell
.\kdoctor.exe probe --config .\kdoctor.yaml
```

Linux：

```bash
./kdoctor probe --config ./kdoctor.yaml
```

## 17. 相关文档

- 简版入口说明：[README.md](/d:/project/project/Kdoctor/README.md:1)
- 设计文档：[doc.md](/d:/project/project/Kdoctor/doc.md:1)
- 工程标准：[architecture.md](/d:/project/project/Kdoctor/architecture.md:1)
- 示例配置：[kdoctor.example.yaml](/d:/project/project/Kdoctor/kdoctor.example.yaml:1)

如果你只是想知道“现在该敲什么命令”，看这份手册就够了。  
如果你想知道“为什么工具这么设计”，再去看设计和架构文档。
