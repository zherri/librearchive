package http

import (
	stdhttp "net/http"
)

func (s *Server) cors(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && s.allowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		}
		if r.Method == stdhttp.MethodOptions {
			w.WriteHeader(stdhttp.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) allowedOrigin(origin string) bool {
	for _, allowed := range s.config.AllowedOrigins {
		if allowed == origin {
			return true
		}
	}
	return false
}
