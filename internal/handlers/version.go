package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/DiarmuidKelly/astrometry-api-server/internal/astrometry"
)

// VersionHandler handles version requests
type VersionHandler struct{}

// NewVersionHandler creates a new version handler
func NewVersionHandler() *VersionHandler {
	return &VersionHandler{}
}

// VersionResponse represents the version response
type VersionResponse struct {
	Version string `json:"version"`
	Error   string `json:"error,omitempty"`
}

// ServeHTTP godoc
//
//	@Summary		Get Astrometry.net solver version
//	@Description	Returns the version of the solve-field binary included in this API server. This is a fast operation that executes solve-field --version and returns the output.
//	@Tags			Health
//	@Produce		json
//	@Success		200	{object}	VersionResponse	"Version information"	example({"version":"0.98-21-gf33d1e76"})
//	@Failure		500	{object}	VersionResponse	"Failed to get version"
//	@Router			/version [get]
func (h *VersionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	output, err := astrometry.Execute(astrometry.SolveField, "--version")

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		json.NewEncoder(w).Encode(VersionResponse{
			Error: err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(VersionResponse{
		Version: output,
	})
}
