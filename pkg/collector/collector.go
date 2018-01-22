package collector

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/datadog"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/store"
)

// Authenticator is used to refresh the authentication token.
type Authenticator interface {
	RefreshAuthToken() (string, error)
}

// AppInfoStore provides a way to find AppInfo for an app GUID.
type AppInfoStore interface {
	Lookup(guids []string) (map[AppGUID]AppInfo, error)
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

// BuildPoints satisfies the datadog PointBuilder interface. It will
// request all the rates from all the known nozzles and sum their counts.
func (c *Collector) BuildPoints(timestamp int64) ([]datadog.Point, error) {
	rate, err := c.Rate(timestamp)
	if err != nil {
		return nil, err
	}

	var top counts
	for k, v := range rate.Counts {
		top = append(top, count{
			guidIndex: k,
			value:     v,
		})
	}
	sort.Sort(top)
	if len(top) > c.reportLimit {
		top = top[0:c.reportLimit]
	}

	var guids []string
	for _, c := range top {
		g := GUIDIndex(c.guidIndex).GUID()
		guids = append(guids, g)
	}
	// The underlying cached store does not return an error and instead simply
	// returns the cache when an error occurs.
	appInfo, _ := c.store.Lookup(guids)

	var ddPoints []datadog.Point
	for _, c := range top {
		gi := GUIDIndex(c.guidIndex)
		tags := []string{
			fmt.Sprintf("application.instance:%s", gi),
		}
		orgSpaceAppName, ok := appInfo[AppGUID(gi.GUID())]
		if ok {
			tags = []string{
				fmt.Sprintf("application.instance:%s/%s", orgSpaceAppName, gi.Index()),
			}
		}
		ddPoints = append(ddPoints, datadog.Point{
			Metric: "application.ingress",
			Points: [][]int64{[]int64{rate.Timestamp, int64(c.value)}},
			Type:   "gauge",
			Tags:   tags,
		})
	}

	return ddPoints, nil
}

// Rate will collect rates from all the nozzles and sum the totals to produce a
// a single Rate struct.
func (c *Collector) Rate(timestamp int64) (store.Rate, error) {
	token, err := c.auth.RefreshAuthToken()
	if err != nil {
		return store.Rate{}, err
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

	var result []store.Rate
	for i := 0; i < len(c.nozzles); i++ {
		r := <-results

		if r.err != nil {
			err = r.err
		}

		result = append(result, r.rate)
	}

	if err != nil {
		return store.Rate{}, err
	}

	return Sum(result), nil
}

func (c *Collector) fetchRate(timestamp int64, index int, addr, token string) (store.Rate, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/rates/%d", addr, timestamp),
		nil,
	)
	if err != nil {
		return store.Rate{}, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	if c.nozzleAppGUID != "" {
		req.Header.Set("X-CF-APP-INSTANCE", fmt.Sprintf("%s:%d", c.nozzleAppGUID, index))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return store.Rate{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return store.Rate{}, fmt.Errorf("failed to get rates, expected status code 200, got %d", resp.StatusCode)
	}

	var rate store.Rate
	err = json.NewDecoder(resp.Body).Decode(&rate)
	if err != nil {
		return store.Rate{}, err
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

// WithHTTPClient sets the the http client that the collector will use to make
// calls to external services.
func WithHTTPClient(client *http.Client) CollectorOption {
	return func(c *Collector) {
		c.httpClient = client
	}
}

// Sum will take a slice of Rate and sum all their counts together to create a
// single Rate.
func Sum(r []store.Rate) store.Rate {
	var timestamp int64
	interim := make(map[string]uint64)
	for _, rate := range r {
		timestamp = rate.Timestamp
		for instance, count := range rate.Counts {
			interim[instance] += count
		}
	}

	return store.Rate{
		Timestamp: timestamp,
		Counts:    interim,
	}
}

type rateResult struct {
	rate store.Rate
	err  error
}

type count struct {
	guidIndex string
	value     uint64
}

type counts []count

func (c counts) Len() int           { return len(c) }
func (c counts) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c counts) Less(i, j int) bool { return c[i].value > c[j].value }
