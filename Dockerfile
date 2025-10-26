# Go build stage
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS backend-builder

# Receive build args for cross-platform builds
ARG TARGETPLATFORM
ARG BUILDPLATFORM

WORKDIR /workspace

# Install build dependencies
RUN apk add --no-cache git upx

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Copy pre-built frontend (built by GitHub Actions workflow)
COPY frontend/dist ./frontend/dist

# Build binary with maximum optimization (supports both amd64 and arm64)
RUN echo "Building for ${TARGETPLATFORM}" && \
  CGO_ENABLED=0 GOOS=linux GOARCH=$(case ${TARGETPLATFORM} in \
  "linux/arm64") echo "arm64" ;; \
  *) echo "amd64" ;; \
  esac) go build \
  -ldflags="-s -w -extldflags=-static" \
  -tags netgo,osusergo \
  -trimpath \
  -o claude-proxy .

# Compress binary with UPX
RUN upx --best --lzma claude-proxy

# Runtime stage
FROM alpine:3.22

WORKDIR /app

# Install runtime dependencies (ca-certificates for HTTPS, tzdata for timezones)
RUN apk add --no-cache ca-certificates tzdata

# Create data directory with restricted permissions
RUN mkdir -p /app/data && chmod 700 /app/data

# Copy binary from backend-builder
COPY --from=backend-builder /workspace/claude-proxy /app/claude-proxy

# Create non-root user
RUN addgroup -g 1000 claude && \
  adduser -D -u 1000 -G claude claude

# Change ownership
RUN chown -R claude:claude /app

# Switch to non-root user
USER claude

# Expose port
EXPOSE 4000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:4000/health || exit 1

# Default command
CMD ["/app/claude-proxy", "server"]
