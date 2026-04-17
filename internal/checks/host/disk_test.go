package host

import (
	"context"
	"testing"

	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

func TestDiskCheckerWarnsWhenDiskIsNearCapacity(t *testing.T) {
	result := DiskChecker{}.Run(context.Background(), &snapshot.Bundle{
		Host: &snapshot.HostSnapshot{
			Collected: true,
			DiskUsages: []snapshot.DiskUsage{
				{Path: "/data", UsedPercent: 89.5, AvailableBytes: 100},
			},
		},
	})

	if result.Status != model.StatusWarn {
		t.Fatalf("expected WARN, got %s", result.Status)
	}
}
