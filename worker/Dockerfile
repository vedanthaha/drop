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
    python3-pip \
    unzip \
    git \
    && rm -rf /var/lib/apt/lists/*

# Install yt-dlp standalone binary
RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp && \
    chmod a+rx /usr/local/bin/yt-dlp

# Install Deno
ENV DENO_INSTALL="/usr/local"
RUN curl -fsSL https://deno.land/install.sh | sh && \
    chmod a+rx /usr/local/bin/deno
ENV DENO_DIR="/tmp/deno_cache"
ENV XDG_CACHE_HOME="/tmp/.cache"
RUN mkdir -p /tmp/.cache && chmod 777 /tmp/.cache
# Install bgutil-ytdlp-pot-provider and clone the PO token generation script
RUN python3 -m pip install bgutil-ytdlp-pot-provider
RUN git clone https://github.com/Brainicism/bgutil-ytdlp-pot-provider.git /app/bgutil
WORKDIR /app/bgutil/server
RUN deno install
RUN chmod -R 777 /app/bgutil

WORKDIR /app
COPY --from=builder /app/worker .

# Make sure we run on the port provided by Render
ENV PORT=8080

EXPOSE $PORT

CMD ["./worker"]
