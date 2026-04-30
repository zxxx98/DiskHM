package api_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/diskhm/internal/api"
)

func TestLoginSetsSessionCookie(t *testing.T) {
	t.Parallel()

	router := api.NewRouter(api.Dependencies{TokenPlaintext: "dev-token"})
	req := httptest.NewRequest(http.MethodPost, "/api/session/login", bytes.NewBufferString(`{"token":"dev-token"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	if len(rec.Result().Cookies()) == 0 {
		t.Fatal("expected at least one cookie to be set")
	}
}

func TestLoginRejectsWrongToken(t *testing.T) {
	t.Parallel()

	router := api.NewRouter(api.Dependencies{TokenPlaintext: "dev-token"})
	req := httptest.NewRequest(http.MethodPost, "/api/session/login", bytes.NewBufferString(`{"token":"wrong-token"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestLoginRejectsEmptyConfiguredToken(t *testing.T) {
	t.Parallel()

	router := api.NewRouter(api.Dependencies{})
	req := httptest.NewRequest(http.MethodPost, "/api/session/login", bytes.NewBufferString(`{"token":""}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
