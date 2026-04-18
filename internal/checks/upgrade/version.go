package upgrade

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

type RollingVersionChecker struct{}

func (RollingVersionChecker) ID() string     { return "UPG-001" }
func (RollingVersionChecker) Name() string   { return "rolling_upgrade_state" }
func (RollingVersionChecker) Module() string { return "upgrade" }

func (RollingVersionChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if bundle == nil || bundle.Compose == nil {
		return rule.NewSkip("UPG-001", "rolling_upgrade_state", "upgrade", "当前没有 compose 快照，无法评估 rolling upgrade 状态")
	}

	versions := map[string][]string{}
	for _, service := range composeutil.KafkaServices(bundle.Compose) {
		version := imageVersion(service.Image)
		if version == "" {
			version = "unknown"
		}
		versions[version] = append(versions[version], service.ServiceName)
	}
	if len(versions) == 0 {
		return rule.NewSkip("UPG-001", "rolling_upgrade_state", "upgrade", "compose 中没有识别到 Kafka 服务版本信息")
	}

	keys := make([]string, 0, len(versions))
	evidence := []string{}
	for version := range versions {
		keys = append(keys, version)
	}
	sort.Strings(keys)
	for _, version := range keys {
		evidence = append(evidence, fmt.Sprintf("version=%s services=%v", version, versions[version]))
	}

	result := rule.NewPass("UPG-001", "rolling_upgrade_state", "upgrade", "Kafka broker 版本一致，未见 rolling upgrade 半完成迹象")
	result.Evidence = evidence
	if len(keys) > 1 {
		result = rule.NewWarn("UPG-001", "rolling_upgrade_state", "upgrade", "检测到多个 Kafka broker 版本并存，可能正处于 rolling upgrade 半完成状态")
		result.Evidence = evidence
		result.NextActions = []string{"确认当前是否处于计划内升级窗口", "如升级已结束，请尽快完成剩余 broker 收口", "结合 feature/finalization 状态继续确认元数据是否已完成收敛"}
	}
	return result
}

func imageVersion(image string) string {
	image = strings.TrimSpace(image)
	if image == "" {
		return ""
	}
	if idx := strings.LastIndex(image, ":"); idx >= 0 && idx < len(image)-1 {
		return image[idx+1:]
	}
	return image
}
