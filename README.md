# AIO-PANEL
All In One Panel

**AIO-PANEL** is a lightweight, powerful, and security-first **All-In-One Linux Server Management Panel**.

It is designed to run directly on the Linux server it manages and provide a complete server-management interface through a web browser and CLI — without requiring a heavy application stack or an external database server.

> **Manage your entire Linux server from one lightweight panel.**

---

## 🚀 Project Vision

AIO-PANEL is designed for administrators and developers who already have applications, services, databases, web servers, and custom configurations running on their Linux servers.

Unlike traditional hosting panels that may assume control over the server software stack, AIO-PANEL follows a different philosophy:

> **AIO-PANEL manages the server; it does not take ownership of the server.**

If a server already has:

* Nginx
* Apache
* Python
* Node.js
* npm
* Git
* PostgreSQL
* MySQL/MariaDB
* Redis
* Docker
* Gunicorn
* Custom systemd services
* Django applications
* Node applications
* React applications
* PHP applications
* Existing websites and domains

AIO-PANEL should detect and manage them without unnecessarily replacing, reinstalling, or modifying them.

---

# ✨ Core Principles

### Lightweight

AIO-PANEL is designed to consume minimal CPU, RAM, disk, and background resources.

The server may already be running many applications and services. AIO-PANEL should add only a small resource footprint.

### Powerful

AIO-PANEL provides comprehensive management of the Linux server:

* System
* Users
* SSH
* Services
* Domains
* SSL
* Applications
* Databases
* Files
* Logs
* Deployments
* Cron
* Backups
* Firewall
* Monitoring
* Docker
* Terminal
* Security

### Security First

AIO-PANEL has privileged access to the server, so security is a primary architectural requirement.

Authentication, authorization, privilege separation, secure configuration handling, audit logging, validation, rollback, and secure updates are core requirements.

### Non-Invasive

Installing AIO-PANEL must not unnecessarily modify existing applications or services.

AIO-PANEL should discover the current server state first.

> **Discovery is read-only. Management is explicit.**

### Reliable

If AIO-PANEL stops or crashes:

* Nginx must continue running.
* PostgreSQL must continue running.
* MySQL must continue running.
* Docker containers must continue running.
* Redis must continue running.
* Existing applications must continue running.
* SSH must remain available.

AIO-PANEL must never become a dependency of the applications it manages.

### Self-Contained

AIO-PANEL should run as a single primary system service.

It should not require users to install a large stack such as:

* PostgreSQL
* Redis
* Node.js
* Python
* Nginx

just to run the panel itself.

---

# 🖥️ Access

The default web interface is exposed on:

```text
http://SERVER_IP:5555
```

Production deployments should support HTTPS:

```text
https://SERVER_IP:5555
```

The port should be configurable.

AIO-PANEL also provides a local CLI:

```bash
aio status
```

The browser and CLI should use the same internal AIO Core.

---

# 🏗️ High-Level Architecture

```text
                         ┌───────────────────┐
                         │      Browser      │
                         └─────────┬─────────┘
                                   │
                              HTTPS :5555
                                   │
                         ┌─────────▼─────────┐
                         │     AIO WEB       │
                         └─────────┬─────────┘
                                   │
                         ┌─────────▼─────────┐
                         │     AIO CORE      │
                         │                   │
                         │ Authentication    │
                         │ Authorization     │
                         │ Configuration     │
                         │ Security          │
                         │ Management        │
                         │ Modules           │
                         └─────────┬─────────┘
                                   │
             ┌─────────────────────┼─────────────────────┐
             │                     │                     │
             ▼                     ▼                     ▼
       System Manager       Service Manager       Security Manager
             │                     │                     │
             └─────────────────────┼─────────────────────┘
                                   │
                                   ▼
                              Linux Server


                         ┌───────────────────┐
                         │      AIO CLI      │
                         │    `aio ...`      │
                         └─────────┬─────────┘
                                   │
                                   ▼
                              AIO CORE
```

