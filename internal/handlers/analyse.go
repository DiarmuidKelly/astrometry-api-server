package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/DiarmuidKelly/astrometry-go-client/pkg/solver/fov"
)

// AnalyseHandler handles image analysis requests (EXIF extraction + FOV calculation)
type AnalyseHandler struct {
	maxUploadSize int64
}

// NewAnalyseHandler creates a new analyse handler
func NewAnalyseHandler(maxUploadSize int64) *AnalyseHandler {
	return &AnalyseHandler{
		maxUploadSize: maxUploadSize,
	}
}

// AnalyseResponse represents the image analysis response
type AnalyseResponse struct {
	Success      bool     `json:"success"`
	Make         string   `json:"make,omitempty"`
	Model        string   `json:"model,omitempty"`
	FocalLength  float64  `json:"focal_length,omitempty"`
	SensorName   string   `json:"sensor_name,omitempty"`
	DetectedFrom string   `json:"detected_from,omitempty"`
	FOV          *FOVData `json:"fov,omitempty"`
	ScaleLow     float64  `json:"scale_low,omitempty"`
	ScaleHigh    float64  `json:"scale_high,omitempty"`
	ScaleUnits   string   `json:"scale_units,omitempty"`
	HasEXIF      bool     `json:"has_exif"`
	Error        string   `json:"error,omitempty"`
}

// FOVData represents field of view information
type FOVData struct {
	WidthDegrees  float64 `json:"width_degrees"`
	HeightDegrees float64 `json:"height_degrees"`
	WidthArcmin   float64 `json:"width_arcmin"`
	HeightArcmin  float64 `json:"height_arcmin"`
	DiagonalDeg   float64 `json:"diagonal_degrees"`
}

// ServeHTTP godoc
//
//	@Summary		Analyse image EXIF and calculate FOV
//	@Description	Extracts camera information from EXIF data and calculates field of view. Returns recommended scale parameters for plate-solving. This is a fast operation (< 1 second) that does NOT perform plate-solving.
//	@Tags			Analysis
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			image	formData	file				true	"Image file (JPG, JPEG, PNG with EXIF)"
//	@Success		200		{object}	AnalyseResponse		"Analysis complete"
//	@Failure		400		{object}	AnalyseResponse		"Bad request"
//	@Failure		405		{string}	string				"Method not allowed"
//	@Failure		413		{string}	string				"File too large"
//	@Router			/analyse [post]
func (h *AnalyseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit upload size
	r.Body = http.MaxBytesReader(w, r.Body, h.maxUploadSize)

	// Parse multipart form
	if err := r.ParseMultipartForm(h.maxUploadSize); err != nil {
		respondAnalyseError(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("image")
	if err != nil {
		respondAnalyseError(w, "Missing or invalid 'image' field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	validExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
	if !validExts[ext] {
		respondAnalyseError(w, "Invalid file type. Supported: jpg, jpeg, png", http.StatusBadRequest)
		return
	}

	// Save to temporary file
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, fmt.Sprintf("analyse_%d%s", os.Getpid(), ext))
	defer os.Remove(tempFile)

	out, err := os.Create(tempFile)
	if err != nil {
		respondAnalyseError(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		respondAnalyseError(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	out.Close()

	// Analyse the image
	log.Printf("Analysing image: %s (%.2f KB)", header.Filename, float64(header.Size)/1024)
	info, err := fov.AnalyzeImage(tempFile)
	if err != nil {
		log.Printf("Analysis failed: %v", err)
		respondAnalyseError(w, fmt.Sprintf("Failed to analyse image: %v", err), http.StatusBadRequest)
		return
	}

	// Prepare response
	response := &AnalyseResponse{
		Success:      true,
		Make:         info.Make,
		Model:        info.Model,
		FocalLength:  info.FocalLength,
		SensorName:   info.Sensor.Name,
		DetectedFrom: info.DetectedFrom,
		HasEXIF:      info.HasEXIF,
		ScaleUnits:   "arcminwidth",
	}

	if info.FOV.WidthDegrees > 0 {
		response.FOV = &FOVData{
			WidthDegrees:  info.FOV.WidthDegrees,
			HeightDegrees: info.FOV.HeightDegrees,
			WidthArcmin:   info.FOV.WidthArcmin,
			HeightArcmin:  info.FOV.HeightArcmin,
			DiagonalDeg:   info.FOV.DiagonalDeg,
		}
		response.ScaleLow = info.ScaleLow
		response.ScaleHigh = info.ScaleHigh
	}

	log.Printf("Analysis complete: Camera=%s %s, FocalLength=%.0fmm, FOV=%.2f°x%.2f°",
		info.Make, info.Model, info.FocalLength, info.FOV.WidthDegrees, info.FOV.HeightDegrees)

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func respondAnalyseError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(&AnalyseResponse{
		Success: false,
		Error:   message,
	})
}
