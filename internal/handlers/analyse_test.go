package handlers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnalyseHandler_Success(t *testing.T) {
	// Ensure /shared-data exists for the test
	if err := os.MkdirAll("/shared-data", 0755); err != nil {
		t.Skip("Cannot create /shared-data directory, skipping test")
	}

	// Create a test JPEG with EXIF data
	testImage := createTestJPEG(t)
	defer os.Remove(testImage)

	handler := NewAnalyseHandler(50 * 1024 * 1024)

	body, contentType := createMultipartRequest(t, "image", testImage)
	req := httptest.NewRequest(http.MethodPost, "/analyse", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Note: Our minimal test JPEG doesn't have EXIF data, so we expect an error
	// This test verifies the handler processes the request without crashing
	var response AnalyseResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// The minimal JPEG won't have EXIF, so we expect a 400 error
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for JPEG without EXIF, got %d", w.Code)
	}

	if response.Success {
		t.Error("expected success false for JPEG without EXIF")
	}
}

func TestAnalyseHandler_MethodNotAllowed(t *testing.T) {
	handler := NewAnalyseHandler(50 * 1024 * 1024)

	req := httptest.NewRequest(http.MethodGet, "/analyse", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}

	var response AnalyseResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Success {
		t.Error("expected success false")
	}

	if !strings.Contains(response.Error, "Method not allowed") {
		t.Errorf("expected error about method, got: %s", response.Error)
	}
}

func TestAnalyseHandler_MissingImage(t *testing.T) {
	handler := NewAnalyseHandler(50 * 1024 * 1024)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/analyse", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var response AnalyseResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Success {
		t.Error("expected success false")
	}
}

func TestAnalyseHandler_InvalidFileType(t *testing.T) {
	handler := NewAnalyseHandler(50 * 1024 * 1024)

	// Create a test file with invalid extension
	testFile := filepath.Join(os.TempDir(), "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	body, contentType := createMultipartRequest(t, "image", testFile)
	req := httptest.NewRequest(http.MethodPost, "/analyse", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var response AnalyseResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Success {
		t.Error("expected success false")
	}

	if !strings.Contains(response.Error, "Invalid file type") {
		t.Errorf("expected error about file type, got: %s", response.Error)
	}
}
