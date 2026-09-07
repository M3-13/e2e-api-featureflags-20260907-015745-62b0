package api

import (
	"net/http"

	"featureflags/internal/store"
)

// ListFlags handles GET /flags.
func ListFlags(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, s.List())
	}
}

// GetFlag handles GET /flags/{key}.
func GetFlag(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")
		f, ok := s.Get(key)
		if !ok {
			WriteError(w, http.StatusNotFound, "flag not found")
			return
		}
		WriteJSON(w, http.StatusOK, f)
	}
}
