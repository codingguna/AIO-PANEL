package webserver

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// CertificateInfo represents TLS certificate metadata
type CertificateInfo struct {
	Domain        string    `json:"domain"`
	Issuer        string    `json:"issuer"`
	ValidFrom     time.Time `json:"valid_from"`
	ValidTo       time.Time `json:"valid_to"`
	DaysRemaining int       `json:"days_remaining"`
	AutoRenew     bool      `json:"auto_renew"`
}

// IssueCertbotCertificate runs certbot to obtain a Let's Encrypt TLS certificate
func IssueCertbotCertificate(ctx context.Context, domain, email string) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	if runtime.GOOS != "linux" {
		return fmt.Errorf("Certbot is only available on Linux")
	}

	if _, err := exec.LookPath("certbot"); err != nil {
		return fmt.Errorf("certbot is not installed; please install certbot to automate SSL")
	}

	args := []string{
		"--nginx",
		"-d", domain,
		"--non-interactive",
		"--agree-tos",
		"--redirect",
	}

	if email != "" {
		args = append(args, "-m", email)
	} else {
		args = append(args, "--register-unsafely-without-email")
	}

	cmd := exec.CommandContext(ctx, "certbot", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("certbot certificate issuance failed: %s (%w)", strings.TrimSpace(stderr.String()), err)
	}

	return nil
}

// RenewAllCertificates triggers Let's Encrypt certificate renewal
func RenewAllCertificates(ctx context.Context) (string, error) {
	if runtime.GOOS != "linux" {
		return "", fmt.Errorf("certbot is only available on Linux")
	}

	cmd := exec.CommandContext(ctx, "certbot", "renew", "--quiet")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("certbot renew failed: %s (%w)", string(out), err)
	}

	return "Certificate renewal check completed successfully.", nil
}

// ListCertificates discovers all installed TLS certificates in /etc/letsencrypt/live in real-time
func ListCertificates(ctx context.Context) ([]CertificateInfo, error) {
	if runtime.GOOS != "linux" {
		return []CertificateInfo{}, nil
	}

	liveDir := "/etc/letsencrypt/live"
	entries, err := os.ReadDir(liveDir)
	if err != nil {
		return []CertificateInfo{}, nil
	}

	var list []CertificateInfo
	for _, e := range entries {
		if !e.IsDir() || e.Name() == "README" {
			continue
		}

		certPath := filepath.Join(liveDir, e.Name(), "cert.pem")
		info, err := parseX509Cert(certPath, e.Name())
		if err == nil {
			list = append(list, *info)
		}
	}

	return list, nil
}

func parseX509Cert(path, domainName string) (*CertificateInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	days := int(time.Until(cert.NotAfter).Hours() / 24)

	return &CertificateInfo{
		Domain:        domainName,
		Issuer:        cert.Issuer.CommonName,
		ValidFrom:     cert.NotBefore,
		ValidTo:       cert.NotAfter,
		DaysRemaining: days,
		AutoRenew:     true,
	}, nil
}
