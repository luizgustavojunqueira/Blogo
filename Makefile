APP_NAME=blog

.PHONY: all build run dev clean

build:
	@echo "Building..."
	go build -o bin/$(APP_NAME) cmd/$(APP_NAME)/main.go

run: build
	@echo "Running..."
	./bin/$(APP_NAME)

dev:
	@echo "Running in dev mode..."
	air

clean:
	@echo "Cleaning..."
	rm -rf bin/*
