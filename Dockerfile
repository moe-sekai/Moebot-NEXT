# ============================================================
# Moebot NEXT — Single Container Deployment
# ============================================================
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files for dependency caching
COPY package.json package-lock.json* ./
COPY packages/shared/package.json packages/shared/
COPY packages/renderer/package.json packages/renderer/
COPY packages/core/package.json packages/core/
COPY packages/console/package.json packages/console/

# Install dependencies
RUN npm install --production=false

# Copy source code
COPY tsconfig.base.json tsconfig.json ./
COPY packages/ packages/

# Build all packages
RUN npm run build

# ---- Runtime ----
FROM node:20-alpine

WORKDIR /app

# Install production deps only
COPY package.json package-lock.json* ./
COPY packages/shared/package.json packages/shared/
COPY packages/renderer/package.json packages/renderer/
COPY packages/core/package.json packages/core/
COPY packages/console/package.json packages/console/
RUN npm install --production

# Copy built files
COPY --from=builder /app/packages/*/dist packages/
COPY --from=builder /app/packages/*/src packages/

# Copy assets
COPY assets/ assets/

# Copy default config
COPY koishi.example.yml koishi.yml

# Create data directory
RUN mkdir -p data/cache data/master

# Ports: Koishi Console (5140) + OneBot WS (6700)
EXPOSE 5140 6700

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s \
  CMD wget -q --spider http://localhost:5140 || exit 1

CMD ["node", "-e", "require('koishi').start()"]
