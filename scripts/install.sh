#!/usr/bin/env bash
# ==============================================================================
# AIO-PANEL Linux Installer
# Philosophy: "Discovery is read-only. Management is explicit."
# AIO manages the server; it does NOT take ownership of the server.
# ==============================================================================
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

INSTALL_DIR="/opt/aio"
CONFIG_DIR="/etc/aio"
DATA_DIR="/var/lib/aio"
LOG_DIR="/var/log/aio"
SYSTEMD_SERVICE="/etc/systemd/system/aio-panel.service"
BINARY_PATH="${INSTALL_DIR}/aio"
SYMLINK_PATH="/usr/local/bin/aio"
PORT="5555"

echo -e "${CYAN}"
echo "  █████╗ ██╗ ██████╗       ██████╗  █████╗ ███╗   ██╗███████╗██╗     "
echo " ██╔══██╗██║██╔═══██╗      ██╔══██╗██╔══██╗████╗  ██║██╔════╝██║     "
echo " ███████║██║██║   ██║█████╗██████╔╝███████║██╔██╗ ██║█████╗  ██║     "
echo " ██╔══██║██║██║   ██║╚════╝██╔═══╝ ██╔══██║██║╚██╗██║██╔══╝  ██║     "
echo " ██║  ██║██║╚██████╔╝      ██║     ██║  ██║██║ ╚████║███████╗███████╗"
echo " ╚═╝  ╚═╝╚═╝ ╚═════╝       ╚═╝     ╚═╝  ╚═╝╚═╝  ╚═══╝╚══════╝╚══════╝"
echo -e "${NC}"
echo -e "${BLUE}All-In-One Linux Server Control Panel Installer${NC}\n"

# 1. Require Root Privileges
if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}❌ Error: This script must be run as root (e.g. sudo ./install.sh)${NC}" 
   exit 1
fi

# 2. Check architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64)  ARCH_BIN="amd64" ;;
    aarch64|arm64) ARCH_BIN="arm64" ;;
    *) echo -e "${RED}❌ Unsupported architecture: $ARCH${NC}"; exit 1 ;;
esac

echo -e "${YELLOW}🔍 Running Read-Only Preflight Discovery...${NC}"
echo "────────────────────────────────────────────────────────────"

# Run Preflight checks
echo -e "OS              ${GREEN}✓ $(uname -s) $(cat /etc/os-release | grep PRETTY_NAME | cut -d= -f2 | tr -d '\"')${NC}"
echo -e "Architecture    ${GREEN}✓ ${ARCH_BIN}${NC}"

# Check systemd
if command -v systemctl >/dev/null 2>&1; then
    echo -e "systemd         ${GREEN}✓ detected${NC}"
else
    echo -e "systemd         ${RED}❌ not found (systemd is required)${NC}"
    exit 1
fi

echo ""
# Check existing software without modifying anything
command -v nginx >/dev/null 2>&1 && echo -e "Nginx           ${GREEN}✓ installed ($(nginx -v 2>&1 | cut -d/ -f2))${NC}" || echo -e "Nginx           ${CYAN}ℹ not installed (optional)${NC}"
command -v python3 >/dev/null 2>&1 && echo -e "Python          ${GREEN}✓ installed ($(python3 --version))${NC}" || echo -e "Python          ${CYAN}ℹ not installed (optional)${NC}"
command -v node >/dev/null 2>&1 && echo -e "Node.js         ${GREEN}✓ installed ($(node --version))${NC}" || echo -e "Node.js         ${CYAN}ℹ not installed (optional)${NC}"
command -v npm >/dev/null 2>&1 && echo -e "npm             ${GREEN}✓ installed ($(npm --version))${NC}" || echo -e "npm             ${CYAN}ℹ not installed (optional)${NC}"
command -v git >/dev/null 2>&1 && echo -e "Git             ${GREEN}✓ installed ($(git --version))${NC}" || echo -e "Git             ${CYAN}ℹ not installed (optional)${NC}"
command -v psql >/dev/null 2>&1 && echo -e "PostgreSQL      ${GREEN}✓ installed${NC}" || echo -e "PostgreSQL      ${CYAN}ℹ not installed (optional)${NC}"
command -v redis-server >/dev/null 2>&1 && echo -e "Redis           ${GREEN}✓ installed${NC}" || echo -e "Redis           ${CYAN}ℹ not installed (optional)${NC}"
command -v docker >/dev/null 2>&1 && echo -e "Docker          ${GREEN}✓ installed${NC}" || echo -e "Docker          ${CYAN}ℹ not installed (optional)${NC}"

# Check Port 5555
echo ""
if ss -tuln | grep -q ":${PORT} "; then
    echo -e "${RED}⚠️  Port ${PORT} is currently in use by another process!${NC}"
    echo "   AIO-PANEL will NOT terminate this process."
    read -p "   Enter an alternate port for AIO-PANEL [e.g. 5556]: " ALT_PORT
    if [[ -n "$ALT_PORT" ]]; then
        PORT="$ALT_PORT"
    else
        echo -e "${RED}Installation cancelled.${NC}"
        exit 1
    fi
