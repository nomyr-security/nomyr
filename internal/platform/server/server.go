package server

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"nomyr/internal/buildinfo"
	"nomyr/internal/demo"
)

func Handler(assets fs.FS, estate demo.Estate) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/healthz", func(response http.ResponseWriter, _ *http.Request) {
		writeJSON(response, http.StatusOK, map[string]any{"status": "ok", "build": buildinfo.Current()})
	})
	mux.HandleFunc("GET /api/v1/demo-estate", func(response http.ResponseWriter, _ *http.Request) {
		writeJSON(response, http.StatusOK, estate)
	})

	fileServer := http.FileServer(http.FS(assets))
	mux.Handle("/", spaFallback(assets, fileServer))
	return mux
}

func spaFallback(assets fs.FS, fileServer http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requested := strings.TrimPrefix(path.Clean(request.URL.Path), "/")
		if requested == "." || requested == "" {
			requested = "index.html"
		}
		if _, statError := fs.Stat(assets, requested); statError == nil {
			fileServer.ServeHTTP(response, request)
			return
		}
		if _, statError := fs.Stat(assets, "index.html"); statError == nil {
			request.URL.Path = "/"
			fileServer.ServeHTTP(response, request)
			return
		}
		request.URL.Path = "/fallback.html"
		fileServer.ServeHTTP(response, request)
	})
}

func writeJSON(response http.ResponseWriter, status int, value any) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(value)
}
