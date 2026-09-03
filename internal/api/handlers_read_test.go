package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pastebin/internal/paste"
)

func newReadTestServer(t *testing.T) (*paste.Store, *http.ServeMux) {
	t.Helper()
	s := paste.NewStore()
	mux := http.NewServeMux()
	RegisterReadRoutes(mux, s)
	return s, mux
}

func TestGetPasteKnownReturnsContent(t *testing.T) {
	s, mux := newReadTestServer(t)
	p, err := s.Create("hello world", "text", 0)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/pastes/"+p.ID, nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got paste.Paste
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.ID != p.ID {
		t.Errorf("id = %q, want %q", got.ID, p.ID)
	}
	if got.Content != "hello world" {
		t.Errorf("content = %q, want %q", got.Content, "hello world")
	}
	if got.Language != "text" {
		t.Errorf("language = %q, want %q", got.Language, "text")
	}
}

func TestGetPasteUnknownReturns404(t *testing.T) {
	_, mux := newReadTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/pastes/deadbeef", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestGetPasteExpiredReturns404(t *testing.T) {
	s, mux := newReadTestServer(t)
	p, err := s.Create("soon gone", "text", time.Millisecond)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	time.Sleep(5 * time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/pastes/"+p.ID, nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestListPastesMetadataWithoutContent(t *testing.T) {
	s, mux := newReadTestServer(t)
	live, err := s.Create("live", "text", 0)
	if err != nil {
		t.Fatalf("Create live: %v", err)
	}
	if _, err := s.Create("expired", "text", time.Millisecond); err != nil {
		t.Fatalf("Create expired: %v", err)
	}
	time.Sleep(5 * time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/pastes", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("len = %d, want 1 (expired omitted)", len(got))
	}
	if got[0]["id"] != live.ID {
		t.Errorf("id = %v, want %q", got[0]["id"], live.ID)
	}
	if _, hasContent := got[0]["content"]; hasContent {
		t.Errorf("list item contains content field: %v", got[0])
	}
}

func TestListEmpty(t *testing.T) {
	_, mux := newReadTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/pastes", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("expected empty list, got %v", rec.Body.String())
	}
}
