package kafka

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type ClusterChecker struct{}

func (ClusterChecker) ID() string     { return "KFK-001" }
func (ClusterChecker) Name() string   { return "cluster_metadata" }
func (ClusterChecker) Module() string { return "kafka" }

func (ClusterChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Kafka == nil {
		return rule.NewError("KFK-001", "cluster_metadata", "kafka", "kafka metadata unavailable", "kafka snapshot missing")
	}

	result := rule.NewPass("KFK-001", "cluster_metadata", "kafka", "cluster metadata retrieved successfully")
	result.Evidence = []string{
		fmt.Sprintf("cluster_id=%s", snap.Kafka.ClusterID),
		fmt.Sprintf("brokers=%d", len(snap.Kafka.Brokers)),
	}
	if snap.Kafka.ControllerID != nil {
		result.Evidence = append(result.Evidence, fmt.Sprintf("controller_id=%d", *snap.Kafka.ControllerID))
	}
	return result
}
