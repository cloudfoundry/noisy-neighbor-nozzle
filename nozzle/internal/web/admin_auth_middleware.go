package web

import "net/http"

var (
	requiredScope = "doppler.firehose"
)

// AdminAuthMiddleware will return HTTP middleware that will authenticate a user
// is authenticated and has proper permissions.
func AdminAuthMiddleware(ct CheckToken) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !ct(r.Header.Get("Authorization"), requiredScope) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}
