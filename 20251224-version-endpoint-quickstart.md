# Version Endpoint Quick Start

## Validation Results

✅ **Container works**: `diarmuidk/astrometry-dockerised-solver:latest`
✅ **Version command works**: `solve-field --version` returns `0.98-21-gf33d1e76`

```bash
docker run --rm diarmuidk/astrometry-dockerised-solver:latest solve-field --version
# Output: 0.98-21-gf33d1e76
```

## MVP Implementation (~15 mins)

### 1. Add Version Handler (`internal/handlers/version.go`)

```go
package handlers

import (
	"encoding/json"
	"net/http"
	"os/exec"
)

type VersionHandler struct{}

func NewVersionHandler() *VersionHandler {
	return &VersionHandler{}
}

type VersionResponse struct {
	Version string `json:"version"`
	Error   string `json:"error,omitempty"`
}

func (h *VersionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("solve-field", "--version")
	output, err := cmd.CombinedOutput()

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		json.NewEncoder(w).Encode(VersionResponse{
			Error: err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(VersionResponse{
		Version: string(output),
	})
}
```

### 2. Wire Up in `cmd/server/main.go`

```go
// Add to imports if needed
// After line 68 (after analyseHandler)
versionHandler := handlers.NewVersionHandler()

// Add route after line 74
mux.Handle("/version", middleware.Logger(versionHandler))
```

### 3. Update Dockerfile to Build FROM Solver

```dockerfile
# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /build/server ./cmd/server

# Runtime stage - BUILD FROM SOLVER IMAGE
FROM diarmuidk/astrometry-dockerised-solver:latest
COPY --from=builder /build/server /app/server
EXPOSE 8080
ENTRYPOINT ["/app/server"]
```

### 4. Test

```bash
# Build
docker build -t astrometry-api-server:test .

# Run
docker run --rm -p 8080:8080 astrometry-api-server:test

# Test endpoint
curl http://localhost:8080/version
# Expected: {"version":"0.98-21-gf33d1e76\n"}
```

## What This Proves

✅ API server can `exec.Command` directly to solver binaries
✅ Building FROM solver base image works
✅ No docker.sock needed
✅ Foundation ready for full `/solve` implementation
