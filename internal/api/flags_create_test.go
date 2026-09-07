package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"featureflags/internal/store"
)

func newCreateServer() (*store.Store, http.Handler) {
	s := store.NewStore()
	return s, CreateFlag(s)
}

func doCreate(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()
	_, h := newCreateServer()
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestCreateFlag_Valid(t *testing.T) {
	rec := doCreate(t, `{"key":"featureA","enabled":true,"description":"desc","rollout_percent":50}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	var flag store.Flag
	if err := json.NewDecoder(rec.Body).Decode(&flag); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if flag.Key != "featureA" || !flag.Enabled || flag.Description != "desc" || flag.RolloutPercent != 50 {
		t.Fatalf("unexpected flag: %+v", flag)
	}
}

func TestCreateFlag_AppearsInList(t *testing.T) {
	s, _ := newCreateServer()
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":"featureA","enabled":true}`))
	rec := httptest.NewRecorder()
	CreateFlag(s).ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", rec.Code, http.StatusCreated)
	}

	lreq := httptest.NewRequest(http.MethodGet, "/flags", nil)
	lrec := httptest.NewRecorder()
	ListFlags(s).ServeHTTP(lrec, lreq)
	if lrec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", lrec.Code, http.StatusOK)
	}
	var flags []store.Flag
	if err := json.NewDecoder(lrec.Body).Decode(&flags); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(flags) != 1 || flags[0].Key != "featureA" {
		t.Fatalf("unexpected list: %+v", flags)
	}
}

func TestCreateFlag_MissingKey(t *testing.T) {
	rec := doCreate(t, `{"enabled":true}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCreateFlag_EmptyKey(t *testing.T) {
	rec := doCreate(t, `{"key":"","enabled":true}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCreateFlag_InvalidJSON(t *testing.T) {
	rec := doCreate(t, `{not json`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCreateFlag_RolloutAbove100(t *testing.T) {
	rec := doCreate(t, `{"key":"featureA","enabled":true,"rollout_percent":101}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCreateFlag_BodyOverLimit(t *testing.T) {
	big := `{"key":"featureA","description":"` + strings.Repeat("a", maxBodyBytes) + `"}`
	rec := doCreate(t, big)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestCreateFlag_DuplicateKey(t *testing.T) {
	s, _ := newCreateServer()
	body := `{"key":"featureA","enabled":true}`
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
		rec := httptest.NewRecorder()
		CreateFlag(s).ServeHTTP(rec, req)
		if i == 0 && rec.Code != http.StatusCreated {
			t.Fatalf("first create status = %d, want %d", rec.Code, http.StatusCreated)
		}
		if i == 1 && rec.Code != http.StatusConflict {
			t.Fatalf("second create status = %d, want %d", rec.Code, http.StatusConflict)
		}
	}
}
