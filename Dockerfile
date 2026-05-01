# ============================================================
# Moebot NEXT — Single Container Deployment (Bun)
# ============================================================
FROM oven/bun:1.3-alpine AS builder

WORKDIR /app

# Copy package files for dependency caching
COPY package.json bun.lock* ./
COPY packages/shared/package.json packages/shared/
COPY packages/renderer/package.json packages/renderer/
COPY packages/core/package.json packages/core/
COPY packages/console/package.json packages/console/

# Install dependencies
RUN bun install --frozen-lockfile

# Copy source code
COPY tsconfig.base.json tsconfig.json ./
COPY packages/ packages/
COPY assets/ assets/

# Build all packages
RUN bun run build

# ---- Runtime ----
FROM oven/bun:1.3-alpine

WORKDIR /app

# Copy package files
COPY package.json bun.lock* ./
COPY packages/shared/package.json packages/shared/
COPY packages/renderer/package.json packages/renderer/
COPY packages/core/package.json packages/core/
COPY packages/console/package.json packages/console/

# Install production dependencies
RUN bun install --frozen-lockfile --production

# Copy built packages and console client sources
COPY --from=builder /app/packages/shared/dist packages/shared/dist
COPY --from=builder /app/packages/renderer/dist packages/renderer/dist
COPY --from=builder /app/packages/core/dist packages/core/dist
COPY --from=builder /app/packages/console/dist packages/console/dist
COPY --from=builder /app/packages/console/client packages/console/client

# Copy assets and config
COPY --from=builder /app/assets assets
COPY koishi.example.yml koishi.yml

# Create data directory
RUN mkdir -p data/cache data/master

# Ports: Koishi Console (5140) + OneBot WS (6700)
EXPOSE 5140 6700

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s \
  CMD wget -q --spider http://localhost:5140 || exit 1

CMD ["bun", "run", "start"]
