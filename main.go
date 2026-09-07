package main

import (
	"log"
	"net/http"

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
	mux.HandleFunc("GET /healthz", api.Healthz)

	handler := api.Logging(mux)

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	log.Printf("featureflags listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
