FROM golang:latest AS build-stage

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o blog ./cmd/blog/main.go

# Deploy the application binary into a lean image
FROM debian:bookworm-slim AS build-release-stage

WORKDIR /

COPY --from=build-stage /app/blog ./blog
COPY internal/repository/migrations ./internal/repository/migrations
COPY internal/static ./internal/static

EXPOSE 8080

ENTRYPOINT ["/blog"]

