# syntax=docker/dockerfile:1.6

############################
# Stage 1: build Vue admin panel
############################
FROM node:22-alpine AS web-builder
WORKDIR /app/web

# 优先复制锁文件以最大化缓存命中
COPY web/package.json web/package-lock.json* ./
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi

COPY web/ ./
# web 源码通过 ../../../assets/moebot.svg 引用仓库根目录下的 assets,需要把它放到相对位置
COPY assets/ /app/assets/
RUN npm run build


############################
# Stage 2: install Bun renderer 依赖(production)
############################
FROM oven/bun:1-alpine AS renderer-deps
WORKDIR /app/renderer

COPY renderer/package.json renderer/bun.lock* ./
RUN bun install --frozen-lockfile --production || bun install --production

# 再复制源码,与 node_modules 合并(node_modules 不会被覆盖,因为源码里没有该目录)
COPY renderer/ ./


############################
# Stage 3: build Go binary(嵌入 web/dist)
############################
FROM golang:1.25-alpine AS go-builder
WORKDIR /app

RUN apk add --no-cache git ca-certificates

# 先解析依赖以利用层缓存
COPY go.mod go.sum* ./
RUN go mod download

COPY . .
# 用 web-builder 产物覆盖空的 web/dist(go:embed 需要 dist 目录存在)
COPY --from=web-builder /app/web/dist ./web/dist

ENV CGO_ENABLED=0 GOOS=linux
RUN go build -trimpath -ldflags="-s -w" -o /out/moebot ./main.go


############################
# Stage 4: 运行时镜像
############################
FROM oven/bun:1-alpine AS runtime
WORKDIR /app

# tini: 正确转发信号给 Go 主进程,避免 bun 子进程僵尸
# ca-certificates / tzdata: HTTPS 与时区
RUN apk add --no-cache ca-certificates tzdata tini wget \
    && addgroup -S moebot && adduser -S moebot -G moebot

ENV TZ=Asia/Shanghai

# Go 二进制
COPY --from=go-builder /out/moebot /app/moebot

# 默认配置(entrypoint 会在首次运行时复制为 data/config.yml)
COPY config.example.yml /app/config.example.yml

# 渲染器(包含 node_modules + 源码 + 字体等资源)
COPY --from=renderer-deps /app/renderer /app/renderer

# 静态资源(角色立绘 / 卡框 / 图标 / 字体 等)
COPY assets/ /app/assets/

# 入口脚本
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh /app/moebot \
    && mkdir -p /app/data /app/data/master /app/data/cache \
    && chown -R moebot:moebot /app

USER moebot

# 8080: Web 控制台 / 6700: OneBot 反向 WS / 13001: 渲染器(默认仅本机访问)
EXPOSE 8080 6700

HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/api/health || exit 1

VOLUME ["/app/data"]

ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/docker-entrypoint.sh"]
CMD ["/app/moebot"]
