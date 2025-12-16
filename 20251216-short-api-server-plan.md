# API Server - Minimal MR-Ready Implementation

## Scope
**ONLY api-server changes** (not touching go-client)

Build FROM solver base image, add HTTP endpoint with direct exec.Command

---

## Time Estimate: 3-4 hours to MR-ready

### Core Code (~1.5-2 hours)
- Basic HTTP server with `/solve` endpoint
- SolveRequest/SolveResponse structs
- Temp file handling + exec.Command("solve-field")
- Basic arg validation (allowlist for security)
- Simple WCS file reading

### Dockerfile (~15 mins)
```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o api-server .

FROM yourname/astrometry-solver:latest
COPY --from=builder /src/api-server /usr/local/bin/
EXPOSE 8082
ENTRYPOINT ["api-server"]
```

### Basic Tests (~1 hour)
- Smoke test: endpoint responds
- Test: arg validation rejects bad input
- Test: basic solve works end-to-end

### Minimal Documentation (~30 mins)
- README with API endpoint spec
- Deployment example (docker-compose)

---

## What This Achieves

✅ No docker.sock dependency
✅ Self-contained API server
✅ Proves the concept works
✅ Ready for review and iteration

---

## Out of Scope (Can Add Later)

- Comprehensive test coverage
- Production error handling
- Rate limiting/queuing
- Authentication
- Structured JSON WCS output
- Health check endpoint

---

## Key Implementation Points

1. **Argument Validation**: Allowlist approach for security
   ```go
   allowedFlags := []string{
       "--scale-low", "--scale-high", "--scale-units",
       "--ra", "--dec", "--radius",
       "--downsample", "--no-plots", "--overwrite",
   }
   ```

2. **Temp File Cleanup**: Always defer cleanup
   ```go
   tmpFile := saveTemp(req.ImageData, req.Filename)
   defer os.Remove(tmpFile)
   ```

3. **Error Handling**: Return solve-field output on failure
   ```go
   output, err := cmd.CombinedOutput()
   // Return output regardless - it contains useful error info
   ```

---

## Revision Notes

- Original estimate of "1-2 days" assumed production-grade implementation
- This plan focuses on minimal viable MR
- Can iterate on polish/features after initial merge
