# Build stage
FROM golang:1.24-alpine AS builder
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
