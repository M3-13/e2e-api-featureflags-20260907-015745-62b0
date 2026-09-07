package api

import (
	"hash/fnv"
	"net/http"

	"featureflags/internal/store"
)

// EvaluateFlag handles GET /flags/{key}/evaluate.
func EvaluateFlag(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")
		user := r.URL.Query().Get("user")
		if user == "" {
			WriteError(w, http.StatusBadRequest, "user is required")
			return
		}

		flag, ok := s.Get(key)
		if !ok {
			WriteError(w, http.StatusNotFound, "flag not found")
			return
		}

		if !flag.Enabled {
			WriteJSON(w, http.StatusOK, map[string]bool{"enabled": false})
			return
		}

		h := fnv.New64a()
		_, _ = h.Write([]byte(key + ":" + user))
		enabled := h.Sum64()%100 < uint64(flag.RolloutPercent)
		WriteJSON(w, http.StatusOK, map[string]bool{"enabled": enabled})
	}
}
