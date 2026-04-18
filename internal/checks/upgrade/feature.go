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

type FeatureChecker struct{}

func (FeatureChecker) ID() string     { return "UPG-002" }
func (FeatureChecker) Name() string   { return "feature_finalization" }
func (FeatureChecker) Module() string { return "upgrade" }

func (FeatureChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if bundle == nil || bundle.Compose == nil {
		return rule.NewSkip("UPG-002", "feature_finalization", "upgrade", "当前没有 compose 快照，无法评估 feature/finalization 状态")
	}

	kraftVersions := map[string][]string{}
	metadataVersions := map[string][]string{}
	bootstrapServers := 0
	for _, service := range composeutil.KafkaServices(bundle.Compose) {
		if value := strings.TrimSpace(service.Environment["KAFKA_KRAFT_VERSION"]); value != "" {
			kraftVersions[value] = append(kraftVersions[value], service.ServiceName)
		}
		if value := strings.TrimSpace(service.Environment["KAFKA_CFG_METADATA_VERSION"]); value != "" {
			metadataVersions[value] = append(metadataVersions[value], service.ServiceName)
		}
		if strings.TrimSpace(service.Environment["KAFKA_CFG_CONTROLLER_QUORUM_BOOTSTRAP_SERVERS"]) != "" {
			bootstrapServers++
		}
	}

	evidence := []string{}
	appendVersionEvidence := func(prefix string, groups map[string][]string) {
		keys := make([]string, 0, len(groups))
		for key := range groups {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			evidence = append(evidence, fmt.Sprintf("%s=%s services=%v", prefix, key, groups[key]))
		}
	}
	appendVersionEvidence("kraft_version", kraftVersions)
	appendVersionEvidence("metadata_version", metadataVersions)
	if bootstrapServers > 0 {
		evidence = append(evidence, fmt.Sprintf("controller_quorum_bootstrap_servers_services=%d", bootstrapServers))
	}

	result := rule.NewPass("UPG-002", "feature_finalization", "upgrade", "未见明显的 KRaft/metadata feature 版本分裂")
	result.Evidence = evidence
	if len(kraftVersions) > 1 || len(metadataVersions) > 1 {
		result = rule.NewWarn("UPG-002", "feature_finalization", "upgrade", "KRaft 或 metadata feature 版本在 broker 之间不一致")
		result.Evidence = evidence
		result.NextActions = []string{"确认升级是否仍在进行中", "如升级已结束，请统一 kraft/metadata 相关版本配置", "避免长期停留在半升级、半 finalization 状态"}
		return result
	}
	if bootstrapServers > 0 {
		result = rule.NewWarn("UPG-002", "feature_finalization", "upgrade", "检测到 controller.quorum.bootstrap.servers 配置，建议确认是否已完成迁移与 finalization 收口")
		result.Evidence = evidence
		result.NextActions = []string{"确认当前 KRaft 配置路径是否已完全迁移", "检查是否仍保留旧的 bootstrap/finalization 兼容配置", "结合版本升级状态继续判断是否需要清理过渡参数"}
	}
	return result
}
