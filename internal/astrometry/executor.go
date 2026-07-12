package astrometry

import (
	"fmt"
	"os/exec"
	"strings"
)

// Binary name constants for astrometry.net solver tools
const (
	SolveField = "solve-field" // Main plate solving command
	Image2XY   = "image2xy"    // Extract sources from images
	FitWCS     = "fit-wcs"     // Fit WCS to xy lists
	WcsXY2RD   = "wcs-xy2rd"   // Convert XY to RA/Dec
	WcsRD2XY   = "wcs-rd2xy"   // Convert RA/Dec to XY
)

// validBinaries is the allowlist of permitted binary names
var validBinaries = map[string]bool{
	SolveField: true,
	Image2XY:   true,
	FitWCS:     true,
	WcsXY2RD:   true,
	WcsRD2XY:   true,
}

// Execute runs an astrometry binary with the given arguments.
// Returns the combined stdout/stderr output and any error encountered.
//
// The binary parameter must be one of the predefined constants (SolveField, Image2XY, etc.).
// Arguments are passed directly to the binary - caller is responsible for validation.
func Execute(binary string, args ...string) (string, error) {
	// Validate binary name against allowlist
	if !validBinaries[binary] {
		return "", fmt.Errorf("invalid binary name: %s", binary)
	}

	// Execute the command
	cmd := exec.Command(binary, args...)
	output, err := cmd.CombinedOutput()

	// Return trimmed output
	return strings.TrimSpace(string(output)), err
}
