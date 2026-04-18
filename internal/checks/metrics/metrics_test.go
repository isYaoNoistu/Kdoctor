package metrics

import (
	"context"
	"testing"

	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

func TestUnderReplicatedCheckerWarns(t *testing.T) {
	bundle := &snapshot.Bundle{
		Metrics: &snapshot.MetricsSnapshot{
			Collected: true,
			Available: true,
			Endpoints: []snapshot.MetricsEndpointStatus{
				{
					Address: "127.0.0.1:5556",
					Metrics: map[string]float64{
						"kafka_server_replicamanager_underreplicatedpartitions": 2,
					},
				},
			},
		},
	}

	result := UnderReplicatedChecker{WarnCount: 1}.Run(context.Background(), bundle)
	if result.Status != model.StatusWarn {
		t.Fatalf("expected warn, got %s", result.Status)
	}
}

func TestMinISRCheckerFailsOnUnderMinISR(t *testing.T) {
	bundle := &snapshot.Bundle{
		Metrics: &snapshot.MetricsSnapshot{
			Collected: true,
			Available: true,
			Endpoints: []snapshot.MetricsEndpointStatus{
				{
					Address: "127.0.0.1:5556",
					Metrics: map[string]float64{
						"kafka_server_replicamanager_underminisrpartitioncount": 1,
						"kafka_server_replicamanager_atminisrpartitioncount":    2,
					},
				},
			},
		},
	}

	result := MinISRChecker{UnderMinISRCrit: 1, AtMinISRWarn: 1}.Run(context.Background(), bundle)
	if result.Status != model.StatusFail {
		t.Fatalf("expected fail, got %s", result.Status)
	}
}

func TestNetworkIdleCheckerWarnsOnLowIdle(t *testing.T) {
	bundle := &snapshot.Bundle{
		Metrics: &snapshot.MetricsSnapshot{
			Collected: true,
			Available: true,
			Endpoints: []snapshot.MetricsEndpointStatus{
				{
					Address: "127.0.0.1:5556",
					Metrics: map[string]float64{
						"kafka_network_socketserver_networkprocessoravgidlepercent": 0.2,
					},
				},
			},
		},
	}

	result := NetworkIdleChecker{Warn: 0.3, Crit: 0.1}.Run(context.Background(), bundle)
	if result.Status != model.StatusWarn {
		t.Fatalf("expected warn, got %s", result.Status)
	}
}
