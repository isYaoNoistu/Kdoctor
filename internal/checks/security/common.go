package security

import (
	"sort"
	"strings"

	"kdoctor/internal/composeutil"
	"kdoctor/internal/snapshot"
)

func services(bundle *snapshot.Bundle) []composeutil.KafkaService {
	if bundle == nil || bundle.Compose == nil {
		return nil
	}
	return composeutil.KafkaServices(bundle.Compose)
}

func selectedClientListeners(service composeutil.KafkaService, executionView string) []string {
	raw := strings.TrimSpace(service.Environment["KAFKA_CFG_ADVERTISED_LISTENERS"])
	if raw == "" {
		raw = strings.TrimSpace(service.Environment["KAFKA_CFG_LISTENERS"])
	}

	listeners, err := composeutil.ParseListeners(raw)
	if err != nil || len(listeners) == 0 {
		return nil
	}

	preferred := preferredListenerName(executionView)
	names := make([]string, 0)
	if preferred != "" {
		for name := range listeners {
			if strings.EqualFold(name, preferred) {
				names = append(names, name)
			}
		}
	}
	if len(names) == 0 {
		for name := range listeners {
			if strings.EqualFold(name, "CONTROLLER") {
				continue
			}
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func preferredListenerName(executionView string) string {
	switch strings.ToLower(strings.TrimSpace(executionView)) {
	case "external":
		return "EXTERNAL"
	case "internal", "host-network", "docker-container", "bastion":
		return "INTERNAL"
	default:
		return ""
	}
}

func normalizeSecurityMode(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "tls":
		return "ssl"
	default:
		return value
	}
}

func protocolMatchesMode(mode string, protocol string) bool {
	mode = normalizeSecurityMode(mode)
	protocol = strings.ToUpper(strings.TrimSpace(protocol))
	switch mode {
	case "", "plaintext":
		return protocol == "PLAINTEXT"
	case "ssl":
		return protocol == "SSL" || protocol == "SASL_SSL"
	case "sasl":
		return strings.HasPrefix(protocol, "SASL_")
	case "sasl_plaintext":
		return protocol == "SASL_PLAINTEXT"
	case "sasl_ssl":
		return protocol == "SASL_SSL"
	default:
		return true
	}
}

func protocolNeedsSASL(protocol string) bool {
	return strings.HasPrefix(strings.ToUpper(strings.TrimSpace(protocol)), "SASL_")
}

func splitUpperCSV(input string) []string {
	values := composeutil.ParseCSV(input)
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.ToUpper(strings.TrimSpace(value))
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func contains(values []string, want string) bool {
	want = strings.TrimSpace(want)
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), want) {
			return true
		}
	}
	return false
}
