//go:build windows

package host

import "context"

func collectSystemSignals(_ context.Context) systemSignals {
	return systemSignals{}
}
