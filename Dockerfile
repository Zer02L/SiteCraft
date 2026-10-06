# Stage 1: Build assets and binary
FROM golang:1.27-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git curl nodejs npm

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@v0.3.1070

# Install standalone tailwindcss CLI
RUN npm install -g tailwindcss

# Copy dependency manifests
COPY go.mod go.sum ./
RUN go mod download

# Copy application sources
COPY . .

# Generate templates & compile CSS
RUN templ generate
RUN tailwindcss -i ui/assets/input.css -o ui/assets/styles.css --minify

# Compile single binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/sitecraft .

# Stage 2: Minimal runtime image
FROM alpine:3.20

WORKDIR /pb

# Install runtime certificates and tzdata
RUN apk add --no-cache ca-certificates tzdata

# Copy binary, assets, and presets
COPY --from=builder /app/bin/sitecraft /pb/sitecraft
COPY --from=builder /app/ui/assets /pb/ui/assets
COPY --from=builder /app/presets /pb/presets

# Expose HTTP port
EXPOSE 8090

# Volume for SQLite database and uploaded files
VOLUME ["/pb/pb_data"]

ENTRYPOINT ["/pb/sitecraft"]
CMD ["serve", "--http=0.0.0.0:8090"]
