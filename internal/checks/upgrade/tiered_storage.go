package upgrade

import (
	"context"
	"fmt"
	"strings"

	"kdoctor/internal/composeutil"
	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type TieredStorageChecker struct{}

func (TieredStorageChecker) ID() string     { return "UPG-003" }
func (TieredStorageChecker) Name() string   { return "tiered_storage_awareness" }
func (TieredStorageChecker) Module() string { return "upgrade" }

func (TieredStorageChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if bundle == nil || bundle.Compose == nil {
		return rule.NewSkip("UPG-003", "tiered_storage_awareness", "upgrade", "当前没有 compose 快照，无法评估 tiered storage 配置")
	}

	enabledServices := []string{}
	evidence := []string{}
	for _, service := range composeutil.KafkaServices(bundle.Compose) {
		if enabled := remoteStorageEnabled(service.Environment); enabled {
			enabledServices = append(enabledServices, service.ServiceName)
			evidence = append(evidence, fmt.Sprintf("service=%s remote_storage_enabled=true", service.ServiceName))
		}
	}

	if len(enabledServices) == 0 {
		return rule.NewSkip("UPG-003", "tiered_storage_awareness", "upgrade", "当前未发现 tiered storage 相关配置")
	}

	result := rule.NewWarn("UPG-003", "tiered_storage_awareness", "upgrade", "检测到 tiered storage 相关配置，建议继续核对 fetch 与远端存储配套参数")
	result.Evidence = evidence
	result.NextActions = []string{"确认远端存储参数与 broker 版本兼容", "检查客户端 fetch 相关参数是否与远端读取场景匹配", "结合日志继续确认没有 remote storage 相关异常"}
	return result
}

func remoteStorageEnabled(env map[string]string) bool {
	keys := []string{
		"KAFKA_CFG_REMOTE_LOG_STORAGE_SYSTEM_ENABLE",
		"KAFKA_CFG_REMOTE_STORAGE_ENABLE",
		"KAFKA_CFG_TIERED_STORAGE_ENABLE",
	}
	for _, key := range keys {
		if strings.EqualFold(strings.TrimSpace(env[key]), "true") {
			return true
		}
	}
	return false
}
