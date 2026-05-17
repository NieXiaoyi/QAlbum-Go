FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o qalbum-server ./cmd/server
RUN go build -o qalbum-admin ./cmd/admin

FROM alpine:3.19

RUN apk add --no-cache sqlite ffmpeg

WORKDIR /app

COPY --from=builder /app/qalbum-server /app/
COPY --from=builder /app/qalbum-admin /app/
COPY config.yaml /app/

RUN mkdir -p /app/data/db /app/data/photos

EXPOSE 8080

CMD ["./qalbum-server"]
