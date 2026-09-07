package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"featureflags/internal/store"
)

func newTestStoreWithFlag() *store.Store {
	s := store.NewStore()
	_ = s.Add(store.Flag{Key: "myflag", Enabled: false, Description: "old", RolloutPercent: 10})
	return s
}

func doUpdate(s *store.Store, key, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPut, "/flags/"+key, strings.NewReader(body))
	req.SetPathValue("key", key)
	rec := httptest.NewRecorder()
	UpdateFlag(s)(rec, req)
	return rec
}

func TestUpdateFlagValid(t *testing.T) {
	s := newTestStoreWithFlag()
	rec := doUpdate(s, "myflag", `{"enabled":true,"description":"new","rollout_percent":80}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got store.Flag
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if got.Key != "myflag" {
		t.Fatalf("key = %q, want %q", got.Key, "myflag")
	}
	if !got.Enabled || got.Description != "new" || got.RolloutPercent != 80 {
		t.Fatalf("got %+v", got)
	}
}

func TestUpdateFlagMissingOptionalFields(t *testing.T) {
	s := newTestStoreWithFlag()
	rec := doUpdate(s, "myflag", `{"enabled":true}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got store.Flag
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if got.Description != "" {
		t.Fatalf("description = %q, want empty", got.Description)
	}
	if got.RolloutPercent != 10 {
		t.Fatalf("rollout_percent = %d, want 10 (previous value)", got.RolloutPercent)
	}
}

func TestUpdateFlagMissingEnabled(t *testing.T) {
	s := newTestStoreWithFlag()
	rec := doUpdate(s, "myflag", `{"description":"x"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestUpdateFlagNotFound(t *testing.T) {
	s := newTestStoreWithFlag()
	rec := doUpdate(s, "missing", `{"enabled":true}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestUpdateFlagRolloutOutOfRange(t *testing.T) {
	s := newTestStoreWithFlag()
	rec := doUpdate(s, "myflag", `{"enabled":true,"rollout_percent":101}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestUpdateFlagInvalidJSON(t *testing.T) {
	s := newTestStoreWithFlag()
	rec := doUpdate(s, "myflag", `{"enabled":`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestUpdateFlagBodyTooLarge(t *testing.T) {
	s := newTestStoreWithFlag()
	big := `{"enabled":true,"description":"` + strings.Repeat("a", 1<<20) + `"}`
	rec := doUpdate(s, "myflag", big)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}
}
