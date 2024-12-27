FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN apk add --no-cache build-base
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/edge/community stress-ng
COPY --from=builder /app/main .
CMD ["./main"]