package storage

import (
	"context"
	"fmt"

	"kdoctor/internal/composeutil"
	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type MountPlanningChecker struct{}

func (MountPlanningChecker) ID() string     { return "STG-005" }
func (MountPlanningChecker) Name() string   { return "storage_mount_planning" }
func (MountPlanningChecker) Module() string { return "storage" }

func (MountPlanningChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	kafkaServices := services(bundle)
	if len(kafkaServices) == 0 {
		return rule.NewSkip("STG-005", "storage_mount_planning", "storage", "当前输入模式下没有可用的 compose Kafka 服务，无法评估存储挂载规划")
	}

	evidence := make([]string, 0)
	failures := 0
	warnings := 0
	for _, service := range kafkaServices {
		logDirs, metadataDir := storagePaths(service)
		required := append([]string(nil), logDirs...)
		if metadataDir != "" {
			required = append(required, metadataDir)
		}
		for _, path := range required {
			mount, ok := coveringVolume(service, path)
			if !ok {
				failures++
				evidence = append(evidence, fmt.Sprintf("服务=%s 路径=%s 没有对应的 volume 挂载", service.ServiceName, path))
				continue
			}

			if mount.NamedVolume {
				warnings++
				evidence = append(evidence, fmt.Sprintf("服务=%s 路径=%s 使用 named volume=%s", service.ServiceName, path, mount.Source))
				continue
			}

			hostPath := composeutil.ResolveHostPath(bundle.Compose.SourcePath, mount.Source)
			evidence = append(evidence, fmt.Sprintf("服务=%s 路径=%s 绑定到宿主机=%s", service.ServiceName, path, hostPath))
		}
	}

	if failures > 0 {
		result := rule.NewFail("STG-005", "storage_mount_planning", "storage", "部分 Kafka 存储路径没有 volume 承载，存在数据持久化风险")
		result.Evidence = evidence
		result.NextActions = []string{"为 Kafka 数据目录与 metadata 目录显式配置 volume", "优先使用 bind mount 映射到宿主机路径", "避免把 Kafka 状态只保存在容器临时层"}
		return result
	}
	if warnings > 0 {
		result := rule.NewWarn("STG-005", "storage_mount_planning", "storage", "Kafka 存储路径已挂载，但部分路径仍依赖 named volume")
		result.Evidence = evidence
		result.NextActions = []string{"评估是否改为 bind mount 以提升排障可见性", "确认 named volume 已纳入备份与容量管理", "结合 Docker 实际 inspect 结果继续核对挂载状态"}
		return result
	}

	result := rule.NewPass("STG-005", "storage_mount_planning", "storage", "Kafka 存储路径已显式挂载到宿主机路径")
	result.Evidence = evidence
	return result
}
