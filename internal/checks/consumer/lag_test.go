package consumer

import (
	"context"
	"testing"

	"kdoctor/internal/config"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

func TestLagCheckerEscalatesWhenLagExceedsThreshold(t *testing.T) {
	result := LagChecker{
		WarnLag: 100,
		CritLag: 500,
		Targets: []config.GroupProbeTarget{
			{Name: "critical-group", Topic: "orders"},
		},
	}.Run(context.Background(), &snapshot.Bundle{
		Group: &snapshot.GroupSnapshot{
			Collected: true,
			Targets: []snapshot.GroupLagSnapshot{
				{
					Name:        "critical-group",
					GroupID:     "critical-group",
					Topic:       "orders",
					State:       "Stable",
					Coordinator: "broker-1:9092",
					TotalLag:    800,
				},
			},
		},
	})

	if result.Status != model.StatusCrit {
		t.Fatalf("expected CRIT, got %s", result.Status)
	}
}

func TestCoordinatorCheckerWarnsOnMissingOffsets(t *testing.T) {
	result := CoordinatorChecker{}.Run(context.Background(), &snapshot.Bundle{
		Group: &snapshot.GroupSnapshot{
			Collected: true,
			Targets: []snapshot.GroupLagSnapshot{
				{
					GroupID:        "group-a",
					Topic:          "payments",
					State:          "Stable",
					Coordinator:    "broker-2:9092",
					MissingOffsets: 1,
				},
			},
		},
	})

	if result.Status != model.StatusWarn {
		t.Fatalf("expected WARN, got %s", result.Status)
	}
}
