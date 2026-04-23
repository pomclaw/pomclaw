package gateway

import (
	"io/fs"
	"net/http"
	"strings"
)

// spaHandler serves a Single Page Application from fsys.
// Exact file matches (assets, favicon, etc.) are served directly.
// Any other path falls back to index.html so the SPA router can handle it.
func spaHandler(fsys fs.FS) http.Handler {
	fileServer := http.FileServerFS(fsys)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Normalise: strip leading slash to get a FS-relative path
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		if _, err := fsys.Open(path); err != nil {
			// File not found – let the SPA client-side router handle it
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/index.html"
			fileServer.ServeHTTP(w, r2)
			return
		}

		fileServer.ServeHTTP(w, r)
	})
}
