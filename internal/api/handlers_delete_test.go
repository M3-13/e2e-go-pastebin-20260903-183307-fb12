package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pastebin/internal/paste"
)

func TestDeleteExistingPaste(t *testing.T) {
	s := paste.NewStore()
	p, err := s.Create("hello", "", 0)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	mux := http.NewServeMux()
	RegisterDeleteRoutes(mux, s)

	req := httptest.NewRequest(http.MethodDelete, "/pastes/"+p.ID, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("DELETE status = %d, want %d", rr.Code, http.StatusNoContent)
	}
	if rr.Body.Len() != 0 {
		t.Fatalf("DELETE body = %q, want empty", rr.Body.String())
	}

	if _, ok := s.Get(p.ID); ok {
		t.Fatalf("paste %s still present after DELETE", p.ID)
	}
}

func TestDeleteUnknownPaste(t *testing.T) {
	s := paste.NewStore()

	mux := http.NewServeMux()
	RegisterDeleteRoutes(mux, s)

	req := httptest.NewRequest(http.MethodDelete, "/pastes/doesnotexist", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("DELETE status = %d, want %d", rr.Code, http.StatusNotFound)
	}
	body := strings.TrimSpace(rr.Body.String())
	if !strings.Contains(body, `"error"`) {
		t.Fatalf("DELETE body = %q, want error JSON", body)
	}
}
