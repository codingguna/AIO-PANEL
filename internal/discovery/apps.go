package discovery

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// DiscoveredApp represents an application detected on the host
type DiscoveredApp struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // Django, FastAPI, Flask, Node.js, Next.js, React SPA, Laravel, PHP, Static
	Path        string `json:"path"`
	Runtime     string `json:"runtime"`     // Python 3.x, Node 20.x, etc.
	Service     string `json:"service"`     // e.g. memotrack.service
	NginxDomain string `json:"nginx_domain"`// e.g. app.memotrack.net
	OwnerType   string `json:"owner_type"`  // 'external' or 'aio'
	Status      string `json:"status"`      // 'running' or 'stopped'
}

// DiscoverApplications scans standard application directories in read-only mode
func DiscoverApplications(ctx context.Context) ([]DiscoveredApp, error) {
	var apps []DiscoveredApp

	searchDirs := []string{}
	if runtime.GOOS == "linux" {
		searchDirs = append(searchDirs, "/var/www", "/srv")
		if homeEntries, err := os.ReadDir("/home"); err == nil {
			for _, e := range homeEntries {
				if e.IsDir() {
					searchDirs = append(searchDirs, filepath.Join("/home", e.Name()))
				}
			}
		}
	} else {
		// On non-Linux, scan current directory subfolders if any
		if cur, err := os.Getwd(); err == nil {
			searchDirs = append(searchDirs, cur)
		}
	}

	for _, baseDir := range searchDirs {
		entries, err := os.ReadDir(baseDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			appPath := filepath.Join(baseDir, entry.Name())
			app, detected := inspectAppDirectory(appPath, entry.Name())
			if detected {
				apps = append(apps, *app)
			}
		}
	}

	return apps, nil
}

func inspectAppDirectory(path, name string) (*DiscoveredApp, bool) {
	app := &DiscoveredApp{
		Name:      name,
		Path:      path,
		OwnerType: "external",
		Status:    "detected",
	}

	// 1. Check Python Django
	if fileExists(filepath.Join(path, "manage.py")) {
		app.Type = "Django"
		app.Runtime = "Python"
		if fileExists(filepath.Join(path, "venv")) || fileExists(filepath.Join(path, ".venv")) {
			app.Runtime = "Python (Virtualenv)"
		}
		app.Service = findProbableService(name)
		return app, true
	}

	// 2. Check Python FastAPI / Flask
	if fileExists(filepath.Join(path, "main.py")) || fileExists(filepath.Join(path, "app.py")) || fileExists(filepath.Join(path, "wsgi.py")) {
		app.Type = "Python WSGI/ASGI"
		app.Runtime = "Python"
		app.Service = findProbableService(name)
		return app, true
	}

	// 3. Check Node.js / Next.js
	if fileExists(filepath.Join(path, "package.json")) {
		if fileExists(filepath.Join(path, "next.config.js")) || fileExists(filepath.Join(path, "next.config.mjs")) {
			app.Type = "Next.js"
		} else {
			app.Type = "Node.js"
		}
		app.Runtime = "Node.js"
		app.Service = findProbableService(name)
		return app, true
	}

	// 4. Check React / Vue SPA built bundle
	if fileExists(filepath.Join(path, "index.html")) && (fileExists(filepath.Join(path, "assets")) || fileExists(filepath.Join(path, "static")) || fileExists(filepath.Join(path, "dist"))) {
		app.Type = "Frontend SPA (React/Vue)"
		app.Runtime = "Static / Nginx"
		app.Service = "nginx.service"
		return app, true
	}

	// 5. Check PHP / Laravel
	if fileExists(filepath.Join(path, "artisan")) {
		app.Type = "Laravel"
		app.Runtime = "PHP-FPM"
		app.Service = "php-fpm.service"
		return app, true
	} else if fileExists(filepath.Join(path, "index.php")) {
		app.Type = "PHP Application"
		app.Runtime = "PHP-FPM"
		app.Service = "php-fpm.service"
		return app, true
	}

	return nil, false
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func findProbableService(appName string) string {
	cleanName := strings.ToLower(appName)
	possibleUnits := []string{
		cleanName + ".service",
		cleanName + "-backend.service",
		cleanName + "-web.service",
		cleanName + "-api.service",
	}

	for _, u := range possibleUnits {
		if _, err := os.Stat(filepath.Join("/etc/systemd/system", u)); err == nil {
			return u
		}
		if _, err := os.Stat(filepath.Join("/lib/systemd/system", u)); err == nil {
			return u
		}
	}

	return cleanName + ".service"
}
