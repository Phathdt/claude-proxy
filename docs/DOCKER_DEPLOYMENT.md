# Docker Deployment & CI/CD Guide

## Quick Start with Docker

### Pull and Run

```bash
# Pull the latest image
docker pull phathdt379/claude-proxy:latest

# Create config file
cp config.example.yaml config.yaml
# Edit config.yaml with your settings

# Run container
docker run -d \
  --name claude-proxy \
  -p 4000:4000 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -v ~/.claude-proxy/data:/app/data \
  phathdt379/claude-proxy:latest
```

### Check Status

```bash
# View logs
docker logs claude-proxy

# Check health
docker exec claude-proxy wget -O - http://localhost:4000/health

# Stop container
docker stop claude-proxy

# Remove container
docker rm claude-proxy
```

## Building Locally

### Prerequisites
- Docker (with BuildKit enabled)
- Go 1.24+
- Node 22+
- pnpm

### Local Build

```bash
# Build image locally
docker build -t claude-proxy:local .

# Run locally built image
docker run -d \
  --name claude-proxy \
  -p 4000:4000 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -v ~/.claude-proxy/data:/app/data \
  claude-proxy:local
```

### Build Without Cache

```bash
docker build --no-cache -t claude-proxy:local .
```

## CI/CD Pipeline

### How It Works

The GitHub Actions workflow automatically:

1. **Triggers**: Manual workflow dispatch (GitHub Actions > Select "Build & Push Docker Image" > Input version)
2. **Builds Frontend**: Node 22 + pnpm
3. **Builds Backend**: Go 1.24 with embedded frontend
4. **Pushes to Docker Hub**: `phathdt379/claude-proxy:<version>` and `phathdt379/claude-proxy:latest`
5. **Creates Release**: GitHub Release with installation instructions and version tag

### Manual Trigger

1. Go to: `https://github.com/yourusername/claude-proxy/actions`
2. Select: `Build & Push Docker Image` workflow
3. Click: `Run workflow`
4. Enter version: `0.1.0` (or your desired version)
5. Click: `Run workflow`

Pipeline will:
- Build Docker image
- Push to Docker Hub as:
  - `phathdt379/claude-proxy:0.1.0`
  - `phathdt379/claude-proxy:latest`
- Create GitHub release: `v0.1.0`

### Required Secrets

For CI/CD to work, set these in GitHub repository settings (Settings > Secrets and Variables > Actions):

```
DOCKER_USERNAME: phathdt379
DOCKER_PASSWORD: <your-docker-hub-token>
```

**To create Docker Hub token:**
1. Go to https://hub.docker.com/settings/security
2. Create new access token
3. Add to GitHub secrets as `DOCKER_PASSWORD`

## Docker Image Details

### Image Structure

```
phathdt379/claude-proxy:0.1.0
├─ Frontend: Built with Node 22
├─ Backend: Built with Go 1.24
├─ Size: ~50MB (multi-stage optimized)
└─ Runtime: Alpine 3.18
```

### Security Features

- **Non-root user**: Runs as `claude:claude` (UID 1000)
- **Read-only data**: Data directory with 0700 permissions
- **Health check**: Automatic health check every 30 seconds
- **No secrets**: Configuration via config.yaml file mount

### Volumes

Mount these for persistence:

```bash
-v $(pwd)/config.yaml:/app/config.yaml    # Configuration file
-v ~/.claude-proxy/data:/app/data        # Account tokens and data
```

### Ports

```bash
-p 4000:4000    # API and admin dashboard
```

## Production Deployment

### Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  claude-proxy:
    image: phathdt379/claude-proxy:0.1.0
    container_name: claude-proxy
    ports:
      - "4000:4000"
    volumes:
      - ./config.yaml:/app/config.yaml:ro
      - claude-proxy-data:/app/data
    environment:
      - LOGGER__LEVEL=info
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-O", "-", "http://localhost:4000/health"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 5s

volumes:
  claude-proxy-data:
    driver: local
```

Run with Docker Compose:

```bash
# Start
docker-compose up -d

# View logs
docker-compose logs -f claude-proxy

# Stop
docker-compose down
```

### Kubernetes Deployment

Example `deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: claude-proxy
  labels:
    app: claude-proxy
