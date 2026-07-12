package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVersionHandler_Success(t *testing.T) {
	handler := NewVersionHandler()

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should return 200 OK
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Should return JSON
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	// Should decode to VersionResponse
	var response VersionResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Version should be set (or error if binary not available)
	if response.Version == "" && response.Error == "" {
		t.Error("expected either version or error to be set")
	}

	// If version is set, it should be non-empty
	if response.Version != "" {
		t.Logf("Solver version: %s", response.Version)
	}

	// If error is set, log it (not a failure - binary might not be available in test env)
	if response.Error != "" {
		t.Logf("Version check returned error (expected in test env): %s", response.Error)
	}
}

func TestVersionHandler_ResponseStructure(t *testing.T) {
	handler := NewVersionHandler()

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	var response VersionResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify the response has the expected structure
	// Even if there's an error, the structure should be correct
	if response.Version == "" && response.Error == "" {
		t.Error("response should have either version or error field populated")
	}
}

func TestVersionResponse_JSONMarshalling(t *testing.T) {
	tests := []struct {
		name     string
		response VersionResponse
		expected string
	}{
		{
			name: "successful version response",
			response: VersionResponse{
				Version: "0.98-21-gf33d1e76",
			},
			expected: `{"version":"0.98-21-gf33d1e76"}`,
		},
		{
			name: "error response",
			response: VersionResponse{
				Error: "command failed",
			},
			expected: `{"version":"","error":"command failed"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.response)
			if err != nil {
				t.Fatalf("failed to marshal response: %v", err)
			}

			// Verify it's valid JSON
			var decoded VersionResponse
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			// Verify fields match
			if decoded.Version != tt.response.Version {
				t.Errorf("expected version %s, got %s", tt.response.Version, decoded.Version)
			}

			if decoded.Error != tt.response.Error {
				t.Errorf("expected error %s, got %s", tt.response.Error, decoded.Error)
			}
		})
	}
}
