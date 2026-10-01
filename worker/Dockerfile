FROM golang:1.21-bullseye AS builder

WORKDIR /app
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o worker main.go

FROM ubuntu:22.04

# Install FFmpeg, CA certificates, curl, Python3 (for yt-dlp dependencies)
RUN apt-get update && apt-get install -y \
    ffmpeg \
    ca-certificates \
    curl \
    python3 \
    && rm -rf /var/lib/apt/lists/*

# Install yt-dlp standalone binary
RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp && \
    chmod a+rx /usr/local/bin/yt-dlp

WORKDIR /app
COPY --from=builder /app/worker .

# Make sure we run on the port provided by Render
ENV PORT=8080

EXPOSE $PORT

CMD ["./worker"]
