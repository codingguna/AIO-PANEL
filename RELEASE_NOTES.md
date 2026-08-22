## 🚀 Welcome to AIO-PANEL v1.0.0

**AIO-PANEL** is a modern, single self-contained server management program. It runs directly on any Linux server, listens on `SERVER_IP:5555`, and provides a complete web UI and CLI for non-invasive server administration.

---

### 📦 Quick Start (1-Line Installation)

Run the following on your Linux server (Ubuntu, Debian, CentOS, AlmaLinux, Rocky):
```bash
# Download and extract the latest release archive
curl -fsSL https://github.com/codingguna/AIO-PANEL/releases/latest/download/aio-panel-linux-amd64.tar.gz -o aio.tar.gz
tar -xzf aio.tar.gz

# Install as an autostart systemd service
sudo ./install.sh

# Create your primary administrative superuser
sudo aio createsuperuser
```

Then open your browser and navigate to: **`http://YOUR_SERVER_IP:5555`**

---

### ✨ Key Features in this Release:

- 🛡️ **Non-Invasive Auto-Discovery**: Discovers and manages existing web servers (Nginx/Apache), databases (PostgreSQL/MySQL/MariaDB/Redis), runtimes (Node.js/Python/PHP), and systemd services without modifying or overwriting your existing configurations.
- 🎨 **Webuzo-Style Clean UI**: Crisp sysadmin dashboard with fixed left-sidebar navigation and full mobile responsiveness (collapsible slide-out drawer on smartphones and tablets).
- 🛍️ **AIO Software & App Store**: 1-Click Package Installer with version selection (e.g. Node 20 LTS vs 22, PHP 8.3 vs 8.2) and live streaming terminal logs for 20+ runtimes, web servers, databases, and DevOps tools.
- 🐘 **Complete Database Engine Suite**: Full CRUD operations, live size metrics, role inspections, and SQL dump backups for **PostgreSQL**, **MySQL**, **MariaDB**, and in-memory caches (**Redis**, **Memcached**).
- 📊 **Kernel-Level Real-Time Telemetry**: 2-second live streaming of CPU deltas (`/proc/stat`), exact RAM & swap byte metrics (`/proc/meminfo`), system load averages (`/proc/loadavg`), and disk storage (`statfs`).
- 🔐 **Bcrypt Authentication & Superuser CLI**: Multi-user administrative access secured by Bcrypt (cost 12), cryptographic session tokens, anti-brute-force rate limiting, and terminal user management (`aio admin create`, `aio admin list`, `aio admin reset-password`).
- 🛠️ **Full Operations Suite**: In-browser File Manager with built-in code editor, interactive Web Terminal, real-time System Log Explorer (`journalctl` & `/var/log`), Scheduled Cron Manager, and Docker container lifecycle controls.
- 🪶 **Single Binary Zero-Dependency**: Ships with the complete React SPA embedded directly inside the Go binary (`//go:embed`). No external Node.js, Python, or database servers required.

---

### 📁 Release Artifacts & Checksums

- **x86_64 (Intel / AMD)**: `aio-panel-linux-amd64.tar.gz`
- **ARM64 (AWS Graviton / Raspberry Pi / Oracle ARM)**: `aio-panel-linux-arm64.tar.gz`
