package server

import (
	"encoding/json"
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

func TestDemoEstateEndpointUsesWatermarkedFixture(t *testing.T) {
	assets := fstest.MapFS{"fallback.html": {Data: []byte("fallback")}}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/demo-estate", nil)
	response := httptest.NewRecorder()
	Handler(fs.FS(assets), demo.Seed()).ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	var estate demo.Estate
	if err := json.Unmarshal(response.Body.Bytes(), &estate); err != nil {
		t.Fatalf("invalid demo estate response: %v", err)
	}
	if estate.Watermark != "demo data" || len(estate.Identities) == 0 {
		t.Fatalf("demo fixture missing watermark or records: %#v", estate)
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

func TestNextStaticAssetIsServedWithoutSPARewrite(t *testing.T) {
	assets := fstest.MapFS{
		"index.html":                  {Data: []byte("app shell")},
		"_next/static/chunks/app.css": {Data: []byte("body { color: white; }")},
	}
	request := httptest.NewRequest(http.MethodGet, "/_next/static/chunks/app.css", nil)
	response := httptest.NewRecorder()

	Handler(fs.FS(assets), demo.Seed()).ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Body.String() != "body { color: white; }" {
		t.Fatalf("unexpected static asset response: %d %q", response.Code, response.Body.String())
	}
	if response.Header().Get("Content-Type") != "text/css; charset=utf-8" {
		t.Fatalf("unexpected static asset content type: %q", response.Header().Get("Content-Type"))
	}
}
