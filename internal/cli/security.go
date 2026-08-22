package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/codingguna/aio-panel/internal/security"
	"github.com/spf13/cobra"
)

var securityCmd = &cobra.Command{
	Use:   "security",
	Short: "Inspect and manage OpenSSH, UFW firewall, and Linux users",
}

var securitySSHCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Show OpenSSH configuration and active sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		sshCfg, err := security.GetSSHConfig(context.Background())
		if err != nil {
			return err
		}

		sessions, _ := security.GetActiveSSHSessions(context.Background())

		fmt.Println("==================================================")
		fmt.Println("             OPENSSH SECURITY STATUS              ")
		fmt.Println("==================================================")
		fmt.Printf("SSH Port              : %d\n", sshCfg.Port)
		fmt.Printf("Permit Root Login     : %s\n", sshCfg.PermitRootLogin)
		fmt.Printf("Password Auth         : %v\n", sshCfg.PasswordAuthentication)
		fmt.Printf("Pubkey Auth           : %v\n", sshCfg.PubkeyAuthentication)
		fmt.Printf("Active SSH Sessions   : %d\n", len(sessions))
		for _, s := range sessions {
			fmt.Printf("  • User: %s, Terminal: %s, Host: %s, Login: %s\n", s.User, s.Terminal, s.Host, s.LoginTime)
		}
		fmt.Println("==================================================")
		return nil
	},
}

var securityFirewallCmd = &cobra.Command{
	Use:   "firewall",
	Short: "Show UFW firewall status and active rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		fw, err := security.GetFirewallStatus(context.Background())
		if err != nil {
			return err
		}

		statusStr := "🔴 INACTIVE"
		if fw.Active {
			statusStr = "🟢 ACTIVE"
		}

		fmt.Printf("\nUFW Firewall Status: %s (Default: %s incoming, %s outgoing)\n\n",
			statusStr, fw.DefaultIncoming, fw.DefaultOutgoing)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "RULE #\tPORT / PROTOCOL\tACTION\tFROM IP\tCOMMENT")
		fmt.Fprintln(w, "------\t---------------\t------\t-------\t-------")

		for _, r := range fw.Rules {
			fmt.Fprintf(w, "[%d]\t%s/%s\t%s\t%s\t%s\n",
				r.ID, r.ToPort, r.Protocol, r.Action, r.FromIP, r.Comment)
		}
		w.Flush()
		fmt.Println("")
		return nil
	},
}

var securityUsersCmd = &cobra.Command{
	Use:   "users",
	Short: "List host Linux users and sudo permissions",
	RunE: func(cmd *cobra.Command, args []string) error {
		users, err := security.ListLinuxUsers(context.Background())
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "USERNAME\tUID:GID\tSUDO\tSSH KEY\tHOME DIR\tSHELL")
		fmt.Fprintln(w, "--------\t-------\t----\t-------\t--------\t-----")

		for _, u := range users {
			if u.IsSystem {
				continue // skip system daemons in simple CLI list
			}
			sudoStr := "No"
			if u.IsSudo {
				sudoStr = "✅ Yes"
			}
			sshStr := "No"
			if u.HasSSHKey {
				sshStr = "🔑 Yes"
			}
			fmt.Fprintf(w, "%s\t%d:%d\t%s\t%s\t%s\t%s\n",
				u.Username, u.UID, u.GID, sudoStr, sshStr, u.HomeDir, u.Shell)
		}
		w.Flush()
		return nil
	},
}

func init() {
	securityCmd.AddCommand(securitySSHCmd)
	securityCmd.AddCommand(securityFirewallCmd)
	securityCmd.AddCommand(securityUsersCmd)
	RootCmd.AddCommand(securityCmd)
}
