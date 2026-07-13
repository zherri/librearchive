package utils

import (
	"io/fs"
	"net/http"
)

func ServeEmbeddedHTML(webFS fs.FS, fileName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		content, err := fs.ReadFile(webFS, fileName)
		if err != nil {
			http.Error(w, "page not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(content)
	}
}
