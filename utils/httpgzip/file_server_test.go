package httpgzip

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestFileServerServesGzipVariantWhenAccepted(t *testing.T) {
	handler := FileServer(fstest.MapFS{
		"app.js":    {Data: []byte("plain")},
		"app.js.gz": {Data: []byte("gzipped")},
	})

	req := httptest.NewRequest(http.MethodGet, "/app.js", nil)
	req.Header.Set("Accept-Encoding", "br, gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	body, _ := io.ReadAll(rec.Result().Body)

	if got := rec.Header().Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("expected gzip content encoding, got %q", got)
	}
	if string(body) != "gzipped" {
		t.Fatalf("expected gzip file content, got %q", string(body))
	}
}

func TestFileServerFallsBackWithoutGzipSupport(t *testing.T) {
	handler := FileServer(fstest.MapFS{
		"app.js":    {Data: []byte("plain")},
		"app.js.gz": {Data: []byte("gzipped")},
	})

	req := httptest.NewRequest(http.MethodGet, "/app.js", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	body, _ := io.ReadAll(rec.Result().Body)

	if got := rec.Header().Get("Content-Encoding"); got != "" {
		t.Fatalf("expected no content encoding, got %q", got)
	}
	if string(body) != "plain" {
		t.Fatalf("expected plain file content, got %q", string(body))
	}
}
