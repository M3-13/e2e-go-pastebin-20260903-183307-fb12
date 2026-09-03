package api

import (
	"net/http"

	"pastebin/internal/paste"
)

// RegisterCreateRoutes wires the paste-creation routes. Implemented by the
// POST /pastes ticket.
func RegisterCreateRoutes(mux *http.ServeMux, s *paste.Store) {
}
