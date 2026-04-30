package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/diskhm/internal/api"
)

func TestDiskListRequiresSession(t *testing.T) {
	t.Parallel()

	router := api.NewRouter(api.Dependencies{})
	req := httptest.NewRequest(http.MethodGet, "/api/disks", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
