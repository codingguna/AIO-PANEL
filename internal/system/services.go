package system

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Service represents a systemd service unit
type Service struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	UnitFile    string `json:"unit_file"`
	ActiveState string `json:"active_state"` // active, inactive, failed, activating, deactivating
	SubState    string `json:"sub_state"`    // running, dead, exited, failed
	LoadState   string `json:"load_state"`   // loaded, not-found, bad-setting
	Enabled     bool   `json:"enabled"`      // enabled, disabled, static
	OwnerType   string `json:"owner_type"`   // 'aio' or 'external'
	PID         int    `json:"pid"`
	MemoryBytes uint64 `json:"memory_bytes"`
	Uptime      string `json:"uptime"`
}

// Valid service name regex to prevent shell injection
var validUnitNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-\.@:]+$`)

// Common core services to proactively detect
var keyServices = []string{
	"nginx",
	"apache2",
	"httpd",
	"postgresql",
	"mysql",
	"mariadb",
	"redis",
	"redis-server",
	"docker",
	"ssh",
	"sshd",
	"ufw",
	"gunicorn",
	"cron",
	"crond",
	"aio-panel",
}

// ListServices returns all relevant system services in real-time
func ListServices(ctx context.Context) ([]Service, error) {
	if runtime.GOOS != "linux" {
		return []Service{}, nil
	}

	// Verify if systemctl is available
	if _, err := exec.LookPath("systemctl"); err != nil {
		return []Service{}, nil
	}

	// Query systemctl for list-units
	cmd := exec.CommandContext(ctx, "systemctl", "list-units", "--type=service", "--all", "--no-pager", "--plain", "--no-legend")
	out, err := cmd.Output()
	if err != nil {
		return checkKeyServices(ctx), nil
	}

	servicesMap := make(map[string]Service)
	scanner := bufio.NewScanner(bytes.NewReader(out))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		rawUnit := fields[0]
		unitName := strings.TrimSuffix(rawUnit, ".service")
		loadState := fields[1]
		activeState := fields[2]
		subState := fields[3]

		desc := ""
		if len(fields) >= 5 {
			desc = strings.Join(fields[4:], " ")
		}

		owner := "external"
		if unitName == "aio-panel" || unitName == "aio" {
			owner = "aio"
		}

		servicesMap[unitName] = Service{
			Name:        unitName,
			DisplayName: formatDisplayName(unitName),
			Description: desc,
			UnitFile:    rawUnit,
			ActiveState: activeState,
			SubState:    subState,
			LoadState:   loadState,
			OwnerType:   owner,
		}
	}

	// Always ensure key services are reported even if not actively loaded
	for _, key := range keyServices {
		if _, exists := servicesMap[key]; !exists {
			svc, err := InspectService(ctx, key)
			if err == nil && svc.LoadState != "not-found" {
				servicesMap[key] = *svc
			}
		}
	}

	var result []Service
	for _, svc := range servicesMap {
		result = append(result, svc)
	}

	return result, nil
}

// InspectService gets comprehensive properties of a specific systemd unit
func InspectService(ctx context.Context, name string) (*Service, error) {
	cleanName := strings.TrimSuffix(name, ".service")
	if !validUnitNameRegex.MatchString(cleanName) {
		return nil, fmt.Errorf("invalid service name: %s", name)
	}

	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf("systemd is only available on Linux")
	}

	unit := cleanName + ".service"
	cmd := exec.CommandContext(ctx, "systemctl", "show", unit, "--no-pager")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to show service %s: %w", name, err)
	}

	svc := &Service{
		Name:        cleanName,
		DisplayName: formatDisplayName(cleanName),
		UnitFile:    unit,
		OwnerType:   "external",
	}
	if cleanName == "aio-panel" || cleanName == "aio" {
		svc.OwnerType = "aio"
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, val := parts[0], parts[1]

		switch key {
		case "Description":
			svc.Description = val
		case "ActiveState":
			svc.ActiveState = val
		case "SubState":
			svc.SubState = val
		case "LoadState":
			svc.LoadState = val
		case "UnitFileState":
			svc.Enabled = val == "enabled"
		case "MainPID":
			if pid, err := strconv.Atoi(val); err == nil {
				svc.PID = pid
			}
		case "MemoryCurrent":
			if mem, err := strconv.ParseUint(val, 10, 64); err == nil && mem != ^uint64(0) {
				svc.MemoryBytes = mem
			}
		case "ActiveEnterTimestamp":
			if t, err := time.Parse("Mon 2006-01-02 15:04:05 MST", val); err == nil && !t.IsZero() {
				svc.Uptime = time.Since(t).Round(time.Second).String()
			}
		}
	}

	return svc, nil
}

// ControlService runs systemctl start|stop|restart|reload|enable|disable
func ControlService(ctx context.Context, name, action string) error {
	cleanName := strings.TrimSuffix(name, ".service")
	if !validUnitNameRegex.MatchString(cleanName) {
		return fmt.Errorf("invalid service name: %s", name)
	}

	validActions := map[string]bool{
		"start":   true,
		"stop":    true,
		"restart": true,
		"reload":  true,
		"enable":  true,
		"disable": true,
	}

	if !validActions[action] {
		return fmt.Errorf("unsupported service action: %s", action)
	}

	if runtime.GOOS != "linux" {
		return fmt.Errorf("service control requires Linux systemd")
	}

	unit := cleanName + ".service"
	cmd := exec.CommandContext(ctx, "systemctl", action, unit)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("systemctl %s %s failed: %s (%w)", action, unit, strings.TrimSpace(stderr.String()), err)
	}

	return nil
}

// GetServiceLogs fetches recent journal logs for a service
func GetServiceLogs(ctx context.Context, name string, lines int) (string, error) {
	cleanName := strings.TrimSuffix(name, ".service")
	if !validUnitNameRegex.MatchString(cleanName) {
		return "", fmt.Errorf("invalid service name: %s", name)
	}

	if lines <= 0 || lines > 500 {
		lines = 50
	}

	if runtime.GOOS != "linux" {
		return "", fmt.Errorf("journalctl is only available on Linux")
	}

	unit := cleanName + ".service"
	cmd := exec.CommandContext(ctx, "journalctl", "-u", unit, "-n", strconv.Itoa(lines), "--no-pager")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("journalctl failed for %s: %w", unit, err)
	}

	return string(out), nil
}

func checkKeyServices(ctx context.Context) []Service {
	var list []Service
	for _, k := range keyServices {
		svc, err := InspectService(ctx, k)
		if err == nil && svc.LoadState != "not-found" {
			list = append(list, *svc)
		}
	}
	return list
}

func formatDisplayName(name string) string {
	switch strings.ToLower(name) {
	case "nginx":
		return "Nginx Web Server"
	case "apache2", "httpd":
		return "Apache Web Server"
	case "postgresql":
		return "PostgreSQL Database"
	case "mysql", "mariadb":
		return "MySQL / MariaDB Database"
	case "redis", "redis-server":
		return "Redis Cache"
	case "docker":
		return "Docker Engine"
	case "ssh", "sshd":
		return "OpenSSH Daemon"
	case "ufw":
		return "UFW Firewall"
	case "gunicorn":
		return "Gunicorn WSGI Server"
	case "aio-panel":
		return "AIO-PANEL Core Daemon"
	default:
		return name
	}
}

// CreateService creates a new systemd unit file and enables/starts it
func CreateService(ctx context.Context, name, description, execStart, workDir, user, restart string, enable bool) error {
	cleanName := strings.TrimSuffix(name, ".service")
	if !validUnitNameRegex.MatchString(cleanName) {
		return fmt.Errorf("invalid service name: %s", name)
	}

	if runtime.GOOS != "linux" {
		return fmt.Errorf("service creation requires Linux systemd")
	}

	if user == "" {
		user = "root"
	}
	if restart == "" {
		restart = "always"
	}
	if description == "" {
		description = fmt.Sprintf("Service %s managed by AIO-PANEL", cleanName)
	}

	unitPath := filepath.Join("/etc/systemd/system", cleanName+".service")
	if _, err := os.Stat(unitPath); err == nil {
		return fmt.Errorf("service unit %s already exists", unitPath)
	}

	unitContent := fmt.Sprintf(`[Unit]
