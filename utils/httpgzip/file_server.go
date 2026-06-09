package httpgzip

import (
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
)

// FileServer serves precompressed .gz files when the client accepts gzip.
func FileServer(root fs.FS) http.Handler {
	plain := http.FileServer(http.FS(root))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if acceptsGzip(r.Header.Get("Accept-Encoding")) && exists(root, name+".gz") {
			gzipReq := r.Clone(r.Context())
			gzipReq.URL.Path = "/" + name + ".gz"
			if contentType := mime.TypeByExtension(path.Ext(name)); contentType != "" {
				w.Header().Set("Content-Type", contentType)
			}
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Add("Vary", "Accept-Encoding")
			plain.ServeHTTP(w, gzipReq)
			return
		}

		plain.ServeHTTP(w, r)
	})
}

func acceptsGzip(value string) bool {
	for _, part := range strings.Split(value, ",") {
		if strings.TrimSpace(strings.Split(part, ";")[0]) == "gzip" {
			return true
		}
	}
	return false
}

func exists(root fs.FS, name string) bool {
	file, err := root.Open(name)
	if err != nil {
		return false
	}
	_ = file.Close()
	return true
}
