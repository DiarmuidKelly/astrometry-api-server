# Astrometry Solver Architecture Proposal

## Current Architecture

```
┌─────────────────────┐
│  Your Application   │
└──────────┬──────────┘
           │ imports
           ▼
┌─────────────────────┐
│  astrometry-go-     │
│  client             │
│  (docker exec)      │
└──────────┬──────────┘
           │ docker.sock 🔓
           ▼
┌─────────────────────┐
│  astrometry-        │
│  dockerised-solver  │
│  (CLI container)    │
└─────────────────────┘
```

**Problem:** Requires docker.sock access, which is a security risk.

---

## Proposed Architecture

```
┌─────────────────────┐
│  Your Application   │
└──────────┬──────────┘
           │ imports
           ▼
┌─────────────────────┐
│  astrometry-go-     │
│  client             │
│  (HTTP client)      │
└──────────┬──────────┘
           │ HTTP (internal network)
           ▼
┌──────────────────────────────────────┐
│  astrometry-api-server               │
│  (FROM astrometry-dockerised-solver) │
│  ┌────────────────────────────────┐  │
│  │ api-server binary (HTTP)       │  │
│  │ └── exec.Command("solve-field")│  │
│  └────────────────────────────────┘  │
│  ┌────────────────────────────────┐  │
│  │ solve-field + deps             │  │
│  └────────────────────────────────┘  │
└──────────────────────────────────────┘

┌─────────────────────┐
│  astrometry-        │
│  dockerised-solver  │  ← Unchanged, still available
│  (CLI container)    │     for community CLI use
└─────────────────────┘
```

**Benefit:** No docker.sock. Solver is local to the API server.

---

## Repository Responsibilities

| Repo                           | Current                              | Proposed                         |
| ------------------------------ | ------------------------------------ | -------------------------------- |
| `astrometry-dockerised-solver` | Pure CLI image                       | Pure CLI image (unchanged)       |
| `astrometry-go-client`         | Docker exec wrapper                  | HTTP client                      |
| `astrometry-api-server`        | Imports go-client, calls docker exec | HTTP server, direct exec.Command |

---

## Build Process

### astrometry-dockerised-solver (unchanged)

```dockerfile
# Multi-stage build for minimal solver image
FROM ubuntu:24.04 AS builder
# ... build solve-field ...

FROM ubuntu:24.04
# ... runtime deps + solve-field binary only ...
ENTRYPOINT ["solve-field"]
```

### astrometry-api-server (new approach)

```dockerfile
# Stage 1: Build Go binary (disposable)
FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o api-server .

# Stage 2: Extend solver image
FROM yourname/astrometry-solver:latest
COPY --from=builder /src/api-server /usr/local/bin/
EXPOSE 8082
ENTRYPOINT ["api-server"]
```

**Result:** Solver image + ~10MB static Go binary. No Go runtime needed.

---

## Concern: API Server Becomes Heavier?

**Yes, but manageable.**

The API server must implement handlers for each solve-field capability you want to expose. However:

### Option A: Thin Passthrough (Recommended)

Expose a generic endpoint that accepts arbitrary (validated) arguments:

```go
type SolveRequest struct {
    ImageData   []byte            `json:"image_data"`
    Filename    string            `json:"filename"`
    Args        []string          `json:"args"`        // e.g., ["--scale-low", "1"]
}

func solveHandler(w http.ResponseWriter, r *http.Request) {
    var req SolveRequest
    json.NewDecoder(r.Body).Decode(&req)

    // Validate args against allowlist
    if !validateArgs(req.Args) {
        http.Error(w, "invalid arguments", 400)
        return
    }

    // Save image to temp file
    tmpFile := saveTemp(req.ImageData, req.Filename)
    defer os.Remove(tmpFile)

    // Build command
    args := append([]string{tmpFile}, req.Args...)
    cmd := exec.Command("solve-field", args...)
    output, err := cmd.CombinedOutput()

    // Return results
    json.NewEncoder(w).Encode(SolveResponse{
        Output: string(output),
        WCS:    readWCSFile(tmpFile),
    })
}
```

**Benefit:** API server stays thin. New solve-field features work automatically.

### Option B: Typed Endpoints

Explicit endpoints for each operation:

```go
POST /solve/blind      // Full blind solve
POST /solve/hint       // With RA/Dec/scale hints
POST /solve/verify     // Verify existing WCS
```

**Benefit:** Better validation, clearer API. More maintenance.

### Recommendation

Start with **Option A** (passthrough with validation). Add typed endpoints later for common operations if needed.

---

## go-client Changes

Current (docker exec):

```go
func (c *Client) Solve(image string, opts Options) (*Result, error) {
    args := []string{"exec", c.containerID, "solve-field", image}
    args = append(args, opts.ToArgs()...)
    cmd := exec.Command("docker", args...)
    // ...
}
```

Proposed (HTTP):

```go
func (c *Client) Solve(image []byte, opts Options) (*Result, error) {
    req := SolveRequest{
        ImageData: image,
        Args:      opts.ToArgs(),
    }
    resp, err := http.Post(c.baseURL+"/solve", "application/json", toJSON(req))
    // ...
}
```

**Same interface to consumers**, different transport underneath.

---

## Deployment Example

```yaml
# docker-compose.yml
services:
  solver-api:
    image: yourname/astrometry-api-server:latest
    expose:
      - "8082" # Internal only
    volumes:
      - ./indexes:/usr/local/data:ro
    networks:
      - internal

  your-app:
    image: your-app
    environment:
      - SOLVER_URL=http://solver-api:8082
    networks:
      - internal

networks:
  internal:
    internal: true # No external access
```

---

## Summary

| Aspect                   | Current                 | Proposed                         |
| ------------------------ | ----------------------- | -------------------------------- |
| Security                 | docker.sock required    | No docker.sock                   |
| Solver image             | Pure CLI                | Pure CLI (unchanged)             |
| API server               | Thin, uses go-client    | Self-contained, local exec       |
| go-client                | Docker exec             | HTTP client                      |
| New solve-field features | Automatic               | Automatic (with passthrough)     |
| Deployment               | Two containers + socket | Two containers, internal network |

---

## Migration Path

1. Update `astrometry-go-client` to HTTP-based client
2. Update `astrometry-api-server` to:
   - Build FROM solver image
   - Direct exec.Command instead of go-client import
3. Update deployment to remove docker.sock mount
4. `astrometry-dockerised-solver` unchanged — still works for CLI users

---

## Open Questions

- [ ] Authentication for API server? (JWT, API key, network isolation only?)
- [ ] Rate limiting / queue for concurrent solves?
- [ ] Health check endpoint?
- [ ] Structured output format? (JSON WCS vs raw files)
