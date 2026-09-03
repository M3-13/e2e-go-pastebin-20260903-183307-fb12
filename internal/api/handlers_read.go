package api

import (
	"net/http"
	"time"

	"pastebin/internal/paste"
)

// pasteMeta is the metadata view of a paste returned by GET /pastes. It
// intentionally omits Content.
type pasteMeta struct {
	ID        string     `json:"id"`
	Language  string     `json:"language"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// RegisterReadRoutes wires the read routes: GET /pastes and GET /pastes/{id}.
func RegisterReadRoutes(mux *http.ServeMux, s *paste.Store) {
	mux.HandleFunc("GET /pastes/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		p, ok := s.Get(id)
		if !ok {
			WriteError(w, http.StatusNotFound, "paste not found")
			return
		}
		WriteJSON(w, http.StatusOK, p)
	})
	mux.HandleFunc("GET /pastes", func(w http.ResponseWriter, r *http.Request) {
		pastes := s.List()
		out := make([]pasteMeta, 0, len(pastes))
		for _, p := range pastes {
			out = append(out, pasteMeta{
				ID:        p.ID,
				Language:  p.Language,
				CreatedAt: p.CreatedAt,
				ExpiresAt: p.ExpiresAt,
			})
		}
		WriteJSON(w, http.StatusOK, out)
	})
}
