package api

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoggingLogsMethodPathAndStatus(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	handler := Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodPost, "/flags/abc/evaluate?user=secret-user-123", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	out := buf.String()
	if !strings.Contains(out, http.MethodPost) {
		t.Fatalf("log %q does not contain method %q", out, http.MethodPost)
	}
	if !strings.Contains(out, "/flags/abc/evaluate") {
		t.Fatalf("log %q does not contain path %q", out, "/flags/abc/evaluate")
	}
	if !strings.Contains(out, "404") {
		t.Fatalf("log %q does not contain status code 404", out)
	}
}

func TestLoggingOmitsQueryString(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	handler := Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/flags/abc/evaluate?user=secret-user-123", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	out := buf.String()
	if strings.Contains(out, "user=") {
		t.Fatalf("log %q must not contain the query string", out)
	}
	if strings.Contains(out, "secret-user-123") {
		t.Fatalf("log %q must not contain the user value", out)
	}
}
