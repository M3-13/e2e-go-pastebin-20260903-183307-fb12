package api

import (
	"net/http"

	"pastebin/internal/paste"
)

// RegisterDeleteRoutes wires the paste-deletion routes. Implemented by the
// DELETE /pastes/{id} ticket.
func RegisterDeleteRoutes(mux *http.ServeMux, s *paste.Store) {
}
