package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSONSetsSecurityHeaders(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteJSON(rec, http.StatusOK, map[string]string{"status": "ok"})

	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want %q", got, "nosniff")
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want %q", got, "no-store")
	}
}
