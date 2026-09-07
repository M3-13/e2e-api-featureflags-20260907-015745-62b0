package api

import (
	"encoding/json"
	"io"
	"net/http"

	"featureflags/internal/store"
)

const maxUpdateBodyBytes = 1 << 20

type updateFlagRequest struct {
	Enabled        *bool   `json:"enabled"`
	Description    *string `json:"description"`
	RolloutPercent *int    `json:"rollout_percent"`
}

// UpdateFlag handles PUT /flags/{key}.
func UpdateFlag(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(io.LimitReader(r.Body, maxUpdateBodyBytes+1))
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if len(body) > maxUpdateBodyBytes {
			WriteError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}

		var req updateFlagRequest
		if err := json.Unmarshal(body, &req); err != nil {
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

		updated, ok := s.Update(key, *req.Enabled, description, rollout)
		if !ok {
			WriteError(w, http.StatusNotFound, "flag not found")
			return
		}
		WriteJSON(w, http.StatusOK, updated)
	}
}
