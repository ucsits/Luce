package rpc

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS_AllowAllOrigin(t *testing.T) {
	s := newTestServer(t)

	// Test a regular GET request from a foreign origin
	req := httptest.NewRequest(http.MethodGet, "/api/v1/blocks", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	got := rec.Header().Get("Access-Control-Allow-Origin")
	if got != "*" {
		t.Fatalf("expected Access-Control-Allow-Origin: *, got %q", got)
	}
}

func TestCORS_Preflight(t *testing.T) {
	s := newTestServer(t)

	// Simulate a CORS preflight OPTIONS request from a foreign origin
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/blocks", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("expected Access-Control-Allow-Origin: *, got %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Fatal("expected non-empty Access-Control-Allow-Methods")
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Fatal("expected non-empty Access-Control-Allow-Headers")
	}
}

func TestCORS_AllowAllMethods(t *testing.T) {
	s := newTestServer(t)

	// OPTIONS preflight checking an uncommon method
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/chain/height", nil)
	req.Header.Set("Origin", "https://other-site.org")
	req.Header.Set("Access-Control-Request-Method", "DELETE")
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	if got := rec.Header().Get("Access-Control-Allow-Methods"); got != "*" {
		t.Fatalf("expected Access-Control-Allow-Methods: *, got %q", got)
	}
}

func TestCORS_AllowAllHeaders(t *testing.T) {
	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/blocks", nil)
	req.Header.Set("Origin", "https://attacker.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "X-Custom-Header,Authorization")
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	if got := rec.Header().Get("Access-Control-Allow-Headers"); got != "*" {
		t.Fatalf("expected Access-Control-Allow-Headers: *, got %q", got)
	}
}

func TestCORS_PresentOnAllEndpoints(t *testing.T) {
	s := newTestServer(t)

	endpoints := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/blocks"},
		{http.MethodGet, "/api/v1/blocks/0"},
		{http.MethodGet, "/api/v1/chain/validate"},
		{http.MethodGet, "/api/v1/chain/height"},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			req := httptest.NewRequest(ep.method, ep.path, nil)
			req.Header.Set("Origin", "https://remote.app")
			rec := httptest.NewRecorder()
			s.echo.ServeHTTP(rec, req)

			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
				t.Fatalf("on %s %s: expected Access-Control-Allow-Origin: *, got %q", ep.method, ep.path, got)
			}
		})
	}
}

func TestCORS_NoOriginRequest(t *testing.T) {
	s := newTestServer(t)

	// Requests without an Origin header should still succeed (no CORS headers needed)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chain/height", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
