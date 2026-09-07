package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimitRejectsAfterLimitExceeded(t *testing.T) {
	const limit = 3
	handler := RateLimit(limit)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// The first `limit` requests are allowed.
	for i := 0; i < limit; i++ {
		req := httptest.NewRequest(http.MethodGet, "/flags", nil)
		req.RemoteAddr = "192.0.2.1:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rec.Code)
		}
	}

	// The next request exceeds the bucket and must be rejected.
	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	req.RemoteAddr = "192.0.2.1:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after exceeding limit, got %d", rec.Code)
	}
}

func TestRateLimitIsPerClient(t *testing.T) {
	const limit = 1
	handler := RateLimit(limit)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	req.RemoteAddr = "192.0.2.1:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	// A different client is not affected by the first client's usage.
	req2 := httptest.NewRequest(http.MethodGet, "/flags", nil)
	req2.RemoteAddr = "192.0.2.2:5678"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 for second client, got %d", rec2.Code)
	}
}
