package web

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// RatesShow gets and renders a single Rate for a given timestamp.
func RatesShow(store RateStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t, ok := mux.Vars(r)["timestamp"]
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
