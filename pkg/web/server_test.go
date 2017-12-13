package web_test

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/store"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/web"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server", func() {
	Describe("/rates/:timestamp", func() {
		It("returns rates from the collector", func() {
			server := web.NewServer(0, checkToken, &rateStore{},
				web.WithLogWriter(GinkgoWriter),
			)
			go server.Serve()
			defer server.Stop()

			req, err := http.NewRequest(
				http.MethodGet,
				fmt.Sprintf("http://%s/rates/1234", server.Addr()),
				nil,
			)
			Expect(err).ToNot(HaveOccurred())

			req.Header.Add("Authorization", "Bearer some-token")

			resp, err := http.DefaultClient.Do(req)
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(200))

			body, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())

			Expect(body).To(MatchJSON(`{
				"timestamp": 1234,
				"counts": {
					"id-1": 9999,
					"id-2": 9999,
					"id-3": 9999
				}
			}`))
		})

		It("returns a 401 if check token fails", func() {
			server := web.NewServer(0, checkTokenFailure, &rateStore{},
				web.WithLogWriter(GinkgoWriter),
			)
			go server.Serve()
			defer server.Stop()

			req, err := http.NewRequest(
				http.MethodGet,
				fmt.Sprintf("http://%s/rates/1234", server.Addr()),
				nil,
			)
			Expect(err).ToNot(HaveOccurred())

			req.Header.Add("Authorization", "Bearer some-token")

			resp, err := http.DefaultClient.Do(req)
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(401))
		})
	})
})

type rateStore struct {
	rateError error
}

func (f *rateStore) Rate(ts int64) (store.Rate, error) {
	return store.Rate{
		Timestamp: 1234,
		Counts: map[string]uint64{
			"id-1": uint64(9999),
			"id-2": uint64(9999),
			"id-3": uint64(9999),
		},
	}, f.rateError
}

func checkToken(_, _ string) bool {
	return true
}

func checkTokenFailure(_, _ string) bool {
	return false
}
