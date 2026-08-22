package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/codingguna/aio-panel/internal/config"
	"github.com/codingguna/aio-panel/internal/db"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Manage AIO-PANEL administrative users and passwords",
}

var createsuperuserCmd = &cobra.Command{
	Use:     "createsuperuser",
	Aliases: []string{"create-admin"},
	Short:   "Create an administrative superuser account for the Web panel",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCreateSuperUser()
	},
}

var adminCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new panel administrator",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCreateSuperUser()
	},
}

var adminListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all panel administrators",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		store, err := db.Open(cfg.Database.Path)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer store.Close()

		users, err := store.ListPanelUsers(context.Background())
		if err != nil {
			return fmt.Errorf("failed to list users: %w", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tUSERNAME\tROLE\tCREATED AT\tLAST LOGIN\tACTIVE")
		fmt.Fprintln(w, "--\t--------\t----\t----------\t----------\t------")

		for _, u := range users {
			lastLogin := "Never"
			if u.LastLoginAt != nil {
				lastLogin = u.LastLoginAt.Format("2006-01-02 15:04:05")
			}
			activeStr := "Yes"
			if !u.IsActive {
				activeStr = "No"
			}

			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
				u.ID,
				u.Username,
				strings.ToUpper(u.Role),
				u.CreatedAt.Format("2006-01-02 15:04"),
				lastLogin,
				activeStr,
			)
		}
		w.Flush()
		return nil
	},
}

var resetUser string

var adminResetPassCmd = &cobra.Command{
	Use:   "reset-password",
	Short: "Reset password for a panel administrator",
	RunE: func(cmd *cobra.Command, args []string) error {
		if resetUser == "" {
			return fmt.Errorf("--user flag is required")
		}

		cfg, err := config.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		store, err := db.Open(cfg.Database.Path)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer store.Close()

		fmt.Printf("Resetting password for '%s'...\n", resetUser)
		fmt.Print("Enter new password: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println("")
		if err != nil {
			return err
		}

		newPass := strings.TrimSpace(string(bytePassword))
		if len(newPass) < 6 {
			return fmt.Errorf("password must be at least 6 characters long")
		}

		if err := store.UpdatePanelUserPassword(context.Background(), resetUser, newPass); err != nil {
			return err
		}

		fmt.Printf("✅ Password for '%s' updated successfully.\n", resetUser)
		return nil
	},
}

func runCreateSuperUser() error {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	store, err := db.Open(cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer store.Close()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("👑 Create AIO-PANEL Superuser Account")
	fmt.Println("──────────────────────────────────────────────────")

	fmt.Print("Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	fmt.Print("Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println("")
	if err != nil {
		return err
	}

	password := strings.TrimSpace(string(bytePassword))
	if len(password) < 6 {
		return fmt.Errorf("password must be at least 6 characters long")
	}

	fmt.Print("Confirm Password: ")
	bytePassword2, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println("")
	if err != nil {
		return err
	}

	if password != strings.TrimSpace(string(bytePassword2)) {
		return fmt.Errorf("passwords do not match")
	}

	user, err := store.CreatePanelUser(context.Background(), username, password, "admin")
	if err != nil {
		return err
	}

	fmt.Println("──────────────────────────────────────────────────")
	fmt.Printf("✅ Superuser '%s' created successfully!\n", user.Username)
	fmt.Println("🌐 You can now log in at: http://YOUR_SERVER_IP:5555")
	return nil
}

func init() {
	adminCmd.AddCommand(adminCreateCmd)
	adminCmd.AddCommand(adminListCmd)
	adminCmd.AddCommand(adminResetPassCmd)

	adminResetPassCmd.Flags().StringVarP(&resetUser, "user", "u", "", "Username to reset password for")

	RootCmd.AddCommand(createsuperuserCmd)
	RootCmd.AddCommand(adminCmd)
}
