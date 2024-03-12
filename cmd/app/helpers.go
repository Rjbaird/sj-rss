package main

import (
	"net/http"
	"runtime/debug"
)

// Error handling

// serverError logs the error and sends a 500 Internal Server Error response to the user.
func (s *server) ServerError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		method = r.Method
		url    = r.URL.String()
		trace  = string(debug.Stack())
	)
	s.logger.Error(err.Error(), "method", method, "url", url, "trace", trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (s *server) notFound404(w http.ResponseWriter, r *http.Request) {
	var (
		method = r.Method
		url    = r.URL.String()
	)
	s.logger.Error("Not Found", "method", method, "url", url)
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}
