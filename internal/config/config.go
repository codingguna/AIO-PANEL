package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

// Config represents all panel settings
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Logging  LoggingConfig  `json:"logging"`
	Paths    PathsConfig    `json:"paths"`
}

// ServerConfig holds HTTP/HTTPS server parameters
type ServerConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	TLS      bool   `json:"tls"`
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
}

// DatabaseConfig holds embedded SQLite parameters
type DatabaseConfig struct {
	Path string `json:"path"`
}

// LoggingConfig holds log output preferences
type LoggingConfig struct {
	Level  string `json:"level"`  // debug, info, warn, error
	Format string `json:"format"` // pretty, text, json
	File   string `json:"file"`
}

// PathsConfig holds standard system directories
type PathsConfig struct {
	ConfigDir string `json:"config_dir"`
	DataDir   string `json:"data_dir"`
	LogDir    string `json:"log_dir"`
	BackupDir string `json:"backup_dir"`
}

// Default returns sensible defaults depending on OS and execution mode
func Default() *Config {
	isLinux := runtime.GOOS == "linux"
	var configDir, dataDir, logDir string

	// Check if running as root on Linux production, or local development
	if isLinux && os.Geteuid() == 0 {
		configDir = "/etc/aio"
		dataDir = "/var/lib/aio"
		logDir = "/var/log/aio"
	} else {
		// Development mode / non-root / Windows fallback
		execDir, err := os.Getwd()
		if err != nil {
			execDir = "."
		}
		dataDir = filepath.Join(execDir, "data")
		configDir = filepath.Join(dataDir, "config")
		logDir = filepath.Join(dataDir, "logs")
	}

	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 5555,
			TLS:  false,
		},
		Database: DatabaseConfig{
			Path: filepath.Join(dataDir, "aio.db"),
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "pretty",
			File:   filepath.Join(logDir, "aio.log"),
		},
		Paths: PathsConfig{
			ConfigDir: configDir,
			DataDir:   dataDir,
			LogDir:    logDir,
			BackupDir: filepath.Join(dataDir, "backups"),
		},
	}
}

// Load loads configuration from disk with env var overrides
func Load(customConfigPath string) (*Config, error) {
	cfg := Default()

	configPath := customConfigPath
	if configPath == "" {
		configPath = filepath.Join(cfg.Paths.ConfigDir, "aio.conf")
	}

	// Read file if it exists
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file at %s: %w", configPath, err)
		}

		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Environment variable overrides
	if host := os.Getenv("AIO_HOST"); host != "" {
		cfg.Server.Host = host
	}
	if portStr := os.Getenv("AIO_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			cfg.Server.Port = p
		}
	}
	if dataDir := os.Getenv("AIO_DATA_DIR"); dataDir != "" {
		cfg.Paths.DataDir = dataDir
		cfg.Database.Path = filepath.Join(dataDir, "aio.db")
		cfg.Paths.BackupDir = filepath.Join(dataDir, "backups")
	}
	if logDir := os.Getenv("AIO_LOG_DIR"); logDir != "" {
		cfg.Paths.LogDir = logDir
		cfg.Logging.File = filepath.Join(logDir, "aio.log")
	}
	if logLevel := os.Getenv("AIO_LOG_LEVEL"); logLevel != "" {
		cfg.Logging.Level = logLevel
	}
	if logFormat := os.Getenv("AIO_LOG_FORMAT"); logFormat != "" {
		cfg.Logging.Format = logFormat
	}

	return cfg, nil
}

// Save writes current configuration to disk
func (c *Config) Save(path string) error {
	if path == "" {
		path = filepath.Join(c.Paths.ConfigDir, "aio.conf")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return os.WriteFile(path, data, 0600)
}

// EnsureDirectories creates required data, log, and config directories
func (c *Config) EnsureDirectories() error {
	dirs := []string{
		c.Paths.ConfigDir,
		c.Paths.DataDir,
		c.Paths.LogDir,
		c.Paths.BackupDir,
	}

	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		if err := os.MkdirAll(dir, 0750); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}
