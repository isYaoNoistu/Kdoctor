package network

import (
	"context"
	"testing"

	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

func TestListenerCheckerWarnsForPrivateControllersFromExternalView(t *testing.T) {
	result := ListenerChecker{}.Run(context.Background(), &snapshot.Bundle{
		Network: &snapshot.NetworkSnapshot{
			BootstrapChecks: []snapshot.EndpointCheck{
				{Kind: "bootstrap", Address: "203.0.113.10:9292", Reachable: true, DurationMs: 10},
			},
			ControllerChecks: []snapshot.EndpointCheck{
				{Kind: "controller", Address: "192.168.1.10:9193", Reachable: false, Error: "dial timeout"},
			},
		},
	})

	if result.Status != model.StatusWarn {
		t.Fatalf("expected WARN, got %s", result.Status)
	}
}
