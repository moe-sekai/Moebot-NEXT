#!/bin/sh
set -e

# 首次运行时,如果挂载/工作目录里没有 config.yml,则用示例配置初始化,确保开箱即用
if [ ! -f /app/config.yml ]; then
  if [ -f /app/config.example.yml ]; then
    echo "[entrypoint] config.yml not found, initializing from config.example.yml"
    cp /app/config.example.yml /app/config.yml
  else
    echo "[entrypoint] WARNING: neither config.yml nor config.example.yml present"
  fi
fi

# 确保运行所需目录存在(即使没有挂载 volume)
mkdir -p /app/data /app/data/master /app/data/cache /app/assets/stickers

exec "$@"
