package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite" // Pure Go SQLite driver (No CGO required)
)

// Store wraps the database handle and exposes high-level database operations
type Store struct {
	db *sql.DB
}

// Open initializes the SQLite database, configures pragmas, and runs migrations
func Open(dbPath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0750); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// SQLite connection string with optimized performance pragmas
	dsn := fmt.Sprintf("%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=synchronous(NORMAL)", dbPath)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// SQLite handles concurrent reads well in WAL mode, but single-writer is standard
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	store := &Store{db: db}
	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("database migration failed: %w", err)
	}

	return store, nil
}

// Close closes the underlying SQLite connection
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// DB returns the underlying *sql.DB instance
func (s *Store) DB() *sql.DB {
	return s.db
}

// AuditEvent represents an audit record
type AuditEvent struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	User      string    `json:"user"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	Result    string    `json:"result"` // SUCCESS, FAILURE, WARNING
	Details   string    `json:"details"`
	IPAddress string    `json:"ip_address"`
}

// LogAudit writes an action to the audit_events table
func (s *Store) LogAudit(ctx context.Context, user, action, target, result, details, ip string) error {
	query := `
		INSERT INTO audit_events (timestamp, user, action, target, result, details, ip_address)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.ExecContext(ctx, query, time.Now().UTC(), user, action, target, result, details, ip)
	return err
}

// GetRecentAuditEvents retrieves the most recent audit records
func (s *Store) GetRecentAuditEvents(ctx context.Context, limit int) ([]AuditEvent, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}

	query := `
		SELECT id, timestamp, user, action, target, result, details, ip_address
		FROM audit_events
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []AuditEvent
	for rows.Next() {
		var e AuditEvent
		var ts time.Time
		if err := rows.Scan(&e.ID, &ts, &e.User, &e.Action, &e.Target, &e.Result, &e.Details, &e.IPAddress); err != nil {
			return nil, err
		}
		e.Timestamp = ts
		events = append(events, e)
	}

	return events, rows.Err()
}

// SystemSnapshot represents historical server stats
type SystemSnapshot struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Hostname  string    `json:"hostname"`
	CPUPct    float64   `json:"cpu_pct"`
	RAMPct    float64   `json:"ram_pct"`
	DiskPct   float64   `json:"disk_pct"`
	LoadAvg   float64   `json:"load_avg"`
}

// SaveSystemSnapshot records a point-in-time metrics sample
func (s *Store) SaveSystemSnapshot(ctx context.Context, snap SystemSnapshot) error {
	query := `
		INSERT INTO system_snapshots (timestamp, hostname, cpu_pct, ram_pct, disk_pct, load_avg)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.ExecContext(ctx, query, time.Now().UTC(), snap.Hostname, snap.CPUPct, snap.RAMPct, snap.DiskPct, snap.LoadAvg)
	return err
}

// GetRecentSnapshots retrieves the most recent system snapshots
func (s *Store) GetRecentSnapshots(ctx context.Context, limit int) ([]SystemSnapshot, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	query := `
		SELECT id, timestamp, hostname, cpu_pct, ram_pct, disk_pct, load_avg
		FROM system_snapshots
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []SystemSnapshot
	for rows.Next() {
		var snap SystemSnapshot
		var ts time.Time
		if err := rows.Scan(&snap.ID, &ts, &snap.Hostname, &snap.CPUPct, &snap.RAMPct, &snap.DiskPct, &snap.LoadAvg); err != nil {
			return nil, err
		}
		snap.Timestamp = ts
		list = append(list, snap)
	}

	return list, rows.Err()
}

// PanelUser represents an administrative user of AIO-PANEL
type PanelUser struct {
	ID          int64      `json:"id"`
	Username    string     `json:"username"`
	Role        string     `json:"role"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at"`
	IsActive    bool       `json:"is_active"`
}

// CountPanelUsers returns the number of active panel administrators
func (s *Store) CountPanelUsers(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM panel_users WHERE is_active = 1").Scan(&count)
	return count, err
}

// CreatePanelUser creates a new panel administrator with Bcrypt password hashing
func (s *Store) CreatePanelUser(ctx context.Context, username, password, role string) (*PanelUser, error) {
	cleanUser := strings.TrimSpace(username)
	if cleanUser == "" || password == "" {
		return nil, fmt.Errorf("username and password are required")
	}

	if role == "" {
		role = "admin"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	query := `
		INSERT INTO panel_users (username, password_hash, role, is_active)
		VALUES (?, ?, ?, 1)
	`
	res, err := s.db.ExecContext(ctx, query, cleanUser, string(hash), role)
	if err != nil {
		return nil, fmt.Errorf("failed to insert user (username may already exist): %w", err)
	}

	id, _ := res.LastInsertId()
	return &PanelUser{
		ID:        id,
		Username:  cleanUser,
		Role:      role,
		CreatedAt: time.Now().UTC(),
		IsActive:  true,
	}, nil
}

// AuthenticateUser verifies username and password, returning user if valid
func (s *Store) AuthenticateUser(ctx context.Context, username, password string) (*PanelUser, error) {
	query := `
		SELECT id, username, password_hash, role, created_at, updated_at, last_login_at, is_active
		FROM panel_users
		WHERE username = ? AND is_active = 1
	`
	var u PanelUser
	var hash string
	var lastLogin sql.NullTime

	err := s.db.QueryRowContext(ctx, query, username).Scan(
		&u.ID, &u.Username, &hash, &u.Role, &u.CreatedAt, &u.UpdatedAt, &lastLogin, &u.IsActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid username or password")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	if lastLogin.Valid {
		u.LastLoginAt = &lastLogin.Time
	}

	// Update last_login_at timestamp
	_, _ = s.db.ExecContext(ctx, "UPDATE panel_users SET last_login_at = ? WHERE id = ?", time.Now().UTC(), u.ID)

	return &u, nil
}

// ListPanelUsers lists all panel administrative users
func (s *Store) ListPanelUsers(ctx context.Context) ([]PanelUser, error) {
	query := `
		SELECT id, username, role, created_at, updated_at, last_login_at, is_active
		FROM panel_users
		ORDER BY id ASC
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []PanelUser
	for rows.Next() {
		var u PanelUser
		var lastLogin sql.NullTime
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.CreatedAt, &u.UpdatedAt, &lastLogin, &u.IsActive); err != nil {
			return nil, err
		}
		if lastLogin.Valid {
			u.LastLoginAt = &lastLogin.Time
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// UpdatePanelUserPassword changes password for a panel user
func (s *Store) UpdatePanelUserPassword(ctx context.Context, username, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	res, err := s.db.ExecContext(ctx, "UPDATE panel_users SET password_hash = ?, updated_at = ? WHERE username = ?", string(hash), time.Now().UTC(), username)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user not found: %s", username)
	}
	return nil
}

// DeletePanelUser removes an admin user
func (s *Store) DeletePanelUser(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM panel_users WHERE id = ?", id)
	return err
}
