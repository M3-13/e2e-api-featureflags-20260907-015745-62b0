package api

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

// RequireAPIKey returns middleware that enforces a Bearer API key on the
// Authorization header. The expected key is passed in explicitly so it is
// never read from the environment at import time.
func RequireAPIKey(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				WriteError(w, http.StatusServiceUnavailable, "service unavailable: API_KEY not configured")
				return
			}

			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			provided := strings.TrimPrefix(auth, "Bearer ")
			if subtle.ConstantTimeCompare([]byte(provided), []byte(key)) != 1 {
				WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
