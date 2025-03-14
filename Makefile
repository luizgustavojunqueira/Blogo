APP_NAME=blog

.PHONY: all build run dev test database clean

build:
	@echo "Building..."
	go build -o bin/$(APP_NAME) cmd/$(APP_NAME)/main.go

run: build database
	@echo "Running..."
	./bin/$(APP_NAME)

dev: database
	@echo "Running in dev mode..."
	air

database:
	@echo "Creating database..."
	docker-compose up database -d

test:
	@echo "Running tests..."
	go test ./...

testv:
	@echo "Running tests with verbose..."
	go test -v ./...

clean:
	@echo "Cleaning..."
	rm -rf bin/*