---

# 📦 Main Modules

## 🖥️ Dashboard

Provides a complete overview of the server.

### System information

* Hostname
* Operating system
* Kernel
* Architecture
* CPU
* CPU cores
* RAM
* Swap
* Disk
* Network interfaces
* IP addresses
* Uptime
* Load average
* Boot time

### Live resource monitoring

* CPU usage
* Memory usage
* Swap usage
* Disk usage
* Network traffic
* Load average
* Processes
* Service health

---

# 👤 Users

Manage Linux users and groups.

Features:

* Create users
* Delete users
* Lock users
* Unlock users
* Change passwords
* Change shells
* Home directories
* Groups
* UID/GID
* Sudo permissions
* User sessions

Linux remains the authoritative source for Linux user information.

---

# 🔑 SSH

Manage OpenSSH configuration and access.

Features:

* SSH status
* SSH port
* SSH configuration
* SSH users
* SSH keys
* Generate SSH keys
* Add authorized keys
* Remove authorized keys
* Password authentication configuration
* Root login configuration
* Active SSH sessions
* SSH logs

Dangerous SSH changes must use configuration backup, validation, health checks, and rollback.

---

# 🌐 Domains

Manage web domains and subdomains.

Features:

* Domains
* Subdomains
* Virtual hosts
* Document roots
* Redirects
* Domain status
* Nginx configuration
* Apache configuration where supported

Existing configurations should be detected before AIO attempts to manage them.

---

# 🔒 SSL

Manage TLS certificates.

Features:

* Let's Encrypt
* Certificate creation
* Certificate installation
* Certificate renewal
* Renewal status
* Expiration monitoring
* Wildcard certificates
* Certificate listing
* HTTPS configuration

Private keys must be protected with strict filesystem permissions and must never appear in logs or normal API/UI responses.

---

# 🚀 Applications

Manage applications running on the server.

Supported application categories include:

### Python

* Django
* Flask
* FastAPI
* WSGI applications
* ASGI applications

### Node.js

* Node.js
* Express
* Next.js
* Other Node applications

### Frontend

* React
* Vue
* Angular
* Static websites

### PHP

* PHP
* PHP-FPM
* Laravel
* WordPress
* Other PHP applications

Application management includes:

* Application path
* Domain
* Runtime
* Environment variables
* Virtual environments
* Build commands
* Start commands
* Process manager
* Logs
* Restart
* Stop/start
* Health status

---

# ⚙️ Services

Manage Linux services through the native service manager.

Examples:

* Nginx
* Apache
* PostgreSQL
* MySQL
* Redis
* Docker
* SSH
* Gunicorn
* Custom systemd services

Operations:

* Status
* Start
* Stop
* Restart
* Reload
* Enable
* Disable
* Logs

AIO-PANEL should discover existing services instead of assuming a fixed service list.

---

# 🗄️ Databases

## PostgreSQL

* Servers
* Databases
* Users
* Roles
* Schemas
* Permissions
* Connections
* Backup
* Restore

## MySQL/MariaDB

* Databases
* Users
* Privileges
* Backup
* Restore

Future database support may include:

* MongoDB
* Redis
* SQLite

AIO-PANEL itself does **not** require PostgreSQL or MySQL.

---

# 📁 File Manager

Manage server files through the web interface.

Features:

* Browse
* Upload
* Download
* Create file
* Edit
* Rename
* Move
* Copy
* Delete
* Create directory
* Archive
* Extract
* Permissions
* Ownership
* Group

Security protections must include:

* Path traversal protection
* Symlink escape protection
* Unauthorized filesystem access prevention
* Permission validation
* Safe file operations

---

# 📜 Logs

Centralized access to existing server logs.

Supported sources may include:

* Nginx access logs
* Nginx error logs
* Gunicorn
* Django
* systemd
* SSH
* PostgreSQL
* MySQL
* Docker
* Application logs
* AIO logs

