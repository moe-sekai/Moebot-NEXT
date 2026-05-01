#!/usr/bin/env bash
set -euo pipefail

# ============================================================
# Moebot NEXT — macOS / Linux Startup Script
# ============================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo ""
echo -e "${CYAN}╔══════════════════════════════════════╗${NC}"
echo -e "${CYAN}║       Moebot NEXT - PJSK BOT        ║${NC}"
echo -e "${CYAN}║         Starting up...               ║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════╝${NC}"
echo ""

# Check Node.js
if ! command -v node &> /dev/null; then
    echo -e "${RED}[ERROR] Node.js not found!${NC}"
    echo "Please install Node.js 20+ from https://nodejs.org/"
    echo ""
    echo "Quick install options:"
    echo "  macOS:  brew install node@20"
    echo "  Ubuntu: curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash - && sudo apt install -y nodejs"
    echo "  Arch:   sudo pacman -S nodejs npm"
    exit 1
fi

# Check Node.js version
NODE_VERSION=$(node -v | sed 's/v//' | cut -d. -f1)
echo -e "${GREEN}[INFO]${NC} Node.js version: $(node -v)"
if [ "$NODE_VERSION" -lt 20 ]; then
    echo -e "${YELLOW}[WARN] Node.js 20+ recommended. Current: v${NODE_VERSION}${NC}"
fi

cd "$PROJECT_DIR"

# Install dependencies if needed
if [ ! -d "node_modules" ]; then
    echo -e "${GREEN}[INFO]${NC} Installing dependencies..."
    npm install
fi

# Create config if not exists
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

# Create data directories
mkdir -p data/cache data/master

echo -e "${GREEN}[INFO]${NC} Starting Moebot NEXT..."
echo -e "${GREEN}[INFO]${NC} Console:   ${CYAN}http://localhost:5140${NC}"
echo -e "${GREEN}[INFO]${NC} OneBot WS: ${CYAN}ws://localhost:6700${NC}"
echo ""

# Start Koishi
exec npx koishi start
