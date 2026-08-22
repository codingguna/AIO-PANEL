package cli

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/codingguna/aio-panel/internal/config"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose panel configuration, permissions, and dependencies",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🏥 Running AIO-PANEL Doctor...")
		fmt.Println("--------------------------------------------------")

		cfg, err := config.Load(cfgPath)
		if err != nil {
			fmt.Printf("❌ Configuration Load: FAILED (%v)\n", err)
			return nil
		}
		fmt.Printf("✅ Configuration Load: OK (%s)\n", cfg.Paths.ConfigDir)

		// Check OS
		fmt.Printf("✅ OS Platform: %s (%s)\n", runtime.GOOS, runtime.GOARCH)

		// Check Port Availability
		addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			fmt.Printf("⚠️  Port %d Binding: IN USE or Permission Denied (%v)\n", cfg.Server.Port, err)
		} else {
			ln.Close()
			fmt.Printf("✅ Port %d: AVAILABLE\n", cfg.Server.Port)
		}

		// Check Data Directory Permissions
		if err := os.MkdirAll(cfg.Paths.DataDir, 0750); err != nil {
			fmt.Printf("❌ Data Directory (%s): FAILED (%v)\n", cfg.Paths.DataDir, err)
		} else {
			testFile := filepath.Join(cfg.Paths.DataDir, ".doctor_test")
			if err := os.WriteFile(testFile, []byte("ok"), 0600); err != nil {
				fmt.Printf("❌ Data Directory Write Test: FAILED (%v)\n", err)
			} else {
				os.Remove(testFile)
				fmt.Printf("✅ Data Directory (%s): WRITABLE\n", cfg.Paths.DataDir)
			}
		}

		// Check SQLite DB
		db, err := sql.Open("sqlite", cfg.Database.Path)
		if err != nil {
			fmt.Printf("❌ SQLite Open (%s): FAILED (%v)\n", cfg.Database.Path, err)
		} else {
			ctx, cancel := contextWithTimeout(2 * time.Second)
			defer cancel()
			if err := db.PingContext(ctx); err != nil {
				fmt.Printf("❌ SQLite Ping: FAILED (%v)\n", err)
			} else {
				fmt.Printf("✅ SQLite Database (%s): CONNECTED & HEALTHY\n", cfg.Database.Path)
			}
			db.Close()
		}

		// Check System Services Availability (on Linux)
		if runtime.GOOS == "linux" {
			checkLinuxCommand("systemctl", "systemd")
			checkLinuxCommand("nginx", "Nginx")
			checkLinuxCommand("psql", "PostgreSQL")
			checkLinuxCommand("docker", "Docker")
			checkLinuxCommand("ufw", "UFW Firewall")
		}

		fmt.Println("--------------------------------------------------")
		fmt.Println("✨ Doctor check completed.")
		return nil
	},
}

func checkLinuxCommand(cmdName, serviceName string) {
	if _, err := os.Stat("/usr/bin/" + cmdName); err == nil {
		fmt.Printf("✅ %s: DETECTED (/usr/bin/%s)\n", serviceName, cmdName)
	} else if _, err := os.Stat("/bin/" + cmdName); err == nil {
		fmt.Printf("✅ %s: DETECTED (/bin/%s)\n", serviceName, cmdName)
	} else {
		fmt.Printf("ℹ️  %s: NOT INSTALLED (Non-invasive mode active)\n", serviceName)
	}
}

func contextWithTimeout(d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d)
}
