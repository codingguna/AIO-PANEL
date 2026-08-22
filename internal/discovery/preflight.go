package discovery

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/codingguna/aio-panel/internal/config"
	"github.com/codingguna/aio-panel/internal/system"
)

// PreflightReport contains the read-only inspection results
type PreflightReport struct {
	OS               string            `json:"os"`
	Architecture     string            `json:"architecture"`
	Systemd          bool              `json:"systemd"`
	PortAvailable    bool              `json:"port_available"`
	Port             int               `json:"port"`
	PortOccupiedBy   string            `json:"port_occupied_by,omitempty"`
	Tools            map[string]bool   `json:"tools"`
	ToolVersions     map[string]string `json:"tool_versions"`
	ExistingSites    []string          `json:"existing_sites"`
	ExistingServices []string          `json:"existing_services"`
	ExistingApps     []DiscoveredApp   `json:"existing_apps"`
}

// RunPreflight performs a complete read-only discovery of the server
func RunPreflight(ctx context.Context, cfg *config.Config) (*PreflightReport, error) {
	info, err := system.GetInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve host info: %w", err)
	}

	report := &PreflightReport{
		OS:               info.OS,
		Architecture:     info.Architecture,
		Port:             cfg.Server.Port,
		Tools:            make(map[string]bool),
		ToolVersions:     make(map[string]string),
		ExistingSites:    make([]string, 0),
		ExistingServices: make([]string, 0),
		ExistingApps:     make([]DiscoveredApp, 0),
	}

	// 1. Check systemd
	if _, err := exec.LookPath("systemctl"); err == nil {
		report.Systemd = true
	} else {
		report.Systemd = false
	}

	// 2. Check Port 5555
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		report.PortAvailable = false
		report.PortOccupiedBy = findProcessOnPort(cfg.Server.Port)
	} else {
		report.PortAvailable = true
		ln.Close()
	}

	// 3. Inspect standard tools
	toolsToCheck := []struct {
		key string
		cmd string
		arg string
	}{
		{"Nginx", "nginx", "-v"},
		{"Apache", "apache2", "-v"},
		{"Python", "python3", "--version"},
		{"Node.js", "node", "--version"},
		{"npm", "npm", "--version"},
		{"Git", "git", "--version"},
		{"PostgreSQL", "psql", "--version"},
		{"MySQL", "mysql", "--version"},
		{"Redis", "redis-server", "--version"},
		{"Docker", "docker", "--version"},
		{"UFW", "ufw", "version"},
	}

	for _, t := range toolsToCheck {
		if path, err := exec.LookPath(t.cmd); err == nil && path != "" {
			report.Tools[t.key] = true
			cmd := exec.CommandContext(ctx, t.cmd, t.arg)
			out, err := cmd.CombinedOutput()
			if err == nil {
				report.ToolVersions[t.key] = cleanVersionOutput(string(out))
			} else {
				report.ToolVersions[t.key] = "installed"
			}
		} else {
			report.Tools[t.key] = false
		}
	}

	// 4. Discover existing Nginx sites
	sites, _ := DiscoverNginxSites(ctx)
	for _, site := range sites {
		report.ExistingSites = append(report.ExistingSites, site.Domain)
	}

	// 5. Discover existing custom systemd services
	services, _ := system.ListServices(ctx)
	for _, s := range services {
		if s.OwnerType == "external" {
			report.ExistingServices = append(report.ExistingServices, s.Name)
		}
	}

	// 6. Discover applications (Django, Node, React, etc.)
	apps, _ := DiscoverApplications(ctx)
	report.ExistingApps = apps

	return report, nil
}

func findProcessOnPort(port int) string {
	if runtime.GOOS == "linux" {
		// Use lsof or ss or fuser if available
		cmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port), "-t")
		out, err := cmd.Output()
		if err == nil {
			pidStr := strings.TrimSpace(string(out))
			if pid, err := strconv.Atoi(pidStr); err == nil {
				// Read process name from /proc/<pid>/comm
				if comm, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid)); err == nil {
					return fmt.Sprintf("%s (PID: %d)", strings.TrimSpace(string(comm)), pid)
				}
				return fmt.Sprintf("PID %d", pid)
			}
		}
	}
	return "another active process"
}

func cleanVersionOutput(raw string) string {
	line := strings.TrimSpace(strings.Split(raw, "\n")[0])
	line = strings.TrimPrefix(line, "nginx version: ")
	return line
}
