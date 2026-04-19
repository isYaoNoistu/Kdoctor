package kafka

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"kdoctor/internal/composeutil"
	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type RegistrationIntegrityChecker struct{}

func (RegistrationIntegrityChecker) ID() string     { return "KFK-006" }
func (RegistrationIntegrityChecker) Name() string   { return "broker_registration_integrity" }
func (RegistrationIntegrityChecker) Module() string { return "kafka" }

func (RegistrationIntegrityChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Kafka == nil {
		return rule.NewSkip("KFK-006", "broker_registration_integrity", "kafka", "当前没有可用的 Kafka metadata，无法评估 broker 注册完整性")
	}

	expectedCount := snap.Kafka.ExpectedBrokerCount
	expectedIDs := make([]int, 0)
	if snap.Compose != nil {
		for _, service := range composeutil.KafkaServices(snap.Compose) {
			if raw := strings.TrimSpace(service.Environment["KAFKA_CFG_NODE_ID"]); raw != "" {
				if id, err := strconv.Atoi(raw); err == nil {
					expectedIDs = append(expectedIDs, id)
				}
			}
		}
	}
	sort.Ints(expectedIDs)

	actualIDs := make([]int, 0, len(snap.Kafka.Brokers))
	for _, broker := range snap.Kafka.Brokers {
		actualIDs = append(actualIDs, int(broker.ID))
	}
	sort.Ints(actualIDs)

	evidence := []string{fmt.Sprintf("实际 broker ID=%v", actualIDs)}
	if expectedCount > 0 {
		evidence = append(evidence, fmt.Sprintf("expected_count=%d", expectedCount))
	}
	if len(expectedIDs) > 0 {
		evidence = append(evidence, fmt.Sprintf("expected_ids=%v", expectedIDs))
	}

	result := rule.NewPass("KFK-006", "broker_registration_integrity", "kafka", "broker 注册集合完整")
	result.Evidence = evidence
	if expectedCount > 0 && len(actualIDs) < expectedCount {
		result = rule.NewFail("KFK-006", "broker_registration_integrity", "kafka", "broker 注册数量低于期望，存在未完成注册或节点缺失")
		result.Evidence = evidence
		result.NextActions = []string{"检查缺失 broker 进程是否运行", "核对 controller、node.id 与 listeners 配置", "查看 broker 日志中的注册失败或元数据收敛异常"}
		return result
	}
	if len(expectedIDs) > 0 && !sameIntSet(expectedIDs, actualIDs) {
		result = rule.NewWarn("KFK-006", "broker_registration_integrity", "kafka", "compose 期望的 broker 集合与 metadata 注册集合不完全一致")
		result.Evidence = evidence
		result.NextActions = []string{"确认环境是否发生 broker 重建或拓扑变更", "检查 compose 与当前 metadata 是否已经脱节", "结合 KFK-009 一起判断拓扑偏移"}
	}
	return result
}

func sameIntSet(left []int, right []int) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
