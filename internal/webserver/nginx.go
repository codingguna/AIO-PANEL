package webserver

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// VHostType defines whether the site is a reverse proxy, static file server, or PHP
type VHostType string

const (
	VHostTypeReverseProxy VHostType = "reverse_proxy"
	VHostTypeStatic       VHostType = "static"
	VHostTypePHP          VHostType = "php"
)

// VHostConfig represents an AIO-managed virtual host
type VHostConfig struct {
	Domain       string    `json:"domain"`
	Aliases      []string  `json:"aliases,omitempty"`
	Type         VHostType `json:"type"`          // reverse_proxy, static, php
	ProxyPass    string    `json:"proxy_pass"`    // e.g. http://127.0.0.1:8000
	DocumentRoot string    `json:"document_root"` // e.g. /var/www/site/dist
	PHPSocket    string    `json:"php_socket"`    // e.g. unix:/var/run/php/php8.2-fpm.sock
	SSL          bool      `json:"ssl"`
	SSLCertPath  string    `json:"ssl_cert_path,omitempty"`
	SSLKeyPath   string    `json:"ssl_key_path,omitempty"`
	OwnerType    string    `json:"owner_type"` // 'aio'
	Enabled      bool      `json:"enabled"`
}

var validDomainRegex = regexp.MustCompile(`^[a-zA-Z0-9.\-]+$`)

// GenerateNginxConfig renders a production-grade Nginx vhost template
func GenerateNginxConfig(cfg VHostConfig) (string, error) {
	if !validDomainRegex.MatchString(cfg.Domain) {
		return "", fmt.Errorf("invalid domain name: %s", cfg.Domain)
	}

	serverNames := cfg.Domain
	if len(cfg.Aliases) > 0 {
		serverNames += " " + strings.Join(cfg.Aliases, " ")
	}

	var sb strings.Builder
	sb.WriteString("# ==============================================================================\n")
	sb.WriteString(fmt.Sprintf("# AIO-PANEL Managed Virtual Host: %s\n", cfg.Domain))
	sb.WriteString(fmt.Sprintf("# Generated at: %s\n", time.Now().UTC().Format(time.RFC3339)))
	sb.WriteString("# Philosophy: Safe, isolated, non-invasive\n")
	sb.WriteString("# ==============================================================================\n\n")

	sb.WriteString("server {\n")
	sb.WriteString("    listen 80;\n")
	sb.WriteString("    listen [::]:80;\n")
	sb.WriteString(fmt.Sprintf("    server_name %s;\n\n", serverNames))

	sb.WriteString("    # Security Headers\n")
	sb.WriteString("    add_header X-Frame-Options \"SAMEORIGIN\" always;\n")
	sb.WriteString("    add_header X-Content-Type-Options \"nosniff\" always;\n")
	sb.WriteString("    add_header X-XSS-Protection \"1; mode=block\" always;\n")
	sb.WriteString("    add_header Referrer-Policy \"strict-origin-when-cross-origin\" always;\n\n")

	sb.WriteString("    # Gzip Compression\n")
	sb.WriteString("    gzip on;\n")
	sb.WriteString("    gzip_types text/plain text/css application/json application/javascript text/xml application/xml;\n\n")

	switch cfg.Type {
	case VHostTypeReverseProxy:
		if cfg.ProxyPass == "" {
			cfg.ProxyPass = "http://127.0.0.1:8000"
		}
		sb.WriteString("    location / {\n")
		sb.WriteString(fmt.Sprintf("        proxy_pass %s;\n", cfg.ProxyPass))
		sb.WriteString("        proxy_http_version 1.1;\n")
		sb.WriteString("        proxy_set_header Upgrade $http_upgrade;\n")
		sb.WriteString("        proxy_set_header Connection \"upgrade\";\n")
		sb.WriteString("        proxy_set_header Host $host;\n")
		sb.WriteString("        proxy_set_header X-Real-IP $remote_addr;\n")
		sb.WriteString("        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n")
		sb.WriteString("        proxy_set_header X-Forwarded-Proto $scheme;\n")
		sb.WriteString("        proxy_read_timeout 90s;\n")
		sb.WriteString("    }\n")

	case VHostTypeStatic:
		if cfg.DocumentRoot == "" {
			cfg.DocumentRoot = "/var/www/" + cfg.Domain
		}
		sb.WriteString(fmt.Sprintf("    root %s;\n", cfg.DocumentRoot))
		sb.WriteString("    index index.html index.htm;\n\n")
		sb.WriteString("    location / {\n")
		sb.WriteString("        try_files $uri $uri/ /index.html;\n")
		sb.WriteString("    }\n\n")
		sb.WriteString("    location ~* \\.(jpg|jpeg|png|gif|ico|css|js|svg|woff|woff2)$ {\n")
		sb.WriteString("        expires 30d;\n")
		sb.WriteString("        add_header Cache-Control \"public, no-transform\";\n")
		sb.WriteString("    }\n")

	case VHostTypePHP:
		if cfg.DocumentRoot == "" {
			cfg.DocumentRoot = "/var/www/" + cfg.Domain
		}
		if cfg.PHPSocket == "" {
			cfg.PHPSocket = "unix:/var/run/php/php-fpm.sock"
		}
		sb.WriteString(fmt.Sprintf("    root %s;\n", cfg.DocumentRoot))
		sb.WriteString("    index index.php index.html;\n\n")
		sb.WriteString("    location / {\n")
		sb.WriteString("        try_files $uri $uri/ /index.php?$query_string;\n")
		sb.WriteString("    }\n\n")
		sb.WriteString("    location ~ \\.php$ {\n")
		sb.WriteString("        include snippets/fastcgi-php.conf;\n")
		sb.WriteString(fmt.Sprintf("        fastcgi_pass %s;\n", cfg.PHPSocket))
		sb.WriteString("    }\n")
	}

	sb.WriteString("}\n")
	return sb.String(), nil
}

