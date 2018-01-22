package datadog

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

const datadogAddr = "https://app.datadoghq.com/api/v1/series"

// Reporter stores configuration for reporting to Datadog.
type Reporter struct {
	apiKey       string
	host         string
	pointBuilder PointBuilder
	httpClient   HTTPClient
	interval     time.Duration
}

// NewReporter initializes and returns a new Reporter.
func NewReporter(
	apiKey string,
	pointBuilder PointBuilder,
	opts ...ReporterOption,
) *Reporter {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}

	r := &Reporter{
		apiKey:       apiKey,
		pointBuilder: pointBuilder,
		httpClient:   httpClient,
		interval:     time.Minute,
	}

	for _, o := range opts {
		o(r)
	}

	return r
}

// Run reports metrics from the configured PointBuilder to Datadog on a
// configured interval.
func (r *Reporter) Run() {
	dURL, err := url.Parse(datadogAddr)
	if err != nil {
		log.Fatalf("Failed to parse datadog URL: %s", err)
	}
	query := url.Values{
		"api_key": []string{r.apiKey},
	}
	dURL.RawQuery = query.Encode()

	ticker := time.NewTicker(r.interval)
	for range ticker.C {
		func() {
			log.Println("datadog reporter ticked")

			body, err := r.buildRequestBody()
			if err != nil {
				log.Printf("failed to build request body for datadog: %s", err)
				return
			}

			response, err := r.httpClient.Post(dURL.String(), "application/json", body)
			if err != nil {
				log.Printf("failed to post to datadog: %s", err)
				return
			}
			defer response.Body.Close()

			if response.StatusCode > 299 || response.StatusCode < 200 {
				respBody, _ := ioutil.ReadAll(response.Body)

				log.Printf("Expected successful status code from Datadog, got %d", response.StatusCode)
				log.Printf("Response: %s", respBody)
				return
			}
		}()
	}
}

func (r *Reporter) buildRequestBody() (io.Reader, error) {
	ts := time.Now().
		Add(-2 * r.interval).
		Truncate(r.interval).
		Unix()
	points, err := r.pointBuilder.BuildPoints(ts)
	if err != nil {
		return nil, err
	}

	for i, p := range points {
		p.Host = r.host

		points[i] = p
	}

	data, err := json.Marshal(map[string][]Point{"series": points})
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(data), nil
}

// ReporterOption is a func that is used to configure optional settings on a
// DatadogReporter.
type ReporterOption func(*Reporter)

// WithHost is a ReporterOption for configuring the host that is applied to
// all metrics.
func WithHost(host string) ReporterOption {
	return func(r *Reporter) {
		r.host = host
	}
}

// WithHTTPClient returns a ReporterOption for configuring the HTTPClient to
// be used for sending metrics via HTTP to Datadog.
func WithHTTPClient(c HTTPClient) ReporterOption {
	return func(r *Reporter) {
		r.httpClient = c
	}
}

// WithInterval returns a ReporterOption for configuring the interval metrics
// will be reported to Datadog.
func WithInterval(d time.Duration) ReporterOption {
	return func(r *Reporter) {
		r.interval = d
	}
}

// Point represents a single metric.
type Point struct {
	Metric string    `json:"metric"`
	Points [][]int64 `json:"points"`
	Type   string    `json:"type"`
	Host   string    `json:"host"`
	Tags   []string  `json:"tags"`
}

// PointBuilder is the interface the DatadogReporter will use to collect
// metrics to send to Datadog.
type PointBuilder interface {
	BuildPoints(int64) ([]Point, error)
}

// HTTPClient is the interface used for sending HTTP POST requests to Datadog.
type HTTPClient interface {
	Post(string, string, io.Reader) (*http.Response, error)
}
