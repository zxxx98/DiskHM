package api

import "net/http"

func registerTopologyRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/topology", requireSession(http.HandlerFunc(topologyHandler)))
}

func topologyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"nodes":[],"edges":[]}`))
}
