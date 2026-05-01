#!/usr/bin/env bash
set -euo pipefail

# ============================================================
# Moebot NEXT — macOS / Linux Startup Script (Bun)
# ============================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

echo ""
echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}       Moebot NEXT - PJSK BOT${NC}"
echo -e "${CYAN}             Starting up${NC}"
echo -e "${CYAN}========================================${NC}"
echo ""

if ! command -v bun &> /dev/null; then
    echo -e "${RED}[ERROR] Bun not found!${NC}"
    echo "Please install Bun from https://bun.sh/"
    echo ""
    echo "Quick install:"
    echo "  curl -fsSL https://bun.sh/install | bash"
    exit 1
fi

echo -e "${GREEN}[INFO]${NC} Bun version: $(bun --version)"

cd "$PROJECT_DIR"

if [ ! -d "node_modules" ]; then
    echo -e "${GREEN}[INFO]${NC} Installing dependencies with Bun..."
    bun install
fi

if [ ! -f "koishi.yml" ]; then
    echo -e "${YELLOW}[INFO]${NC} Creating default configuration..."
    cp koishi.example.yml koishi.yml
    echo -e "${YELLOW}[INFO]${NC} Please edit koishi.yml before running!"
    echo "  At minimum, set your QQ bot selfId in the adapter-onebot section."
    echo ""
    echo "  nano koishi.yml"
    echo "  # or"
    echo "  vim koishi.yml"
    exit 0
fi

if [ ! -f "packages/core/dist/index.js" ]; then
    echo -e "${GREEN}[INFO]${NC} Building workspace packages..."
    bun run build
fi

mkdir -p data/cache data/master

echo -e "${GREEN}[INFO]${NC} Starting Moebot NEXT..."
echo -e "${GREEN}[INFO]${NC} Console:   ${CYAN}http://localhost:5140${NC}"
echo -e "${GREEN}[INFO]${NC} OneBot WS: ${CYAN}ws://localhost:6700${NC}"
echo ""

exec bun run start
