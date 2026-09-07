package api

import "net/http"

// Healthz reports service liveness with 200 {"status":"ok"}.
func Healthz(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
