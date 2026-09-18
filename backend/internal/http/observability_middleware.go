package http

import (
	"net"
	stdhttp "net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type loginRateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
}

func newLoginRateLimiter() *loginRateLimiter {
	return &loginRateLimiter{attempts: make(map[string][]time.Time)}
}
func (s *Server) limitAuthentication(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		key, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			key = r.RemoteAddr
		}
		now := time.Now()
		s.rateLimiter.mu.Lock()
		recent := s.rateLimiter.attempts[key][:0]
		for _, attempt := range s.rateLimiter.attempts[key] {
			if now.Sub(attempt) < time.Minute {
				recent = append(recent, attempt)
			}
		}
		if len(recent) >= 10 {
			s.rateLimiter.attempts[key] = recent
			s.rateLimiter.mu.Unlock()
			fail(w, 429, "too many authentication attempts")
			return
		}
		s.rateLimiter.attempts[key] = append(recent, now)
		s.rateLimiter.mu.Unlock()
		next.ServeHTTP(w, r)
	})
}
func (s *Server) requestLogger(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		recorder := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		started := time.Now()
		next.ServeHTTP(recorder, r)
		s.logger.Info("http request", "request_id", middleware.GetReqID(r.Context()), "method", r.Method, "path", r.URL.Path, "status", recorder.Status(), "duration_ms", time.Since(started).Milliseconds())
	})
}
