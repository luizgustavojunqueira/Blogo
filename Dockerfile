FROM node:18-alpine AS tailwind-stage
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY tailwind.config.js ./
COPY internal/static ./internal/static 
COPY internal/templates ./internal/templates
RUN npm run tailwind

# Install dependencies
FROM golang:latest AS fetch-stage
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# Generate SQL code using sqlc
FROM golang:latest AS sqlc-stage
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . /app
RUN curl -sSL https://github.com/sqlc-dev/sqlc/releases/latest/download/sqlc_1.29.0_linux_amd64.tar.gz \
 | tar -xz -C /usr/local/bin sqlc
RUN sqlc generate

# Generate templates using templ
FROM ghcr.io/a-h/templ:latest AS templ-stage
COPY --from=sqlc-stage --chown=65532:65532 /app /app
WORKDIR /app
RUN ["templ", "generate"]

# Build the Go application
FROM golang:latest AS build-stage
WORKDIR /app
COPY --from=fetch-stage /go/pkg /go/pkg
COPY --from=templ-stage /app /app
COPY --from=tailwind-stage /app/internal/static /app/internal/static
RUN GOOS=linux go build -o blog ./cmd/blog/main.go

# Create a minimal image for the release
FROM debian:bookworm-slim AS build-release-stage
WORKDIR /
COPY --from=build-stage /app/blog ./blog
COPY --from=build-stage /app/internal/repository/migrations ./internal/repository/migrations
COPY --from=build-stage /app/internal/static ./internal/static
EXPOSE 8080
ENTRYPOINT ["/blog"]
