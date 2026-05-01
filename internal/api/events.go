package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func registerEventRoutes(mux *http.ServeMux, deps Dependencies) {
	mux.Handle("GET /api/events", requireSession(http.HandlerFunc(listEventsHandler(deps))))
	mux.Handle("GET /api/events/stream", requireSession(http.HandlerFunc(streamEventsHandler(deps))))
}

func listEventsHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		events, err := deps.Runtime.ListEvents(r.Context(), 100)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "failed to load events")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"items": events})
	}
}

func streamEventsHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		events, unsubscribe := deps.Runtime.SubscribeEvents()
		defer unsubscribe()

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)
		flusher.Flush()

		for {
			select {
			case <-r.Context().Done():
				return
			case event, ok := <-events:
				if !ok {
					return
				}
				payload, err := json.Marshal(event)
				if err != nil {
					return
				}
				_, _ = fmt.Fprintf(w, "data: %s\n\n", payload)
				flusher.Flush()
			}
		}
	}
}
