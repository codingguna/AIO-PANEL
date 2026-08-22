package security

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// SSHConfig represents key parameters from sshd_config
type SSHConfig struct {
	Port                   int      `json:"port"`
	PermitRootLogin        string   `json:"permit_root_login"`        // yes, no, prohibit-password
	PasswordAuthentication bool     `json:"password_authentication"`  // true/false
	PubkeyAuthentication   bool     `json:"pubkey_authentication"`    // true/false
	AllowUsers             []string `json:"allow_users,omitempty"`
	ConfigPath             string   `json:"config_path"`
}

// SSHSession represents an active user session connected via SSH
type SSHSession struct {
	User      string `json:"user"`
	Terminal  string `json:"terminal"`
	Host      string `json:"host"`
	LoginTime string `json:"login_time"`
}

// SSHKey represents an authorized public key
type SSHKey struct {
	Index       int    `json:"index"`
	Type        string `json:"type"` // ssh-ed25519, ssh-rsa, ecdsa-sha2-nistp256
	Fingerprint string `json:"fingerprint"`
	Comment     string `json:"comment"`
	KeyData     string `json:"key_data"`
}

var (
	portRegex         = regexp.MustCompile(`(?i)^\s*Port\s+(\d+)`)
	rootLoginRegex    = regexp.MustCompile(`(?i)^\s*PermitRootLogin\s+([a-zA-Z0-9\-]+)`)
	passwordAuthRegex = regexp.MustCompile(`(?i)^\s*PasswordAuthentication\s+(yes|no)`)
	pubkeyAuthRegex   = regexp.MustCompile(`(?i)^\s*PubkeyAuthentication\s+(yes|no)`)
)

// GetSSHConfig parses the host sshd_config in read-only mode
func GetSSHConfig(ctx context.Context) (*SSHConfig, error) {
	configPath := "/etc/ssh/sshd_config"
	if runtime.GOOS != "linux" {
		return &SSHConfig{
			Port:                   22,
			PermitRootLogin:        "unknown",
			PasswordAuthentication: false,
			PubkeyAuthentication:   true,
			ConfigPath:             configPath,
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return &SSHConfig{
			Port:                   22,
			PermitRootLogin:        "unknown",
			PasswordAuthentication: false,
			PubkeyAuthentication:   true,
			ConfigPath:             configPath,
		}, nil
	}

	cfg := &SSHConfig{
		Port:                   22, // default
		PermitRootLogin:        "yes",
		PasswordAuthentication: true,
		PubkeyAuthentication:   true,
		ConfigPath:             configPath,
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		if match := portRegex.FindStringSubmatch(line); len(match) > 1 {
			if p, err := strconv.Atoi(match[1]); err == nil {
				cfg.Port = p
			}
		} else if match := rootLoginRegex.FindStringSubmatch(line); len(match) > 1 {
			cfg.PermitRootLogin = strings.ToLower(match[1])
		} else if match := passwordAuthRegex.FindStringSubmatch(line); len(match) > 1 {
			cfg.PasswordAuthentication = strings.ToLower(match[1]) == "yes"
		} else if match := pubkeyAuthRegex.FindStringSubmatch(line); len(match) > 1 {
			cfg.PubkeyAuthentication = strings.ToLower(match[1]) == "yes"
		}
	}

	return cfg, nil
}

// UpdateSSHConfig applies changes safely using: Backup -> Validate (sshd -t) -> Reload -> Rollback on error
func UpdateSSHConfig(ctx context.Context, newPort int, permitRoot string, passwordAuth bool, backupDir string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("ssh configuration update requires Linux environment")
	}

	configPath := "/etc/ssh/sshd_config"
	originalContent, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read current sshd_config: %w", err)
	}

	// 1. Create Timestamped Backup
	if backupDir == "" {
		backupDir = "/var/lib/aio/backups"
	}
	_ = os.MkdirAll(backupDir, 0700)
	backupPath := filepath.Join(backupDir, fmt.Sprintf("sshd_config.%d.bak", time.Now().Unix()))
	if err := os.WriteFile(backupPath, originalContent, 0600); err != nil {
		return fmt.Errorf("failed to create sshd backup: %w", err)
	}

	// 2. Generate New Config with Modified Directives
	lines := strings.Split(string(originalContent), "\n")
	var updatedLines []string
	portSet, rootSet, passSet := false, false, false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if portRegex.MatchString(trimmed) || strings.HasPrefix(trimmed, "#Port ") {
			if !portSet {
				updatedLines = append(updatedLines, fmt.Sprintf("Port %d", newPort))
				portSet = true
			}
		} else if rootLoginRegex.MatchString(trimmed) || strings.HasPrefix(trimmed, "#PermitRootLogin ") {
			if !rootSet {
				updatedLines = append(updatedLines, fmt.Sprintf("PermitRootLogin %s", permitRoot))
				rootSet = true
			}
		} else if passwordAuthRegex.MatchString(trimmed) || strings.HasPrefix(trimmed, "#PasswordAuthentication ") {
			if !passSet {
				val := "no"
				if passwordAuth {
					val = "yes"
				}
				updatedLines = append(updatedLines, fmt.Sprintf("PasswordAuthentication %s", val))
				passSet = true
			}
		} else {
			updatedLines = append(updatedLines, line)
		}
	}

	if !portSet {
		updatedLines = append(updatedLines, fmt.Sprintf("Port %d", newPort))
	}
	if !rootSet {
		updatedLines = append(updatedLines, fmt.Sprintf("PermitRootLogin %s", permitRoot))
	}
	if !passSet {
		val := "no"
		if passwordAuth {
			val = "yes"
		}
		updatedLines = append(updatedLines, fmt.Sprintf("PasswordAuthentication %s", val))
	}

	newContent := strings.Join(updatedLines, "\n")

	// 3. Write Proposed Config
	if err := os.WriteFile(configPath, []byte(newContent), 0600); err != nil {
		return fmt.Errorf("failed to write proposed sshd_config: %w", err)
	}

	// 4. Validate with sshd -t
	validateCmd := exec.CommandContext(ctx, "sshd", "-t")
	var validateErr bytes.Buffer
	validateCmd.Stderr = &validateErr

	if err := validateCmd.Run(); err != nil {
		// ROLLBACK IMMEDIATELY
		_ = os.WriteFile(configPath, originalContent, 0600)
		return fmt.Errorf("validation failed (sshd -t), automatically rolled back to backup: %s (%w)",
			strings.TrimSpace(validateErr.String()), err)
	}

	// 5. Reload SSH Service
	reloadCmd := exec.CommandContext(ctx, "systemctl", "reload", "ssh")
	if err := reloadCmd.Run(); err != nil {
		// Try sshd service name
		_ = exec.CommandContext(ctx, "systemctl", "reload", "sshd").Run()
	}

	return nil
}

