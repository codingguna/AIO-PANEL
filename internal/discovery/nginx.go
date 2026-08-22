package discovery

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// NginxSite represents a discovered Nginx virtual host configuration
type NginxSite struct {
	Domain       string   `json:"domain"`
	Aliases      []string `json:"aliases"`
	ConfigFile   string   `json:"config_file"`
	DocumentRoot string   `json:"document_root"`
	ProxyPass    string   `json:"proxy_pass"`
	SSL          bool     `json:"ssl"`
	OwnerType    string   `json:"owner_type"` // 'external' or 'aio'
	Enabled      bool     `json:"enabled"`
}

var (
	serverNameRegex = regexp.MustCompile(`(?i)^\s*server_name\s+([^;]+);`)
	rootRegex       = regexp.MustCompile(`(?i)^\s*root\s+([^;]+);`)
	proxyPassRegex  = regexp.MustCompile(`(?i)^\s*proxy_pass\s+([^;]+);`)
	sslCertRegex    = regexp.MustCompile(`(?i)^\s*ssl_certificate\s+([^;]+);`)
)

// DiscoverNginxSites scans existing Nginx virtual hosts in read-only mode
func DiscoverNginxSites(ctx context.Context) ([]NginxSite, error) {
	if runtime.GOOS != "linux" {
		return []NginxSite{}, nil
	}

	var sites []NginxSite

	searchDirs := []string{
		"/etc/nginx/sites-enabled",
		"/etc/nginx/conf.d",
	}

	for _, dir := range searchDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, e := range entries {
			if e.IsDir() {
				continue
			}

			filePath := filepath.Join(dir, e.Name())
			parsedSites := parseNginxFile(filePath)
			sites = append(sites, parsedSites...)
		}
	}

	return sites, nil
}

func parseNginxFile(filePath string) []NginxSite {
	var sites []NginxSite
	f, err := os.Open(filePath)
	if err != nil {
		return sites
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var currentSite *NginxSite

	owner := "external"
	if strings.Contains(filePath, "aio_") {
		owner = "aio"
	}

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "server {") {
			if currentSite != nil && currentSite.Domain != "" {
				sites = append(sites, *currentSite)
			}
			currentSite = &NginxSite{
				ConfigFile: filePath,
				OwnerType:  owner,
				Enabled:    true,
				Aliases:    make([]string, 0),
			}
		}

		if currentSite == nil {
			continue
		}

		if match := serverNameRegex.FindStringSubmatch(line); len(match) > 1 {
			names := strings.Fields(match[1])
			if len(names) > 0 {
				currentSite.Domain = names[0]
				if len(names) > 1 {
					currentSite.Aliases = names[1:]
				}
			}
		}

		if match := rootRegex.FindStringSubmatch(line); len(match) > 1 {
			currentSite.DocumentRoot = strings.TrimSpace(match[1])
		}

		if match := proxyPassRegex.FindStringSubmatch(line); len(match) > 1 {
			currentSite.ProxyPass = strings.TrimSpace(match[1])
		}

		if sslCertRegex.MatchString(line) {
			currentSite.SSL = true
		}
	}

	if currentSite != nil && currentSite.Domain != "" {
		sites = append(sites, *currentSite)
	}

	return sites
}
