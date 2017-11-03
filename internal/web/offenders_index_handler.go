package web

import (
	"encoding/json"
	"log"
	"net/http"
)

// OffendersIndex returns a HTTP handler for returning a list of the top
// offenders.
func OffendersIndex(st State) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(st()); err != nil {
			log.Printf("failed to write response: %s", err)
		}
	})
}