// CreateVHost creates and validates an AIO-managed Nginx virtual host
func CreateVHost(ctx context.Context, cfg VHostConfig, backupDir string) error {
	content, err := GenerateNginxConfig(cfg)
	if err != nil {
		return err
	}

	if runtime.GOOS != "linux" {
		return fmt.Errorf("Nginx management requires Linux environment")
	}

	availFile := fmt.Sprintf("/etc/nginx/sites-available/aio_%s.conf", cfg.Domain)
	enabledFile := fmt.Sprintf("/etc/nginx/sites-enabled/aio_%s.conf", cfg.Domain)

	// 1. Write to sites-available
	if err := os.WriteFile(availFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", availFile, err)
	}

	// 2. Create symlink in sites-enabled
	_ = os.Remove(enabledFile)
	if err := os.Symlink(availFile, enabledFile); err != nil {
		_ = os.Remove(availFile)
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	// 3. Test with nginx -t
	testCmd := exec.CommandContext(ctx, "nginx", "-t")
	var stderr bytes.Buffer
	testCmd.Stderr = &stderr

	if err := testCmd.Run(); err != nil {
		// ROLLBACK
		_ = os.Remove(enabledFile)
		_ = os.Remove(availFile)
		return fmt.Errorf("nginx validation failed (nginx -t), automatically rolled back: %s (%w)",
			strings.TrimSpace(stderr.String()), err)
	}

	// 4. Reload Nginx
	reloadCmd := exec.CommandContext(ctx, "systemctl", "reload", "nginx")
	if err := reloadCmd.Run(); err != nil {
		return fmt.Errorf("failed to reload nginx: %w", err)
	}

	return nil
}

// DeleteVHost safely deletes ONLY an AIO-managed virtual host
func DeleteVHost(ctx context.Context, domain string) error {
	if !validDomainRegex.MatchString(domain) {
		return fmt.Errorf("invalid domain: %s", domain)
	}

	if runtime.GOOS != "linux" {
		return fmt.Errorf("Nginx is only available on Linux")
	}

	availFile := fmt.Sprintf("/etc/nginx/sites-available/aio_%s.conf", domain)
	enabledFile := fmt.Sprintf("/etc/nginx/sites-enabled/aio_%s.conf", domain)

	_ = os.Remove(enabledFile)
	_ = os.Remove(availFile)

	_ = exec.CommandContext(ctx, "nginx", "-t").Run()
	_ = exec.CommandContext(ctx, "systemctl", "reload", "nginx").Run()

	return nil
}

// ListAIOVHosts lists virtual hosts created specifically by AIO-PANEL in real-time
func ListAIOVHosts() ([]VHostConfig, error) {
	if runtime.GOOS != "linux" {
		return []VHostConfig{}, nil
	}

	dir := "/etc/nginx/sites-available"
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []VHostConfig{}, nil
	}

	var list []VHostConfig
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "aio_") && strings.HasSuffix(e.Name(), ".conf") {
			domain := strings.TrimSuffix(strings.TrimPrefix(e.Name(), "aio_"), ".conf")
			list = append(list, VHostConfig{
				Domain:    domain,
				OwnerType: "aio",
				Enabled:   true,
			})
		}
	}

	return list, nil
}
