package cli

import (
	"context"
	"fmt"

	"github.com/codingguna/aio-panel/internal/config"
	"github.com/codingguna/aio-panel/internal/discovery"
	"github.com/spf13/cobra"
)

var preflightCmd = &cobra.Command{
	Use:   "preflight",
	Short: "Run a read-only environment preflight inspection",
	Long:  "Discovers OS, tools, applications, virtual hosts, and port availability without modifying the host.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return err
		}

		fmt.Println("\n🔍 AIO-PANEL Preflight Inspection")
		fmt.Println("──────────────────────────────────────────────────")

		report, err := discovery.RunPreflight(context.Background(), cfg)
		if err != nil {
			return fmt.Errorf("preflight failed: %w", err)
		}

		fmt.Printf("OS              ✓ %s\n", report.OS)
		fmt.Printf("Architecture    ✓ %s\n", report.Architecture)
		if report.Systemd {
			fmt.Printf("systemd         ✓ detected\n")
		} else {
			fmt.Printf("systemd         ⚠️  not found\n")
		}

		fmt.Println("")
		for tool, installed := range report.Tools {
			ver := report.ToolVersions[tool]
			if installed {
				if ver != "" && ver != "installed" {
					fmt.Printf("%-15s ✓ %s\n", tool, ver)
				} else {
					fmt.Printf("%-15s ✓ installed\n", tool)
				}
			} else {
				fmt.Printf("%-15s ℹ️  not installed (optional)\n", tool)
			}
		}

		fmt.Println("")
		if len(report.ExistingSites) > 0 {
			fmt.Printf("Existing sites  ✓ %d detected (%v)\n", len(report.ExistingSites), report.ExistingSites)
		} else {
			fmt.Println("Existing sites  ℹ️  none detected")
		}

		if len(report.ExistingApps) > 0 {
			fmt.Printf("Existing apps   ✓ %d detected\n", len(report.ExistingApps))
			for _, app := range report.ExistingApps {
				fmt.Printf("  • %s (%s at %s) -> Service: %s\n", app.Name, app.Type, app.Path, app.Service)
			}
		}

		if len(report.ExistingServices) > 0 {
			fmt.Printf("Existing svcs   ✓ %d custom services detected\n", len(report.ExistingServices))
		}

		fmt.Println("")
		if report.PortAvailable {
			fmt.Printf("Port %d       ✓ available\n", report.Port)
		} else {
			fmt.Printf("Port %d       ❌ OCCUPIED by %s\n", report.Port, report.PortOccupiedBy)
			fmt.Println("  -> AIO-PANEL will NOT terminate this process.")
			fmt.Println("  -> Please specify another port using: aio server --port <PORT>")
		}

		fmt.Println("\n🛡️  Non-Invasive Guarantee:")
		fmt.Println("  AIO-PANEL will NOT modify or replace:")
		fmt.Println("  • Nginx & Virtual Hosts")
		fmt.Println("  • Python, Node.js, and npm runtimes")
		fmt.Println("  • PostgreSQL, MySQL, and Redis databases")
		fmt.Println("  • Existing applications & user files")
		fmt.Println("  • Existing systemd services")
		fmt.Println("──────────────────────────────────────────────────")

		return nil
	},
}

func init() {
	RootCmd.AddCommand(preflightCmd)
}
