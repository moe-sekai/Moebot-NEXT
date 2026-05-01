#!/usr/bin/env bash
set -euo pipefail

# ============================================================
# Moebot NEXT — One-Click Installation Script (Bun)
# Usage: curl -fsSL https://raw.githubusercontent.com/xxx/moebot-next/main/scripts/install.sh | bash
# ============================================================

REPO_URL="https://github.com/xxx/moebot-next.git"
INSTALL_DIR="${MOEBOT_DIR:-$HOME/moebot-next}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

echo ""
echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}    Moebot NEXT — Installation${NC}"
echo -e "${CYAN}========================================${NC}"
echo ""

if ! command -v git &> /dev/null; then
    echo -e "${RED}[ERROR] git is required but not found.${NC}"
    exit 1
fi

if ! command -v bun &> /dev/null; then
    echo -e "${YELLOW}[INFO]${NC} Bun not found, installing Bun..."
    curl -fsSL https://bun.sh/install | bash
    export BUN_INSTALL="$HOME/.bun"
    export PATH="$BUN_INSTALL/bin:$PATH"
fi

if ! command -v bun &> /dev/null; then
    echo -e "${RED}[ERROR] Bun installation failed or PATH not updated.${NC}"
    echo "Please install Bun manually from https://bun.sh/ and retry."
    exit 1
fi

echo -e "${GREEN}[✓]${NC} Bun version: $(bun --version)"

if [ -d "$INSTALL_DIR" ]; then
    echo -e "${YELLOW}[INFO]${NC} Directory exists, pulling latest..."
    cd "$INSTALL_DIR"
    git pull --ff-only
else
    echo -e "${GREEN}[INFO]${NC} Cloning repository..."
    git clone "$REPO_URL" "$INSTALL_DIR"
    cd "$INSTALL_DIR"
fi

echo -e "${GREEN}[INFO]${NC} Installing dependencies with Bun..."
bun install

echo -e "${GREEN}[INFO]${NC} Building packages..."
bun run build

if [ ! -f "koishi.yml" ]; then
    cp koishi.example.yml koishi.yml
    echo -e "${GREEN}[✓]${NC} Configuration file created at koishi.yml"
fi

mkdir -p data/cache data/master

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}    Installation Complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "Next steps:"
echo -e "  1. ${CYAN}cd $INSTALL_DIR${NC}"
echo -e "  2. Edit ${CYAN}koishi.yml${NC} — set your bot's selfId"
echo -e "  3. Run ${CYAN}./scripts/start.sh${NC} (Linux/macOS) or ${CYAN}scripts\\start.bat${NC} (Windows)"
echo -e "  4. Open ${CYAN}http://localhost:5140${NC} for the management panel"
echo ""
