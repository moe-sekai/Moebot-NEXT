#!/bin/sh
set -e

# 首次运行时,如果 data 目录里没有 config.yml,则用示例配置初始化,确保开箱即用
if [ ! -f /app/data/config.yml ]; then
  if [ -f /app/config.example.yml ]; then
    echo "[entrypoint] config.yml not found, initializing from config.example.yml"
    mkdir -p /app/data
    cp /app/config.example.yml /app/data/config.yml
  else
    echo "[entrypoint] WARNING: neither config.yml nor config.example.yml present"
  fi
fi

# 确保运行所需目录存在(即使没有挂载 volume)
mkdir -p /app/data /app/data/master /app/data/cache /app/assets/stickers

exec "$@"
