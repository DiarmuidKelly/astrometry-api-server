package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLogger_Success(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(log.Writer())

	handler := Logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test-path", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Verify log output
	logOutput := buf.String()
	if !strings.Contains(logOutput, "GET") {
		t.Errorf("expected log to contain 'GET', got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "/test-path") {
		t.Errorf("expected log to contain '/test-path', got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "200") {
		t.Errorf("expected log to contain '200', got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "127.0.0.1:12345") {
		t.Errorf("expected log to contain '127.0.0.1:12345', got: %s", logOutput)
	}
}

func TestLogger_ErrorResponse(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(log.Writer())

	handler := Logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodPost, "/missing", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}

	// Verify log output contains error status
	logOutput := buf.String()
	if !strings.Contains(logOutput, "404") {
		t.Errorf("expected log to contain '404', got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "POST") {
		t.Errorf("expected log to contain 'POST', got: %s", logOutput)
	}
}

func TestLogger_DefaultStatusCode(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(log.Writer())

	// Handler that doesn't explicitly set status code
	handler := Logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("no explicit status"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Verify default status is logged as 200
	logOutput := buf.String()
	if !strings.Contains(logOutput, "200") {
		t.Errorf("expected log to contain '200', got: %s", logOutput)
	}
}

func TestResponseWriter_WriteHeader(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}

	rw.WriteHeader(http.StatusCreated)

	if rw.statusCode != http.StatusCreated {
		t.Errorf("expected statusCode 201, got %d", rw.statusCode)
	}

	if w.Code != http.StatusCreated {
		t.Errorf("expected wrapped response code 201, got %d", w.Code)
	}
}
