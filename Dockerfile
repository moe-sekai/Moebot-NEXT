# Stage 1: build Vue admin panel
FROM node:22-alpine AS web-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm install
COPY web/ ./
RUN npm run build

# Stage 2: install Bun renderer deps
FROM oven/bun:1-alpine AS renderer-deps
WORKDIR /app/renderer
COPY renderer/package.json ./
RUN bun install --production
COPY renderer/ ./

# Stage 3: build Go binary
FROM golang:1.25-alpine AS go-builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
COPY --from=web-builder /app/web/dist ./web/dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o moebot ./main.go

# Stage 4: runtime
FROM oven/bun:1-alpine
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY --from=go-builder /app/moebot ./moebot
COPY --from=go-builder /app/config.example.yml ./config.yml
COPY --from=renderer-deps /app/renderer ./renderer
COPY assets/ ./assets/
EXPOSE 8080 6700
CMD ["./moebot"]
