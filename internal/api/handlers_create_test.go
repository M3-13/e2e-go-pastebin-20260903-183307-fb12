package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"pastebin/internal/paste"
)

func newCreateMux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", Health)
	RegisterCreateRoutes(mux, paste.NewStore())
	return JSONErrorHandler(mux)
}

var hex32 = regexp.MustCompile(`^[0-9a-f]{32}$`)

func doCreate(body string) *httptest.ResponseRecorder {
	h := newCreateMux()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/pastes", strings.NewReader(body))
	h.ServeHTTP(rec, req)
	return rec
}

func TestCreatePasteValid(t *testing.T) {
	rec := doCreate(`{"content":"hello","language":"go","expires_in_seconds":60}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (body %s)", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not JSON: %v", err)
	}
	if !hex32.MatchString(body["id"]) {
		t.Fatalf("expected a 32-char hex id, got %q", body["id"])
	}
}

func TestCreatePasteMissingContent(t *testing.T) {
	rec := doCreate(`{"language":"go"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreatePasteEmptyContent(t *testing.T) {
	rec := doCreate(`{"content":""}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreatePasteNonPositiveExpiry(t *testing.T) {
	for _, v := range []string{"0", "-1"} {
		rec := doCreate(`{"content":"x","expires_in_seconds":` + v + `}`)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expires_in_seconds=%s: expected 400, got %d", v, rec.Code)
		}
	}
}

func TestCreatePasteInvalidJSON(t *testing.T) {
	rec := doCreate(`{not json`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreatePasteBodyTooLarge(t *testing.T) {
	big := strings.Repeat("x", maxBodyBytes+1)
	rec := doCreate(`{"content":"` + big + `"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
