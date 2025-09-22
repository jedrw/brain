package server

import (
	"net/http"
	"time"

	"github.com/charmbracelet/log"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w}

		next.ServeHTTP(rw, r)

		elapsed := time.Since(start)

		entry := log.Default().With(
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.statusCode,
			"duration", elapsed,
			"size", rw.size,
		)

		switch {
		case rw.statusCode >= 500:
			entry.Error("http request")
		case rw.statusCode >= 400:
			entry.Warn("http request")
		default:
			entry.Info("http request")
		}
	})
}

func NewHttpServer(handler http.Handler) *http.Server {
	httpMux := http.NewServeMux()
	httpMux.Handle("/", loggingMiddleware(handler))
	httpServer := &http.Server{
		Handler: httpMux,
	}

	return httpServer
}
