package kafka

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"

	"kdoctor/internal/composeutil"
	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type BrokerIdentityChecker struct{}

func (BrokerIdentityChecker) ID() string     { return "KFK-007" }
func (BrokerIdentityChecker) Name() string   { return "broker_identity_conflict" }
func (BrokerIdentityChecker) Module() string { return "kafka" }

func (BrokerIdentityChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Kafka == nil {
		return rule.NewSkip("KFK-007", "broker_identity_conflict", "kafka", "当前没有可用的 Kafka metadata，无法评估 broker 身份冲突")
	}

	evidence := []string{}
	metadataByHostPort := map[string]int32{}
	for _, broker := range snap.Kafka.Brokers {
		host, port, err := net.SplitHostPort(broker.Address)
		if err != nil {
			host = broker.Address
			port = ""
		}
		key := strings.ToLower(host) + ":" + port
		if previous, ok := metadataByHostPort[key]; ok {
			result := rule.NewFail("KFK-007", "broker_identity_conflict", "kafka", "metadata 中出现 broker ID 与地址复用冲突")
			result.Evidence = []string{fmt.Sprintf("address=%s broker_ids=%d,%d", broker.Address, previous, broker.ID)}
			result.NextActions = []string{"检查 broker 是否错误复用了 node.id 或 advertised.listeners", "确认是否存在旧 broker 残留地址", "排查集群拓扑变更后的注册地址漂移"}
			return result
		}
		metadataByHostPort[key] = broker.ID
		evidence = append(evidence, fmt.Sprintf("metadata broker_id=%d address=%s", broker.ID, broker.Address))
	}

	if snap.Compose != nil {
		conflicts := []string{}
		for _, service := range composeutil.KafkaServices(snap.Compose) {
			advertised, err := composeutil.ParseListeners(service.Environment["KAFKA_CFG_ADVERTISED_LISTENERS"])
			if err != nil {
				continue
			}
			names := make([]string, 0, len(advertised))
			for name := range advertised {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				endpoint := advertised[name]
				address := fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
				evidence = append(evidence, fmt.Sprintf("compose service=%s listener=%s address=%s", service.ServiceName, name, address))
			}
			if nodeID := strings.TrimSpace(service.Environment["KAFKA_CFG_NODE_ID"]); nodeID != "" && len(advertised) == 0 {
				conflicts = append(conflicts, fmt.Sprintf("service=%s node.id=%s has no advertised.listeners", service.ServiceName, nodeID))
			}
		}
		if len(conflicts) > 0 {
			result := rule.NewWarn("KFK-007", "broker_identity_conflict", "kafka", "compose 中部分 broker 身份声明不完整，可能导致注册地址漂移")
			result.Evidence = append(evidence, conflicts...)
			result.NextActions = []string{"补齐 advertised.listeners", "确认每个 node.id 都映射到稳定的 broker 地址", "避免 broker 重建后复用错误地址"}
			return result
		}
	}

	result := rule.NewPass("KFK-007", "broker_identity_conflict", "kafka", "broker ID 与地址映射稳定，未见明显身份冲突")
	result.Evidence = evidence
	return result
}
