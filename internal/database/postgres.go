package database

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// PostgresDB represents a PostgreSQL database
type PostgresDB struct {
	Name      string `json:"name"`
	Owner     string `json:"owner"`
	Encoding  string `json:"encoding"`
	Collate   string `json:"collate"`
	SizeBytes uint64 `json:"size_bytes"`
	SizeHuman string `json:"size_human"`
}

// PostgresUser represents a PostgreSQL role/user
type PostgresUser struct {
	Name        string `json:"name"`
	Superuser   bool   `json:"superuser"`
	CreateDB    bool   `json:"create_db"`
	CreateRole  bool   `json:"create_role"`
	Replication bool   `json:"replication"`
}

// ListPostgresDatabases discovers all PostgreSQL databases in real-time
func ListPostgresDatabases(ctx context.Context) ([]PostgresDB, error) {
	if runtime.GOOS != "linux" {
		return []PostgresDB{}, nil
	}

	query := `SELECT datname, pg_get_userbyid(datdba), pg_encoding_to_char(encoding), datcollate, pg_database_size(datname) FROM pg_database WHERE datistemplate = false;`
	cmd := exec.CommandContext(ctx, "sudo", "-u", "postgres", "psql", "-t", "-A", "-F", "|", "-c", query)
	out, err := cmd.Output()
	if err != nil {
		return []PostgresDB{}, nil // PostgreSQL service not running
	}

	var list []PostgresDB
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 5 {
			var sz uint64
			fmt.Sscanf(parts[4], "%d", &sz)
			list = append(list, PostgresDB{
				Name:      parts[0],
				Owner:     parts[1],
				Encoding:  parts[2],
				Collate:   parts[3],
				SizeBytes: sz,
				SizeHuman: formatBytes(sz),
			})
		}
	}

	return list, nil
}

// ListPostgresUsers discovers all PostgreSQL roles
func ListPostgresUsers(ctx context.Context) ([]PostgresUser, error) {
	if runtime.GOOS != "linux" {
		return []PostgresUser{}, nil
	}

	query := `SELECT rolname, rolsuper, rolcreatedb, rolcreaterole, rolreplication FROM pg_roles;`
	cmd := exec.CommandContext(ctx, "sudo", "-u", "postgres", "psql", "-t", "-A", "-F", "|", "-c", query)
	out, err := cmd.Output()
	if err != nil {
		return []PostgresUser{}, nil
	}

	var list []PostgresUser
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 5 {
			list = append(list, PostgresUser{
				Name:        parts[0],
				Superuser:   parts[1] == "t",
				CreateDB:    parts[2] == "t",
				CreateRole:  parts[3] == "t",
				Replication: parts[4] == "t",
			})
		}
	}

	return list, nil
}

// BackupPostgresDatabase creates a pg_dump backup SQL archive
func BackupPostgresDatabase(ctx context.Context, dbName, backupDir string) (string, error) {
	if backupDir == "" {
		backupDir = "/var/lib/aio/backups"
	}
	_ = os.MkdirAll(backupDir, 0750)

	timestamp := time.Now().Format("20060102_150405")
	backupFile := filepath.Join(backupDir, fmt.Sprintf("postgres_%s_%s.sql", dbName, timestamp))

	if runtime.GOOS != "linux" {
		return "", fmt.Errorf("PostgreSQL pg_dump is only available on Linux")
	}

	cmd := exec.CommandContext(ctx, "sudo", "-u", "postgres", "pg_dump", "-Fc", dbName, "-f", backupFile)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pg_dump failed: %s (%w)", strings.TrimSpace(stderr.String()), err)
	}

	return backupFile, nil
}

// RestorePostgresDatabase restores a database from a pg_dump archive
func RestorePostgresDatabase(ctx context.Context, dbName, backupFile string) error {
	if _, err := os.Stat(backupFile); err != nil {
		return fmt.Errorf("backup file not found: %s", backupFile)
	}

	if runtime.GOOS != "linux" {
		return fmt.Errorf("pg_restore is only available on Linux")
	}

	cmd := exec.CommandContext(ctx, "sudo", "-u", "postgres", "pg_restore", "-d", dbName, "--clean", "--if-exists", backupFile)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pg_restore failed: %s (%w)", strings.TrimSpace(stderr.String()), err)
	}

	return nil
}

// CreatePostgresDatabase creates a new PostgreSQL database
func CreatePostgresDatabase(ctx context.Context, dbName, owner string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("PostgreSQL database creation is only available on Linux")
	}

	cleanDb := strings.TrimSpace(dbName)
	if cleanDb == "" {
		return fmt.Errorf("database name cannot be empty")
	}

	query := fmt.Sprintf(`CREATE DATABASE "%s"`, cleanDb)
	if owner != "" {
		query += fmt.Sprintf(` OWNER "%s"`, owner)
	}
	query += ";"

	cmd := exec.CommandContext(ctx, "sudo", "-u", "postgres", "psql", "-c", query)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create database: %s (%w)", strings.TrimSpace(stderr.String()), err)
	}

	return nil
}

// DeletePostgresDatabase drops a PostgreSQL database
func DeletePostgresDatabase(ctx context.Context, dbName string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("PostgreSQL database deletion is only available on Linux")
	}

	cleanDb := strings.TrimSpace(dbName)
	if cleanDb == "" || cleanDb == "postgres" || cleanDb == "template0" || cleanDb == "template1" {
		return fmt.Errorf("invalid or protected database name: %s", dbName)
	}

	query := fmt.Sprintf(`DROP DATABASE "%s";`, cleanDb)
	cmd := exec.CommandContext(ctx, "sudo", "-u", "postgres", "psql", "-c", query)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to drop database: %s (%w)", strings.TrimSpace(stderr.String()), err)
	}

	return nil
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
