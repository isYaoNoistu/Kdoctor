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
	totalInodes := int64(stat.Files)
	availableInodes := int64(stat.Ffree)
	usedInodes := totalInodes - availableInodes
	inodePercent := 0.0
	if totalInodes > 0 {
		inodePercent = float64(usedInodes) * 100 / float64(totalInodes)
	}

	return Usage{
		Path:            absolute,
		TotalBytes:      total,
		AvailableBytes:  available,
		UsedBytes:       used,
		UsedPercent:     percent,
		TotalInodes:     totalInodes,
		AvailableInodes: availableInodes,
		UsedInodes:      usedInodes,
		UsedInodePct:    inodePercent,
	}, nil
}
