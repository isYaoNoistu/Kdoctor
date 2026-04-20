package host

import (
	"context"
	"testing"

	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

func TestFDCheckerPrefersContainerOpenFileLimit(t *testing.T) {
	result := FDChecker{}.Run(context.Background(), &snapshot.Bundle{
		Host: &snapshot.HostSnapshot{
			Collected: true,
			FD: &snapshot.FDStats{
				SoftLimit: 1024,
			},
			ContainerFD: []snapshot.ContainerFDStat{
				{Name: "kafka1", SoftLimit: 1048576, HardLimit: 1048576},
				{Name: "kafka2", SoftLimit: 1048576, HardLimit: 1048576},
			},
		},
	})

	if result.Status != model.StatusPass {
		t.Fatalf("expected PASS when container limits are healthy, got %s", result.Status)
	}
}
