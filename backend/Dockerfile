FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server cmd/server/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server .
# Copy .env if it exists, but usually we use env vars in docker-compose
# COPY .env . 

EXPOSE 8080

CMD ["./server"]
