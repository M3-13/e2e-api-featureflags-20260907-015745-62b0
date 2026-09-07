package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"featureflags/internal/store"
)

func newReadServer() (*store.Store, *http.ServeMux) {
	s := store.NewStore()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /flags", ListFlags(s))
	mux.HandleFunc("GET /flags/{key}", GetFlag(s))
	return s, mux
}

func TestGetFlag_Found(t *testing.T) {
	s, mux := newReadServer()
	if err := s.Add(store.Flag{Key: "featureA", Enabled: true, Description: "d", RolloutPercent: 30}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/flags/featureA", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var flag store.Flag
	if err := json.NewDecoder(rec.Body).Decode(&flag); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if flag.Key != "featureA" {
		t.Fatalf("unexpected flag: %+v", flag)
	}
}

func TestGetFlag_NotFound(t *testing.T) {
	_, mux := newReadServer()
	req := httptest.NewRequest(http.MethodGet, "/flags/missing", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if _, ok := body["error"]; !ok {
		t.Fatalf("error field missing: %+v", body)
	}
}

func TestListFlags_Empty(t *testing.T) {
	_, mux := newReadServer()
	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var flags []store.Flag
	if err := json.NewDecoder(rec.Body).Decode(&flags); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if flags == nil {
		t.Fatalf("expected JSON array, got null")
	}
}

func TestListFlags_Sorted(t *testing.T) {
	s, mux := newReadServer()
	for _, k := range []string{"b", "a", "c"} {
		if err := s.Add(store.Flag{Key: k}); err != nil {
			t.Fatalf("seed %s: %v", k, err)
		}
	}
	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	var flags []store.Flag
	if err := json.NewDecoder(rec.Body).Decode(&flags); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	want := []string{"a", "b", "c"}
	for i, f := range flags {
		if f.Key != want[i] {
			t.Fatalf("flags[%d].Key = %q, want %q", i, f.Key, want[i])
		}
	}
}