System services should preferably use native Linux logging facilities such as `journald` rather than duplicating all logs into AIO storage.

---

# 🔄 Deployments

AIO-PANEL can manage application deployments.

Typical workflow:

```text
Git repository
      ↓
Git pull
      ↓
Install dependencies
      ↓
Build
      ↓
Database migration
      ↓
Static collection
      ↓
Configuration validation
      ↓
Restart service
      ↓
Health check
```

Supported operations may include:

* Git repositories
* Branch selection
* Pull
* Build
* `npm install`
* `npm build`
* `pip install`
* Virtual environments
* Django migrations
* `collectstatic`
* Service restart
* Deployment history

---

# ⏰ Cron

Manage scheduled tasks.

Features:

* System cron
* User cron
* `/etc/cron.d`
* Create jobs
* Edit jobs
* Delete jobs
* Enable/disable
* Execution history

AIO-PANEL should use the existing Linux scheduler rather than creating another unnecessary scheduler service.

---

# 💾 Backups

Backup management includes:

### Files

* Application files
* Media
* Uploads
* Configuration

### Databases

* PostgreSQL
* MySQL/MariaDB

### Server configuration

* Nginx
* SSH
* Firewall
* AIO configuration
* Application configuration

### Destinations

Initially:

* Local storage

Future:

* S3-compatible storage
* Cloud storage
* Other remote storage providers

### Restore

```text
Select backup
      ↓
Validate
      ↓
Preview
      ↓
Confirm
      ↓
Restore
      ↓
Verify
```

---

# 🔥 Firewall

Manage the server firewall.

Initial target:

* UFW

Potential future support:

* nftables
* iptables where necessary

Features:

* Firewall status
* Allow port
* Deny port
* Delete rule
* IP allowlist
* IP blocklist
* Protocol
* Port ranges
* Rule management

Firewall changes must have safeguards to prevent administrator lockout.

---

# 📊 Monitoring

Monitor:

* CPU
* RAM
* Swap
* Disk
* Network
* Load
* Processes
* Services
* Disk health where supported

AIO-PANEL should avoid deploying a heavyweight monitoring stack merely to display basic server statistics.

Monitoring should use efficient native Linux interfaces wherever possible.

---

# 🐳 Docker

If Docker is already installed, AIO can manage it.

Features:

* Containers
* Images
* Volumes
* Networks
* Logs
* Start
* Stop
* Restart
* Remove
* Inspect
* Docker Compose

AIO should detect Docker rather than automatically installing it.

---

# 💻 Terminal

AIO may provide a browser-based terminal.

Because this provides direct server-level access, it is considered a high-risk feature.

Requirements include:

* Strong authentication
* MFA/re-authentication
* Short-lived sessions
* Secure WebSocket handling
* PTY management
* Session termination
* Audit logging

---

# 🛡️ Security

Security is a first-class AIO-PANEL module.

Potential security checks:

```text
HTTPS                         ✓
MFA                           ✓
Firewall                      ✓
SSH key authentication       ✓
Password SSH authentication  ⚠
Root SSH login               ⚠
Public database ports        ⚠
AIO updates                  ✓
Secure permissions           ✓
```

AIO should detect insecure configurations and report them.

It should not automatically make dangerous security changes without explicit administrator authorization.

---

# 🧾 Audit & History

AIO will use **SQLite** for its own historical data.

Proposed database:

```text
/var/lib/aio/aio.db
```

Potential tables:

* `audit_events`
* `login_history`
* `security_events`
* `deployment_history`
* `backup_history`
* `restore_history`
* `configuration_history`
* `terminal_sessions`
* `system_events`

Example audit event:

```text
Timestamp:
2026-08-22 15:00:00

User:
admin

Action:
Restart PostgreSQL

Target:
postgresql.service

Result:
SUCCESS
```

AIO should not use SQLite as a duplicate source of truth for Linux state.

