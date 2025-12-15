package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	client "github.com/DiarmuidKelly/astrometry-go-client"
)

// AstrometryClient interface for testing
type AstrometryClient interface {
	Solve(ctx context.Context, imagePath string, opts *client.SolveOptions) (*client.Result, error)
}

// SolveHandler handles plate-solving requests
type SolveHandler struct {
	client        AstrometryClient
	maxUploadSize int64
}

// NewSolveHandler creates a new solve handler
func NewSolveHandler(c AstrometryClient, maxUploadSize int64) *SolveHandler {
	return &SolveHandler{
		client:        c,
		maxUploadSize: maxUploadSize,
	}
}

// SolveRequest represents the solve request parameters
type SolveRequest struct {
	ScaleLow         float64 `json:"scale_low"`
	ScaleHigh        float64 `json:"scale_high"`
	ScaleUnits       string  `json:"scale_units"`
	DownsampleFactor int     `json:"downsample_factor"`
	DepthLow         int     `json:"depth_low"`
	DepthHigh        int     `json:"depth_high"`
	RA               float64 `json:"ra"`
	Dec              float64 `json:"dec"`
	Radius           float64 `json:"radius"`
}

// SolveResponse represents the solve response
type SolveResponse struct {
	Solved      bool              `json:"solved"`
	RA          float64           `json:"ra,omitempty"`
	Dec         float64           `json:"dec,omitempty"`
	PixelScale  float64           `json:"pixel_scale,omitempty"`
	Rotation    float64           `json:"rotation,omitempty"`
	FieldWidth  float64           `json:"field_width,omitempty"`
	FieldHeight float64           `json:"field_height,omitempty"`
	WCSHeader   map[string]string `json:"wcs_header,omitempty"`
	SolveTime   float64           `json:"solve_time,omitempty"`
	RawOutput   string            `json:"raw_output,omitempty"`
	Error       string            `json:"error,omitempty"`
}

// ServeHTTP godoc
//
//	@Summary		Plate-solve an astronomical image using offline Astrometry.net engine
//	@Description	Performs plate-solving using the offline Astrometry.net solving engine to determine celestial coordinates and orientation. Recommended: First call /analyse to get optimal scale parameters for 3-5x faster solving.
//	@Tags			Solving
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			image				formData	file			true	"Image file (JPG, JPEG, PNG, FITS, FIT)"
//	@Param			scale_low			formData	number			false	"Lower bound of image scale"
//	@Param			scale_high			formData	number			false	"Upper bound of image scale"
//	@Param			scale_units			formData	string			false	"Units for scale bounds (degwidth, arcminwidth, arcsecperpix)"	default(arcminwidth)
//	@Param			downsample_factor	formData	int				false	"Downsample factor (higher = faster but less accurate)"	default(2)
//	@Param			depth_low			formData	int				false	"Minimum number of quads to try"	default(10)
//	@Param			depth_high			formData	int				false	"Maximum number of quads to try"	default(20)
//	@Param			ra					formData	number			false	"Right Ascension hint in degrees (J2000)"
//	@Param			dec					formData	number			false	"Declination hint in degrees (J2000)"
//	@Param			radius				formData	number			false	"Search radius in degrees (requires ra/dec)"
//	@Param			keep_temp_files		formData	boolean			false	"Preserve temporary files for debugging"	default(false)
//	@Success		200					{object}	SolveResponse	"Solve complete (check solved field)"
//	@Failure		400					{object}	SolveResponse	"Bad request"
//	@Failure		405					{object}	SolveResponse	"Method not allowed"
//	@Failure		413					{object}	SolveResponse	"File too large"
//	@Failure		500					{object}	SolveResponse	"Internal server error"
//	@Router			/solve [post]
func (h *SolveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit upload size
	r.Body = http.MaxBytesReader(w, r.Body, h.maxUploadSize)

	// Parse multipart form
	if err := r.ParseMultipartForm(h.maxUploadSize); err != nil {
		respondError(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("image")
	if err != nil {
		respondError(w, "Missing or invalid 'image' field", http.StatusBadRequest)
		return
	}
	defer file.Close() //nolint:errcheck // Error from Close on read is not critical

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	validExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".fits": true, ".fit": true}
	if !validExts[ext] {
		respondError(w, "Invalid file type. Supported: jpg, jpeg, png, fits, fit", http.StatusBadRequest)
		return
	}

	// Save to temporary file in shared directory (must match client's TempDir config)
	tempDir := "/shared-data"
	tempFile := filepath.Join(tempDir, fmt.Sprintf("astro_%d%s", os.Getpid(), ext))
	defer os.Remove(tempFile) //nolint:errcheck // Cleanup failure is not critical

	out, err := os.Create(tempFile)
	if err != nil {
		respondError(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer out.Close() //nolint:errcheck // Deferred close errors are not critical

	if _, err := io.Copy(out, file); err != nil {
		respondError(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	if err := out.Close(); err != nil {
		respondError(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Parse solve options from form fields
	opts := h.parseSolveOptions(r)

	// Solve the image
	log.Printf("Solving image: %s (%.2f KB)", header.Filename, float64(header.Size)/1024)
	result, err := h.client.Solve(r.Context(), tempFile, opts)

	// Prepare response
	response := &SolveResponse{}
	if err != nil {
		log.Printf("Solve failed: %v", err)
		response.Solved = false
		response.Error = err.Error()
	} else {
		response.Solved = result.Solved
		response.SolveTime = result.SolveTime
		response.RawOutput = result.RawOutput
		if result.Solved {
			response.RA = result.RA
			response.Dec = result.Dec
			response.PixelScale = result.PixelScale
			response.Rotation = result.Rotation
			response.FieldWidth = result.FieldWidth
			response.FieldHeight = result.FieldHeight
			response.WCSHeader = result.WCSHeader
			log.Printf("Solved: RA=%.6f, Dec=%.6f, PixelScale=%.2f, Time=%.2fs",
				result.RA, result.Dec, result.PixelScale, result.SolveTime)
		} else {
			log.Printf("No solution found (Time=%.2fs)", result.SolveTime)
		}
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func (h *SolveHandler) parseSolveOptions(r *http.Request) *client.SolveOptions {
	opts := client.DefaultSolveOptions()

	// Parse optional parameters
	if val := r.FormValue("scale_low"); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			opts.ScaleLow = f
		}
	}
	if val := r.FormValue("scale_high"); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			opts.ScaleHigh = f
		}
	}
	if val := r.FormValue("scale_units"); val != "" {
		opts.ScaleUnits = val
	}
	if val := r.FormValue("downsample_factor"); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			opts.DownsampleFactor = i
		}
	}
	if val := r.FormValue("depth_low"); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			opts.DepthLow = i
		}
	}
	if val := r.FormValue("depth_high"); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			opts.DepthHigh = i
		}
	}
	if val := r.FormValue("ra"); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			opts.RA = f
		}
	}
	if val := r.FormValue("dec"); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			opts.Dec = f
		}
	}
	if val := r.FormValue("radius"); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			opts.Radius = f
		}
	}
	if val := r.FormValue("keep_temp_files"); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			opts.KeepTempFiles = b
		}
	}

	return opts
}

func respondError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(&SolveResponse{ //nolint:errcheck // Already in error path, encoding failure indicates connection issue
		Solved: false,
		Error:  message,
	})
}
