package web_test

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"code.cloudfoundry.org/noisyneighbor/internal/store"
	"code.cloudfoundry.org/noisyneighbor/internal/web"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server", func() {
	Describe("/stats", func() {
		It("returns a list of top offenders", func() {
			checkToken := func(token, scope string) bool {
				return true
			}
			fakeRates := func() store.Rates {
				return []store.Rate{
					{
						Timestamp: 1234,
						Counts: map[string]uint64{
							"id-1": uint64(9999),
							"id-2": uint64(9999),
							"id-3": uint64(9999),
						},
					},
					{
						Timestamp: 12345,
						Counts: map[string]uint64{
							"id-1": uint64(9999),
							"id-2": uint64(9999),
							"id-3": uint64(9999),
						},
					},
				}
			}

			server := web.NewServer(0, fakeRates, checkToken)
			go server.Serve()
			defer server.Stop()

			var resp *http.Response
			Eventually(func() error {
				req, err := http.NewRequest(
					http.MethodGet,
					fmt.Sprintf("http://%s/offenders", server.Addr()),
					nil,
				)
				Expect(err).ToNot(HaveOccurred())

				resp, err = http.DefaultClient.Do(req)
				if err != nil {
					return err
				}

				return nil
			}).Should(Succeed())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			body, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())

			Expect(body).To(MatchJSON(`[
				{
					"timestamp": 1234,
					"counts": {
					"id-1": 9999,
					"id-2": 9999,
					"id-3": 9999
					}
				},
				{
					"timestamp": 12345,
					"counts": {
					"id-1": 9999,
					"id-2": 9999,
					"id-3": 9999
					}
				}
			]`))
		})
	})

	Describe("authentication", func() {
		It("returns a 401 Unauthorized when authentication fails", func() {
			checkToken := func(token, scope string) bool {
				return false
			}
			fakeRates := func() store.Rates {
				return []store.Rate{}
			}

			server := web.NewServer(0, fakeRates, checkToken)
			go server.Serve()
			defer server.Stop()

			var resp *http.Response
			Eventually(func() error {
				var err error
				resp, err = http.Get(fmt.Sprintf("http://%s/offenders", server.Addr()))
				if err != nil {
					return err
				}

				return nil
			}).Should(Succeed())

			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
		})
	})
})
