package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTopologyRequiresSession(t *testing.T) {
	t.Parallel()

	router := NewRouter(Dependencies{})
	req := httptest.NewRequest(http.MethodGet, "/api/topology", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestTopologyAllowsValidSessionCookie(t *testing.T) {
	t.Parallel()

	router := NewRouter(Dependencies{})
	req := httptest.NewRequest(http.MethodGet, "/api/topology", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionCookieValue})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", got, "application/json")
	}

	if got := rec.Body.String(); got != "{\"nodes\":[],\"edges\":[]}\n" {
		t.Fatalf("body = %q, want %q", got, "{\"nodes\":[],\"edges\":[]}\n")
	}
}
