//go:build !linux && !windows

package system

import "runtime"

func collectOSMetrics(m *Metrics) {
	var rtMem runtime.MemStats
	runtime.ReadMemStats(&rtMem)

	m.Memory.UsedBytes = rtMem.Alloc
	m.Memory.TotalBytes = rtMem.Sys
	m.Memory.FreeBytes = m.Memory.TotalBytes - m.Memory.UsedBytes
	m.Memory.AvailableBytes = m.Memory.FreeBytes
	if m.Memory.TotalBytes > 0 {
		m.Memory.UsagePercent = (float64(m.Memory.UsedBytes) / float64(m.Memory.TotalBytes)) * 100.0
	}
	m.Processes = runtime.NumGoroutine()
}
