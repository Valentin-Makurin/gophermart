package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

func Compress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz := gzip.NewWriter(w)
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		gzResponseWriter := &gzipResponseWriter{Writer: gz, ResponseWriter: w}

		next.ServeHTTP(gzResponseWriter, r)
	})
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	return g.Writer.Write(b)
}
