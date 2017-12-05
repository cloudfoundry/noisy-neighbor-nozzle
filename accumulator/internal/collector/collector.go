package collector

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/accumulator/internal/app"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/accumulator/internal/datadogreporter"
)

// Rate stores data for a single polling interval.
type Rate struct {
	Timestamp int64             `json:"timestamp"`
	Counts    map[string]uint64 `json:"counts"`
}

// Points is used to sort datadogreporter Points by count
type Points []datadogreporter.Point

func (p Points) Len() int           { return len(p) }
func (p Points) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p Points) Less(i, j int) bool { return p[i].Points[0][1] > p[j].Points[0][1] }

// Authenticator is used to refresh the authentication token.
type Authenticator interface {
	RefreshAuthToken() (string, error)
}

// AppInfoStore provides a way to find AppInfo for an app GUID.
type AppInfoStore interface {
	Lookup(guids []string) (map[app.AppGUID]AppInfo, error)
}

// Collector handles fetch rates form multiple nozzles and summing their
// rates.
type Collector struct {
	nozzles       []string
	auth          Authenticator
	httpClient    *http.Client
	reportLimit   int
	nozzleAppGUID string
	store         AppInfoStore
}

// New initializes and returns a new Collector.
func New(
	nozzles []string,
	auth Authenticator,
	nozzleAppGUID string,
	store AppInfoStore,
	opts ...CollectorOption,
) *Collector {
	c := &Collector{
		nozzles: nozzles,
		auth:    auth,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				DisableKeepAlives: true,
			},
		},
		reportLimit:   250,
		nozzleAppGUID: nozzleAppGUID,
		store:         store,
	}

	for _, o := range opts {
		o(c)
	}

	return c
}

// BuildPoints satisfies the datadogreporter PointBuilder interface. It will
// request all the rates from all the known nozzles and sum their counts.
func (c *Collector) BuildPoints(timestamp int64) ([]datadogreporter.Point, error) {
	rate, err := c.Rate(timestamp)
	if err != nil {
		return nil, err
	}

	var guids []string
	for k := range rate.Counts {
		g := GUIDIndex(k).GUID()
		guids = append(guids, g)
	}
	// The underlying cached store does not return an error and instead simply
	// returns the cache when an error occurs.
	appInfo, _ := c.store.Lookup(guids)

	ddPoints := make([]datadogreporter.Point, 0, len(rate.Counts))
	for instance, count := range rate.Counts {
		tags := []string{
			fmt.Sprintf("application.instance:%s", GUIDIndex(instance)),
		}
		gi := GUIDIndex(instance)
		orgSpaceAppName, ok := appInfo[AppGUID(gi.GUID())]
		if ok {
			tags = []string{
				fmt.Sprintf("application.instance:%s/%s", orgSpaceAppName, gi.Index()),
			}
		}
		ddPoints = append(ddPoints, datadogreporter.Point{
			Metric: "application.ingress",
			Points: [][]int64{[]int64{rate.Timestamp, int64(count)}},
			Type:   "gauge",
			Tags:   tags,
		})
	}

	sort.Sort(Points(ddPoints))
	if len(ddPoints) <= c.reportLimit {
		return ddPoints, nil
	}

	return ddPoints[0:c.reportLimit], nil
}

// Rate will collect rates from all the nozzles and sum the totals to produce a
// a single Rate struct.
func (c *Collector) Rate(timestamp int64) (Rate, error) {
	token, err := c.auth.RefreshAuthToken()
	if err != nil {
		return Rate{}, err
	}

	results := make(chan rateResult, len(c.nozzles))
	defer close(results)
	for i, n := range c.nozzles {
		go func(idx int, addr string) {
			rate, err := c.fetchRate(timestamp, idx, addr, token)
			results <- rateResult{
				rate: rate,
				err:  err,
			}
		}(i, n)
	}

	var result []Rate
	for i := 0; i < len(c.nozzles); i++ {
		r := <-results

		if r.err != nil {
			err = r.err
		}

		result = append(result, r.rate)
	}

	if err != nil {
		return Rate{}, err
	}

	return Sum(result), nil
}

func (c *Collector) fetchRate(timestamp int64, index int, addr, token string) (Rate, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/state/%d", addr, timestamp),
		nil,
	)
	if err != nil {
		return Rate{}, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	if c.nozzleAppGUID != "" {
		req.Header.Set("X-CF-APP-INSTANCE", fmt.Sprintf("%s:%d", c.nozzleAppGUID, index))
	}

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

	return rate, nil
}

// CollectorOption is a type of func that can be used for optional configuration
// settings for the Collector.
type CollectorOption func(c *Collector)

// WithReportLimit sets the limit of application instances to report to datadog.
// Example: If report limit is set to 100, only the 100 noisest application
// instances will be reported.
func WithReportLimit(n int) CollectorOption {
	return func(c *Collector) {
		c.reportLimit = n
	}
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

type rateResult struct {
	rate Rate
	err  error
}
