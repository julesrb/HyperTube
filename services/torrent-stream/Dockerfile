FROM golang:1.23-alpine AS builder
WORKDIR /app
RUN apk add --no-cache ffmpeg
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN go build -o torrent-stream .

FROM alpine:3.20
RUN apk add --no-cache ffmpeg
WORKDIR /app
COPY --from=builder /app/torrent-stream .
COPY player.html .
COPY internal/transcode/test/ ./internal/transcode/test/
EXPOSE 8081
CMD ["./torrent-stream"]
