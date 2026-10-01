.PHONY: run build test cover fmt up down

run:
	go run cmd/api/main.go

build:
	go build -o bin/api cmd/api/main.go

test:
	go test ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

fmt:
	gofmt -w .

up:
	docker-compose up -d

down:
	docker-compose down