For example:

* Linux users → Linux
* Services → systemd
* Firewall → firewall subsystem
* Nginx → Nginx configuration
* PostgreSQL → PostgreSQL
* Docker → Docker

SQLite stores **AIO-specific state, history, audit information, and metadata**.

---

# 📂 AIO Filesystem

A proposed Linux layout:

```text
/etc/aio/
├── aio.conf
├── security/
├── modules/
└── certificates/

/var/lib/aio/
├── aio.db
├── state/
├── audit/
├── history/
├── backups/
└── cache/

/var/log/aio/
├── aio.log
├── security.log
├── audit.log
└── error.log

/opt/aio/
└── aio

/etc/systemd/system/
└── aio-panel.service
```

The exact layout may change during implementation.

---

# 🧠 Existing Server Compatibility

AIO-PANEL is specifically designed for servers that already contain applications and services.

Example:

```text
Ubuntu
│
├── Nginx
├── Python
├── Node.js
├── npm
├── Git
├── PostgreSQL
├── Redis
├── Docker
│
├── Custom systemd services
│
├── Django applications
├── React applications
├── Node applications
└── Other websites
```

Installing AIO should add only:

```text
AIO
└── aio-panel.service
       └── :5555
```

It should not unnecessarily reinstall:

* Nginx
* Python
* Node.js
* npm
* Git
* PostgreSQL
* MySQL
* Redis
* Docker

---

# 🔍 Discovery Before Modification

AIO follows:

> **Discovery is read-only. Management is explicit.**

Installation should first detect:

```text
Operating system
Architecture
systemd
Existing AIO installation
Port 5555
Nginx
Apache
Python
Node.js
npm
Git
PostgreSQL
MySQL
Redis
Docker
Existing applications
Existing services
Firewall
SSH
```

AIO should report what it found before making changes.

---

# 🏷️ Resource Ownership

AIO should distinguish between:

```text
AIO Managed
Externally Managed
Detected
Unknown
```

For example:

```text
nginx.service
Status: Running
Owner: External

memotrack.service
Status: Running
Owner: External

aio-panel.service
Status: Running
Owner: AIO
```

AIO should not automatically take ownership of externally managed resources.

---

# 🛠️ Safe Configuration Changes

Critical operations should follow:

```text
Request
   ↓
Authenticate
   ↓
Authorize
   ↓
Validate
   ↓
Backup
   ↓
Apply
   ↓
Health Check
   ↓
Success?
  / \
Yes  No
 |    |
Keep Rollback
```

This applies particularly to:

* SSH
* Nginx
* Firewall
* PostgreSQL
* MySQL
* systemd
* DNS
* SSL

---

# 💥 Failure Isolation

AIO-PANEL must not become a dependency for the server.

If AIO stops:

```text
AIO              ❌
Nginx            ✓
PostgreSQL       ✓
MySQL            ✓
Redis            ✓
Docker           ✓
Gunicorn         ✓
Applications     ✓
SSH              ✓
```

AIO is a control layer, not a dependency layer.

---

# 📦 Technology Direction

The current proposed stack is:

| Component           | Planned technology         |
| ------------------- | -------------------------- |
| Core                | **Go**                     |
| Web server          | Go HTTP server             |
| CLI                 | Go                         |
| Database            | SQLite                     |
| Frontend            | TypeScript + React         |
| Frontend build      | Vite                       |
| Production frontend | Embedded into AIO binary   |
| Service manager     | systemd                    |
| Packaging           | Linux binary + archive     |
| CI/CD               | GitHub Actions             |
| Releases            | GitHub Releases            |
| Installer           | Shell script               |
| Uninstaller         | AIO CLI + uninstall script |

> The technology stack is currently a proposal. Final technology selection will follow architecture, security, resource-usage, and implementation research.

---

# 📦 Single-Program Architecture

AIO should ultimately behave like a single application:

