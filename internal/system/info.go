package system

import (
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/codingguna/aio-panel/internal/version"
)

// Info represents static and semi-static host machine information
type Info struct {
	Hostname     string `json:"hostname"`
	OS           string `json:"os"`
	Platform     string `json:"platform"`
	Kernel       string `json:"kernel"`
	Architecture string `json:"architecture"`
	CPUModel     string `json:"cpu_model"`
	CPUCores     int    `json:"cpu_cores"`
	TotalMemory  uint64 `json:"total_memory"`
	TotalDisk    uint64 `json:"total_disk"`
	Uptime       uint64 `json:"uptime_seconds"`
	BootTime     string `json:"boot_time"`
	GoVersion    string `json:"go_version"`
	PanelVersion string `json:"panel_version"`
}

// GetInfo collects current system overview
func GetInfo() (*Info, error) {
	hostname, _ := os.Hostname()
	cores := runtime.NumCPU()
	arch := runtime.GOARCH
	goos := runtime.GOOS

	info := &Info{
		Hostname:     hostname,
		OS:           goos,
		Platform:     runtime.GOOS,
		Kernel:       "unknown",
		Architecture: arch,
		CPUModel:     runtime.GOARCH + " Processor",
		CPUCores:     cores,
		GoVersion:    runtime.Version(),
		PanelVersion: version.Version,
	}

	// Platform-specific enhancements
	populatePlatformInfo(info)

	return info, nil
}

func populatePlatformInfo(info *Info) {
	if runtime.GOOS == "linux" {
		// Read /etc/os-release for pretty OS name
		if data, err := os.ReadFile("/etc/os-release"); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "PRETTY_NAME=") {
					info.OS = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
					break
				}
			}
		}

		// Read /proc/version for kernel
		if data, err := os.ReadFile("/proc/version"); err == nil {
			fields := strings.Fields(string(data))
			if len(fields) >= 3 {
				info.Kernel = fields[2]
			}
		}

		// Read /proc/cpuinfo for CPU model
		if data, err := os.ReadFile("/proc/cpuinfo"); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "model name") {
					parts := strings.Split(line, ":")
					if len(parts) == 2 {
						info.CPUModel = strings.TrimSpace(parts[1])
						break
					}
				}
			}
		}

		// Read uptime from /proc/uptime
		if data, err := os.ReadFile("/proc/uptime"); err == nil {
			fields := strings.Fields(string(data))
			if len(fields) > 0 {
				var up float64
				if _, err := parseNumber(fields[0], &up); err == nil {
					info.Uptime = uint64(up)
					boot := time.Now().Add(-time.Duration(up) * time.Second)
					info.BootTime = boot.Format(time.RFC3339)
				}
			}
		}
	} else if runtime.GOOS == "windows" {
		info.OS = "Microsoft Windows"
		info.Kernel = "NT"
	}
}

func parseNumber(s string, out *float64) (bool, error) {
	var f float64
	_, err := strings.NewReader(s).Read([]byte(s))
	if err != nil {
		return false, err
	}
	for _, c := range s {
		if c >= '0' && c <= '9' {
			f = f*10 + float64(c-'0')
		} else if c == '.' {
			break
		}
	}
	*out = f
	return true, nil
}
