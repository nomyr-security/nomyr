package server

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"nomyr/internal/demo"
)

func TestHealthEndpoint(t *testing.T) {
	assets := fstest.MapFS{"fallback.html": {Data: []byte("fallback")}}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/healthz", nil)
	response := httptest.NewRecorder()

	Handler(fs.FS(assets), demo.Seed()).ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
}

func TestMissingStaticRouteUsesFallback(t *testing.T) {
	assets := fstest.MapFS{"fallback.html": {Data: []byte("fallback")}}
	request := httptest.NewRequest(http.MethodGet, "/identities/demo-service-account", nil)
	response := httptest.NewRecorder()

	Handler(fs.FS(assets), demo.Seed()).ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Body.String() != "fallback" {
		t.Fatalf("unexpected fallback response: %d %q", response.Code, response.Body.String())
	}
}
