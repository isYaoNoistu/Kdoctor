package metrics

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type NetworkIdleChecker struct {
	Warn float64
	Crit float64
}

func (NetworkIdleChecker) ID() string     { return "JVM-001" }
func (NetworkIdleChecker) Name() string   { return "network_processor_idle" }
func (NetworkIdleChecker) Module() string { return "jvm" }

func (c NetworkIdleChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if result, skip := skipIfUnavailable("JVM-001", "network_processor_idle", "jvm", bundle); skip {
		return result
	}

	if c.Warn <= 0 {
		c.Warn = 0.3
	}
	if c.Crit <= 0 {
		c.Crit = 0.1
	}

	value, ok, evidence := aggregateMin(metricsSnap(bundle), "kafka_network_socketserver_networkprocessoravgidlepercent")
	if !ok {
		return rule.NewSkip("JVM-001", "network_processor_idle", "jvm", "当前 JMX 指标中没有 NetworkProcessorAvgIdlePercent")
	}
	evidence = append(evidence, fmt.Sprintf("聚合最小 network_idle=%.3f", value))

	result := rule.NewPass("JVM-001", "network_processor_idle", "jvm", "网络线程空闲率处于安全范围")
	result.Evidence = evidence
	if value <= c.Crit {
		result = rule.NewFail("JVM-001", "network_processor_idle", "jvm", "网络线程空闲率持续偏低，broker 可能已接近网络处理瓶颈")
		result.Evidence = evidence
		result.NextActions = []string{"检查连接数、请求延迟与流量高峰", "结合 listener、quota 与网络路径继续排查", "观察是否伴随请求线程或 ISR 异常"}
		return result
	}
	if value <= c.Warn {
		result = rule.NewWarn("JVM-001", "network_processor_idle", "jvm", "网络线程空闲率已经偏低，建议提前关注 broker 网络压力")
		result.Evidence = evidence
		result.NextActions = []string{"持续观察网络线程 idle 百分比", "排查短时流量尖峰与连接风暴", "结合请求处理线程 idle 一起判断 broker 压力"}
	}
	return result
}
