package kafka

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"kdoctor/internal/composeutil"
	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type TopologyMismatchChecker struct{}

func (TopologyMismatchChecker) ID() string     { return "KFK-009" }
func (TopologyMismatchChecker) Name() string   { return "topology_mismatch" }
func (TopologyMismatchChecker) Module() string { return "kafka" }

func (TopologyMismatchChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Kafka == nil {
		return rule.NewSkip("KFK-009", "topology_mismatch", "kafka", "当前没有可用的 Kafka metadata，无法评估集群拓扑是否偏离")
	}
	if snap.Compose == nil {
		return rule.NewSkip("KFK-009", "topology_mismatch", "kafka", "当前没有 compose 快照，暂不评估静态拓扑与运行态偏离")
	}

	services := composeutil.KafkaServices(snap.Compose)
	if len(services) == 0 {
		return rule.NewSkip("KFK-009", "topology_mismatch", "kafka", "compose 中没有识别到 Kafka 服务")
	}

	expectedIDs := map[int32]struct{}{}
	for _, service := range services {
		nodeID := strings.TrimSpace(service.Environment["KAFKA_CFG_NODE_ID"])
		if nodeID == "" {
			continue
		}
		var id int
		fmt.Sscanf(nodeID, "%d", &id)
		if id > 0 {
			expectedIDs[int32(id)] = struct{}{}
		}
	}

	missing := make([]int32, 0)
	extra := make([]int32, 0)
	actualIDs := map[int32]struct{}{}
	for _, broker := range snap.Kafka.Brokers {
		actualIDs[broker.ID] = struct{}{}
	}
	for id := range expectedIDs {
		if _, ok := actualIDs[id]; !ok {
			missing = append(missing, id)
		}
	}
	for id := range actualIDs {
		if len(expectedIDs) > 0 {
			if _, ok := expectedIDs[id]; !ok {
				extra = append(extra, id)
			}
		}
	}
	sort.Slice(missing, func(i, j int) bool { return missing[i] < missing[j] })
	sort.Slice(extra, func(i, j int) bool { return extra[i] < extra[j] })

	result := rule.NewPass("KFK-009", "topology_mismatch", "kafka", "运行态拓扑与 compose 描述一致")
	result.Evidence = []string{
		fmt.Sprintf("compose_brokers=%d metadata_brokers=%d", len(services), len(snap.Kafka.Brokers)),
	}
	if len(missing) > 0 || len(extra) > 0 || len(services) != len(snap.Kafka.Brokers) {
		result = rule.NewWarn("KFK-009", "topology_mismatch", "kafka", "运行态拓扑与 compose 描述存在偏离")
		result.Evidence = []string{
			fmt.Sprintf("compose_brokers=%d metadata_brokers=%d", len(services), len(snap.Kafka.Brokers)),
			fmt.Sprintf("missing_ids=%v", missing),
			fmt.Sprintf("extra_ids=%v", extra),
		}
		result.NextActions = []string{"确认 compose 描述是否仍代表当前真实环境", "核对 broker 是否有新增/下线/重建", "结合 node.id、listeners 与 metadata 一起检查拓扑漂移"}
	}
	return result
}
