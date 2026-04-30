package api

import (
	"errors"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/example/diskhm/internal/webassets"
)

func registerStaticRoutes(mux *http.ServeMux) {
	distFS, err := fs.Sub(webassets.Dist, "dist")
	if err != nil {
		panic(err)
	}

	mux.Handle("GET /", staticHandler(distFS))
}

func staticHandler(distFS fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api" || strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if name == "" || name == "." || !strings.Contains(path.Base(name), ".") {
			name = "index.html"
		}

		if _, err := fs.Stat(distFS, name); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				http.NotFound(w, r)
				return
			}

			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.ServeFileFS(w, r, distFS, name)
	})
}
