@echo off
setlocal EnableExtensions
title Moebot NEXT

echo.
echo ========================================
echo        Moebot NEXT - PJSK BOT
echo              Starting up
echo ========================================
echo.

rem Check Bun
where bun >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Bun not found! Please install Bun first.
    echo Download: https://bun.sh/
    echo Windows PowerShell install:
    echo   powershell -c "irm bun.sh/install.ps1 ^| iex"
    pause
    exit /b 1
)

for /f "usebackq delims=" %%v in (`bun --version`) do set "BUN_VERSION=%%v"
echo [INFO] Bun version: %BUN_VERSION%

rem Navigate to project root
cd /d "%~dp0.."

rem Check if node_modules exists
if not exist "node_modules" (
    echo [INFO] Installing dependencies with Bun...
    call bun install
    if errorlevel 1 (
        echo [ERROR] Failed to install dependencies
        pause
        exit /b 1
    )
)

rem Check if koishi.yml exists
if not exist "koishi.yml" (
    echo [INFO] Creating default configuration...
    copy "koishi.example.yml" "koishi.yml" >nul
    echo [INFO] Please edit koishi.yml before running!
    echo [INFO] At minimum, set your QQ bot selfId in the adapter-onebot section.
    notepad "koishi.yml"
    pause
    exit /b 0
)

rem Build packages if dist output is missing
if not exist "packages\core\dist\index.js" (
    echo [INFO] Building workspace packages...
    call bun run build
    if errorlevel 1 (
        echo [ERROR] Build failed
        pause
        exit /b 1
    )
)

rem Create data directories
if not exist "data" mkdir data
if not exist "data\cache" mkdir data\cache
if not exist "data\master" mkdir data\master

echo [INFO] Starting Moebot NEXT...
echo [INFO] Console: http://localhost:5140
echo [INFO] OneBot WS: ws://localhost:6700
echo.

call bun run start

pause
