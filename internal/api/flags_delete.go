package api

import (
	"net/http"

	"featureflags/internal/store"
)

// DeleteFlag handles DELETE /flags/{key}.
func DeleteFlag(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		WriteError(w, http.StatusNotImplemented, "not implemented")
	}
}
