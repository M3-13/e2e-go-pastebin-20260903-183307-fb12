package api

import (
	"net/http"

	"pastebin/internal/paste"
)

// RegisterDeleteRoutes wires the paste-deletion routes.
func RegisterDeleteRoutes(mux *http.ServeMux, s *paste.Store) {
	mux.HandleFunc("DELETE /pastes/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if !s.Delete(id) {
			WriteError(w, http.StatusNotFound, "paste not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
