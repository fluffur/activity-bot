FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY migrate .

CMD ["./migrate"]