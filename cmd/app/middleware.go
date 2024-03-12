package main

import "net/http"

// logRequest logs each request with the method, url, remote address, and user agent
func (s *server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info("Request", "method", r.Method, "url", r.URL.Path, "remote_addr", r.RemoteAddr, "user_agent", r.UserAgent())
		next.ServeHTTP(w, r)
	})
}
