package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/codingguna/aio-panel/internal/system"
	"github.com/spf13/cobra"
)

var servicesCmd = &cobra.Command{
	Use:   "services",
	Short: "Manage systemd services and view daemon status",
	Long:  "Discover, inspect, start, stop, restart, and view logs for systemd units.",
}

var servicesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all discovered system services and their statuses",
	RunE: func(cmd *cobra.Command, args []string) error {
		services, err := system.ListServices(context.Background())
		if err != nil {
			return fmt.Errorf("failed to list services: %w", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "NAME\tDISPLAY NAME\tSTATUS\tSUBSTATE\tOWNERSHIP\tPID\tMEMORY")
		fmt.Fprintln(w, "----\t------------\t------\t--------\t---------\t---\t------")

		for _, s := range services {
			statusIcon := "🟢"
			if s.ActiveState != "active" {
				statusIcon = "🔴"
			}
			memStr := "-"
			if s.MemoryBytes > 0 {
				memStr = formatBytes(s.MemoryBytes)
			}
			pidStr := "-"
			if s.PID > 0 {
				pidStr = fmt.Sprintf("%d", s.PID)
			}

			fmt.Fprintf(w, "%s\t%s\t%s %s\t%s\t%s\t%s\t%s\n",
				s.Name,
				s.DisplayName,
				statusIcon,
				s.ActiveState,
				s.SubState,
				strings.ToUpper(s.OwnerType),
				pidStr,
				memStr,
			)
		}
		w.Flush()
		return nil
	},
}

var servicesStartCmd = &cobra.Command{
	Use:   "start [service]",
	Short: "Start a system service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		fmt.Printf("Starting service %s...\n", name)
		if err := system.ControlService(context.Background(), name, "start"); err != nil {
			return err
		}
		fmt.Printf("✅ Service %s started successfully.\n", name)
		return nil
	},
}

var servicesStopCmd = &cobra.Command{
	Use:   "stop [service]",
	Short: "Stop a system service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		fmt.Printf("Stopping service %s...\n", name)
		if err := system.ControlService(context.Background(), name, "stop"); err != nil {
			return err
		}
		fmt.Printf("✅ Service %s stopped successfully.\n", name)
		return nil
	},
}

var servicesRestartCmd = &cobra.Command{
	Use:   "restart [service]",
	Short: "Restart a system service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		fmt.Printf("Restarting service %s...\n", name)
		if err := system.ControlService(context.Background(), name, "restart"); err != nil {
			return err
		}
		fmt.Printf("✅ Service %s restarted successfully.\n", name)
		return nil
	},
}

var servicesLogsCmd = &cobra.Command{
	Use:   "logs [service]",
	Short: "View recent journal logs for a service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		logs, err := system.GetServiceLogs(context.Background(), name, 50)
		if err != nil {
			return err
		}
		fmt.Println(logs)
		return nil
	},
}

var (
	newSvcName    string
	newSvcExec    string
	newSvcDir     string
	newSvcUser    string
	newSvcDesc    string
	newSvcRestart string
	newSvcEnable  bool
)

var servicesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create and register a new systemd service",
	RunE: func(cmd *cobra.Command, args []string) error {
		if newSvcName == "" || newSvcExec == "" {
			return fmt.Errorf("--name and --exec are required")
		}
		fmt.Printf("Creating systemd service: %s.service...\n", newSvcName)
		err := system.CreateService(context.Background(), newSvcName, newSvcDesc, newSvcExec, newSvcDir, newSvcUser, newSvcRestart, newSvcEnable)
		if err != nil {
			return err
		}
		fmt.Printf("✅ Service %s.service successfully created and registered.\n", newSvcName)
		return nil
	},
}

func init() {
	servicesCmd.AddCommand(servicesListCmd)
	servicesCmd.AddCommand(servicesStartCmd)
	servicesCmd.AddCommand(servicesStopCmd)
	servicesCmd.AddCommand(servicesRestartCmd)
	servicesCmd.AddCommand(servicesLogsCmd)
	servicesCmd.AddCommand(servicesCreateCmd)

	servicesCreateCmd.Flags().StringVarP(&newSvcName, "name", "n", "", "Service name (e.g. my-app)")
	servicesCreateCmd.Flags().StringVarP(&newSvcExec, "exec", "e", "", "Command to execute (e.g. /usr/bin/node /var/www/app/index.js)")
	servicesCreateCmd.Flags().StringVarP(&newSvcDir, "dir", "d", "", "Working directory")
	servicesCreateCmd.Flags().StringVarP(&newSvcUser, "user", "u", "root", "User to run the service as")
	servicesCreateCmd.Flags().StringVar(&newSvcDesc, "desc", "", "Description for the service")
	servicesCreateCmd.Flags().StringVarP(&newSvcRestart, "restart", "r", "always", "Restart policy (always, on-failure, no)")
	servicesCreateCmd.Flags().BoolVar(&newSvcEnable, "enable", true, "Enable autostart on boot and start immediately")
}
