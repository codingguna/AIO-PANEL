package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/codingguna/aio-panel/internal/config"
	"github.com/codingguna/aio-panel/internal/system"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current AIO-PANEL and server status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return err
		}

		info, err := system.GetInfo()
		if err != nil {
			return fmt.Errorf("failed to retrieve system info: %w", err)
		}

		metrics, err := system.GetLiveMetrics()
		if err != nil {
			return fmt.Errorf("failed to sample metrics: %w", err)
		}

		// Check if daemon is active
		daemonURL := fmt.Sprintf("http://127.0.0.1:%d/health", cfg.Server.Port)
		client := http.Client{Timeout: 1 * time.Second}
		resp, err := client.Get(daemonURL)

		daemonStatus := "🔴 Stopped"
		if err == nil && resp.StatusCode == http.StatusOK {
			daemonStatus = "🟢 Running"
			resp.Body.Close()
		}

		fmt.Println("==================================================")
		fmt.Println("             AIO-PANEL SERVER STATUS              ")
		fmt.Println("==================================================")
		fmt.Printf("Daemon Status : %s (Port: %d)\n", daemonStatus, cfg.Server.Port)
		fmt.Printf("Hostname      : %s\n", info.Hostname)
		fmt.Printf("OS / Kernel   : %s (%s / %s)\n", info.OS, info.Kernel, info.Architecture)
		fmt.Printf("CPU Model     : %s (%d Cores)\n", info.CPUModel, info.CPUCores)
		fmt.Printf("CPU Usage     : %.1f%%\n", metrics.CPU.UsagePercent)
		fmt.Printf("Memory Usage  : %.1f%% (%s / %s)\n",
			metrics.Memory.UsagePercent,
			formatBytes(metrics.Memory.UsedBytes),
			formatBytes(metrics.Memory.TotalBytes))
		fmt.Printf("Disk Usage    : %.1f%% (%s / %s)\n",
			metrics.Disk.UsagePercent,
			formatBytes(metrics.Disk.UsedBytes),
			formatBytes(metrics.Disk.TotalBytes))
		fmt.Printf("Load Average  : %.2f, %.2f, %.2f\n",
			metrics.LoadAverage[0], metrics.LoadAverage[1], metrics.LoadAverage[2])
		fmt.Println("==================================================")

		return nil
	},
}

func formatBytes(bytes uint64) string {
	if bytes == 0 {
		return "0 B"
	}
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func prettyPrint(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))
}
