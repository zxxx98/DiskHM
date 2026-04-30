package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDiskListRequiresSession(t *testing.T) {
	t.Parallel()

	router := NewRouter(Dependencies{})
	req := httptest.NewRequest(http.MethodGet, "/api/disks", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestDiskListRejectsForgedSessionCookie(t *testing.T) {
	t.Parallel()

	router := NewRouter(Dependencies{})
	req := httptest.NewRequest(http.MethodGet, "/api/disks", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "forged"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestDiskListAllowsValidSessionCookie(t *testing.T) {
	t.Parallel()

	router := NewRouter(Dependencies{})
	req := httptest.NewRequest(http.MethodGet, "/api/disks", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionCookieValue})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestEventStreamRequiresSession(t *testing.T) {
	t.Parallel()

	router := NewRouter(Dependencies{})
	req := httptest.NewRequest(http.MethodGet, "/api/events/stream", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireSessionRejectsMissingCSRFTokenOnProtectedWrite(t *testing.T) {
	t.Parallel()

	nextCalled := false
	handler := requireSession(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/disks", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionCookieValue})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}

	if nextCalled {
		t.Fatal("expected middleware to stop request before next handler")
	}
}
