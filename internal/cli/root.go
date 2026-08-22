package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgPath string
	verbose bool
)

// RootCmd is the base command for the AIO CLI
var RootCmd = &cobra.Command{
	Use:   "aio",
	Short: "AIO-PANEL - All-In-One Linux Server Control Panel",
	Long: `AIO-PANEL is a lightweight, powerful, and security-first 
All-In-One Linux Server Management Panel.

Manage your entire Linux server from one lightweight panel.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&cfgPath, "config", "c", "", "Path to configuration file (default: /etc/aio/aio.conf)")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose/debug logging output")

	RootCmd.AddCommand(serverCmd)
	RootCmd.AddCommand(statusCmd)
	RootCmd.AddCommand(doctorCmd)
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(servicesCmd)
}