spec:
  replicas: 1
  selector:
    matchLabels:
      app: claude-proxy
  template:
    metadata:
      labels:
        app: claude-proxy
    spec:
      containers:
      - name: claude-proxy
        image: phathdt379/claude-proxy:0.1.0
        ports:
        - containerPort: 4000
        volumeMounts:
        - name: config
          mountPath: /app/config.yaml
          subPath: config.yaml
          readOnly: true
        - name: data
          mountPath: /app/data
        livenessProbe:
          httpGet:
            path: /health
            port: 4000
          initialDelaySeconds: 5
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health
            port: 4000
          initialDelaySeconds: 5
          periodSeconds: 10
      volumes:
      - name: config
        configMap:
          name: claude-proxy-config
      - name: data
        persistentVolumeClaim:
          claimName: claude-proxy-data
```

## Environment Variables

Override config.yaml values with environment variables:

```bash
docker run -d \
  -p 4000:4000 \
  -e SERVER__PORT=8080 \
  -e LOGGER__LEVEL=debug \
  -e OAUTH__CLIENT_ID=your-client-id \
  phathdt379/claude-proxy:latest
```

Common variables:

```
SERVER__HOST=0.0.0.0
SERVER__PORT=4000
LOGGER__LEVEL=info
LOGGER__FORMAT=text
OAUTH__CLIENT_ID=your-client-id
STORAGE__DATA_FOLDER=/app/data
```

## Troubleshooting

### Container fails to start

```bash
# View logs
docker logs claude-proxy

# Common issues:
# 1. Port 4000 already in use: use -p 8080:4000
# 2. Config file not found: mount config.yaml correctly
# 3. Data directory permissions: ensure ~/.claude-proxy/data exists
```

### Health check failing

```bash
# Manual health check
docker exec claude-proxy wget -O - http://localhost:4000/health

# If fails, check:
# 1. Container logs: docker logs claude-proxy
# 2. Port binding: docker port claude-proxy
# 3. Network: docker network inspect bridge
```

### Data not persisting

```bash
# Verify volume mount
docker inspect claude-proxy | grep -A 10 Mounts

# Ensure host directory exists and has write permissions
mkdir -p ~/.claude-proxy/data
chmod 700 ~/.claude-proxy/data
```

## Version Management

### Release Workflow

1. **Local testing**: `docker build -t claude-proxy:test .`
2. **Trigger CI/CD**: Provide version (e.g., `0.1.0`)
3. **Automatic push**: Pushed as `phathdt379/claude-proxy:0.1.0` and `latest`
4. **GitHub Release**: Created with installation instructions
5. **Git tag**: `v0.1.0` tag created automatically

### Pulling specific versions

```bash
# Latest (recommended for stable)
docker pull phathdt379/claude-proxy:latest

# Specific version
docker pull phathdt379/claude-proxy:0.1.0

# List available tags
# https://hub.docker.com/r/phathdt379/claude-proxy/tags
```

## Performance Tuning

### Resource Limits

```bash
docker run -d \
  --memory=512m \
  --cpus="1.0" \
  -p 4000:4000 \
  phathdt379/claude-proxy:latest
```

### Build Optimization

Multi-stage build reduces image size by:
- Building frontend separately
- Using Alpine for runtime (~5MB base)
- Embedding frontend in Go binary
- Stripping symbols from binary

Final image size: ~50MB (including Go runtime)

## Best Practices

1. **Always use tags**: `phathdt379/claude-proxy:0.1.0` not `latest` in production
2. **Mount config**: Use read-only mount for config files: `--config.yaml:ro`
3. **Persistent data**: Always mount `/app/data` volume
4. **Health checks**: Let Docker manage restart on failure
5. **Logs**: Monitor with `docker logs -f container-name`
6. **Security**: Run with `--user nobody` if needed (though uses claude user)
7. **Updates**: Pull new versions explicitly, don't rely on auto-pull

## Support

For issues or questions:
- Check container logs: `docker logs claude-proxy`
- View health status: `docker exec claude-proxy curl http://localhost:4000/health`
- Read documentation: `docs/` folder in repository
