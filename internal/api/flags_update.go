package api

import (
	"net/http"

	"featureflags/internal/store"
)

// UpdateFlag handles PUT /flags/{key}.
func UpdateFlag(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		WriteError(w, http.StatusNotImplemented, "not implemented")
	}
}
