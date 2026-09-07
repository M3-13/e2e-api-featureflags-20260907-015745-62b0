package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"featureflags/internal/api"
	"featureflags/internal/store"
)

func main() {
	s := store.NewStore()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /flags", api.CreateFlag(s))
	mux.HandleFunc("GET /flags", api.ListFlags(s))
	mux.HandleFunc("GET /flags/{key}", api.GetFlag(s))
	mux.HandleFunc("PUT /flags/{key}", api.UpdateFlag(s))
	mux.HandleFunc("DELETE /flags/{key}", api.DeleteFlag(s))
	mux.HandleFunc("GET /flags/{key}/evaluate", api.EvaluateFlag(s))

	apiKey := os.Getenv("API_KEY")
	rateLimit := rateLimitPerMinute()

	protected := api.RateLimit(rateLimit)(api.RequireAPIKey(apiKey)(mux))

	// /healthz is deliberately kept outside auth and rate-limiting so liveness
	// probes always succeed; every other route goes through the protected chain.
	root := http.NewServeMux()
	root.HandleFunc("GET /healthz", api.Healthz)
	root.Handle("/", protected)

	handler := api.Logging(root)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("featureflags listening on %s", server.Addr)

	certFile := os.Getenv("TLS_CERT_FILE")
	keyFile := os.Getenv("TLS_KEY_FILE")

	if certFile != "" && keyFile != "" {
		server.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		log.Printf("featureflags serving TLS")
		if err := server.ListenAndServeTLS(certFile, keyFile); err != nil {
			log.Fatal(err)
		}
		return
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// rateLimitPerMinute reads RATE_LIMIT_PER_MINUTE and returns it, falling back
// to 120 when the variable is absent or unparsable.
func rateLimitPerMinute() int {
	const defaultLimit = 120

	v := os.Getenv("RATE_LIMIT_PER_MINUTE")
	if v == "" {
		return defaultLimit
	}

	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return defaultLimit
	}
	return n
}
