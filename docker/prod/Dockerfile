FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o bot cmd/bot/main.go
RUN go build -o migrate cmd/migrate/main.go

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/bot .
COPY --from=builder /app/migrate .
COPY migrations ./migrations
CMD ["./bot"]
