//go:build !windows

package disk

import (
	"fmt"
	"path/filepath"
	"syscall"
)

func Stat(path string) (Usage, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return Usage{}, fmt.Errorf("resolve path: %w", err)
	}

	var stat syscall.Statfs_t
	if err := syscall.Statfs(absolute, &stat); err != nil {
		return Usage{}, fmt.Errorf("statfs: %w", err)
	}

	total := int64(stat.Blocks) * int64(stat.Bsize)
	available := int64(stat.Bavail) * int64(stat.Bsize)
	used := total - available
	percent := 0.0
	if total > 0 {
		percent = float64(used) * 100 / float64(total)
	}

	return Usage{
		Path:           absolute,
		TotalBytes:     total,
		AvailableBytes: available,
		UsedBytes:      used,
		UsedPercent:    percent,
	}, nil
}
