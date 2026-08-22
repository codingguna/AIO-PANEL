//go:build linux

package system

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

var (
	linuxCPUMu   sync.Mutex
	linuxLastTot uint64
	linuxLastIdl uint64
)

func collectOSMetrics(m *Metrics) {
	// 1. Real CPU usage from /proc/stat
	if f, err := os.Open("/proc/stat"); err == nil {
		scanner := bufio.NewScanner(f)
		if scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) >= 5 && fields[0] == "cpu" {
				var user, nice, sys, idle, iowait, irq, softirq, steal uint64
				user, _ = strconv.ParseUint(fields[1], 10, 64)
				nice, _ = strconv.ParseUint(fields[2], 10, 64)
				sys, _ = strconv.ParseUint(fields[3], 10, 64)
				idle, _ = strconv.ParseUint(fields[4], 10, 64)
				if len(fields) > 5 {
					iowait, _ = strconv.ParseUint(fields[5], 10, 64)
				}
				if len(fields) > 6 {
					irq, _ = strconv.ParseUint(fields[6], 10, 64)
				}
				if len(fields) > 7 {
					softirq, _ = strconv.ParseUint(fields[7], 10, 64)
				}
				if len(fields) > 8 {
					steal, _ = strconv.ParseUint(fields[8], 10, 64)
				}

				total := user + nice + sys + idle + iowait + irq + softirq + steal
				idleTotal := idle + iowait

				linuxCPUMu.Lock()
				if linuxLastTot > 0 && total > linuxLastTot {
					diffTotal := float64(total - linuxLastTot)
					diffIdle := float64(idleTotal - linuxLastIdl)
					usage := ((diffTotal - diffIdle) / diffTotal) * 100.0
					if usage >= 0 && usage <= 100 {
						m.CPU.UsagePercent = usage
					}
				}
				linuxLastTot = total
				linuxLastIdl = idleTotal
				linuxCPUMu.Unlock()
			}
		}
		f.Close()
	}

	// 2. Real Memory & Swap from /proc/meminfo
	if f, err := os.Open("/proc/meminfo"); err == nil {
		scanner := bufio.NewScanner(f)
		var memTotal, memFree, memAvailable, buffers, cached uint64
		var swapTotal, swapFree uint64

		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			val, _ := strconv.ParseUint(fields[1], 10, 64)
			valBytes := val * 1024 // /proc/meminfo is in kB

			switch fields[0] {
			case "MemTotal:":
				memTotal = valBytes
			case "MemFree:":
				memFree = valBytes
			case "MemAvailable:":
				memAvailable = valBytes
			case "Buffers:":
				buffers = valBytes
			case "Cached:":
				cached = valBytes
			case "SwapTotal:":
				swapTotal = valBytes
			case "SwapFree:":
				swapFree = valBytes
			}
		}
		f.Close()

		m.Memory.TotalBytes = memTotal
		if memAvailable > 0 {
			m.Memory.AvailableBytes = memAvailable
			m.Memory.UsedBytes = memTotal - memAvailable
		} else {
			m.Memory.UsedBytes = memTotal - (memFree + buffers + cached)
			m.Memory.AvailableBytes = memFree + buffers + cached
		}
		m.Memory.FreeBytes = memFree
		if memTotal > 0 {
			m.Memory.UsagePercent = (float64(m.Memory.UsedBytes) / float64(memTotal)) * 100.0
		}

		m.Swap.TotalBytes = swapTotal
		m.Swap.FreeBytes = swapFree
		m.Swap.UsedBytes = swapTotal - swapFree
		if swapTotal > 0 {
			m.Swap.UsagePercent = (float64(m.Swap.UsedBytes) / float64(swapTotal)) * 100.0
		}
	}

	// 3. Real Load averages & Processes from /proc/loadavg
	if data, err := os.ReadFile("/proc/loadavg"); err == nil {
		fields := strings.Fields(string(data))
		if len(fields) >= 3 {
			m.LoadAverage[0], _ = strconv.ParseFloat(fields[0], 64)
			m.LoadAverage[1], _ = strconv.ParseFloat(fields[1], 64)
			m.LoadAverage[2], _ = strconv.ParseFloat(fields[2], 64)
		}
		if len(fields) >= 4 {
			procParts := strings.Split(fields[3], "/")
			if len(procParts) == 2 {
				m.Processes, _ = strconv.Atoi(procParts[1])
			}
		}
	}

	// 4. Real Disk Usage via statfs
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err == nil {
		total := stat.Blocks * uint64(stat.Bsize)
		free := stat.Bavail * uint64(stat.Bsize)
		used := total - free
		m.Disk.TotalBytes = total
		m.Disk.FreeBytes = free
		m.Disk.UsedBytes = used
		if total > 0 {
			m.Disk.UsagePercent = (float64(used) / float64(total)) * 100.0
		}
	}
}
