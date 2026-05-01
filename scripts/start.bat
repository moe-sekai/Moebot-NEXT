@echo off
chcp 65001 >nul 2>&1
title Moebot NEXT

echo.
echo  ╔══════════════════════════════════════╗
echo  ║       Moebot NEXT - PJSK BOT        ║
echo  ║         Starting up...               ║
echo  ╚══════════════════════════════════════╝
echo.

:: Check Node.js
where node >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Node.js not found! Please install Node.js 20+
    echo Download: https://nodejs.org/
    pause
    exit /b 1
)

:: Check Node.js version
for /f "tokens=1 delims=v." %%a in ('node -v') do set NODE_MAJOR=%%a
for /f "tokens=2 delims=v." %%a in ('node -v') do set NODE_MAJOR=%%a
echo [INFO] Node.js version: 
node -v

:: Navigate to project root
cd /d "%~dp0.."

:: Check if node_modules exists
if not exist "node_modules" (
    echo [INFO] Installing dependencies...
    call npm install
    if %errorlevel% neq 0 (
        echo [ERROR] Failed to install dependencies
        pause
        exit /b 1
    )
)

:: Check if koishi.yml exists
if not exist "koishi.yml" (
    echo [INFO] Creating default configuration...
    copy "koishi.example.yml" "koishi.yml"
    echo [INFO] Please edit koishi.yml before running!
    echo [INFO] At minimum, set your QQ bot selfId in the adapter-onebot section.
    notepad "koishi.yml"
    pause
    exit /b 0
)

:: Create data directories
if not exist "data" mkdir data
if not exist "data\cache" mkdir data\cache
if not exist "data\master" mkdir data\master

echo [INFO] Starting Moebot NEXT...
echo [INFO] Console: http://localhost:5140
echo [INFO] OneBot WS: ws://localhost:6700
echo.

:: Start Koishi
call npx koishi start

pause
