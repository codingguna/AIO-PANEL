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

// MySQLDB represents a MySQL / MariaDB database
type MySQLDB struct {
	Name      string `json:"name"`
	SizeBytes uint64 `json:"size_bytes"`
	SizeHuman string `json:"size_human"`
}

// ListMySQLDatabases discovers all MySQL databases in real-time
func ListMySQLDatabases(ctx context.Context) ([]MySQLDB, error) {
	if runtime.GOOS != "linux" {
		return []MySQLDB{}, nil
	}

	cmd := exec.CommandContext(ctx, "mysql", "-e", "SHOW DATABASES;", "--batch", "--skip-column-names")
	out, err := cmd.Output()
	if err != nil {
		return []MySQLDB{}, nil // MySQL not running
	}

	var list []MySQLDB
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if name != "" && name != "information_schema" && name != "performance_schema" {
			list = append(list, MySQLDB{
				Name:      name,
				SizeHuman: "N/A",
			})
		}
	}

	return list, nil
}

// BackupMySQLDatabase dumps a MySQL database using mysqldump
func BackupMySQLDatabase(ctx context.Context, dbName, backupDir string) (string, error) {
	if backupDir == "" {
		backupDir = "/var/lib/aio/backups"
	}
	_ = os.MkdirAll(backupDir, 0750)

	timestamp := time.Now().Format("20060102_150405")
	backupFile := filepath.Join(backupDir, fmt.Sprintf("mysql_%s_%s.sql", dbName, timestamp))

	if runtime.GOOS != "linux" {
		return "", fmt.Errorf("mysqldump is only available on Linux")
	}

	f, err := os.Create(backupFile)
	if err != nil {
		return "", err
	}
	defer f.Close()

	cmd := exec.CommandContext(ctx, "mysqldump", dbName)
	cmd.Stdout = f
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		_ = os.Remove(backupFile)
		return "", fmt.Errorf("mysqldump failed: %s (%w)", strings.TrimSpace(stderr.String()), err)
	}

	return backupFile, nil
}

// CreateMySQLDatabase creates a new MySQL database
func CreateMySQLDatabase(ctx context.Context, dbName string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("MySQL database creation is only available on Linux")
	}

	cleanDb := strings.TrimSpace(dbName)
	if cleanDb == "" {
		return fmt.Errorf("database name cannot be empty")
	}

	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`;", cleanDb)
	cmd := exec.CommandContext(ctx, "mysql", "-e", query)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create MySQL database: %s (%w)", strings.TrimSpace(stderr.String()), err)
	}

	return nil
}

// DeleteMySQLDatabase drops a MySQL database
func DeleteMySQLDatabase(ctx context.Context, dbName string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("MySQL database deletion is only available on Linux")
	}

	cleanDb := strings.TrimSpace(dbName)
	if cleanDb == "" || cleanDb == "mysql" || cleanDb == "sys" || cleanDb == "information_schema" || cleanDb == "performance_schema" {
		return fmt.Errorf("invalid or protected database name: %s", dbName)
	}

	query := fmt.Sprintf("DROP DATABASE IF EXISTS `%s`;", cleanDb)
	cmd := exec.CommandContext(ctx, "mysql", "-e", query)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to drop MySQL database: %s (%w)", strings.TrimSpace(stderr.String()), err)
	}

	return nil
}
