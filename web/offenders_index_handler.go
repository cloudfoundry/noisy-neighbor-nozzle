package web

import (
	"encoding/json"
	"log"
	"net/http"
)

// OffendersIndex returns a HTTP handler for returning a list of the top
// offenders.
func OffendersIndex(tn TopN) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		topN := tn(5)

		if err := json.NewEncoder(w).Encode(topN); err != nil {
			log.Printf("failed to write response: %s", err)
		}
	})
}
