package consumer

import (
	"fmt"
	"strings"

	"kdoctor/internal/config"
	"kdoctor/internal/snapshot"
)

func groupSnap(bundle *snapshot.Bundle) *snapshot.GroupSnapshot {
	if bundle == nil {
		return nil
	}
	return bundle.Group
}

func targetKey(groupID, topic string) string {
	return strings.TrimSpace(groupID) + "|" + strings.TrimSpace(topic)
}

func thresholdsForTarget(target snapshot.GroupLagSnapshot, defaults map[string]config.GroupProbeTarget, warnLag, critLag int64) (int64, int64) {
	cfg, ok := defaults[targetKey(target.GroupID, target.Topic)]
	if ok {
		if cfg.LagWarn > 0 {
			warnLag = cfg.LagWarn
		}
		if cfg.LagCrit > 0 {
			critLag = cfg.LagCrit
		}
	}
	return warnLag, critLag
}

func buildTargetMap(targets []config.GroupProbeTarget) map[string]config.GroupProbeTarget {
	result := make(map[string]config.GroupProbeTarget, len(targets))
	for _, target := range targets {
		groupID := strings.TrimSpace(target.GroupID)
		if groupID == "" {
			groupID = strings.TrimSpace(target.Name)
		}
		if groupID == "" || strings.TrimSpace(target.Topic) == "" {
			continue
		}
		result[targetKey(groupID, target.Topic)] = target
	}
	return result
}

func summarizeTargets(groups *snapshot.GroupSnapshot) []string {
	if groups == nil {
		return nil
	}
	out := make([]string, 0, len(groups.Targets)+len(groups.Errors))
	for _, target := range groups.Targets {
		out = append(out, fmt.Sprintf("group_id=%s topic=%s state=%s lag=%d coordinator=%s missing_offsets=%d", target.GroupID, target.Topic, target.State, target.TotalLag, target.Coordinator, target.MissingOffsets))
	}
	for _, err := range groups.Errors {
		out = append(out, fmt.Sprintf("collector_error=%s", err))
	}
	return out
}
