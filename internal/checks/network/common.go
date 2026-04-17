package network

import (
	"net"
	"net/netip"

	"kdoctor/internal/snapshot"
)

func isExternalProbeView(snap *snapshot.Bundle) bool {
	if snap == nil || snap.Network == nil || len(snap.Network.BootstrapChecks) == 0 {
		return false
	}
	for _, check := range snap.Network.BootstrapChecks {
		if !isPrivateEndpoint(check.Address) {
			return true
		}
	}
	return false
}

func isPrivateEndpoint(address string) bool {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		host = address
	}
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}
	return addr.IsPrivate() || addr.IsLoopback() || addr.IsLinkLocalUnicast()
}
