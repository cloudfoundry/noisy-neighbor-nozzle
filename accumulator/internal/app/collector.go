package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/accumulator/internal/datadogreporter"
)

// Rate stores data for a single polling interval.
type Rate struct {
	Timestamp int64             `json:"timestamp"`
	Counts    map[string]uint64 `json:"counts"`
}

type Authenticator interface {
	RefreshAuthToken() (string, error)
}

// Collector handles fetch rates form multiple nozzles and summing their
// rates.
type Collector struct {
	nozzles    []string
	auth       Authenticator
	httpClient *http.Client
}

// NewCollector initializes and returns a new Collector.
func NewCollector(nozzles []string, auth Authenticator) *Collector {
	return &Collector{
		nozzles: nozzles,
		auth:    auth,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				DisableKeepAlives: true,
			},
		},
	}
}

// BuildPoints satisfies the datadogreporter PointBuilder interface. It will
// request all the rates from all the known nozzles and sum their counts.
func (c *Collector) BuildPoints(timestamp int64) ([]datadogreporter.Point, error) {
	rate, err := c.rates(timestamp)
	if err != nil {
		return nil, err
	}

	ddPoints := make([]datadogreporter.Point, 0, len(rate.Counts))
	for instance, count := range rate.Counts {
		ddPoints = append(ddPoints, datadogreporter.Point{
			Metric: "application.ingress",
			Points: [][]int64{[]int64{rate.Timestamp, int64(count)}},
			Type:   "gauge",
			Tags: []string{
				fmt.Sprintf("application.instance:%s", instance),
			},
		})
	}

	return ddPoints, nil
}

// rates will collect all the rates from all the nozzles.
func (c *Collector) rates(timestamp int64) (Rate, error) {
	token, err := c.auth.RefreshAuthToken()
	if err != nil {
		return Rate{}, err
	}

	var results []Rate
	for _, n := range c.nozzles {
		req, err := http.NewRequest(
			http.MethodGet,
			fmt.Sprintf("%s/state/%d", n, timestamp),
			nil,
		)
		if err != nil {
			return Rate{}, err
		}

		tokenWithAuthMethod := fmt.Sprintf("Bearer %s", token)
		req.Header.Set("Authorization", tokenWithAuthMethod)
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return Rate{}, err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return Rate{}, fmt.Errorf("failed to get rates, expected status code 200, got %d", resp.StatusCode)
		}

		var rate Rate
		err = json.NewDecoder(resp.Body).Decode(&rate)
		if err != nil {
			return Rate{}, err
		}

		results = append(results, rate)
	}

	return Sum(results), nil
}

// Sum will take a slice of Rate and sum all their counts together to create a
// single Rate.
func Sum(r []Rate) Rate {
	var timestamp int64
	interim := make(map[string]uint64)
	for _, rate := range r {
		timestamp = rate.Timestamp
		for instance, count := range rate.Counts {
			interim[instance] += count
		}
	}

	return Rate{
		Timestamp: timestamp,
		Counts:    interim,
	}
}
