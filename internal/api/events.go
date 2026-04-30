package api

import "net/http"

func registerEventRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/events/stream", requireSession(http.HandlerFunc(streamEventsHandler)))
}

func streamEventsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.WriteHeader(http.StatusOK)
}
