package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// RatesShow gets and renders a single Rate for a given timestamp.
func RatesShow(store RateStore, rateInterval time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t, ok := mux.Vars(r)["timestamp"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		timestamp, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if strings.ToLower(r.URL.Query().Get("truncate_timestamp")) == "true" {
			ts := time.Unix(timestamp, 0)
			timestamp = ts.Truncate(rateInterval).Unix()
		}

		rate, err := store.Rate(timestamp)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Encode will never fail with known data.
		_ = json.NewEncoder(w).Encode(rate)
	})
}
