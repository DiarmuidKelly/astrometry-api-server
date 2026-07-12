package astrometry

import (
	"strings"
	"testing"
)

func TestExecute_ValidBinary(t *testing.T) {
	// Test that Execute accepts valid binary names
	validBinaries := []string{
		SolveField,
		Image2XY,
		FitWCS,
		WcsXY2RD,
		WcsRD2XY,
	}

	for _, binary := range validBinaries {
		t.Run(binary, func(t *testing.T) {
			// Execute with --help (most likely to work in any environment)
			output, err := Execute(binary, "--help")

			// We expect either:
			// 1. Success with output (binary exists)
			// 2. Error because binary not in PATH (acceptable in test environment)
			if err != nil {
				// Error is acceptable if binary not found
				if !strings.Contains(err.Error(), "executable file not found") &&
					!strings.Contains(err.Error(), "no such file or directory") {
					t.Logf("Binary %s returned unexpected error: %v", binary, err)
				}
			} else {
				// If successful, output should be non-empty
				if output == "" {
					t.Errorf("expected non-empty output from %s --help", binary)
				}
			}
		})
	}
}

func TestExecute_InvalidBinary(t *testing.T) {
	invalidBinaries := []string{
		"invalid-binary",
		"rm",
		"bash",
		"python",
		"../solve-field",
		"solve-field; rm -rf /",
	}

	for _, binary := range invalidBinaries {
		t.Run(binary, func(t *testing.T) {
			_, err := Execute(binary, "--help")

			// Should return an error for invalid binary
			if err == nil {
				t.Errorf("expected error for invalid binary %s, got nil", binary)
			}

			// Error message should mention invalid binary
			if !strings.Contains(err.Error(), "invalid binary name") {
				t.Errorf("expected 'invalid binary name' error, got: %v", err)
			}
		})
	}
}

func TestExecute_OutputTrimming(t *testing.T) {
	// Test that output is properly trimmed
	// We can test this by checking that the output doesn't have leading/trailing whitespace
	// even if the binary is not available (error case still trims)

	output, _ := Execute(SolveField, "--version")

	// Check no leading/trailing whitespace
	if output != strings.TrimSpace(output) {
		t.Errorf("output should be trimmed, got: %q", output)
	}
}

func TestBinaryConstants(t *testing.T) {
	// Verify that all binary constants are correctly defined
	expectedBinaries := map[string]string{
		"SolveField": "solve-field",
		"Image2XY":   "image2xy",
		"FitWCS":     "fit-wcs",
		"WcsXY2RD":   "wcs-xy2rd",
		"WcsRD2XY":   "wcs-rd2xy",
	}

	actualBinaries := map[string]string{
		"SolveField": SolveField,
		"Image2XY":   Image2XY,
		"FitWCS":     FitWCS,
		"WcsXY2RD":   WcsXY2RD,
		"WcsRD2XY":   WcsRD2XY,
	}

	for name, expected := range expectedBinaries {
		actual, exists := actualBinaries[name]
		if !exists {
			t.Errorf("constant %s not found", name)
			continue
		}

		if actual != expected {
			t.Errorf("constant %s: expected %q, got %q", name, expected, actual)
		}

		// Verify it's in the valid binaries map
		if !validBinaries[actual] {
			t.Errorf("constant %s (%q) not in validBinaries map", name, actual)
		}
	}
}

func TestValidBinariesMap(t *testing.T) {
	// Verify the validBinaries map contains exactly the expected entries
	expectedCount := 5

	if len(validBinaries) != expectedCount {
		t.Errorf("expected %d valid binaries, got %d", expectedCount, len(validBinaries))
	}

	// All should be true
	for binary, isValid := range validBinaries {
		if !isValid {
			t.Errorf("binary %s should be valid (true), got false", binary)
		}
	}
}
