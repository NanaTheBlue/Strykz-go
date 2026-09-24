# Build the Sidecar Binary
FROM golang:1.26.0-bookworm AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /sidecar ./cmd/sidecar

# Build the CSS Dedicated Server Image
FROM cm2network/steamcmd:root

# Install dependencies (tmux, netcat, etc. if needed)
RUN apt-get update && apt-get install -y tmux netcat-traditional && rm -rf /var/lib/apt/lists/*

USER steam
WORKDIR /home/steam/steamcmd

# Download Counter-Strike: Source (AppID 232330)
RUN ./steamcmd.sh +force_install_dir /home/steam/css-dedicated \
    +login anonymous \
    +app_update 232330 validate \
    +quit

WORKDIR /home/steam/css-dedicated

# Copy Sidecar from builder
COPY --from=builder /sidecar /home/steam/css-dedicated/sidecar

# Copy start script
COPY --chown=steam:steam docker/css-server/start.sh /home/steam/css-dedicated/start.sh
RUN chmod +x /home/steam/css-dedicated/start.sh

# Expose SRCDS ports
EXPOSE 27015/tcp 27015/udp 27020/udp

ENTRYPOINT ["/home/steam/css-dedicated/start.sh"]
