package metrics

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type MinISRChecker struct {
	UnderMinISRCrit int
	AtMinISRWarn    int
}

func (MinISRChecker) ID() string     { return "MET-002" }
func (MinISRChecker) Name() string   { return "under_min_isr_partitions" }
func (MinISRChecker) Module() string { return "metrics" }

func (c MinISRChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if result, skip := skipIfUnavailable("MET-002", "under_min_isr_partitions", "metrics", bundle); skip {
		return result
	}

	if c.UnderMinISRCrit <= 0 {
		c.UnderMinISRCrit = 1
	}
	if c.AtMinISRWarn <= 0 {
		c.AtMinISRWarn = 1
	}

	underMinISR, underFound, underEvidence := aggregateMax(metricsSnap(bundle), "kafka_server_replicamanager_underminisrpartitioncount")
	atMinISR, atFound, atEvidence := aggregateMax(metricsSnap(bundle), "kafka_server_replicamanager_atminisrpartitioncount")
	if !underFound && !atFound {
		return rule.NewSkip("MET-002", "under_min_isr_partitions", "metrics", "当前 JMX 指标中没有 UnderMinISR/AtMinISR 相关指标")
	}

	evidence := append([]string{}, underEvidence...)
	evidence = append(evidence, atEvidence...)
	if underFound {
		evidence = append(evidence, fmt.Sprintf("聚合 under_min_isr=%.0f", underMinISR))
	}
	if atFound {
		evidence = append(evidence, fmt.Sprintf("聚合 at_min_isr=%.0f", atMinISR))
	}

	result := rule.NewPass("MET-002", "under_min_isr_partitions", "metrics", "JMX 未发现 UnderMinISR / AtMinISR 压力")
	result.Evidence = evidence
	if int(underMinISR) >= c.UnderMinISRCrit {
		result = rule.NewFail("MET-002", "under_min_isr_partitions", "metrics", "JMX 检测到 UnderMinISRPartitionCount 大于 0")
		result.Evidence = evidence
		result.NextActions = []string{"优先检查 ISR、磁盘和复制链路", "确认 acks=all 业务是否已经受到影响", "结合 TOP-005 与客户端探针交叉判断"}
		return result
	}
	if int(atMinISR) >= c.AtMinISRWarn {
		result = rule.NewWarn("MET-002", "under_min_isr_partitions", "metrics", "JMX 检测到 AtMinISRPartitionCount 大于 0")
		result.Evidence = evidence
		result.NextActions = []string{"当前写入链路已经接近 min.insync.replicas 边界", "提前排查 follower broker 与复制延迟", "在流量上升前持续观察 ISR 恢复情况"}
	}
	return result
}
