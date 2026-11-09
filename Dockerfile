# Stage 1: Build assets and binary
FROM golang:1.27-alpine AS builder

WORKDIR /app

# Install build tools
RUN apk add --no-cache git curl ca-certificates

# Install templ CLI
RUN go install github.com/a-h/templ/cmd/templ@v0.3.1100

# Install standalone Tailwind CSS v4 CLI (Zero-Node.js)
RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 && \
    chmod +x tailwindcss-linux-x64 && \
    mv tailwindcss-linux-x64 /usr/local/bin/tailwindcss

# Copy dependency manifests
COPY go.mod go.sum ./
RUN go mod download

# Copy application sources
COPY . .

# Generate Templ templates & compile Tailwind CSS
RUN templ generate
RUN tailwindcss -i ui/assets/input.css -o ui/assets/styles.css --minify

# Compile pure Go binary (CGO=0)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/sitecraft main.go

# Stage 2: Minimal runtime image
FROM alpine:3.21

WORKDIR /app

# Install runtime certificates and timezone data
RUN apk add --no-cache ca-certificates tzdata

# Copy binary and presets
COPY --from=builder /app/bin/sitecraft /app/sitecraft
COPY --from=builder /app/presets /app/presets

# Expose HTTP port
EXPOSE 8090

# Volume for data (SQLite database & uploaded media)
VOLUME ["/data"]

ENV DATABASE_URL="file:/data/sitecraft.db?_pragma=busy_timeout(5000)" \
    UPLOADS_DIR="/data/uploads" \
    PORT=":8090"

ENTRYPOINT ["/app/sitecraft"]
CMD ["serve", "--port=:8090"]
