#!/usr/bin/env bash
# ==============================================================================
# AIO-PANEL Safe Uninstaller
# Philosophy: Removes ONLY AIO-owned resources.
# All existing applications, web servers, databases & services remain untouched.
# ==============================================================================

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}❌ Error: This script must be run as root (sudo ./uninstall.sh)${NC}" 
   exit 1
fi

echo -e "${YELLOW}⚠️  AIO-PANEL Uninstaller${NC}"
echo "────────────────────────────────────────────────────────────"
echo "The following AIO-owned resources will be removed:"
echo -e "  ${RED}•${NC} AIO Binary (/opt/aio & /usr/local/bin/aio)"
echo -e "  ${RED}•${NC} AIO systemd service (aio-panel.service)"
echo -e "  ${RED}•${NC} AIO configuration (/etc/aio)"
echo -e "  ${RED}•${NC} AIO SQLite metadata database (/var/lib/aio)"
echo -e "  ${RED}•${NC} AIO logs (/var/log/aio)"
echo ""
echo -e "${GREEN}🛡️  Untouched Resources (WILL NOT BE MODIFIED OR REMOVED):${NC}"
echo "  ✓ Nginx & existing websites"
echo "  ✓ PostgreSQL, MySQL & Redis databases"
echo "  ✓ Docker containers & images"
echo "  ✓ Python & Node.js runtimes"
echo "  ✓ User directories & custom systemd services"
echo "────────────────────────────────────────────────────────────"

read -p "Are you sure you want to uninstall AIO-PANEL? [y/N]: " CONFIRM
if [[ "$CONFIRM" != "y" && "$CONFIRM" != "Y" ]]; then
    echo "Uninstallation aborted."
    exit 0
fi

echo -e "\n${YELLOW}Stopping and disabling aio-panel.service...${NC}"
if systemctl is-active --quiet aio-panel.service 2>/dev/null; then
    systemctl stop aio-panel.service
fi
if systemctl is-enabled --quiet aio-panel.service 2>/dev/null; then
    systemctl disable aio-panel.service
fi

rm -f /etc/systemd/system/aio-panel.service
systemctl daemon-reload

echo -e "${YELLOW}Removing AIO binary and directories...${NC}"
rm -f /usr/local/bin/aio
rm -rf /opt/aio
rm -rf /etc/aio
rm -rf /var/lib/aio
rm -rf /var/log/aio

echo -e "\n${GREEN}============================================================${NC}"
echo -e "${GREEN}✅ AIO-PANEL has been cleanly uninstalled.${NC}"
echo -e "All your applications and system services remain 100% untouched."
echo -e "${GREEN}============================================================${NC}\n"