// ListAuthorizedKeys reads a user's authorized_keys file
func ListAuthorizedKeys(username string) ([]SSHKey, error) {
	if username == "" {
		username = "root"
	}

	var authFile string
	if runtime.GOOS == "linux" {
		u, err := user.Lookup(username)
		if err != nil {
			return []SSHKey{}, nil
		}
		authFile = filepath.Join(u.HomeDir, ".ssh", "authorized_keys")
	} else {
		// Non-linux: look up user home directory
		u, err := user.Current()
		if err != nil {
			return []SSHKey{}, nil
		}
		authFile = filepath.Join(u.HomeDir, ".ssh", "authorized_keys")
	}

	if _, err := os.Stat(authFile); err != nil {
		return []SSHKey{}, nil
	}

	data, err := os.ReadFile(authFile)
	if err != nil {
		return []SSHKey{}, nil
	}

	var keys []SSHKey
	scanner := bufio.NewScanner(bytes.NewReader(data))
	idx := 1

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 2 {
			comment := ""
			if len(fields) >= 3 {
				comment = strings.Join(fields[2:], " ")
			}
			keys = append(keys, SSHKey{
				Index:   idx,
				Type:    fields[0],
				Comment: comment,
				KeyData: line,
			})
			idx++
		}
	}

	return keys, nil
}

// AddAuthorizedKey appends a new public key to a user's authorized_keys file
func AddAuthorizedKey(username, rawKey string) error {
	rawKey = strings.TrimSpace(rawKey)
	if rawKey == "" {
		return fmt.Errorf("public key cannot be empty")
	}

	validTypes := []string{"ssh-rsa", "ssh-ed25519", "ecdsa-sha2-nistp256", "ecdsa-sha2-nistp384", "ecdsa-sha2-nistp521"}
	isValid := false
	for _, t := range validTypes {
		if strings.HasPrefix(rawKey, t) {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("unsupported key type; must start with ssh-ed25519, ssh-rsa, or ecdsa")
	}

	var homeDir string
	if runtime.GOOS == "linux" {
		u, err := user.Lookup(username)
		if err != nil {
			return fmt.Errorf("user %s not found: %w", username, err)
		}
		homeDir = u.HomeDir
	} else {
		u, err := user.Current()
		if err != nil {
			return err
		}
		homeDir = u.HomeDir
	}

	sshDir := filepath.Join(homeDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	authFile := filepath.Join(sshDir, "authorized_keys")
	f, err := os.OpenFile(authFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("failed to open authorized_keys: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(rawKey + "\n"); err != nil {
		return fmt.Errorf("failed to append public key: %w", err)
	}

	return nil
}

// GetActiveSSHSessions returns active logged-in terminal sessions
func GetActiveSSHSessions(ctx context.Context) ([]SSHSession, error) {
	if runtime.GOOS != "linux" {
		return []SSHSession{}, nil
	}

	cmd := exec.CommandContext(ctx, "who")
	out, err := cmd.Output()
	if err != nil {
		return []SSHSession{}, nil
	}

	var sessions []SSHSession
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 3 {
			s := SSHSession{
				User:      fields[0],
				Terminal:  fields[1],
				LoginTime: fields[2],
			}
			if len(fields) >= 5 {
				s.Host = strings.Trim(fields[4], "()")
			}
			sessions = append(sessions, s)
		}
	}

	return sessions, nil
}
