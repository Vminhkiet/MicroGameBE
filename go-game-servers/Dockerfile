# Stage 1: build
FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o game-server ./cmd/server/main.go

# Stage 2: run
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/game-server /game-server
EXPOSE 8080/udp
EXPOSE 2112
CMD ["/game-server"]
