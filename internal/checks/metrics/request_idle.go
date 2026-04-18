package metrics

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type RequestIdleChecker struct {
	Warn float64
	Crit float64
}

func (RequestIdleChecker) ID() string     { return "JVM-002" }
func (RequestIdleChecker) Name() string   { return "request_handler_idle" }
func (RequestIdleChecker) Module() string { return "jvm" }

func (c RequestIdleChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if result, skip := skipIfUnavailable("JVM-002", "request_handler_idle", "jvm", bundle); skip {
		return result
	}

	if c.Warn <= 0 {
		c.Warn = 0.3
	}
	if c.Crit <= 0 {
		c.Crit = 0.1
	}

	value, ok, evidence := aggregateMin(metricsSnap(bundle), "kafka_server_kafkarequesthandlerpool_requesthandleravgidlepercent")
	if !ok {
		return rule.NewSkip("JVM-002", "request_handler_idle", "jvm", "当前 JMX 指标中没有 RequestHandlerAvgIdlePercent")
	}
	evidence = append(evidence, fmt.Sprintf("聚合最小 request_idle=%.3f", value))

	result := rule.NewPass("JVM-002", "request_handler_idle", "jvm", "请求处理线程空闲率处于安全范围")
	result.Evidence = evidence
	if value <= c.Crit {
		result = rule.NewFail("JVM-002", "request_handler_idle", "jvm", "请求处理线程空闲率持续偏低，broker 可能已接近处理瓶颈")
		result.Evidence = evidence
		result.NextActions = []string{"检查请求排队、磁盘与副本压力", "结合 JVM/GC 与网络线程 idle 一起判断", "关注是否已经影响 produce/fetch 延迟"}
		return result
	}
	if value <= c.Warn {
		result = rule.NewWarn("JVM-002", "request_handler_idle", "jvm", "请求处理线程空闲率已经偏低，建议提前关注 broker 压力")
		result.Evidence = evidence
		result.NextActions = []string{"持续观察请求处理线程 idle 百分比", "结合磁盘、ISR 与请求延迟继续排查", "在流量高峰前评估 broker 处理余量"}
	}
	return result
}