Description=%s
After=network.target

[Service]
Type=simple
User=%s
WorkingDirectory=%s
ExecStart=%s
Restart=%s
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
`, description, user, workDir, execStart, restart)

	if err := os.WriteFile(unitPath, []byte(unitContent), 0644); err != nil {
		return fmt.Errorf("failed to write unit file: %w", err)
	}

	_ = exec.CommandContext(ctx, "systemctl", "daemon-reload").Run()

	if enable {
		_ = exec.CommandContext(ctx, "systemctl", "enable", cleanName+".service").Run()
		_ = exec.CommandContext(ctx, "systemctl", "start", cleanName+".service").Run()
	}

	return nil
}

// DeleteService stops, disables, and deletes a systemd unit file
func DeleteService(ctx context.Context, name string) error {
	cleanName := strings.TrimSuffix(name, ".service")
	if !validUnitNameRegex.MatchString(cleanName) {
		return fmt.Errorf("invalid service name: %s", name)
	}

	if runtime.GOOS != "linux" {
		return fmt.Errorf("service deletion requires Linux systemd")
	}

	// Stop & disable service
	_ = exec.CommandContext(ctx, "systemctl", "stop", cleanName+".service").Run()
	_ = exec.CommandContext(ctx, "systemctl", "disable", cleanName+".service").Run()

	unitPath := filepath.Join("/etc/systemd/system", cleanName+".service")
	if err := os.Remove(unitPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove unit file: %w", err)
	}

	_ = exec.CommandContext(ctx, "systemctl", "daemon-reload").Run()
	return nil
}
