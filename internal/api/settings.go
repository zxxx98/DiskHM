package api

import "net/http"

func registerSettingsRoutes(mux *http.ServeMux, deps Dependencies) {
	mux.Handle("GET /api/settings", requireSession(http.HandlerFunc(settingsHandler(deps))))
}

func settingsHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		settings, err := deps.Runtime.Settings(r.Context())
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "failed to load settings")
			return
		}

		writeJSON(w, http.StatusOK, settings)
	}
}
