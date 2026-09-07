package api

import (
	"net/http"
	"sync"
	"time"
)

// RateLimit returns a per-client token-bucket middleware. Each client is
// identified by its remote address; once a client exceeds the bucket capacity
// within a minute it is rejected with 429 until the window refills.
func RateLimit(limit int) func(http.Handler) http.Handler {
	if limit <= 0 {
		limit = 1
	}

	mu := &sync.Mutex{}
	clients := make(map[string]*bucket)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			now := time.Now()
			mu.Lock()
			b, ok := clients[r.RemoteAddr]
			if !ok {
				b = &bucket{tokens: float64(limit), last: now}
				clients[r.RemoteAddr] = b
			}
			// Refill tokens based on elapsed time (limit tokens per minute).
			elapsed := now.Sub(b.last).Seconds()
			b.tokens += elapsed * (float64(limit) / 60.0)
			if b.tokens > float64(limit) {
				b.tokens = float64(limit)
			}
			b.last = now

			if b.tokens < 1 {
				mu.Unlock()
				WriteError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}
			b.tokens--
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

// bucket tracks the token count and last refill time for a single client.
type bucket struct {
	tokens float64
	last   time.Time
}
