# Astrometry API Server

![Version](https://img.shields.io/github/v/release/DiarmuidKelly/Astrometry-API-Server?label=version)
![License](https://img.shields.io/badge/license-GPL--3.0-blue.svg)
![Go Version](https://img.shields.io/github/go-mod/go-version/DiarmuidKelly/Astrometry-API-Server)
[![Go Report Card](https://goreportcard.com/badge/github.com/DiarmuidKelly/Astrometry-API-Server)](https://goreportcard.com/report/github.com/DiarmuidKelly/Astrometry-API-Server)

A production-ready HTTP API server for astrometric plate-solving. Built on top of the [Astrometry-Go-Client](https://github.com/DiarmuidKelly/Astrometry-Go-Client) library, this server provides a RESTful interface for solving astronomical images.

## Features

- RESTful HTTP API for plate-solving
- Multipart file upload support
- Configurable solve parameters (scale bounds, downsampling, RA/Dec hints)
- Docker-based deployment
- CORS support for web applications
- Health check endpoint
- Request logging and monitoring
- Graceful shutdown
- Multi-platform Docker images (amd64, arm64)

## Quick Start

### Using Docker (Recommended)

```bash
# Pull the image from GitHub Container Registry
docker pull ghcr.io/diarmuidkelly/astrometry-api-server:latest

# Run the server
docker run -d \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /path/to/astrometry-data:/data/indexes:ro \
  -e ASTROMETRY_INDEX_PATH=/data/indexes \
  ghcr.io/diarmuidkelly/astrometry-api-server:latest

# Test the server
curl http://localhost:8080/health
```

### Using Docker Compose

```bash
# Clone the repository
git clone https://github.com/DiarmuidKelly/Astrometry-API-Server.git
cd Astrometry-API-Server

# Create .env file
cp .env.example .env
# Edit .env and set ASTROMETRY_INDEX_PATH

# Start the server
docker-compose up -d

# View logs
docker-compose logs -f api
```

### Building from Source

```bash
# Prerequisites: Go 1.21+, Docker

# Clone the repository
git clone https://github.com/DiarmuidKelly/Astrometry-API-Server.git
cd Astrometry-API-Server

# Download dependencies
go mod download

# Build the binary
make build

# Run the server
export ASTROMETRY_INDEX_PATH=/path/to/astrometry-data
./bin/astrometry-api-server
```

## API Reference

### Endpoints

#### `GET /health`

Health check endpoint.

**Response:**
```json
{
  "status": "healthy",
  "uptime_seconds": 123.45,
  "version": "0.1.0"
}
```

#### `POST /solve`

Solve an astronomical image.

**Request:**
- Method: `POST`
- Content-Type: `multipart/form-data`
- Max upload size: 50 MB

**Form Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `image` | file | Yes | Image file (jpg, png, fits) |
| `scale_low` | float | No | Lower bound of image scale |
| `scale_high` | float | No | Upper bound of image scale |
| `scale_units` | string | No | Scale units: "degwidth", "arcminwidth", "arcsecperpix" (default: "arcminwidth") |
| `downsample_factor` | int | No | Downsample factor (default: 2) |
| `depth_low` | int | No | Min quads to try (default: 10) |
| `depth_high` | int | No | Max quads to try (default: 20) |
| `ra` | float | No | RA hint in degrees |
| `dec` | float | No | Dec hint in degrees |
| `radius` | float | No | Search radius in degrees |

**Response:**
```json
{
  "solved": true,
  "ra": 120.123456,
  "dec": 45.987654,
  "pixel_scale": 1.23,
  "rotation": 15.5,
  "field_width": 2.5,
  "field_height": 1.8,
  "wcs_header": {
    "CRVAL1": "120.123456",
    "CRVAL2": "45.987654",
    ...
  },
  "solve_time": 12.34
}
```

If solving fails:
```json
{
  "solved": false,
  "error": "error description"
}
```

### Example Usage

#### cURL

```bash
# Basic solve
curl -X POST \
  -F "image=@photo.jpg" \
  -F "scale_low=1" \
  -F "scale_high=5" \
  http://localhost:8080/solve

# Solve with RA/Dec hints
curl -X POST \
  -F "image=@photo.jpg" \
  -F "scale_low=1" \
  -F "scale_high=3" \
  -F "ra=120.5" \
  -F "dec=45.2" \
  -F "radius=5" \
  http://localhost:8080/solve
```

#### Python

```python
import requests

# Solve an image
with open('photo.jpg', 'rb') as f:
    files = {'image': f}
    data = {
        'scale_low': 1.0,
        'scale_high': 5.0,
        'downsample_factor': 2
    }
    response = requests.post('http://localhost:8080/solve', files=files, data=data)
    result = response.json()

if result['solved']:
    print(f"RA: {result['ra']}, Dec: {result['dec']}")
    print(f"Pixel Scale: {result['pixel_scale']} arcsec/pixel")
else:
    print(f"Failed: {result.get('error', 'Unknown error')}")
```

#### JavaScript (Fetch API)

```javascript
async function solveImage(imageFile) {
  const formData = new FormData();
  formData.append('image', imageFile);
  formData.append('scale_low', '1');
  formData.append('scale_high', '5');

  const response = await fetch('http://localhost:8080/solve', {
    method: 'POST',
    body: formData
  });

  const result = await response.json();

  if (result.solved) {
    console.log(`RA: ${result.ra}, Dec: ${result.dec}`);
    console.log(`Pixel Scale: ${result.pixel_scale} arcsec/pixel`);
  } else {
    console.error(`Failed: ${result.error}`);
  }
}
```

## Configuration

The server is configured via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `ASTROMETRY_INDEX_PATH` | `/data/indexes` | Path to astrometry index files |
| `PORT` | `8080` | HTTP server port |

## Prerequisites

### Docker Socket Access

The API server needs access to the Docker socket to spawn astrometry containers. When running in Docker, mount the socket:

```bash
-v /var/run/docker.sock:/var/run/docker.sock
```

### Astrometry Index Files

Download appropriate index files for your field of view. See the [Astrometry-Go-Client Index Files Guide](https://github.com/DiarmuidKelly/Astrometry-Go-Client#index-files-guide) for details.

**Quick solver setup (50mm-300mm focal lengths):**
```bash
mkdir -p astrometry-data && cd astrometry-data
wget http://data.astrometry.net/4100/index-4110.fits  # 3.0° - 4.2°
wget http://data.astrometry.net/4100/index-4111.fits  # 2.2° - 3.0°
wget http://data.astrometry.net/4100/index-4112.fits  # 1.6° - 2.2°
wget http://data.astrometry.net/4100/index-4113.fits  # 1.1° - 1.6°
```

## Deployment

### Production Deployment

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  api:
    image: ghcr.io/diarmuidkelly/astrometry-api-server:latest
    restart: always
    ports:
      - "8080:8080"
    environment:
      - ASTROMETRY_INDEX_PATH=/data/indexes
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - /opt/astrometry-data:/data/indexes:ro
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 2G
```

### Kubernetes

Example deployment (see `examples/kubernetes/` for full manifests):

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: astrometry-api-server
spec:
  replicas: 2
  template:
    spec:
      containers:
      - name: api
        image: ghcr.io/diarmuidkelly/astrometry-api-server:latest
        ports:
        - containerPort: 8080
        env:
        - name: ASTROMETRY_INDEX_PATH
          value: /data/indexes
        volumeMounts:
        - name: docker-socket
          mountPath: /var/run/docker.sock
        - name: indexes
          mountPath: /data/indexes
          readOnly: true
```

## Development

### Running Tests

```bash
make test              # Run all tests
make test-coverage     # Generate coverage report
make lint              # Run linter
```

### Building

```bash
make build             # Build binary
make docker-build      # Build Docker image
make all               # Run all checks and build
```

### Project Structure

```
.
├── cmd/
│   └── server/          # Main server application
├── internal/
│   ├── handlers/        # HTTP handlers
│   └── middleware/      # HTTP middleware
├── scripts/             # Build and release scripts
├── .github/
│   └── workflows/       # CI/CD workflows
├── Dockerfile           # Multi-stage Docker build
├── docker-compose.yml   # Local development setup
└── Makefile            # Build commands
```

## Performance Tips

1. **Use scale bounds**: Providing `scale_low` and `scale_high` significantly speeds up solving
2. **Downsample large images**: Use `downsample_factor=2` or higher for megapixel images
3. **Provide RA/Dec hints**: If approximate coordinates are known, use `ra`, `dec`, and `radius` parameters
4. **Choose appropriate indexes**: Only download index files matching your typical field of view

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for workflow details.

**Quick Start:**
1. Fork the repository
2. Create a feature branch (`git checkout -b feat/my-feature`)
3. Make your changes with conventional commits
4. Push and create a PR with `[MAJOR]`, `[MINOR]`, or `[PATCH]` prefix

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

## Links

- [Changelog](CHANGELOG.md)
- [Contributing Guide](CONTRIBUTING.md)
- [Issues](https://github.com/DiarmuidKelly/Astrometry-API-Server/issues)
- [Astrometry-Go-Client Library](https://github.com/DiarmuidKelly/Astrometry-Go-Client)
- [Astrometry.net](http://astrometry.net/)

## Acknowledgments

- [Astrometry.net](http://astrometry.net/) project for the plate-solving engine
- [dm90/astrometry](https://hub.docker.com/r/dm90/astrometry) for the containerized version
- [Astrometry-Go-Client](https://github.com/DiarmuidKelly/Astrometry-Go-Client) for the Go library
