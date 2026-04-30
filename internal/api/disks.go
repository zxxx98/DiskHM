package api

import "net/http"

func registerDiskRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/disks", requireSession(http.HandlerFunc(listDisksHandler)))
}

func listDisksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"items":[]}`))
}
