package runner

import (
	"context"
	"errors"
	"testing"
)

func TestSplitTaskMessages(t *testing.T) {
	messages := []string{
		"采集任务 compose_snapshot 超时，已降级跳过: context deadline exceeded",
		"采集任务 docker_runtime 已降级: docker unavailable",
		"no bootstrap brokers configured",
	}

	degraded, hardErrors := splitTaskMessages(messages)
	if len(degraded) != 2 {
		t.Fatalf("expected 2 degraded messages, got %d", len(degraded))
	}
	if len(hardErrors) != 1 {
		t.Fatalf("expected 1 hard error, got %d", len(hardErrors))
	}
	if hardErrors[0] != "no bootstrap brokers configured" {
		t.Fatalf("unexpected hard error: %q", hardErrors[0])
	}
}

func TestRunTasksReturnsStableSortedErrors(t *testing.T) {
	errs := runTasks(context.Background(),
		taskSpec{
			Name: "zeta",
			Soft: true,
			Run: func(context.Context) error {
				return errors.New("zeta down")
			},
		},
		taskSpec{
			Name: "alpha",
			Soft: true,
			Run: func(context.Context) error {
				return errors.New("alpha down")
			},
		},
	)

	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(errs))
	}
	if errs[0] > errs[1] {
		t.Fatalf("expected stable sorted errors, got %v", errs)
	}
}
