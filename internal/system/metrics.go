package system

import (
	"runtime"
	"time"
)

// Metrics represents a real-time snapshot of system resources
type Metrics struct {
	Timestamp   time.Time   `json:"timestamp"`
	CPU         CPUMetrics  `json:"cpu"`
	Memory      MemMetrics  `json:"memory"`
	Swap        SwapMetrics `json:"swap"`
	Disk        DiskMetrics `json:"disk"`
	LoadAverage [3]float64  `json:"load_average"`
	Processes   int         `json:"processes"`
}

type CPUMetrics struct {
	UsagePercent float64 `json:"usage_percent"`
	Cores        int     `json:"cores"`
}

type MemMetrics struct {
	TotalBytes     uint64  `json:"total_bytes"`
	UsedBytes      uint64  `json:"used_bytes"`
	FreeBytes      uint64  `json:"free_bytes"`
	AvailableBytes uint64  `json:"available_bytes"`
	UsagePercent   float64 `json:"usage_percent"`
}

type SwapMetrics struct {
	TotalBytes   uint64  `json:"total_bytes"`
	UsedBytes    uint64  `json:"used_bytes"`
	FreeBytes    uint64  `json:"free_bytes"`
	UsagePercent float64 `json:"usage_percent"`
}

type DiskMetrics struct {
	TotalBytes   uint64  `json:"total_bytes"`
	UsedBytes    uint64  `json:"used_bytes"`
	FreeBytes    uint64  `json:"free_bytes"`
	UsagePercent float64 `json:"usage_percent"`
	Path         string  `json:"path"`
}

// GetLiveMetrics samples current CPU, RAM, Disk, and Load in real-time
func GetLiveMetrics() (*Metrics, error) {
	m := &Metrics{
		Timestamp: time.Now().UTC(),
		CPU: CPUMetrics{
			Cores: runtime.NumCPU(),
		},
		Disk: DiskMetrics{
			Path: "/",
		},
	}

	collectOSMetrics(m)
	return m, nil
}
