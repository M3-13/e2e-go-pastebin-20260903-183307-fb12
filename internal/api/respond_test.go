package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestMux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", Health)
	return JSONErrorHandler(mux)
}

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteJSON(rec, http.StatusOK, map[string]string{"status": "ok"})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}
	if rec.Body.String() != `{"status":"ok"}` {
		t.Fatalf("expected body {\"status\":\"ok\"}, got %q", rec.Body.String())
	}
}

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteError(rec, http.StatusBadRequest, "bad input")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not JSON: %v", err)
	}
	if body["error"] != "bad input" {
		t.Fatalf("expected error \"bad input\", got %v", body)
	}
	if len(body) != 1 {
		t.Fatalf("expected exactly the error field, got %v", body)
	}
}

func TestHealthEndpoint(t *testing.T) {
	h := newTestMux()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}
	if rec.Body.String() != `{"status":"ok"}` {
		t.Fatalf("expected body {\"status\":\"ok\"}, got %q", rec.Body.String())
	}
}

func TestUnknownPathReturnsJSON404(t *testing.T) {
	h := newTestMux()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not JSON: %v", err)
	}
	if body["error"] == "" {
		t.Fatalf("expected non-empty error field, got %v", body)
	}
}

func TestWrongMethodReturnsJSON405(t *testing.T) {
	h := newTestMux()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}
	if rec.Header().Get("Allow") == "" {
		t.Fatal("expected an Allow header on 405")
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not JSON: %v", err)
	}
	if body["error"] == "" {
		t.Fatalf("expected non-empty error field, got %v", body)
	}
}