```text
AIO-PANEL
│
├── Web UI
├── Web server
├── AIO Core
├── Authentication
├── Security
├── Modules
├── CLI
└── SQLite
```

The production server should not require:

```text
Node.js
npm
Python
Redis
PostgreSQL
MySQL
Nginx
```

just to operate AIO-PANEL.

---

# 🔨 Build Architecture

Development:

```text
Source Code
     │
     ├── Go Core
     │
     └── TypeScript/React
              │
              ▼
         Frontend Build
              │
              ▼
       Embedded Static Files
              │
              ▼
          Go Compiler
              │
              ▼
        AIO Executable
```

Production:

```text
AIO executable
     │
     ├── Core
     ├── Web server
     ├── CLI
     └── Embedded Web UI
```

---

# 🚀 GitHub CI/CD

The source code will be hosted in GitHub.

A release can be created using a version tag:

```text
v0.1.0
```

GitHub Actions will:

1. Check source code
2. Run tests
3. Run security checks
4. Build frontend
5. Embed frontend
6. Compile AIO
7. Build Linux AMD64
8. Build Linux ARM64
9. Package release artifacts
10. Generate checksums
11. Generate release metadata
12. Publish GitHub Release

Example artifacts:

```text
aio-panel-linux-amd64.tar.gz
aio-panel-linux-arm64.tar.gz

aio-panel-linux-amd64.tar.gz.sha256
aio-panel-linux-arm64.tar.gz.sha256
```

Future architectures can be added.

---

# 📥 Installation

A release artifact will contain everything required to install AIO:

```text
aio-panel-linux-amd64.tar.gz
│
├── aio
├── install.sh
├── uninstall.sh
├── VERSION
├── LICENSE
└── checksums
```

The server does not need development tools to build AIO.

Example:

```bash
tar -xzf aio-panel-linux-amd64.tar.gz
cd aio-panel
sha256sum -c aio.sha256
sudo ./install.sh
```

---

# 🔎 Installation Preflight

The installer should perform a read-only preflight.

Example:

```text
AIO-PANEL Preflight
────────────────────────────

OS              ✓ Ubuntu 24.04
Architecture    ✓ amd64
systemd         ✓ detected

Nginx           ✓ installed
Python          ✓ installed
Node.js         ✓ installed
npm             ✓ installed
Git             ✓ installed
PostgreSQL      ✓ installed
Redis           ✓ installed
Docker          ✓ installed

Existing sites  ✓ detected
Existing services ✓ detected

Port 5555       ✓ available

Existing services will NOT be modified.

Continue? [y/N]
```

---

# 🔁 Idempotent Installation

Running the installer multiple times must not corrupt the server.

Example:

```bash
sudo ./install.sh
```

If AIO already exists:

```text
AIO-PANEL is already installed.

Installed version: 0.1.0

Use:
  aio update
  aio repair
```

The installer should distinguish between:

* Fresh installation
* Upgrade
* Repair
* Already installed
* Unsupported environment

---

# 🧹 Uninstallation

AIO can be removed through CLI:

```bash
sudo aio uninstall
```

or through the included script:

```bash
sudo ./uninstall.sh
```

The uninstaller must first show what will be removed.

It should remove only AIO-owned resources:

```text
✓ AIO binary
✓ AIO systemd service
✓ AIO configuration
✓ AIO SQLite database
✓ AIO logs
✓ AIO state
```

It must NOT automatically remove:

```text
✗ Nginx
✗ PostgreSQL
✗ MySQL
✗ Docker
✗ Redis
✗ Applications
✗ Websites
✗ User files
✗ Existing systemd services
✗ SSH configuration
```

---

# 🔄 Installation Rollback

If installation fails:

```text
Install
   ↓
Failure
   ↓
Rollback AIO changes
   ↓
Verify server
```

Existing server services must remain untouched.

---

# 🩺 Repair & Diagnostics

AIO should eventually provide:

