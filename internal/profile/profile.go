package profile

import "kdoctor/internal/config"

func BuiltinProfiles() map[string]config.ProfileConfig {
	return map[string]config.ProfileConfig{
		"generic-bootstrap": {
			ExecutionView:     "auto",
			SecurityMode:      "plaintext",
			PlaintextExternal: true,
		},
		"single-host-3broker-kraft-prod": {
			BootstrapExternal:         []string{"203.0.113.10:9392", "203.0.113.10:9394", "203.0.113.10:9396"},
			BootstrapInternal:         []string{"192.168.1.10:9192", "192.168.1.10:9194", "192.168.1.10:9196"},
			ControllerEndpoints:       []string{"192.168.1.10:9193", "192.168.1.10:9195", "192.168.1.10:9197"},
			BrokerCount:               3,
			ExpectedMinISR:            2,
			ExpectedReplicationFactor: 3,
			ExecutionView:             "external",
			SecurityMode:              "plaintext",
			HostNetwork:               true,
			PlaintextExternal:         true,
		},
		"single-host-3broker-kraft-uat": {
			BootstrapExternal:         []string{"203.0.113.10:9292", "203.0.113.10:9294", "203.0.113.10:9296"},
			BootstrapInternal:         []string{"192.168.1.10:9192", "192.168.1.10:9194", "192.168.1.10:9196"},
			ControllerEndpoints:       []string{"192.168.1.10:9193", "192.168.1.10:9195", "192.168.1.10:9197"},
			BrokerCount:               3,
			ExpectedMinISR:            2,
			ExpectedReplicationFactor: 3,
			ExecutionView:             "external",
			SecurityMode:              "plaintext",
			HostNetwork:               true,
			PlaintextExternal:         true,
		},
	}
}
