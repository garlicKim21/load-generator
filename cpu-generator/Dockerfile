FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN apk add --no-cache build-base
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM debian:bookworm-slim
WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends stress-ng && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/main .
CMD ["./main"]