.PHONY: run build test test-integration lint generate sqlc openapi tidy

run:
	go run ./cmd/api

build:
	CGO_ENABLED=0 go build -o bin/api ./cmd/api

test:
	go test ./...

test-integration:
	go test -tags=integration ./...

lint:
	golangci-lint run

generate: sqlc openapi

sqlc:
	cd db && sqlc generate

openapi:
	swag init -g cmd/api/main.go -o docs --outputTypes json --parseInternal

tidy:
	go mod tidy
