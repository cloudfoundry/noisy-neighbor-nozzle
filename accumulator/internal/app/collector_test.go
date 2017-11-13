package app_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/accumulator/internal/app"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/accumulator/internal/datadogreporter"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Collector", func() {
	Describe("BuildPoints", func() {
		It("returns Points", func() {
			ts1 := time.Now().Add(time.Minute).Unix()
			testServer, requests := setupTestServer(ts1, http.StatusOK)
			defer testServer.Close()

			c := app.NewCollector([]string{testServer.URL})

			points, err := c.BuildPoints(ts1)
			Expect(err).ToNot(HaveOccurred())
			Expect(points).To(HaveLen(2))

			var request request
			Expect(requests).To(Receive(&request))
			Expect(request.url.Path).To(Equal(fmt.Sprintf("/state/%d", ts1)))

			point := findPointWithTag("application.instance:app-1/0", points)
			Expect(point).ToNot(BeZero())
			Expect(point.Metric).To(Equal("application.ingress"))
			Expect(point.Points).To(Equal([][]int64{
				[]int64{ts1, 1186},
			}))

			point = findPointWithTag("application.instance:app-1/1", points)
			Expect(point).ToNot(BeZero())
			Expect(point.Metric).To(Equal("application.ingress"))
			Expect(point.Points).To(Equal([][]int64{
				[]int64{ts1, 966},
			}))
		})

		It("sums counts from multiple nozzles", func() {
			ts1 := time.Now().Add(time.Minute).Unix()

			serverA, requestsA := setupTestServer(ts1, http.StatusOK)
			serverB, requestsB := setupTestServer(ts1, http.StatusOK)
			defer serverA.Close()
			defer serverB.Close()

			_ = requestsA
			_ = requestsB

			c := app.NewCollector([]string{serverA.URL, serverB.URL})

			points, err := c.BuildPoints(ts1)
			Expect(err).ToNot(HaveOccurred())

			point := findPointWithTag("application.instance:app-1/0", points)
			Expect(point).ToNot(BeZero())
			Expect(point.Metric).To(Equal("application.ingress"))
			Expect(point.Points).To(Equal([][]int64{
				[]int64{ts1, 2372},
			}))

			point = findPointWithTag("application.instance:app-1/1", points)
			Expect(point).ToNot(BeZero())
			Expect(point.Metric).To(Equal("application.ingress"))
			Expect(point.Points).To(Equal([][]int64{
				[]int64{ts1, 1932},
			}))
		})

		It("returns an error if any of the nozzles return a non 200 status code", func() {
			ts1 := time.Now().Add(time.Minute).Unix()
			serverA, _ := setupTestServer(ts1, http.StatusOK)
			serverB, _ := setupTestServer(ts1, http.StatusNotFound)
			defer serverA.Close()
			defer serverB.Close()

			c := app.NewCollector([]string{serverA.URL, serverB.URL})

			_, err := c.BuildPoints(ts1)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Sum", func() {
		It("aggregates rates into a single list of rates", func() {
			rates := []app.Rate{
				{
					Timestamp: 60,
					Counts: map[string]uint64{
						"app-1/0": 10,
						"app-1/1": 20,
						"app-2/0": 30,
					},
				},
				{
					Timestamp: 60,
					Counts: map[string]uint64{
						"app-1/0": 10,
						"app-1/1": 20,
						"app-2/0": 30,
					},
				},
			}

			results := app.Sum(rates)

			Expect(results).To(Equal(
				app.Rate{
					Timestamp: 60,
					Counts: map[string]uint64{
						"app-1/0": 20,
						"app-1/1": 40,
						"app-2/0": 60,
					},
				},
			))
		})
	})
})

type request struct {
	url *url.URL
}

func setupTestServer(ts1 int64, statusCode int) (*httptest.Server, chan request) {
	requests := make(chan request, 100)

	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests <- request{
				url: r.URL,
			}

			w.WriteHeader(statusCode)
			w.Write([]byte(fmt.Sprintf(`
				{
					"counts": {
					  "app-1/1": 966,
					  "app-1/0": 1186
					},
					"timestamp": %d
				},
			]`, ts1)))
		}),
	), requests
}

func findPointWithTag(tag string, points []datadogreporter.Point) datadogreporter.Point {
	for _, p := range points {
		if len(p.Tags) < 1 {
			continue
		}

		if p.Tags[0] == tag {
			return p
		}
	}

	return datadogreporter.Point{}
}
