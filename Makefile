.PHONY: run build test up down

run:
	go run cmd/server/main.go

build:
	go build -o bin/server cmd/server/main.go

test:
	go test ./... -v

up:
	docker-compose up -d

down:
	docker-compose down
