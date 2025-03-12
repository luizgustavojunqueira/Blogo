APP_NAME=blog

.PHONY: all build run dev test clean

build:
	@echo "Building..."
	go build -o bin/$(APP_NAME) cmd/$(APP_NAME)/main.go

run: build
	@echo "Running..."
	./bin/$(APP_NAME)

dev:
	@echo "Running in dev mode..."
	air

test:
	@echo "Running tests..."
	go test ./...

clean:
	@echo "Cleaning..."
	rm -rf bin/*
