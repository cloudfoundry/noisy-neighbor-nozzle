package web

import (
	"net/http"
	"strings"
)

var (
	requiredScope = "doppler.firehose"
)

// AdminAuthMiddleware will return HTTP middleware that will authenticate a user
// is authenticated and has proper permissions.
func AdminAuthMiddleware(ct CheckToken) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			items := strings.Split(r.Header.Get("Authorization"), " ")
			if len(items) != 2 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			token := items[1]
			if !ct(token, requiredScope) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}
