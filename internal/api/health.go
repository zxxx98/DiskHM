package api

import "net/http"

func registerHealthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/health", healthHandler)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
