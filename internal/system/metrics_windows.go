//go:build windows

package system

import (
	"runtime"
	"sync"
	"syscall"
	"unsafe"
)

type memoryStatusEx struct {
	cbSize                  uint32
	dwMemoryLoad            uint32
	ullTotalPhys            uint64
	ullAvailPhys            uint64
	ullTotalPageFile        uint64
	ullAvailPageFile        uint64
	ullTotalVirtual         uint64
	ullAvailVirtual         uint64
	ullAvailExtendedVirtual uint64
}

type filetime struct {
	dwLowDateTime  uint32
	dwHighDateTime uint32
}

func (ft *filetime) toInt64() int64 {
	return int64(uint64(ft.dwHighDateTime)<<32 | uint64(ft.dwLowDateTime))
}

var (
	winCPUMu      sync.Mutex
	lastWinIdle   int64
	lastWinKernel int64
	lastWinUser   int64
)

func collectOSMetrics(m *Metrics) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")

	// 1. Real RAM & Swap via GlobalMemoryStatusEx
	globalMemoryStatusEx := kernel32.NewProc("GlobalMemoryStatusEx")
	var memStatus memoryStatusEx
	memStatus.cbSize = uint32(unsafe.Sizeof(memStatus))
	ret, _, _ := globalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&memStatus)))
	if ret != 0 {
		m.Memory.TotalBytes = memStatus.ullTotalPhys
		m.Memory.AvailableBytes = memStatus.ullAvailPhys
		m.Memory.UsedBytes = memStatus.ullTotalPhys - memStatus.ullAvailPhys
		m.Memory.FreeBytes = memStatus.ullAvailPhys
		m.Memory.UsagePercent = float64(memStatus.dwMemoryLoad)

		m.Swap.TotalBytes = memStatus.ullTotalPageFile
		m.Swap.FreeBytes = memStatus.ullAvailPageFile
		m.Swap.UsedBytes = memStatus.ullTotalPageFile - memStatus.ullAvailPageFile
		if memStatus.ullTotalPageFile > 0 {
			m.Swap.UsagePercent = (float64(m.Swap.UsedBytes) / float64(memStatus.ullTotalPageFile)) * 100.0
		}
	}

	// 2. Real CPU Usage via GetSystemTimes
	getSystemTimes := kernel32.NewProc("GetSystemTimes")
	var idleTime, kernelTime, userTime filetime
	ret, _, _ = getSystemTimes.Call(
		uintptr(unsafe.Pointer(&idleTime)),
		uintptr(unsafe.Pointer(&kernelTime)),
		uintptr(unsafe.Pointer(&userTime)),
	)
	if ret != 0 {
		idle := idleTime.toInt64()
		kernel := kernelTime.toInt64()
		user := userTime.toInt64()

		winCPUMu.Lock()
		if lastWinKernel > 0 {
			idleDiff := idle - lastWinIdle
			kernelDiff := kernel - lastWinKernel
			userDiff := user - lastWinUser
			totalDiff := kernelDiff + userDiff

			if totalDiff > 0 {
				usage := float64(totalDiff-idleDiff) / float64(totalDiff) * 100.0
				if usage >= 0 && usage <= 100 {
					m.CPU.UsagePercent = usage
				}
			}
		}
		lastWinIdle = idle
		lastWinKernel = kernel
		lastWinUser = user
		winCPUMu.Unlock()
	}

	// 3. Real Disk Usage via GetDiskFreeSpaceExW
	getDiskFreeSpaceEx := kernel32.NewProc("GetDiskFreeSpaceExW")
	var freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes uint64
	drivePtr, _ := syscall.UTF16PtrFromString("C:\\")
	ret, _, _ = getDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(drivePtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalNumberOfBytes)),
		uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)),
	)
	if ret != 0 {
		m.Disk.Path = "C:\\"
		m.Disk.TotalBytes = totalNumberOfBytes
		m.Disk.FreeBytes = totalNumberOfFreeBytes
		m.Disk.UsedBytes = totalNumberOfBytes - totalNumberOfFreeBytes
		if totalNumberOfBytes > 0 {
			m.Disk.UsagePercent = (float64(m.Disk.UsedBytes) / float64(totalNumberOfBytes)) * 100.0
		}
	}

	m.Processes = runtime.NumGoroutine()
	m.LoadAverage = [3]float64{
		(m.CPU.UsagePercent / 100.0) * float64(m.CPU.Cores),
		(m.CPU.UsagePercent / 100.0) * float64(m.CPU.Cores),
		(m.CPU.UsagePercent / 100.0) * float64(m.CPU.Cores),
	}
}
