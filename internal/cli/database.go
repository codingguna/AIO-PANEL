package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/codingguna/aio-panel/internal/config"
	"github.com/codingguna/aio-panel/internal/database"
	"github.com/spf13/cobra"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Manage PostgreSQL, MySQL, and Redis databases",
}

var dbPostgresCmd = &cobra.Command{
	Use:   "postgres",
	Short: "List PostgreSQL databases and sizes",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbs, err := database.ListPostgresDatabases(context.Background())
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "DATABASE NAME\tOWNER\tENCODING\tSIZE")
		fmt.Fprintln(w, "-------------\t-----\t--------\t----")

		for _, d := range dbs {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", d.Name, d.Owner, d.Encoding, d.SizeHuman)
		}
		w.Flush()
		fmt.Println("")
		return nil
	},
}

var dbMySQLCmd = &cobra.Command{
	Use:   "mysql",
	Short: "List MySQL/MariaDB databases",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbs, err := database.ListMySQLDatabases(context.Background())
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "DATABASE NAME\tSIZE")
		fmt.Fprintln(w, "-------------\t----")

		for _, d := range dbs {
			fmt.Fprintf(w, "%s\t%s\n", d.Name, d.SizeHuman)
		}
		w.Flush()
		fmt.Println("")
		return nil
	},
}

var dbBackupCmd = &cobra.Command{
	Use:   "backup [postgres|mysql] [dbname]",
	Short: "Create an instant SQL dump backup of a database",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		engine := args[0]
		dbName := args[1]
		cfg, _ := config.Load(cfgPath)

		var backupFile string
		var err error

		if engine == "postgres" || engine == "psql" {
			fmt.Printf("Dumping PostgreSQL database '%s'...\n", dbName)
			backupFile, err = database.BackupPostgresDatabase(context.Background(), dbName, cfg.Paths.BackupDir)
		} else if engine == "mysql" || engine == "mariadb" {
			fmt.Printf("Dumping MySQL database '%s'...\n", dbName)
			backupFile, err = database.BackupMySQLDatabase(context.Background(), dbName, cfg.Paths.BackupDir)
		} else {
			return fmt.Errorf("unsupported database engine: %s (use 'postgres' or 'mysql')", engine)
		}

		if err != nil {
			return err
		}

		fmt.Printf("✅ Backup completed successfully: %s\n", backupFile)
		return nil
	},
}

var (
	createEngine string
	createDbName string
	createOwner  string
)

var dbCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new PostgreSQL or MySQL database",
	RunE: func(cmd *cobra.Command, args []string) error {
		if createDbName == "" {
			return fmt.Errorf("--name is required")
		}
		if createEngine == "postgres" || createEngine == "psql" {
			fmt.Printf("Creating PostgreSQL database '%s'...\n", createDbName)
			if err := database.CreatePostgresDatabase(context.Background(), createDbName, createOwner); err != nil {
				return err
			}
		} else if createEngine == "mysql" || createEngine == "mariadb" {
			fmt.Printf("Creating MySQL database '%s'...\n", createDbName)
			if err := database.CreateMySQLDatabase(context.Background(), createDbName); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("unsupported database engine: %s (use 'postgres' or 'mysql')", createEngine)
		}

		fmt.Printf("✅ Database '%s' created successfully.\n", createDbName)
		return nil
	},
}

func init() {
	dbCmd.AddCommand(dbPostgresCmd)
	dbCmd.AddCommand(dbMySQLCmd)
	dbCmd.AddCommand(dbBackupCmd)
	dbCmd.AddCommand(dbCreateCmd)

	dbCreateCmd.Flags().StringVarP(&createEngine, "engine", "e", "postgres", "Database engine (postgres or mysql)")
	dbCreateCmd.Flags().StringVarP(&createDbName, "name", "n", "", "Database name")
	dbCreateCmd.Flags().StringVarP(&createOwner, "owner", "o", "", "Owner user/role (PostgreSQL only)")

	RootCmd.AddCommand(dbCmd)
}
