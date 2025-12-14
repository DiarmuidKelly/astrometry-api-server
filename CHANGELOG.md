# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2024-01-01

### Added

- Initial release of Astrometry API Server
- RESTful HTTP API for plate-solving astronomical images
- `POST /solve` endpoint for solving images via multipart upload
- `GET /health` endpoint for health checks
- Support for configurable solve parameters:
  - Scale bounds (scale_low, scale_high, scale_units)
  - Downsampling factor
  - Depth parameters (depth_low, depth_high)
  - RA/Dec position hints with search radius
- Request logging middleware
- CORS support for web applications
- Docker-based deployment with multi-stage builds
- Multi-platform Docker images (linux/amd64, linux/arm64)
- Docker Compose setup for local development
- Graceful shutdown handling
- Comprehensive documentation:
  - README with API reference and examples
  - CONTRIBUTING guide with PR workflow
  - Example usage in cURL, Python, and JavaScript
- Automated CI/CD workflows:
  - PR title validation
  - Automatic PR labeling
  - Go tests on multiple platforms (Ubuntu, macOS)
  - Go linting with golangci-lint
  - Docker image building and publishing
  - Automatic releases with version bumping
- Integration with Astrometry-Go-Client library
- Support for multiple image formats (JPG, PNG, FITS)
- JSON response format with solve results and WCS headers

[0.1.0]: https://github.com/DiarmuidKelly/Astrometry-API-Server/releases/tag/v0.1.0
