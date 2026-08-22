package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	installPort int
	purgeData   bool
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install AIO-PANEL as a systemd service with autostart on boot",
	Long: `Installs the AIO-PANEL single binary to /usr/local/bin/aio, sets up
data directories in /var/lib/aio, installs /etc/systemd/system/aio-panel.service,
and enables automatic startup on server boot.`,
	Run: func(cmd *cobra.Command, args []string) {
		if runtime.GOOS != "linux" {
			fmt.Println("❌ Installation as a systemd service is only supported on Linux.")
			return
		}

		if os.Geteuid() != 0 {
			fmt.Println("❌ Error: You must run 'sudo aio install' with root privileges.")
			return
		}

		fmt.Println("🚀 Installing AIO-PANEL Single-Binary Server...")
		fmt.Println("──────────────────────────────────────────────────")

		// 1. Locate current executable
		execPath, err := os.Executable()
		if err != nil {
			fmt.Printf("❌ Failed to resolve binary path: %v\n", err)
			return
		}

		targetBin := "/usr/local/bin/aio"
		if execPath != targetBin {
			fmt.Printf("📦 Installing binary to %s...\n", targetBin)
			src, err := os.Open(execPath)
			if err != nil {
				fmt.Printf("❌ Failed to read binary: %v\n", err)
				return
			}
			defer src.Close()

			dst, err := os.OpenFile(targetBin, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				fmt.Printf("❌ Failed to copy to %s: %v\n", targetBin, err)
				return
			}
			if _, err := io.Copy(dst, src); err != nil {
				dst.Close()
				fmt.Printf("❌ Failed to write binary: %v\n", err)
				return
			}
			dst.Close()
			_ = os.Chmod(targetBin, 0755)
		}

		// 2. Create required directories
		dirs := []string{
			"/etc/aio",
			"/var/lib/aio",
			"/var/lib/aio/data",
			"/var/lib/aio/backups",
			"/var/log/aio",
		}
		for _, d := range dirs {
			if err := os.MkdirAll(d, 0750); err != nil {
				fmt.Printf("⚠️ Warning creating %s: %v\n", d, err)
			}
		}

		// 3. Write default configuration if not present
		configPath := "/etc/aio/aio.conf"
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			defaultConf := fmt.Sprintf(`{
  "server": {
    "host": "0.0.0.0",
    "port": %d,
    "tls": false
  },
  "database": {
    "path": "/var/lib/aio/data/aio.db"
  },
  "logging": {
    "level": "info",
    "format": "pretty",
    "file": "/var/log/aio/aio.log"
  },
  "paths": {
    "config_dir": "/etc/aio",
    "data_dir": "/var/lib/aio",
    "log_dir": "/var/log/aio",
    "backup_dir": "/var/lib/aio/backups"
  }
}
`, installPort)
			_ = os.WriteFile(configPath, []byte(defaultConf), 0600)
			fmt.Println("✅ Provisioned /etc/aio/aio.conf")
		}

		// 4. Write /etc/systemd/system/aio-panel.service
		unitContent := fmt.Sprintf(`[Unit]
Description=AIO-PANEL Server Management Daemon
After=network.target network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
WorkingDirectory=/var/lib/aio
ExecStart=%s server --config /etc/aio/aio.conf
Restart=always
RestartSec=5
LimitNOFILE=65535
TimeoutStopSec=20
AmbientCapabilities=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
`, targetBin)

		servicePath := "/etc/systemd/system/aio-panel.service"
		if err := os.WriteFile(servicePath, []byte(unitContent), 0644); err != nil {
			fmt.Printf("❌ Failed to write systemd unit file: %v\n", err)
			return
		}
		fmt.Println("✅ Installed /etc/systemd/system/aio-panel.service")

		// 5. Reload systemd, enable & start service
		_ = exec.Command("systemctl", "daemon-reload").Run()
		if err := exec.Command("systemctl", "enable", "aio-panel.service").Run(); err != nil {
			fmt.Printf("⚠️ Warning enabling service: %v\n", err)
		} else {
			fmt.Println("✅ Enabled autostart on boot (aio-panel.service)")
		}

		if err := exec.Command("systemctl", "restart", "aio-panel.service").Run(); err != nil {
			fmt.Printf("❌ Failed to start service: %v\n", err)
			return
		}

		// 6. Verify service status
		fmt.Println("🔍 Verifying service status and health...")
		if err := exec.Command("systemctl", "is-active", "--quiet", "aio-panel.service").Run(); err != nil {
			fmt.Println("⚠️ Warning: aio-panel.service is starting up or encountered an issue.")
			fmt.Println("Inspect logs with: journalctl -u aio-panel.service -n 30 --no-pager")
		} else {
			fmt.Println("✅ Service is active and running")
		}

		fmt.Println("──────────────────────────────────────────────────")
		fmt.Println("🎉 AIO-PANEL is now installed & running!")
		fmt.Println("──────────────────────────────────────────────────")
		fmt.Println("👑 Step 1: Create your superuser login:")
		fmt.Println("   sudo aio createsuperuser")
		fmt.Println("")
		fmt.Printf("🌐 Step 2: Access your web panel at:\n")
		fmt.Printf("   http://YOUR_SERVER_IP:%d\n", installPort)
		fmt.Println("──────────────────────────────────────────────────")
	},
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Stop and remove AIO-PANEL systemd service",
	Run: func(cmd *cobra.Command, args []string) {
		if runtime.GOOS != "linux" {
			fmt.Println("❌ Uninstallation is only supported on Linux.")
			return
		}

		if os.Geteuid() != 0 {
			fmt.Println("❌ Error: You must run 'sudo aio uninstall' with root privileges.")
			return
		}

		fmt.Println("🛑 Stopping and removing AIO-PANEL systemd service...")
		_ = exec.Command("systemctl", "stop", "aio-panel.service").Run()
		_ = exec.Command("systemctl", "disable", "aio-panel.service").Run()
		_ = os.Remove("/etc/systemd/system/aio-panel.service")
		_ = exec.Command("systemctl", "daemon-reload").Run()
		_ = os.Remove("/usr/local/bin/aio")

		if purgeData {
			_ = os.RemoveAll("/etc/aio")
			_ = os.RemoveAll("/var/lib/aio")
			_ = os.RemoveAll("/var/log/aio")
			fmt.Println("🧹 Removed configuration, logs, and database directories.")
		}

		fmt.Println("✅ AIO-PANEL has been cleanly uninstalled.")
		fmt.Println("🛡️ All existing applications, web servers, and databases remain untouched.")
	},
}

func init() {
	installCmd.Flags().IntVarP(&installPort, "port", "p", 5555, "Port for AIO-PANEL to listen on")
	uninstallCmd.Flags().BoolVar(&purgeData, "purge", false, "Remove all database files, logs, and configuration directories")
	RootCmd.AddCommand(installCmd)
	RootCmd.AddCommand(uninstallCmd)
}
