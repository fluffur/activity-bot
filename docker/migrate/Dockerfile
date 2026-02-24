FROM golang:1.25-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
ENV CGO_ENABLED=0
RUN --mount=type=cache,target=/root/.cache/go-build go build -o migrate cmd/migrate/main.go

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/migrate .
COPY migrations ./migrations
CMD ["./migrate"]