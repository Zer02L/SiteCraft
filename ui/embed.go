package ui

import (
	"embed"
	"io/fs"
	"os"
)

//go:embed all:assets
var embeddedAssetsFS embed.FS

// AssetsFS returns an fs.FS for serving static assets.
// If the local directory ./ui/assets exists (development mode), it uses os.DirFS
// to allow instant reloading of styles and assets without recompiling the Go binary.
// Otherwise, it falls back to the embedded filesystem for standalone binary deployments.
func AssetsFS() fs.FS {
	if fi, err := os.Stat("./ui/assets"); err == nil && fi.IsDir() {
		return os.DirFS("./ui/assets")
	}
	sub, err := fs.Sub(embeddedAssetsFS, "assets")
	if err != nil {
		return embeddedAssetsFS
	}
	return sub
}

// EmbeddedAssetsFS returns the pure in-memory embedded filesystem directly.
func EmbeddedAssetsFS() fs.FS {
	sub, err := fs.Sub(embeddedAssetsFS, "assets")
	if err != nil {
		return embeddedAssetsFS
	}
	return sub
}
