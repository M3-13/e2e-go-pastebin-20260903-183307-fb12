package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"pastebin/internal/paste"
)

// maxBodyBytes is the maximum accepted request body size for POST /pastes.
const maxBodyBytes = 1 << 20

type createRequest struct {
	Content          string          `json:"content"`
	Language         string          `json:"language"`
	ExpiresInSeconds json.RawMessage `json:"expires_in_seconds"`
}

// RegisterCreateRoutes wires the paste-creation route.
func RegisterCreateRoutes(mux *http.ServeMux, s *paste.Store) {
	mux.HandleFunc("POST /pastes", createPasteHandler(s))
}

func createPasteHandler(s *paste.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

		var req createRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}

		if req.Content == "" {
			WriteError(w, http.StatusBadRequest, "content is required")
			return
		}

		var expiresIn time.Duration
		if len(req.ExpiresInSeconds) > 0 {
			raw := bytes.TrimSpace(req.ExpiresInSeconds)
			var secs int
			if bytes.Equal(raw, []byte("null")) || json.Unmarshal(raw, &secs) != nil || secs <= 0 {
				WriteError(w, http.StatusBadRequest, "expires_in_seconds must be positive")
				return
			}
			expiresIn = time.Duration(secs) * time.Second
		}

		p, err := s.Create(req.Content, req.Language, expiresIn)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		WriteJSON(w, http.StatusCreated, map[string]string{"id": p.ID})
	}
}
