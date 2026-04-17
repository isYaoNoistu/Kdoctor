package logs

import (
	"context"
	"testing"

	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

func TestSourcesCheckerWarnsWhenSourcesAreStaleOrSparse(t *testing.T) {
	result := SourcesChecker{}.Run(context.Background(), &snapshot.Bundle{
		Logs: &snapshot.LogSnapshot{
			Collected: true,
			Sources:   []string{"file:/tmp/server.log"},
			SourceStats: []snapshot.LogSourceStat{
				{
					Source:           "file:/tmp/server.log",
					Kind:             "file",
					Lines:            3,
					Bytes:            128,
					LastModifiedUnix: 1700000000,
					Fresh:            false,
					SufficientLines:  false,
				},
			},
		},
	})

	if result.Status != model.StatusWarn {
		t.Fatalf("expected WARN, got %s", result.Status)
	}
}
