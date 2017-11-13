package web

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// StateIndex returns a HTTP handler for returning a list of the top
// offenders.
func StateIndex(store RateStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(store.Rates()); err != nil {
			log.Printf("failed to write response: %s", err)
		}
	})
}

func StateShow(store RateStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t, ok := r.Context().Value("timestamp").(string)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		timestamp, err := strconv.Atoi(t)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		rate, err := store.Rate(int64(timestamp))
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Encode will never fail with known data.
		_ = json.NewEncoder(w).Encode(rate)
	})
}
