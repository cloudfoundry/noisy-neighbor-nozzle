package web

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// OffendersIndex returns a HTTP handler for returning a list of the top
// offenders.
func OffendersIndex(tn TopN) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := r.URL.Query().Get("count")
		if count == "" {
			count = "5"
		}

		n, err := strconv.Atoi(count)
		if err != nil {
			log.Println("unable to parse count")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		topN := tn(n)

		if err := json.NewEncoder(w).Encode(topN); err != nil {
			log.Printf("failed to write response: %s", err)
		}
	})
}
