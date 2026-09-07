package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"featureflags/internal/store"
)

func TestDeleteFlagRemovesFlag(t *testing.T) {
	s := store.NewStore()
	if err := s.Add(store.Flag{Key: "a", Enabled: true}); err != nil {
		t.Fatalf("Add: %v", err)
	}

	handler := DeleteFlag(s)
	req := httptest.NewRequest(http.MethodDelete, "/flags/a", nil)
	req.SetPathValue("key", "a")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if body := rec.Body.String(); body != "" {
		t.Fatalf("body = %q, want empty", body)
	}
	if _, ok := s.Get("a"); ok {
		t.Fatal("flag still present after DELETE")
	}
}

func TestDeleteFlagUnknownKey(t *testing.T) {
	s := store.NewStore()

	handler := DeleteFlag(s)
	req := httptest.NewRequest(http.MethodDelete, "/flags/missing", nil)
	req.SetPathValue("key", "missing")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