```bash
aio doctor
```

Example:

```text
AIO Doctor

✓ Binary
✓ Configuration
✓ SQLite
✓ Permissions
✓ systemd
✓ Port 5555
✓ Web server
✓ Database integrity

AIO is healthy.
```

And:

```bash
aio repair
```

for fixing AIO-specific installation problems.

---

# 🔐 Release Security

Because AIO can control the entire server, release integrity is critical.

Release artifacts should eventually provide:

* SHA-256 checksums
* Cryptographic signatures
* Trusted release metadata

Installation should ideally follow:

```text
Download
   ↓
Verify checksum
   ↓
Verify signature
   ↓
Inspect package
   ↓
Install
```

Convenience installation methods may be added later.

---

# 📈 Future Features

Potential future modules:

* Multi-server management
* Cloud backups
* S3-compatible storage
* Advanced monitoring
* Alerts
* Notifications
* WebAuthn/passkeys
* Advanced RBAC
* Container management
* Docker Compose management
* DNS management
* Mail server management
* Security hardening
* Vulnerability checks
* Application templates
* Plugin system
* Optional API
* Mobile management
* Remote server management

These should not compromise the lightweight core.

---

# 🗺️ Development Roadmap

## Phase 0 — Research

Research:

* Webuzo
* Webmin
* Plesk
* cPanel/WHM
* HestiaCP
* ISPConfig
* aaPanel
* Virtualmin
* CloudPanel
* CyberPanel

Study:

* Architecture
* Technology stack
* CLI
* Service management
* Configuration management
* Security
* Privilege model
* Installation
* Uninstallation
* Update mechanisms
* Resource usage

---

## Phase 1 — Architecture

Define:

* AIO Core
* Module system
* Security architecture
* Configuration format
* SQLite schema
* Filesystem layout
* CLI design
* Web architecture
* Installation architecture
* Update architecture

---

## Phase 2 — Core

Implement:

* AIO executable
* Configuration
* Logging
* SQLite
* Authentication
* Security foundation
* CLI
* Web server
* systemd integration

---

## Phase 3 — Dashboard

Implement:

* CPU
* RAM
* Disk
* Network
* Uptime
* Load
* Processes
* Services

---

## Phase 4 — Server Management

Implement:

* Users
* Groups
* SSH
* Services
* Logs
* Files
* Firewall

---

## Phase 5 — Web & Applications

Implement:

* Domains
* Nginx
* SSL
* Django
* Node.js
* React
* PHP
* Deployments

---

## Phase 6 — Databases

Implement:

* PostgreSQL
* MySQL/MariaDB
* Redis
* Backup
* Restore

---

## Phase 7 — Operations

Implement:

* Cron
* Backups
* Docker
* Monitoring
* Terminal
* Audit

---

## Phase 8 — Distribution

Implement:

* GitHub Actions
* Linux AMD64 build
* Linux ARM64 build
* Release artifacts
* Checksums
* Signatures
* Installer
* Uninstaller
* Upgrade
* Repair
* Rollback

---

# 🎯 Final Goal

AIO-PANEL should ultimately provide this experience:

```text
                    AIO-PANEL
              All-In-One Linux Control
                         │
                         │ :5555
                         ▼
                   ┌───────────┐
                   │ Web UI    │
                   └─────┬─────┘
                         │
                     AIO CORE
                         │
      ┌──────────────────┼──────────────────┐
      │                  │                  │
   Server             Web/App           Operations
      │                  │                  │
   Users              Domains            Backups
   SSH                SSL                Deployments
   Services           Apps               Cron
   Firewall           Databases          Monitoring
   Files              Docker             Logs
   Processes
                         │
                         ▼
                    Linux Server
```

**AIO-PANEL should be lightweight enough to run alongside a production server, powerful enough to manage the entire machine, secure enough to control privileged operations, and non-invasive enough to install on an already-configured Linux server without breaking existing applications or services.**
