package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSettingsRequiresSession(t *testing.T) {
	t.Parallel()

	router := NewRouter(Dependencies{})
	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestSettingsAllowsValidSessionCookie(t *testing.T) {
	t.Parallel()

	router := NewRouter(Dependencies{})
	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionCookieValue})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", got, "application/json")
	}

	if got := rec.Body.String(); got != `{"quiet_grace_seconds":10}` {
		t.Fatalf("body = %q, want %q", got, `{"quiet_grace_seconds":10}`)
	}
}
