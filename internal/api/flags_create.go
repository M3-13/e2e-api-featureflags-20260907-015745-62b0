package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"

	"featureflags/internal/store"
)

const maxBodyBytes = 1 << 20

var keyPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)

// CreateFlag handles POST /flags.
func CreateFlag(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

		var in store.Flag
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				WriteError(w, http.StatusRequestEntityTooLarge, "request body too large")
				return
			}
			WriteError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		if in.Key == "" {
			WriteError(w, http.StatusBadRequest, "key is required")
			return
		}
		if !keyPattern.MatchString(in.Key) {
			WriteError(w, http.StatusBadRequest, "invalid key format")
			return
		}
		if in.RolloutPercent < 0 || in.RolloutPercent > 100 {
			WriteError(w, http.StatusBadRequest, "rollout_percent must be between 0 and 100")
			return
		}

		if err := s.Add(in); err != nil {
			if errors.Is(err, store.ErrFlagExists) {
				WriteError(w, http.StatusConflict, "flag already exists")
				return
			}
			WriteError(w, http.StatusInternalServerError, "internal error")
			return
		}

		WriteJSON(w, http.StatusCreated, in)
	}
}
