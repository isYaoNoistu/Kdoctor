package tcp

import (
	"context"
	"net"
	"time"
)

type Result struct {
	Address   string
	Reachable bool
	Duration  time.Duration
	Error     string
}

func Dial(ctx context.Context, address string, timeout time.Duration) Result {
	startedAt := time.Now()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", address)
	result := Result{
		Address:  address,
		Duration: time.Since(startedAt),
	}
	if err != nil {
		result.Error = err.Error()
		return result
	}
	_ = conn.Close()
	result.Reachable = true
	return result
}
