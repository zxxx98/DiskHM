package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/diskhm/internal/api"
)

func TestHealthEndpointReturnsJSON(t *testing.T) {
	t.Parallel()

	router := api.NewRouter(api.Dependencies{})
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", got, "application/json")
	}
}
