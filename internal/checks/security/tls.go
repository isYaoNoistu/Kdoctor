package security

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"strings"
	"time"

	"kdoctor/internal/composeutil"
	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type TLSChecker struct {
	ExecutionView    string
	CertWarnDays     int
	HandshakeTimeout time.Duration
}

func (TLSChecker) ID() string     { return "SEC-003" }
func (TLSChecker) Name() string   { return "tls_certificate_health" }
func (TLSChecker) Module() string { return "security" }

func (c TLSChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	kafkaServices := services(bundle)
	if len(kafkaServices) == 0 {
		return rule.NewSkip("SEC-003", "tls_certificate_health", "security", "compose Kafka services are not available for TLS certificate inspection")
	}

	targets := make([]string, 0)
	for _, service := range kafkaServices {
		protocols := composeutil.ParseListenerProtocolMap(service.Environment["KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP"])
		if len(protocols) == 0 {
			continue
		}
		listeners, raw := securityListeners(service, c.ExecutionView)
		if len(listeners) == 0 {
			continue
		}
		for _, name := range selectedClientListeners(service, c.ExecutionView) {
			protocol := strings.TrimSpace(protocols[name])
			if !strings.EqualFold(protocol, "SSL") && !strings.EqualFold(protocol, "SASL_SSL") {
				continue
			}
			listener, ok := listeners[name]
			if !ok {
				continue
			}
			targets = append(targets, listener.Raw)
		}
		_ = raw
	}

	if len(targets) == 0 {
		return rule.NewSkip("SEC-003", "tls_certificate_health", "security", "the current listener set does not expose SSL or SASL_SSL endpoints")
	}

	if c.CertWarnDays <= 0 {
		c.CertWarnDays = 30
	}
	if c.HandshakeTimeout <= 0 {
		c.HandshakeTimeout = 5 * time.Second
	}

	failures := 0
	warnings := 0
	evidence := []string{}
	for _, target := range targets {
		host, port, err := net.SplitHostPort(target)
		if err != nil {
			failures++
			evidence = append(evidence, fmt.Sprintf("target=%s parse_error=%v", target, err))
			continue
		}
		result, err := inspectTLS(host, net.JoinHostPort(host, port), c.HandshakeTimeout)
		if err != nil {
			failures++
			evidence = append(evidence, fmt.Sprintf("target=%s tls_error=%v", target, err))
			continue
		}
		evidence = append(evidence, fmt.Sprintf("target=%s subject=%s not_after=%s", target, result.Subject, result.NotAfter.Format(time.RFC3339)))
		if days := int(time.Until(result.NotAfter).Hours() / 24); days < 0 {
			failures++
		} else if days < c.CertWarnDays {
			warnings++
			evidence = append(evidence, fmt.Sprintf("target=%s days_until_expiry=%d", target, days))
		}
	}

	result := rule.NewPass("SEC-003", "tls_certificate_health", "security", "TLS certificate chain and expiry look healthy for the current listener set")
	result.Evidence = evidence
	switch {
	case failures > 0:
		result = rule.NewFail("SEC-003", "tls_certificate_health", "security", "some SSL listeners failed certificate validation, SAN matching, or expiry checks")
		result.Evidence = evidence
		result.NextActions = []string{"verify the certificate chain, CA trust, and SAN entries for the listener hostnames", "compare advertised.listeners hostnames with the certificate subject and SAN set", "renew or replace expired certificates before retrying client access"}
	case warnings > 0:
		result = rule.NewWarn("SEC-003", "tls_certificate_health", "security", "some SSL listener certificates are approaching expiry")
		result.Evidence = evidence
		result.NextActions = []string{"schedule certificate renewal before the warning window closes", "confirm every advertised hostname or IP is covered by SAN", "verify clients use the same hostname that the certificate presents"}
	}
	return result
}

type tlsInspectResult struct {
	Subject  string
	NotAfter time.Time
}

func inspectTLS(serverName string, address string, timeout time.Duration) (tlsInspectResult, error) {
	dialer := &net.Dialer{Timeout: timeout}
	cfg := &tls.Config{
		ServerName:         serverName,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false,
	}
	conn, err := tls.DialWithDialer(dialer, "tcp", address, cfg)
	if err != nil {
		return tlsInspectResult{}, err
	}
	defer conn.Close()

	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return tlsInspectResult{}, fmt.Errorf("no peer certificates returned")
	}
	leaf := state.PeerCertificates[0]
	if err := leaf.VerifyHostname(serverName); err != nil {
		return tlsInspectResult{}, err
	}
	if _, err := leaf.Verify(x509.VerifyOptions{DNSName: serverName}); err != nil {
		// The TLS handshake already validated the chain. Keep SAN/host validation strict,
		// but tolerate environments where the full local trust roots are not exported here.
		if !strings.Contains(strings.ToLower(err.Error()), "unknown authority") {
			return tlsInspectResult{}, err
		}
	}
	return tlsInspectResult{
		Subject:  leaf.Subject.String(),
		NotAfter: leaf.NotAfter,
	}, nil
}

func securityListeners(service composeutil.KafkaService, executionView string) (map[string]composeutil.ListenerEndpoint, string) {
	raw := strings.TrimSpace(service.Environment["KAFKA_CFG_ADVERTISED_LISTENERS"])
	if raw == "" {
		raw = strings.TrimSpace(service.Environment["KAFKA_CFG_LISTENERS"])
	}
	listeners, err := composeutil.ParseListeners(raw)
	if err != nil {
		return nil, raw
	}
	return listeners, raw
}
