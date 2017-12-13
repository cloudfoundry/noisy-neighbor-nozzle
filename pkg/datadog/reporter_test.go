package datadog_test

import (
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/datadog"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DatadogReporter", func() {
	It("sends data points to datadog on an interval", func() {
		pointBuilder := &spyPointBuilder{}
		httpClient := &spyHTTPClient{}

		reporter := datadog.NewReporter(
			"api-key",
			pointBuilder,
			datadog.WithHost("abcdefg"),
			datadog.WithInterval(50*time.Millisecond),
			datadog.WithHTTPClient(httpClient),
		)
		go reporter.Run()

		Eventually(pointBuilder.buildCalled).Should(BeNumerically(">", 1))
		Expect(pointBuilder.buildPointsTimestamp()).To(BeNumerically("~",
			time.Now().Add(-2*(50*time.Millisecond)).Truncate(50*time.Millisecond).Unix(),
			1,
		))
		Eventually(httpClient.postCount).Should(BeNumerically(">", 1))
		Eventually(httpClient.url).Should(Equal(
			"https://app.datadoghq.com/api/v1/series?api_key=api-key",
		))
		Eventually(httpClient.contentType).Should(Equal("application/json"))
		Eventually(httpClient.body).Should(MatchJSON(`{
			"series": [
				{
					"metric": "application.ingress",
					"points": [[1234, 4321]],
					"type": "gauge",
					"host": "abcdefg",
					"tags": [
						"application.instance:app-id/2"
					]
				},
				{
					"metric": "application.ingress",
					"points": [[1234, 4321]],
					"type": "gauge",
					"host": "abcdefg",
					"tags": [
						"application.instance:app-id/1"
					]
				}
			]
		}`))
	})
})

type spyPointBuilder struct {
	mu                    sync.Mutex
	_buildCalled          int
	_buildPointsTimestamp int64
}

func (s *spyPointBuilder) BuildPoints(timestamp int64) ([]datadog.Point, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s._buildCalled++
	s._buildPointsTimestamp = timestamp

	return []datadog.Point{
		{
			Metric: "application.ingress",
			Points: [][]int64{
				[]int64{int64(1234), int64(4321)},
			},
			Type: "gauge",
			Tags: []string{
				"application.instance:app-id/2",
			},
		},
		{
			Metric: "application.ingress",
			Points: [][]int64{
				[]int64{int64(1234), int64(4321)},
			},
			Type: "gauge",
			Tags: []string{
				"application.instance:app-id/1",
			},
		},
	}, nil
}

func (s *spyPointBuilder) buildCalled() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s._buildCalled
}

func (s *spyPointBuilder) buildPointsTimestamp() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s._buildPointsTimestamp
}

type spyReadCloser struct{}

func (s *spyReadCloser) Close() error {
	return nil
}

func (s *spyReadCloser) Read([]byte) (int, error) {
	return 0, nil
}

type spyHTTPClient struct {
	mu           sync.Mutex
	_postCount   int
	_url         string
	_contentType string
	_body        string
}

func (s *spyHTTPClient) Post(url string, contentType string, r io.Reader) (*http.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s._postCount++

	body, err := ioutil.ReadAll(r)
	Expect(err).ToNot(HaveOccurred())

	s._url = url
	s._contentType = contentType
	s._body = string(body)

	return &http.Response{StatusCode: 201, Body: &spyReadCloser{}}, nil
}

func (s *spyHTTPClient) url() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s._url
}

func (s *spyHTTPClient) contentType() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s._contentType
}

func (s *spyHTTPClient) body() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s._body
}

func (s *spyHTTPClient) postCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s._postCount
}
