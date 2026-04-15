FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY bot .

RUN mkdir -p /app/data

CMD ["./bot"]