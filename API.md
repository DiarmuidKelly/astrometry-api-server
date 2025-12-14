# Astrometry API Server - API Documentation

Version: 0.1.0

Base URL: `http://localhost:8080`

## Table of Contents

- [Overview](#overview)
- [Endpoints](#endpoints)
  - [POST /solve](#post-solve)
  - [GET /health](#get-health)
- [Data Models](#data-models)
- [Error Handling](#error-handling)
- [Examples](#examples)

---

## Overview

The Astrometry API Server provides a REST API for plate-solving astronomical images using Astrometry.net. Upload an image and receive celestial coordinates, field of view, and World Coordinate System (WCS) information.

**Features:**
- Plate-solving for astronomical images
- Support for JPEG, PNG, and FITS formats
- Automatic WCS header extraction
- Field of view calculations
- Configurable solve parameters

---

## Endpoints

### POST /analyse

Analyzes an image to extract camera details from EXIF and calculate field of view. This is a fast operation that does NOT perform plate-solving - use it to determine recommended scale parameters before calling `/solve`.

**URL:** `/analyse`

**Method:** `POST`

**Content-Type:** `multipart/form-data`

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `image` | file | **Yes** | Image file to analyze (jpg, jpeg, png) |

**Response:**

**Success (200 OK):**

```json
{
  "success": true,
  "make": "Canon",
  "model": "Canon EOS M50m2",
  "focal_length": 200,
  "sensor_name": "APS-C Canon",
  "detected_from": "exif",
  "fov": {
    "width_degrees": 6.38,
    "height_degrees": 4.27,
    "width_arcmin": 382.9,
    "height_arcmin": 256.0,
    "diagonal_degrees": 7.67
  },
  "scale_low": 319,
  "scale_high": 459,
  "scale_units": "arcminwidth",
  "has_exif": true
}
```

**No EXIF (200 OK):**

```json
{
  "success": false,
  "has_exif": false,
  "error": "failed to decode EXIF data: ..."
}
```

**Status Codes:**

| Code | Description |
|------|-------------|
| 200 | Request processed (check `success` field) |
| 400 | Bad request (invalid file or missing image field) |
| 405 | Method not allowed (use POST) |
| 413 | File too large (max 50MB) |

---

### POST /solve

Performs plate-solving on an uploaded astronomical image.

**URL:** `/solve`

**Method:** `POST`

**Content-Type:** `multipart/form-data`

**Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `image` | file | **Yes** | - | Image file to solve (jpg, jpeg, png, fits, fit) |
| `scale_low` | float | No | - | Lower bound of image scale |
| `scale_high` | float | No | - | Upper bound of image scale |
| `scale_units` | string | No | `arcminwidth` | Units for scale bounds (`degwidth`, `arcminwidth`, `arcsecperpix`) |
| `downsample_factor` | int | No | `2` | Downsample factor (higher = faster but less accurate) |
| `depth_low` | int | No | `10` | Minimum number of quads to try |
| `depth_high` | int | No | `20` | Maximum number of quads to try |
| `ra` | float | No | - | Right Ascension hint in degrees (J2000) |
| `dec` | float | No | - | Declination hint in degrees (J2000) |
| `radius` | float | No | - | Search radius in degrees (requires ra/dec) |
| `keep_temp_files` | boolean | No | `false` | Preserve temporary files for debugging |

**Response:**

**Success (200 OK):**

```json
{
  "solved": true,
  "ra": 82.853594079,
  "dec": -6.19791638337,
  "pixel_scale": 3.66389661672,
  "rotation": 15.5,
  "field_width": 6.1064943612,
  "field_height": 4.0709962408,
  "wcs_header": {
    "CRVAL1": "82.853594079",
    "CRVAL2": "-6.19791638337",
    "CD1_1": "-0.0010177490602",
    ...
  },
  "solve_time": 6.304580955
}
```

**No Solution (200 OK):**

```json
{
  "solved": false
}
```

**Error (4xx/5xx):**

```json
{
  "solved": false,
  "error": "Error description"
}
```

**Status Codes:**

| Code | Description |
|------|-------------|
| 200 | Request processed (check `solved` field for success) |
| 400 | Bad request (invalid parameters or file) |
| 405 | Method not allowed (use POST) |
| 413 | File too large (max 50MB) |
| 500 | Internal server error |

---

### GET /health

Health check endpoint.

**URL:** `/health`

**Method:** `GET`

**Response:**

```json
{
  "status": "healthy",
  "uptime_seconds": 123.45,
  "version": "0.1.0"
}
```

**Status Codes:**

| Code | Description |
|------|-------------|
| 200 | Service is healthy |
| 405 | Method not allowed (use GET) |

---

## Data Models

### SolveResponse

| Field | Type | Description |
|-------|------|-------------|
| `solved` | boolean | Whether the image was successfully plate-solved |
| `ra` | float | Right Ascension of image center in degrees (J2000) |
| `dec` | float | Declination of image center in degrees (J2000) |
| `pixel_scale` | float | Image scale in arcseconds per pixel |
| `rotation` | float | Field rotation in degrees |
| `field_width` | float | Field of view width in degrees |
| `field_height` | float | Field of view height in degrees |
| `wcs_header` | object | Raw WCS header fields from FITS file |
| `solve_time` | float | Duration of solve operation in seconds |
| `error` | string | Error message (only present if solve failed) |

### HealthResponse

| Field | Type | Description |
|-------|------|-------------|
| `status` | string | Health status ("healthy") |
| `uptime_seconds` | float | Server uptime in seconds |
| `version` | string | API version |

---

## Error Handling

All errors are returned with appropriate HTTP status codes and a JSON response:

```json
{
  "solved": false,
  "error": "Error description"
}
```

**Common Errors:**

| Error | Status | Cause |
|-------|--------|-------|
| `Missing or invalid 'image' field` | 400 | No image uploaded or wrong field name |
| `Invalid file type` | 400 | Unsupported file format |
| `Failed to parse form` | 400 | Malformed multipart request |
| `Failed to save file` | 500 | Server I/O error |
| `no solution found` | 200 | Image could not be solved (not an error) |
| `solve operation timed out` | 200 | Solve took longer than 5 minutes |

---

## Examples

### Recommended Workflow: Analyse then Solve

The recommended workflow is to first analyse the image to get optimal solve parameters:

**Step 1: Analyse the image**

```bash
curl -X POST -F "image=@photo.jpg" http://localhost:8080/analyse | jq
```

Response:
```json
{
  "success": true,
  "make": "Canon",
  "model": "Canon EOS M50m2",
  "focal_length": 200,
  "sensor_name": "APS-C Canon",
  "scale_low": 319,
  "scale_high": 459,
  "scale_units": "arcminwidth",
  "fov": { ... }
}
```

**Step 2: Use recommended parameters to solve**

```bash
curl -X POST \
  -F "image=@photo.jpg" \
  -F "scale_low=319" \
  -F "scale_high=459" \
  -F "scale_units=arcminwidth" \
  http://localhost:8080/solve | jq
```

This workflow is 3-5x faster than solving without scale hints!

### One-Line Analyse + Solve

Using jq to pipe the scale parameters automatically:

```bash
# Get scale parameters from analyse
SCALE=$(curl -s -X POST -F "image=@photo.jpg" http://localhost:8080/analyse | jq -r '"\(.scale_low) \(.scale_high)"')
read SCALE_LOW SCALE_HIGH <<< "$SCALE"

# Solve with those parameters
curl -X POST \
  -F "image=@photo.jpg" \
  -F "scale_low=$SCALE_LOW" \
  -F "scale_high=$SCALE_HIGH" \
  -F "scale_units=arcminwidth" \
  http://localhost:8080/solve | jq
```

### Basic Solve (without scale hints)

```bash
curl -X POST \
  -F "image=@photo.jpg" \
  http://localhost:8080/solve | jq
```

Note: This will be slower and may not solve if the image is outside the index coverage (0.1° - 11.0°).

### Solve with Position Hint

If you know approximately where the image points:

```bash
curl -X POST \
  -F "image=@m42.jpg" \
  -F "ra=83.8" \
  -F "dec=-5.4" \
  -F "radius=10" \
  http://localhost:8080/solve | jq
```

### Calculate Scale Parameters from EXIF

Use the Go client library to analyze your image first:

```go
package main

import (
    "fmt"
    "log"
    "github.com/DiarmuidKelly/Astrometry-Go-Client/pkg/solver/fov"
)

func main() {
    info, err := fov.AnalyzeImage("photo.jpg")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Camera: %s %s\n", info.Make, info.Model)
    fmt.Printf("Focal Length: %.0fmm\n", info.FocalLength)
    fmt.Printf("Recommended scale: %.0f-%.0f arcminwidth\n",
        info.ScaleLow, info.ScaleHigh)
}
```

Then use the recommended parameters:

```bash
curl -X POST \
  -F "image=@photo.jpg" \
  -F "scale_low=319" \
  -F "scale_high=459" \
  -F "scale_units=arcminwidth" \
  http://localhost:8080/solve | jq
```

### Health Check

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "healthy",
  "uptime_seconds": 3600.5,
  "version": "0.1.0"
}
```

### Extract Specific Fields with jq

Get only coordinates:

```bash
curl -s -X POST -F "image=@photo.jpg" \
  http://localhost:8080/solve | jq '{solved, ra, dec}'
```

Check if solve succeeded:

```bash
curl -s -X POST -F "image=@photo.jpg" \
  http://localhost:8080/solve | jq '.solved'
```

---

## Field of View Coverage

The current index files (4107-4119) cover:

- **FOV Range:** 0.1° to 11.0°
- **Focal Length (APS-C):** ~100mm to 2000mm
- **Focal Length (Full Frame):** ~135mm to 2700mm

For wider fields of view, additional 4200-series index files are required.

### Determine Required Indexes

Use the FOV calculator:

```go
package main

import (
    "fmt"
    "github.com/DiarmuidKelly/Astrometry-Go-Client/pkg/solver/fov"
)

func main() {
    // Calculate FOV for 50-300mm zoom on APS-C
    recommendation := fov.RecommendIndexesForLens(50, 300, fov.APSCCanon, 1.2)
    fmt.Println(recommendation.String())
}
```

---

## Rate Limiting

Currently no rate limiting is implemented. Each solve operation can take 2-30 seconds depending on image complexity and parameters.

**Recommended:** Implement your own rate limiting if exposing publicly.

---

## Authentication

Currently no authentication is required.

**Recommended:** Add authentication (API keys, JWT, etc.) before deploying to production.

---

## CORS

CORS is enabled for all origins in development. Configure appropriately for production use.

---

## Timeout

Solve operations timeout after **5 minutes**. If your images consistently timeout:

1. Use scale hints (`scale_low`, `scale_high`)
2. Increase `downsample_factor` (2-4)
3. Use position hints if known (`ra`, `dec`, `radius`)
4. Ensure you have the correct index files for your FOV

---

## Support

For issues or questions:
- GitHub: https://github.com/DiarmuidKelly/Astrometry-API-Server
- Documentation: See README.md

---

## License

GPL-3.0 license
