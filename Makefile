PORT ?= 8080

.PHONY: run build test

run:
	PORT=$(PORT) go run ./cmd/server

build:
	go build -o ./bin/server ./cmd/server

test:
	go test ./... -v
