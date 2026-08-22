package db

import (
	"database/sql"
	"fmt"
)

var migrations = []struct {
	version int
	name    string
	sql     string
}{
	{
		version: 1,
		name:    "initial_schema",
		sql: `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS panel_users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'admin',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_login_at DATETIME,
			is_active BOOLEAN NOT NULL DEFAULT 1
		);

		CREATE TABLE IF NOT EXISTS audit_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME NOT NULL,
			user TEXT NOT NULL,
			action TEXT NOT NULL,
			target TEXT NOT NULL,
			result TEXT NOT NULL,
			details TEXT,
			ip_address TEXT
		);

		CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_events (timestamp DESC);
		CREATE INDEX IF NOT EXISTS idx_audit_user ON audit_events (user);
		CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_events (action);

		CREATE TABLE IF NOT EXISTS system_snapshots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME NOT NULL,
			hostname TEXT NOT NULL,
			cpu_pct REAL NOT NULL,
			ram_pct REAL NOT NULL,
			disk_pct REAL NOT NULL,
			load_avg REAL NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_snapshots_timestamp ON system_snapshots (timestamp DESC);

		CREATE TABLE IF NOT EXISTS managed_resources (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT NOT NULL,          -- 'service', 'domain', 'app', 'database', 'user'
			name TEXT NOT NULL,
			path TEXT,
			owner_type TEXT NOT NULL,    -- 'aio' or 'external'
			status TEXT,                 -- 'active', 'inactive', 'stopped', 'error'
			metadata TEXT,               -- JSON metadata
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE UNIQUE INDEX IF NOT EXISTS idx_resources_type_name ON managed_resources (type, name);
		`,
	},
}

func (s *Store) migrate() error {
	// Ensure migration table exists
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	for _, m := range migrations {
		var applied int
		err := s.db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", m.version).Scan(&applied)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to check migration %d status: %w", m.version, err)
		}

		if applied == 0 {
			tx, err := s.db.Begin()
			if err != nil {
				return fmt.Errorf("failed to start migration transaction for v%d: %w", m.version, err)
			}

			if _, err := tx.Exec(m.sql); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed executing migration v%d (%s): %w", m.version, m.name, err)
			}

			if _, err := tx.Exec("INSERT INTO schema_migrations (version, name) VALUES (?, ?)", m.version, m.name); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed recording migration v%d: %w", m.version, err)
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed committing migration v%d: %w", m.version, err)
			}
		}
	}

	return nil
}
