//go:build windows

package disk

import (
	"fmt"
	"path/filepath"
	"syscall"
	"unsafe"
)

func Stat(path string) (Usage, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return Usage{}, fmt.Errorf("resolve path: %w", err)
	}

	target, err := syscall.UTF16PtrFromString(absolute)
	if err != nil {
		return Usage{}, fmt.Errorf("encode path: %w", err)
	}

	var freeBytesAvailable uint64
	var totalNumberOfBytes uint64
	var totalNumberOfFreeBytes uint64

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("GetDiskFreeSpaceExW")
	ret, _, callErr := proc.Call(
		uintptr(unsafe.Pointer(target)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalNumberOfBytes)),
		uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)),
	)
	if ret == 0 {
		if callErr != syscall.Errno(0) {
			return Usage{}, fmt.Errorf("get disk usage: %w", callErr)
		}
		return Usage{}, fmt.Errorf("get disk usage failed")
	}

	total := int64(totalNumberOfBytes)
	available := int64(totalNumberOfFreeBytes)
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
