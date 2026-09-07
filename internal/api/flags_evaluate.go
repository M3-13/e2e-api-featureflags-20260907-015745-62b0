package api

import (
	"net/http"

	"featureflags/internal/store"
)

// EvaluateFlag handles GET /flags/{key}/evaluate.
func EvaluateFlag(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		WriteError(w, http.StatusNotImplemented, "not implemented")
	}
}
