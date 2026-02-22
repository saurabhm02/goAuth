.PHONY: build run dev

BINARY := bin/goAuth
MAIN   := cmd/server/main.go

build:
	go build -o $(BINARY) $(MAIN)

run: build
	(set -a && [ -f .env ] && . ./.env; set +a; ./$(BINARY))

dev:
	(set -a && [ -f .env ] && . ./.env; set +a; go run $(MAIN))
