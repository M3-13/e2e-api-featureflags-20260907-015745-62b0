package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"featureflags/internal/store"
)

func newEvaluateTestStore() *store.Store {
	s := store.NewStore()
	_ = s.Add(store.Flag{Key: "full", Enabled: true, Description: "", RolloutPercent: 100})
	_ = s.Add(store.Flag{Key: "none", Enabled: true, Description: "", RolloutPercent: 0})
	_ = s.Add(store.Flag{Key: "half", Enabled: true, Description: "", RolloutPercent: 50})
	_ = s.Add(store.Flag{Key: "disabled", Enabled: false, Description: "", RolloutPercent: 100})
	return s
}

func evaluate(t *testing.T, s *store.Store, key, user string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/flags/"+key+"/evaluate?user="+user, nil)
	req.SetPathValue("key", key)
	rec := httptest.NewRecorder()
	EvaluateFlag(s)(rec, req)
	return rec
}

func TestEvaluateSameKeyUserDeterministic(t *testing.T) {
	s := newEvaluateTestStore()
	var first *httptest.ResponseRecorder
	for i := 0; i < 10; i++ {
		rec := evaluate(t, s, "half", "alice")
		if i == 0 {
			first = rec
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if rec.Body.String() != first.Body.String() {
			t.Fatalf("result not deterministic: %q vs %q", first.Body.String(), rec.Body.String())
		}
	}
}

func TestEvaluateRollout100AlwaysEnabled(t *testing.T) {
	s := newEvaluateTestStore()
	for _, u := range []string{"a", "b", "c", "d", "e"} {
		rec := evaluate(t, s, "full", u)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if rec.Body.String() != "{\"enabled\":true}\n" {
			t.Fatalf("rollout 100 should always enable, got %q", rec.Body.String())
		}
	}
}

func TestEvaluateRollout0NeverEnabled(t *testing.T) {
	s := newEvaluateTestStore()
	for _, u := range []string{"a", "b", "c", "d", "e"} {
		rec := evaluate(t, s, "none", u)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if rec.Body.String() != "{\"enabled\":false}\n" {
			t.Fatalf("rollout 0 should never enable, got %q", rec.Body.String())
		}
	}
}

func TestEvaluateDisabledFlagAlwaysFalse(t *testing.T) {
	s := newEvaluateTestStore()
	rec := evaluate(t, s, "disabled", "alice")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "{\"enabled\":false}\n" {
		t.Fatalf("disabled flag should be false, got %q", rec.Body.String())
	}
}

func TestEvaluateMissingUserBadRequest(t *testing.T) {
	s := newEvaluateTestStore()
	req := httptest.NewRequest(http.MethodGet, "/flags/full/evaluate", nil)
	rec := httptest.NewRecorder()
	EvaluateFlag(s)(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/flags/full/evaluate?user=", nil)
	rec = httptest.NewRecorder()
	EvaluateFlag(s)(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty user, got %d", rec.Code)
	}
}

func TestEvaluateUnknownKeyNotFound(t *testing.T) {
	s := newEvaluateTestStore()
	rec := evaluate(t, s, "does-not-exist", "alice")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
