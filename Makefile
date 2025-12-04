.PHONY: run build test up down proto

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

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/v1/*.proto
