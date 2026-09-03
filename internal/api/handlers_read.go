package api

import (
	"net/http"

	"pastebin/internal/paste"
)

// RegisterReadRoutes wires the paste-read routes. Implemented by the
// GET /pastes and GET /pastes/{id} ticket.
func RegisterReadRoutes(mux *http.ServeMux, s *paste.Store) {
}
