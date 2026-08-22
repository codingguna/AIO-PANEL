package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/codingguna/aio-panel/internal/ops"
	"github.com/spf13/cobra"
)

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Inspect Docker containers and state",
	RunE: func(cmd *cobra.Command, args []string) error {
		containers, err := ops.ListDockerContainers(context.Background())
		if err != nil {
			return err
		}

		if len(containers) == 0 {
			fmt.Println("No Docker containers found (or Docker not installed).")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "CONTAINER ID\tNAME\tIMAGE\tSTATE\tSTATUS\tPORTS")
		fmt.Fprintln(w, "------------\t----\t-----\t-----\t------\t-----")

		for _, c := range containers {
			icon := "🟢"
			if c.State != "running" {
				icon = "🔴"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s %s\t%s\t%s\n",
				c.ID, c.Names, c.Image, icon, c.State, c.Status, c.Ports)
		}
		w.Flush()
		fmt.Println("")
		return nil
	},
}

var cronCmd = &cobra.Command{
	Use:   "cron",
	Short: "List scheduled cron jobs",
	RunE: func(cmd *cobra.Command, args []string) error {
		jobs, err := ops.ListCronJobs(context.Background())
		if err != nil {
			return err
		}

		if len(jobs) == 0 {
			fmt.Println("No cron jobs found for current user.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "JOB #\tSCHEDULE\tCOMMAND\tUSER")
		fmt.Fprintln(w, "-----\t--------\t-------\t----")

		for _, j := range jobs {
			fmt.Fprintf(w, "[%d]\t%s\t%s\t%s\n", j.ID, j.Schedule, j.Command, j.User)
		}
		w.Flush()
		fmt.Println("")
		return nil
	},
}

func init() {
	RootCmd.AddCommand(dockerCmd)
	RootCmd.AddCommand(cronCmd)
}
