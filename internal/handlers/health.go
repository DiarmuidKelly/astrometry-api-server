package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	startTime time.Time
}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status  string  `json:"status"`
	Uptime  float64 `json:"uptime_seconds"`
	Version string  `json:"version"`
}

// ServeHTTP godoc
//
//	@Summary		Health check
//	@Description	Returns server health status and uptime
//	@Tags			Health
//	@Produce		json
//	@Success		200	{object}	HealthResponse	"Server is healthy"
//	@Failure		405	{string}	string			"Method not allowed"
//	@Router			/health [get]
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	uptime := time.Since(h.startTime).Seconds()

	response := &HealthResponse{
		Status:  "healthy",
		Uptime:  uptime,
		Version: "0.1.0", // This should be read from VERSION file in production
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
