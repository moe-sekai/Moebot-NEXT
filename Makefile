.PHONY: dev build build-web renderer-install tidy test clean docker docker-init docker-up docker-down docker-logs

dev:
	go run ./main.go

renderer-install:
	cd renderer && bun install

build-web:
	cd web && npm install && npm run build

build: build-web
	go build -o moebot ./main.go

tidy:
	go mod tidy

test:
	go test ./...

clean:
	rm -f moebot moebot.exe
	rm -rf web/dist

docker:
	docker build -t moebot-next-go .

# 首次使用 docker compose 前执行:确保 data 目录存在
docker-init:
	@mkdir -p data

docker-up: docker-init
	docker compose up -d --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f --tail=200
