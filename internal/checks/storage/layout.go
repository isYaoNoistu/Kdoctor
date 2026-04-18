package storage

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type LayoutChecker struct{}

func (LayoutChecker) ID() string     { return "STG-003" }
func (LayoutChecker) Name() string   { return "metadata_logdir_layout" }
func (LayoutChecker) Module() string { return "storage" }

func (LayoutChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	kafkaServices := services(bundle)
	if len(kafkaServices) == 0 {
		return rule.NewSkip("STG-003", "metadata_logdir_layout", "storage", "当前输入模式下没有可用的 compose Kafka 服务，无法评估存储目录规划")
	}

	evidence := make([]string, 0)
	failures := 0
	warnings := 0
	for _, service := range kafkaServices {
		logDirs, metadataDir := storagePaths(service)
		serviceRoles := roles(service)
		if len(logDirs) == 0 {
			failures++
			evidence = append(evidence, fmt.Sprintf("服务=%s 缺少 KAFKA_CFG_LOG_DIRS", service.ServiceName))
			continue
		}

		evidence = append(evidence, fmt.Sprintf("服务=%s log.dirs=%v", service.ServiceName, logDirs))
		if hasRole(serviceRoles, "controller") {
			if metadataDir == "" {
				warnings++
				evidence = append(evidence, fmt.Sprintf("服务=%s 未显式配置 KAFKA_CFG_METADATA_LOG_DIR", service.ServiceName))
				continue
			}

			shared := false
			for _, logDir := range logDirs {
				if overlaps(logDir, metadataDir) {
					shared = true
					break
				}
			}
			if shared {
				warnings++
				evidence = append(evidence, fmt.Sprintf("服务=%s metadata.log.dir=%s 与数据目录重叠", service.ServiceName, metadataDir))
				continue
			}
			evidence = append(evidence, fmt.Sprintf("服务=%s metadata.log.dir=%s 与数据目录分离", service.ServiceName, metadataDir))
		}
	}

	if failures > 0 {
		result := rule.NewFail("STG-003", "metadata_logdir_layout", "storage", "Kafka 存储目录配置不完整，至少有服务缺少 log.dirs")
		result.Evidence = evidence
		result.NextActions = []string{"补齐 KAFKA_CFG_LOG_DIRS", "为 KRaft 节点规划明确的数据目录与 metadata 目录", "复核 compose 与 broker 实际目录一致性"}
		return result
	}
	if warnings > 0 {
		result := rule.NewWarn("STG-003", "metadata_logdir_layout", "storage", "KRaft metadata 目录与数据目录的规划存在风险")
		result.Evidence = evidence
		result.NextActions = []string{"为 controller/broker 节点单独设置 KAFKA_CFG_METADATA_LOG_DIR", "避免 metadata 目录与数据目录共用同一路径", "结合挂载检查确认宿主机目录规划清晰"}
		return result
	}

	result := rule.NewPass("STG-003", "metadata_logdir_layout", "storage", "Kafka 数据目录与 KRaft metadata 目录规划清晰")
	result.Evidence = evidence
	return result
}
