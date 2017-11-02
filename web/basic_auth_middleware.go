package web

import "net/http"

// BasicAuthMiddleware returns HTTP middleware for handling basic authentication.
func BasicAuthMiddleware(username, password string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, p, ok := r.BasicAuth()
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			_, _ = u, p

			if u != username || p != password {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}
