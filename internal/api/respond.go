package api

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// WriteJSON writes v as JSON with the given status code and sets Content-Type
// to application/json.
func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(data)
}

// WriteError writes {"error": msg} with the given status code and sets
// Content-Type to application/json.
func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

// Health answers GET /health with 200 {"status":"ok"}.
func Health(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// JSONErrorHandler rewrites net/http's plain-text 404 (unknown path) and 405
// (wrong method) responses into JSON error responses. For a 405 it preserves
// the Allow header. Responses already produced as JSON (application/json) are
// passed through untouched.
func JSONErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &recorder{header: make(http.Header)}
		next.ServeHTTP(rec, r)

		status := rec.status
		if status == 0 {
			status = http.StatusOK
		}

		if rec.header.Get("Content-Type") != "application/json" &&
			(status == http.StatusNotFound || status == http.StatusMethodNotAllowed) {
			for k, vs := range rec.header {
				for _, v := range vs {
					w.Header().Add(k, v)
				}
			}
			msg := "not found"
			if status == http.StatusMethodNotAllowed {
				msg = "method not allowed"
			}
			WriteError(w, status, msg)
			return
		}

		rec.commit(w)
	})
}

type recorder struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func (r *recorder) Header() http.Header { return r.header }

func (r *recorder) WriteHeader(code int) {
	if r.status == 0 {
		r.status = code
	}
}

func (r *recorder) Write(p []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.body.Write(p)
}

func (r *recorder) commit(w http.ResponseWriter) {
	for k, vs := range r.header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	if r.status == 0 {
		r.status = http.StatusOK
	}
	w.WriteHeader(r.status)
	_, _ = w.Write(r.body.Bytes())
}
