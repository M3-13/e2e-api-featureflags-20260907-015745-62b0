package api

import (
	"net/http"

	"featureflags/internal/store"
)

// CreateFlag handles POST /flags.
func CreateFlag(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		WriteError(w, http.StatusNotImplemented, "not implemented")
	}
}
