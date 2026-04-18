//go:build !windows

package host

import (
	"context"
	"os"
	"sort"
	"strconv"
	"strings"

	"kdoctor/internal/snapshot"
	shelltransport "kdoctor/internal/transport/shell"
)

func collectSystemSignals(ctx context.Context) systemSignals {
	out := systemSignals{}

	if softLimit, err := readSoftLimit(ctx); err == nil && softLimit > 0 {
		out.FD = &snapshot.FDStats{SoftLimit: softLimit}
	} else if err != nil {
		out.Errors = append(out.Errors, err.Error())
	}

	if used, max, err := readFileNR(); err == nil {
		if out.FD == nil {
			out.FD = &snapshot.FDStats{}
		}
		out.FD.SystemUsed = used
		out.FD.SystemMax = max
	} else if err != nil {
		out.Errors = append(out.Errors, err.Error())
	}

	if memory, err := readMemoryStats(); err == nil && memory != nil {
		out.Memory = memory
	} else if err != nil {
		out.Errors = append(out.Errors, err.Error())
	}

	if ports, err := readListeningPorts(ctx); err == nil {
		out.ListenPorts = ports
	} else if err != nil {
		out.Errors = append(out.Errors, err.Error())
	}

	return out
}

func readSoftLimit(ctx context.Context) (uint64, error) {
	output, err := shelltransport.Run(ctx, "sh", "-c", "ulimit -n")
	if err != nil {
		return 0, err
	}
	value, err := strconv.ParseUint(strings.TrimSpace(output), 10, 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func readFileNR() (uint64, uint64, error) {
	data, err := os.ReadFile("/proc/sys/fs/file-nr")
	if err != nil {
		return 0, 0, err
	}
	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return 0, 0, nil
	}
	used, err := strconv.ParseUint(fields[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	max, err := strconv.ParseUint(fields[2], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return used, max, nil
}

func readMemoryStats() (*snapshot.MemoryStats, error) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return nil, err
	}

	values := map[string]uint64{}
	for _, line := range strings.Split(string(data), "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		fields := strings.Fields(strings.TrimSpace(parts[1]))
		if len(fields) == 0 {
			continue
		}
		value, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil {
			continue
		}
		values[strings.TrimSpace(parts[0])] = value * 1024
	}

	total := int64(values["MemTotal"])
	available := int64(values["MemAvailable"])
	if total <= 0 {
		return nil, nil
	}
	used := total - available
	usedPercent := 0.0
	if total > 0 {
		usedPercent = float64(used) * 100 / float64(total)
	}
	return &snapshot.MemoryStats{
		TotalBytes:     total,
		AvailableBytes: available,
		UsedBytes:      used,
		UsedPercent:    usedPercent,
	}, nil
}

func readListeningPorts(ctx context.Context) ([]int, error) {
	output, err := shelltransport.Run(ctx, "ss", "-ltnH")
	if err != nil {
		return nil, err
	}

	seen := map[int]struct{}{}
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 4 {
			continue
		}
		localAddress := fields[3]
		port := parsePort(localAddress)
		if port > 0 {
			seen[port] = struct{}{}
		}
	}

	ports := make([]int, 0, len(seen))
	for port := range seen {
		ports = append(ports, port)
	}
	sort.Ints(ports)
	return ports, nil
}

func parsePort(address string) int {
	address = strings.TrimSpace(address)
	if address == "" {
		return 0
	}
	if idx := strings.LastIndex(address, ":"); idx >= 0 && idx < len(address)-1 {
		value, err := strconv.Atoi(address[idx+1:])
		if err == nil {
			return value
		}
	}
	return 0
}
