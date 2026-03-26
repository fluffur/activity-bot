# syntax=docker/dockerfile:1.7

FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN --mount=type=cache,target=/root/.cache/go-build \
    GOGC=20 go build -ldflags="-s -w" -p 1 -o bot cmd/bot/main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/bot .

RUN mkdir -p /app/data

CMD ["./bot"]