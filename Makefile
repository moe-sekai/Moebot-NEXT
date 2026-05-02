.PHONY: dev build build-web renderer-install tidy test clean docker

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