else
    echo -e "Port ${PORT}       ${GREEN}✓ available${NC}"
fi

echo ""
echo -e "${GREEN}🛡️  Non-Invasive Guarantee:${NC}"
echo "   AIO-PANEL will install alongside your existing applications."
echo "   It will NOT modify, replace, or reinstall:"
echo "   • Nginx configuration or sites"
echo "   • Python, Node.js, or npm runtimes"
echo "   • PostgreSQL, MySQL, or Redis databases"
echo "   • Custom applications or systemd services"
echo "────────────────────────────────────────────────────────────"

read -p "Continue with AIO-PANEL installation? [y/N]: " CONFIRM
if [[ "$CONFIRM" != "y" && "$CONFIRM" != "Y" ]]; then
    echo "Installation aborted."
    exit 0
fi

echo -e "\n${YELLOW}📦 Installing AIO-PANEL...${NC}"

# 3. Create Directories
mkdir -p "${INSTALL_DIR}"
mkdir -p "${CONFIG_DIR}"
mkdir -p "${DATA_DIR}"
mkdir -p "${LOG_DIR}"

# 4. Copy Binary
if [[ -f "./aio-linux-${ARCH_BIN}" ]]; then
    cp "./aio-linux-${ARCH_BIN}" "${BINARY_PATH}"
elif [[ -f "./aio" ]]; then
    cp "./aio" "${BINARY_PATH}"
else
    echo -e "${RED}❌ Binary not found in current directory.${NC}"
    exit 1
fi
chmod 750 "${BINARY_PATH}"
ln -sf "${BINARY_PATH}" "${SYMLINK_PATH}"

# 5. Create default configuration if not present
if [[ ! -f "${CONFIG_DIR}/aio.conf" ]]; then
    cat <<EOF > "${CONFIG_DIR}/aio.conf"
{
  "server": {
    "host": "0.0.0.0",
    "port": ${PORT},
    "tls": false
  },
  "database": {
    "path": "${DATA_DIR}/aio.db"
  },
  "logging": {
    "level": "info",
    "format": "pretty",
    "file": "${LOG_DIR}/aio.log"
  },
  "paths": {
    "config_dir": "${CONFIG_DIR}",
    "data_dir": "${DATA_DIR}",
    "log_dir": "${LOG_DIR}",
    "backup_dir": "${DATA_DIR}/backups"
  }
}
EOF
    chmod 600 "${CONFIG_DIR}/aio.conf"
fi

# 6. Install systemd Service
cat <<EOF > "${SYSTEMD_SERVICE}"
[Unit]
Description=AIO-PANEL Server Daemon
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=${DATA_DIR}
ExecStart=${BINARY_PATH} server --config ${CONFIG_DIR}/aio.conf
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable aio-panel.service
systemctl restart aio-panel.service

# 7. Post-Installation Health Verification
echo -e "${YELLOW}🔍 Verifying service status and health endpoint...${NC}"
sleep 1.5

if ! systemctl is-active --quiet aio-panel.service; then
    echo -e "${RED}❌ Error: aio-panel.service failed to start!${NC}"
    echo -e "Inspect logs with: ${CYAN}journalctl -u aio-panel -n 30 --no-pager${NC}"
    exit 1
fi
echo -e "systemd service   ${GREEN}✓ active & running${NC}"

HEALTH_OK=false
for i in {1..10}; do
    if curl -s -f "http://127.0.0.1:${PORT}/health" >/dev/null 2>&1 || curl -s -f "http://127.0.0.1:${PORT}/api/v1/health" >/dev/null 2>&1; then
        HEALTH_OK=true
        break
    fi
    sleep 0.5
done

if [[ "$HEALTH_OK" == "true" ]]; then
    echo -e "HTTP API Health   ${GREEN}✓ 200 OK (http://127.0.0.1:${PORT}/health)${NC}"
else
    echo -e "HTTP API Health   ${YELLOW}⚠️ Service started, waiting for port ${PORT}${NC}"
fi

# 8. Success Banner
SERVER_IP=$(hostname -I | awk '{print $1}' || echo "SERVER_IP")
echo -e "\n${GREEN}============================================================${NC}"
echo -e "${GREEN}🎉 AIO-PANEL successfully installed & verified!${NC}"
echo -e "${GREEN}============================================================${NC}"
echo -e "Web Access : ${CYAN}http://${SERVER_IP}:${PORT}${NC}"
echo -e "Superuser  : ${CYAN}sudo aio createsuperuser${NC} (Run this to set admin login)"
echo -e "CLI Tool   : ${CYAN}aio status${NC} or ${CYAN}aio doctor${NC}"
echo -e "Logs       : ${CYAN}journalctl -u aio-panel -f${NC}"
echo -e "${GREEN}============================================================${NC}\n"
