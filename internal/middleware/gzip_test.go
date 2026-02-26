package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzipCompress_NoAcceptEncoding(t *testing.T) {
	Called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Called = true
		w.WriteHeader(http.StatusOK)
	})

	wrapped := GzipCompress(handler)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if !Called {
		t.Error("handler was not called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestGzipCompress_WithAcceptEncoding(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"test": "data"}`))
	})

	wrapped := GzipCompress(handler)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	// Check that response is gzip compressed
	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Error("Content-Encoding should be gzip")
	}

	// Decompress and verify
	gzReader, err := gzip.NewReader(bytes.NewReader(rec.Body.Bytes()))
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	decompressed, err := io.ReadAll(gzReader)
	if err != nil {
		t.Fatalf("failed to decompress: %v", err)
	}
	gzReader.Close()

	if string(decompressed) != `{"test": "data"}` {
		t.Errorf("expected decompressed data '{\"test\": \"data\"}', got '%s'", string(decompressed))
	}
}

func TestGzipCompress_NotCompressibleContentType(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<xml>data</xml>`))
	})

	wrapped := GzipCompress(handler)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	// Should NOT be compressed
	if rec.Header().Get("Content-Encoding") == "gzip" {
		t.Error("Content-Encoding should not be gzip for application/xml")
	}
}

func TestGzipDecompress_NoContentEncoding(t *testing.T) {
	Called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Called = true
		w.WriteHeader(http.StatusOK)
	})

	wrapped := GzipDecompress(handler)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if !Called {
		t.Error("handler was not called")
	}
}

func TestGzipDecompress_WithGzipContent(t *testing.T) {
	Called := false
	var receivedBody []byte

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Called = true
		body, _ := io.ReadAll(r.Body)
		receivedBody = body
		w.WriteHeader(http.StatusOK)
	})

	wrapped := GzipDecompress(handler)

	// Create gzip compressed data
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	gzWriter.Write([]byte(`test request body`))
	gzWriter.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Encoding", "gzip")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if !Called {
		t.Error("handler was not called")
	}

	if string(receivedBody) != "test request body" {
		t.Errorf("expected 'test request body', got '%s'", string(receivedBody))
	}
}

func TestGzipDecompress_InvalidGzip(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := GzipDecompress(handler)

	// Send invalid gzip data
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte("not gzip data")))
	req.Header.Set("Content-Encoding", "gzip")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestGzipResponseWriter_Header(t *testing.T) {
	rec := httptest.NewRecorder()
	gz := &gzipResponseWriter{
		ResponseWriter: rec,
	}

	header := gz.Header()
	if header == nil {
		t.Error("Header returned nil")
	}
}

func TestGzipResponseWriter_WriteHeader_Twice(t *testing.T) {
	rec := httptest.NewRecorder()
	gz := &gzipResponseWriter{
		ResponseWriter: rec,
		wroteHeader:    true, // Simulate first WriteHeader was called
	}

	// Even though wroteHeader is true, it should still call WriteHeader on ResponseWriter
	gz.WriteHeader(http.StatusOK)

	// Status should still be written because we can't prevent it at the ResponseWriter level
	_ = rec.Code // Just verify no panic
}

func TestGzipResponseWriter_Write_WithoutHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	gz := &gzipResponseWriter{
		ResponseWriter: rec,
		contentType:    "application/json",
	}

	data := []byte(`{"test": "data"}`)
	n, err := gz.Write(data)

	if err != nil {
		t.Errorf("Write returned error: %v", err)
	}
	if n != len(data) {
		t.Errorf("expected %d bytes written, got %d", len(data), n)
	}
}

func TestGzipResponseWriter_Write_WithCompression(t *testing.T) {
	rec := httptest.NewRecorder()
	gz := &gzipResponseWriter{
		ResponseWriter: rec,
		contentType:    "application/json",
	}

	data := []byte(`{"test": "data"}`)
	n, err := gz.Write(data)

	if err != nil {
		t.Errorf("Write returned error: %v", err)
	}
	if n != len(data) {
		t.Errorf("expected %d bytes written, got %d", len(data), n)
	}

	// Close the gzip writer to flush data
	if gz.Writer != nil {
		gz.Writer.Close()
	}

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Error("Content-Encoding should be gzip")
	}
}

func TestShouldCompress(t *testing.T) {
	tests := []struct {
		contentType string
		expected    bool
	}{
		{"application/json", true},
		{"application/json; charset=utf-8", true},
		{"text/html", true},
		{"text/html; charset=utf-8", true},
		{"text/plain", false},
		{"application/xml", false},
		{"image/png", false},
		{"", false},
	}

	for _, tt := range tests {
		result := shouldCompress(tt.contentType)
		if result != tt.expected {
			t.Errorf("shouldCompress(%q) = %v, expected %v", tt.contentType, result, tt.expected)
		}
	}
}

func TestGzipCompress_EmptyResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	})

	wrapped := GzipCompress(handler)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", rec.Code)
	}
}
