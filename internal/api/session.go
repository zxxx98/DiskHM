package api

import (
	"encoding/json"
	"net/http"
)

const (
	sessionCookieName  = "diskhm_session"
	sessionCookieValue = "dev-session"
	csrfHeaderName     = "X-CSRF-Token"
	csrfTokenValue     = "dev-csrf"
)

type loginRequest struct {
	Token string `json:"token"`
}

func registerSessionRoutes(mux *http.ServeMux, deps Dependencies) {
	mux.HandleFunc("POST /api/session/login", loginHandler(deps))
}

func loginHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		if deps.TokenPlaintext == "" || req.Token != deps.TokenPlaintext {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     sessionCookieName,
			Value:    sessionCookieValue,
			Path:     "/",
			HttpOnly: true,
		})
		w.Header().Set(csrfHeaderName, csrfTokenValue)
		w.WriteHeader(http.StatusNoContent)
	}
}
