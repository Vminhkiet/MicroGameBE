# Stage 1: Build
FROM golang:1.22 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server main.go

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE 9999/udp

CMD ["./server"]
