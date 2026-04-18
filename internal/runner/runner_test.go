package runner

import "testing"

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
