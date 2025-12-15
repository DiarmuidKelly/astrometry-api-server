package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	client "github.com/DiarmuidKelly/astrometry-go-client"
)

func TestSolveHandler_Success(t *testing.T) {
	// Ensure /shared-data exists for the test
	if err := os.MkdirAll("/shared-data", 0755); err != nil {
		t.Skip("Cannot create /shared-data directory, skipping test")
	}

	mockClient := &MockAstroClient{
		SolveFunc: func(ctx context.Context, imagePath string, opts *client.SolveOptions) (*client.Result, error) {
			return &client.Result{
				Solved:     true,
				RA:         83.421,
				Dec:        -5.891,
				PixelScale: 3.96,
				Rotation:   22.43,
				SolveTime:  5.5,
			}, nil
		},
	}

	handler := NewSolveHandler(mockClient, 50*1024*1024)

	testImage := createTestJPEG(t)
	defer os.Remove(testImage)

	body, contentType := createMultipartRequest(t, "image", testImage)
	req := httptest.NewRequest(http.MethodPost, "/solve", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response SolveResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !response.Solved {
		t.Errorf("expected solved true, got error: %s", response.Error)
	}

	if response.RA != 83.421 {
		t.Errorf("expected RA 83.421, got %f", response.RA)
	}
}

func TestSolveHandler_NoSolution(t *testing.T) {
	// Ensure /shared-data exists for the test
	if err := os.MkdirAll("/shared-data", 0755); err != nil {
		t.Skip("Cannot create /shared-data directory, skipping test")
	}

	mockClient := &MockAstroClient{
		SolveFunc: func(ctx context.Context, imagePath string, opts *client.SolveOptions) (*client.Result, error) {
			return &client.Result{
				Solved:    false,
				SolveTime: 7.5,
				RawOutput: "Did not solve",
			}, nil
		},
	}

	handler := NewSolveHandler(mockClient, 50*1024*1024)

	testImage := createTestJPEG(t)
	defer os.Remove(testImage)

	body, contentType := createMultipartRequest(t, "image", testImage)
	req := httptest.NewRequest(http.MethodPost, "/solve", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response SolveResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Solved {
		t.Error("expected solved false")
	}

	if response.SolveTime != 7.5 {
		t.Errorf("expected solve time 7.5, got %f", response.SolveTime)
	}

	if response.RawOutput == "" {
		t.Error("expected raw output to be populated")
	}
}

func TestSolveHandler_MethodNotAllowed(t *testing.T) {
	mockClient := &MockAstroClient{}
	handler := NewSolveHandler(mockClient, 50*1024*1024)

	req := httptest.NewRequest(http.MethodGet, "/solve", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}

	var response SolveResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Solved {
		t.Error("expected solved false")
	}

	if !strings.Contains(response.Error, "Method not allowed") {
		t.Errorf("expected error about method, got: %s", response.Error)
	}
}

func TestSolveHandler_MissingImage(t *testing.T) {
	mockClient := &MockAstroClient{}
	handler := NewSolveHandler(mockClient, 50*1024*1024)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/solve", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var response SolveResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Solved {
		t.Error("expected solved false")
	}
}

func TestSolveHandler_InvalidFileType(t *testing.T) {
	mockClient := &MockAstroClient{}
	handler := NewSolveHandler(mockClient, 50*1024*1024)

	// Create a test file with invalid extension
	testFile := filepath.Join(os.TempDir(), "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	body, contentType := createMultipartRequest(t, "image", testFile)
	req := httptest.NewRequest(http.MethodPost, "/solve", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var response SolveResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Solved {
		t.Error("expected solved false")
	}

	if !strings.Contains(response.Error, "Invalid file type") {
		t.Errorf("expected error about file type, got: %s", response.Error)
	}
}

func TestSolveHandler_ParameterParsing(t *testing.T) {
	// Ensure /shared-data exists for the test
	if err := os.MkdirAll("/shared-data", 0755); err != nil {
		t.Skip("Cannot create /shared-data directory, skipping test")
	}

	var capturedOpts *client.SolveOptions
	mockClient := &MockAstroClient{
		SolveFunc: func(ctx context.Context, imagePath string, opts *client.SolveOptions) (*client.Result, error) {
			capturedOpts = opts
			return &client.Result{Solved: true}, nil
		},
	}

	handler := NewSolveHandler(mockClient, 50*1024*1024)

	testImage := createTestJPEG(t)
	defer os.Remove(testImage)

	body, contentType := createMultipartRequestWithParams(t, "image", testImage, map[string]string{
		"scale_low":  "320",
		"scale_high": "460",
		"ra":         "83.5",
		"dec":        "-5.9",
	})
	req := httptest.NewRequest(http.MethodPost, "/solve", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if capturedOpts == nil {
		t.Fatal("expected opts to be captured")
	}

	if capturedOpts.ScaleLow != 320 {
		t.Errorf("expected scale_low 320, got %f", capturedOpts.ScaleLow)
	}

	if capturedOpts.ScaleHigh != 460 {
		t.Errorf("expected scale_high 460, got %f", capturedOpts.ScaleHigh)
	}

	if capturedOpts.RA != 83.5 {
		t.Errorf("expected RA 83.5, got %f", capturedOpts.RA)
	}

	if capturedOpts.Dec != -5.9 {
		t.Errorf("expected Dec -5.9, got %f", capturedOpts.Dec)
	}
}

// Helper function to create multipart request with additional params
func createMultipartRequestWithParams(t *testing.T, fieldName, filePath string, params map[string]string) (io.Reader, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		t.Fatalf("failed to copy file: %v", err)
	}

	// Add params
	for key, value := range params {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("failed to write field %s: %v", key, err)
		}
	}

	writer.Close()
	return body, writer.FormDataContentType()
}
