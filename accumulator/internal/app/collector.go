package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/accumulator/internal/datadogreporter"
)

// Rate stores data for a single polling interval.
type Rate struct {
	Timestamp int64             `json:"timestamp"`
	Counts    map[string]uint64 `json:"counts"`
}

// Collector handles fetch rates form multiple nozzles and summing their
// rates.
type Collector struct {
	nozzles []string
}

// NewCollector initializes and returns a new Collector.
func NewCollector(nozzles []string) *Collector {
	return &Collector{
		nozzles: nozzles,
	}
}

// BuildPoints satisfies the datadogreporter PointBuilder interface. It will
// request all the rates from all the known nozzles and sum their counts.
func (c *Collector) BuildPoints(timestamp int64) ([]datadogreporter.Point, error) {
	rate, err := c.rates(timestamp)
	if err != nil {
		log.Printf("failed to get rates: %s", err)
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
	var results []Rate
	for _, n := range c.nozzles {
		resp, err := http.Get(fmt.Sprintf("%s/state/%d", n, timestamp))
		if err != nil {
			return Rate{}, err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return Rate{}, fmt.Errorf("Failed to get rates. Expected status code 200, got %d", resp.StatusCode)
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
