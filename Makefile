.PHONY: build run dev

BINARY := bin/goAuth
MAIN   := cmd/server/main.go

build:
	go build -o $(BINARY) $(MAIN)

run: build
	./$(BINARY)

dev:
	go run $(MAIN)
