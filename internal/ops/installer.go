package ops

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

// StorePackage represents a software package available in the AIO App Store
type StorePackage struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Category    string   `json:"category"` // Web Server, Runtime, Database, Containers, Security, DevOps
	Description string   `json:"description"`
	Icon        string   `json:"icon"`
	Installed   bool     `json:"installed"`
	Version     string   `json:"version"`
	Versions    []string `json:"versions"`
	InstallCmd  string   `json:"install_cmd"`
}

// PackageInstaller manages package installation jobs
type PackageInstaller struct {
	mu   sync.RWMutex
	jobs map[string]*InstallJob
}

type InstallJob struct {
	PackageID string    `json:"package_id"`
	Status    string    `json:"status"` // RUNNING, SUCCESS, FAILED
	Output    string    `json:"output"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
	Error     string    `json:"error,omitempty"`
}

var globalInstaller = &PackageInstaller{
	jobs: make(map[string]*InstallJob),
}

func GetPackageInstaller() *PackageInstaller {
	return globalInstaller
}

// GetCatalog returns all available packages with their live installation status
func (pi *PackageInstaller) GetCatalog(ctx context.Context) []StorePackage {
	catalog := []StorePackage{
		// === WEB SERVERS & REVERSE PROXIES ===
		{
			ID:          "nginx",
			Name:        "Nginx Web Server",
			Category:    "Web Server",
			Description: "High-performance HTTP server, reverse proxy, and load balancer with low memory footprint.",
			Icon:        "Globe",
			Versions:    []string{"Latest Stable (1.24+)"},
			InstallCmd:  "DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -y nginx && systemctl enable nginx && systemctl start nginx",
		},
		{
			ID:          "apache2",
			Name:        "Apache2 HTTP Server",
			Category:    "Web Server",
			Description: "Robust, enterprise-tested, and modular HTTP web server powering millions of websites.",
			Icon:        "Server",
			Versions:    []string{"Latest Stable (2.4+)"},
			InstallCmd:  "DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -y apache2 && systemctl enable apache2 && systemctl start apache2",
		},
		{
			ID:          "caddy",
			Name:        "Caddy Web Server",
			Category:    "Web Server",
			Description: "Modern, secure open-source web server with automatic HTTPS via Let's Encrypt written in Go.",
			Icon:        "Globe",
			Versions:    []string{"Latest Stable (2.7+)"},
			InstallCmd:  "DEBIAN_FRONTEND=noninteractive apt-get install -y debian-keyring debian-archive-keyring apt-transport-https curl && curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg && curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list && apt-get update && apt-get install -y caddy && systemctl enable caddy && systemctl start caddy",
		},

		// === RUNTIMES & ENVIRONMENTS ===
		{
			ID:          "nodejs",
			Name:        "Node.js & NPM Runtime",
			Category:    "Runtime",
			Description: "Event-driven asynchronous JavaScript & TypeScript runtime built on Chrome's V8 engine.",
			Icon:        "Zap",
			Versions:    []string{"Node.js 20 LTS (Recommended)", "Node.js 22 Current", "Node.js 18 LTS"},
			InstallCmd:  "curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && DEBIAN_FRONTEND=noninteractive apt-get install -y nodejs",
		},
		{
			ID:          "python3",
			Name:        "Python 3, PIP & Venv",
			Category:    "Runtime",
			Description: "General-purpose programming language for backend web frameworks (Django, FastAPI, Flask) and AI.",
			Icon:        "Layers",
			Versions:    []string{"Python 3.12", "Python 3.11", "Python 3.10"},
			InstallCmd:  "DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -y python3 python3-pip python3-venv python3-dev build-essential",
		},
		{
			ID:          "php",
			Name:        "PHP-FPM & Extensions",
			Category:    "Runtime",
			Description: "FastCGI Process Manager and complete core extensions (MySQL, PgSQL, Curl, GD, MBString, ZIP, XML).",
			Icon:        "FileCode",
			Versions:    []string{"PHP 8.3", "PHP 8.2", "PHP 8.1"},
			InstallCmd:  "DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -y php-fpm php-mysql php-pgsql php-curl php-gd php-mbstring php-xml php-zip php-bcmath php-intl",
		},
		{
			ID:          "golang",
			Name:        "Go Programming Language",
			Category:    "Runtime",
			Description: "Fast, statically typed, compiled language designed for scalable concurrent server applications.",
			Icon:        "Zap",
			Versions:    []string{"Go 1.22", "Go 1.21"},
			InstallCmd:  "DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -y golang-go",
		},
		{
			ID:          "openjdk",
			Name:        "Java OpenJDK",
			Category:    "Runtime",
			Description: "Open-source implementation of the Java Platform, Standard Edition for enterprise JVM applications.",
			Icon:        "Layers",
			Versions:    []string{"OpenJDK 21 (LTS)", "OpenJDK 17 (LTS)"},
			InstallCmd:  "DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -y openjdk-21-jdk",
		},
		{
			ID:          "ruby",
			Name:        "Ruby & Bundler",
			Category:    "Runtime",
			Description: "Dynamic, open-source programming language with a focus on simplicity and productivity (Rails).",
			Icon:        "FileCode",
			Versions:    []string{"Latest Stable"},
			InstallCmd:  "DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -y ruby-full build-essential && gem install bundler",
		},

		// === DATABASES & IN-MEMORY CACHE ===
		{
			ID:          "postgresql",
			Name:        "PostgreSQL Database",
			Category:    "Database",
			Description: "Enterprise-class open-source relational SQL database with JSONB, extensions, and spatial support.",
			Icon:        "Database",
			Versions:    []string{"PostgreSQL 16", "PostgreSQL 15"},
			InstallCmd:  "DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -y postgresql postgresql-contrib && systemctl enable postgresql && systemctl start postgresql",
		},
		{
			ID:          "mysql",
			Name:        "MySQL Server",
			Category:    "Database",
			Description: "The world's most popular open-source relational database management system (RDBMS).",
			Icon:        "Database",
			Versions:    []string{"MySQL 8.0"},
			InstallCmd:  "DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -y mysql-server && systemctl enable mysql && systemctl start mysql",
		},
		{
			ID:          "mariadb",
			Name:        "MariaDB Server",
			Category:    "Database",
			Description: "High-performance drop-in replacement for MySQL created by original MySQL developers.",
			Icon:        "Database",
			Versions:    []string{"MariaDB 10.11 LTS"},
			InstallCmd:  "DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -y mariadb-server && systemctl enable mariadb && systemctl start mariadb",
		},
		{
			ID:          "redis",
			Name:        "Redis In-Memory Cache",
			Category:    "Database",
			Description: "Ultra-fast in-memory data structure store used as a database, cache, message broker, and queue.",
			Icon:        "Activity",
			Versions:    []string{"Latest Stable (7.x)"},
			InstallCmd:  "DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -y redis-server && systemctl enable redis-server && systemctl start redis-server",
		},
		{
			ID:          "memcached",
			Name:        "Memcached Distributed Cache",
			Category:    "Database",
			Description: "High-performance, distributed memory object caching system for speeding up dynamic web applications.",
			Icon:        "Activity",
			Versions:    []string{"Latest Stable"},
			InstallCmd:  "DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -y memcached && systemctl enable memcached && systemctl start memcached",
		},

		// === DEVOPS & PROCESS MANAGERS ===
		{
			ID:          "docker",
			Name:        "Docker Engine & Compose",
			Category:    "Containers",
			Description: "Enterprise containerization platform to build, ship, and run applications in isolated environments.",
			Icon:        "Box",
			Versions:    []string{"Docker Community Edition (26.x)"},
			InstallCmd:  "DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -y docker.io docker-compose-plugin && systemctl enable docker && systemctl start docker",
		},
		{
			ID:          "pm2",
			Name:        "PM2 Process Manager",
			Category:    "DevOps",
			Description: "Production process manager for Node.js applications with built-in load balancer and autostart.",
			Icon:        "Activity",
			Versions:    []string{"Latest (5.x)"},
			InstallCmd:  "npm install -g pm2 && pm2 startup",
		},
		{
			ID:          "supervisor",
			Name:        "Supervisor Process Control",
			Category:    "DevOps",
			Description: "Client/server system that allows users to monitor and control a number of processes on UNIX-like OS.",
			Icon:        "Server",
			Versions:    []string{"Latest Stable"},
			InstallCmd:  "DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -y supervisor && systemctl enable supervisor && systemctl start supervisor",
		},
		{
			ID:          "composer",
			Name:        "Composer (PHP Dependency Manager)",
			Category:    "DevOps",
			Description: "Standard package and dependency manager for PHP software and modern frameworks like Laravel.",
			Icon:        "Layers",
			Versions:    []string{"Composer 2.x"},
			InstallCmd:  "curl -sS https://getcomposer.org/installer | php -- --install-dir=/usr/local/bin --filename=composer",
		},
		{
			ID:          "git",
			Name:        "Git & Build Essentials",
			Category:    "DevOps",
			Description: "Distributed version control system and compilation toolchain (`gcc`, `g++`, `make`, `curl`, `wget`).",
			Icon:        "GitBranch",
			Versions:    []string{"Latest Stable"},
			InstallCmd:  "DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -y git build-essential curl wget unzip",
		},

		// === SECURITY & SYSTEM TOOLS ===
		{
			ID:          "certbot",
			Name:        "Certbot (Let's Encrypt SSL)",
			Category:    "Security",
			Description: "Automated tool to fetch and automatically renew free Let's Encrypt SSL/TLS certificates.",
			Icon:        "Shield",
			Versions:    []string{"Latest"},
			InstallCmd:  "DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -y certbot python3-certbot-nginx python3-certbot-apache",
		},
		{
			ID:          "ufw",
			Name:        "UFW Firewall",
			Category:    "Security",
			Description: "Uncomplicated Firewall program for managing a netfilter/iptables firewall on Linux hosts.",
			Icon:        "Shield",
			Versions:    []string{"Latest"},
			InstallCmd:  "DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -y ufw",
		},
		{
			ID:          "fail2ban",
			Name:        "Fail2ban Intrusion Prevention",
			Category:    "Security",
			Description: "Scans log files and bans IPs that show malicious signs like too many password failures.",
			Icon:        "Shield",
			Versions:    []string{"Latest"},
			InstallCmd:  "DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -y fail2ban && systemctl enable fail2ban && systemctl start fail2ban",
		},
		{
			ID:          "htop",
			Name:        "Htop & Sysstat Monitor",
			Category:    "DevOps",
			Description: "Interactive, real-time process viewer and system activity data collection utilities.",
			Icon:        "Activity",
			Versions:    []string{"Latest"},
			InstallCmd:  "DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -y htop sysstat iftop iotop",
		},
	}

	// Inspect real-time status of each package
	for i := range catalog {
		pkg := &catalog[i]
		installed, ver := checkPackageStatus(ctx, pkg.ID)
		pkg.Installed = installed
		pkg.Version = ver
	}

	return catalog
}

// InstallPackage initiates a package installation job
func (pi *PackageInstaller) InstallPackage(ctx context.Context, pkgID, version string) (*InstallJob, error) {
	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf("package installation is only supported on Linux")
	}

	catalog := pi.GetCatalog(ctx)
	var target *StorePackage
	for _, p := range catalog {
		if p.ID == pkgID {
			target = &p
			break
		}
	}

	if target == nil {
		return nil, fmt.Errorf("package not found in store: %s", pkgID)
	}

	cmdStr := target.InstallCmd
	if pkgID == "nodejs" && strings.Contains(version, "18") {
		cmdStr = "curl -fsSL https://deb.nodesource.com/setup_18.x | bash - && DEBIAN_FRONTEND=noninteractive apt-get install -y nodejs"
	} else if pkgID == "nodejs" && strings.Contains(version, "22") {
		cmdStr = "curl -fsSL https://deb.nodesource.com/setup_22.x | bash - && DEBIAN_FRONTEND=noninteractive apt-get install -y nodejs"
	}

	job := &InstallJob{
		PackageID: pkgID,
		Status:    "RUNNING",
		Output:    fmt.Sprintf("🚀 Starting installation of %s (%s)...\n", target.Name, version),
		StartedAt: time.Now(),
	}

	pi.mu.Lock()
	pi.jobs[pkgID] = job
	pi.mu.Unlock()

	// Run installation asynchronously
	go func() {
		cmd := exec.Command("bash", "-c", cmdStr)
		var outBuf, errBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf

		err := cmd.Run()

		pi.mu.Lock()
		defer pi.mu.Unlock()

		job.EndedAt = time.Now()
		job.Output += outBuf.String() + "\n" + errBuf.String()

		if err != nil {
			job.Status = "FAILED"
			job.Error = err.Error()
			job.Output += fmt.Sprintf("\n❌ Installation failed: %v\n", err)
		} else {
			job.Status = "SUCCESS"
			job.Output += fmt.Sprintf("\n✨ %s successfully installed!\n", target.Name)
		}
	}()

	return job, nil
}

// GetJobStatus retrieves current status of an installation job
func (pi *PackageInstaller) GetJobStatus(pkgID string) *InstallJob {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	return pi.jobs[pkgID]
}

func checkPackageStatus(ctx context.Context, pkgID string) (bool, string) {
	var cmdName string
	var args []string

	switch pkgID {
	case "nginx":
		cmdName = "nginx"
		args = []string{"-v"}
	case "apache2":
		cmdName = "apache2"
		args = []string{"-v"}
	case "caddy":
		cmdName = "caddy"
		args = []string{"version"}
	case "nodejs":
		cmdName = "node"
		args = []string{"-v"}
	case "python3":
		cmdName = "python3"
		args = []string{"--version"}
	case "php":
		cmdName = "php"
		args = []string{"-v"}
	case "golang":
		cmdName = "go"
		args = []string{"version"}
	case "openjdk":
		cmdName = "java"
		args = []string{"-version"}
	case "ruby":
		cmdName = "ruby"
		args = []string{"-v"}
	case "postgresql":
		cmdName = "psql"
		args = []string{"--version"}
	case "mysql":
		cmdName = "mysql"
		args = []string{"--version"}
	case "mariadb":
		cmdName = "mariadb"
		args = []string{"--version"}
	case "redis":
		cmdName = "redis-server"
		args = []string{"--version"}
	case "memcached":
		cmdName = "memcached"
		args = []string{"-h"}
	case "docker":
		cmdName = "docker"
		args = []string{"-v"}
	case "pm2":
		cmdName = "pm2"
		args = []string{"-v"}
	case "supervisor":
		cmdName = "supervisord"
		args = []string{"-v"}
	case "composer":
		cmdName = "composer"
		args = []string{"--version"}
	case "certbot":
		cmdName = "certbot"
		args = []string{"--version"}
	case "ufw":
		cmdName = "ufw"
		args = []string{"version"}
	case "fail2ban":
		cmdName = "fail2ban-client"
		args = []string{"version"}
	case "htop":
		cmdName = "htop"
		args = []string{"--version"}
	case "git":
		cmdName = "git"
		args = []string{"--version"}
	default:
		return false, ""
	}

	cmd := exec.CommandContext(ctx, cmdName, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, ""
	}

	firstLine := strings.TrimSpace(strings.Split(string(out), "\n")[0])
	return true, firstLine
}
