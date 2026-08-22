package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed all:dist
var distEmbedFS embed.FS

// GetFileSystem returns an http.FileSystem serving the compiled React UI
func GetFileSystem() http.FileSystem {
	sub, err := fs.Sub(distEmbedFS, "dist")
	if err != nil {
		return nil
	}
	return http.FS(sub)
}
