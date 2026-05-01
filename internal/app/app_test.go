package app_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/diskhm/internal/app"
	"github.com/example/diskhm/internal/config"
)

func TestAppUsesConfiguredTokenForLogin(t *testing.T) {
	t.Parallel()

	cfg := config.Default()
	cfg.Security.TokenHash = "dev-token"

	application := app.New(cfg)
	req := httptest.NewRequest(http.MethodPost, "/api/session/login", bytes.NewBufferString(`{"token":"dev-token"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	application.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}
