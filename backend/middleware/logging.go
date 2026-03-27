package middleware

import (
	"log"
	"net/http"
	"time"
)

// responseCapture wraps http.ResponseWriter to record the status code.
type responseCapture struct {
	http.ResponseWriter
	status int
}

func (c *responseCapture) WriteHeader(code int) {
	if c.status == 0 {
		c.status = code
	}
	c.ResponseWriter.WriteHeader(code)
}

func (c *responseCapture) statusOrOK() int {
	if c.status == 0 {
		return http.StatusOK
	}
	return c.status
}

// RequestLog logs one line per request: method, path, status, duration.
// GET /api/health is omitted to reduce noise from probes.
func RequestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/health" {
			next.ServeHTTP(w, r)
			return
		}
		start := time.Now()
		cap := &responseCapture{ResponseWriter: w, status: 0}
		next.ServeHTTP(cap, r)
		path := r.URL.Path
		if r.URL.RawQuery != "" {
			path = path + "?" + r.URL.RawQuery
		}
		// Truncate very long query strings (e.g. future filters)
		if len(path) > 120 {
			path = path[:117] + "..."
		}
		log.Printf("[http] %s %s %d %s %s",
			r.Method, path, cap.statusOrOK(), time.Since(start).Round(time.Millisecond), r.RemoteAddr)
	})
}
