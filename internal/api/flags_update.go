package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"featureflags/internal/store"
)

const maxBodyBytes = 1 << 20

type updateFlagRequest struct {
	Enabled        *bool   `json:"enabled"`
	Description    *string `json:"description"`
	RolloutPercent *int    `json:"rollout_percent"`
}

// UpdateFlag handles PUT /flags/{key}.
func UpdateFlag(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req updateFlagRequest
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				WriteError(w, http.StatusRequestEntityTooLarge, "request body too large")
				return
			}
			WriteError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		if req.Enabled == nil {
			WriteError(w, http.StatusBadRequest, "enabled is required")
			return
		}
		if req.RolloutPercent != nil && (*req.RolloutPercent < 0 || *req.RolloutPercent > 100) {
			WriteError(w, http.StatusBadRequest, "rollout_percent must be between 0 and 100")
			return
		}

		key := r.PathValue("key")
		existing, ok := s.Get(key)
		if !ok {
			WriteError(w, http.StatusNotFound, "flag not found")
			return
		}

		description := ""
		if req.Description != nil {
			description = *req.Description
		}

		rollout := existing.RolloutPercent
		if req.RolloutPercent != nil {
			rollout = *req.RolloutPercent
		}

		updated, _ := s.Update(key, *req.Enabled, description, rollout)
		WriteJSON(w, http.StatusOK, updated)
	}
}
