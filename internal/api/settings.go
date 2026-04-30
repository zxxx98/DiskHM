package api

import "net/http"

func registerSettingsRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/settings", requireSession(http.HandlerFunc(settingsHandler)))
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"quiet_grace_seconds":10}`))
}
