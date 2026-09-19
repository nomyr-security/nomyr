package webui

import (
	"embed"
	"io/fs"
)

// Assets contains the static Next.js export after `make web-build`.
//
//go:embed assets
var assets embed.FS

func Public() fs.FS {
	generated, generatedError := fs.Sub(assets, "assets/generated")
	if generatedError == nil {
		if _, indexError := fs.Stat(generated, "index.html"); indexError == nil {
			return generated
		}
	}
	fallback, fallbackError := fs.Sub(assets, "assets")
	if fallbackError != nil {
		panic(fallbackError)
	}
	return fallback
}
