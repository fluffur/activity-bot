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
    go build -o bot cmd/bot/main.go

FROM scratch

WORKDIR /app

COPY --from=builder /app/bot .

CMD ["./bot"]