package web_test

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"code.cloudfoundry.org/noisyneighbor/internal/web"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server", func() {
	Describe("/stats", func() {
		It("returns a list of top offenders", func() {
			fakestoreTopN := func() map[int64]map[string]uint64 {
				return map[int64]map[string]uint64{
					1234: {
						"id-1": uint64(9999),
						"id-2": uint64(9999),
						"id-3": uint64(9999),
					},
					12345: {
						"id-1": uint64(9999),
						"id-2": uint64(9999),
						"id-3": uint64(9999),
					},
				}
			}

			server := web.NewServer(0, "username", "password", fakestoreTopN)
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

				req.SetBasicAuth("username", "password")

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

			Expect(body).To(MatchJSON(`{
				"1234": {
					"id-1": 9999,
					"id-2": 9999,
					"id-3": 9999
				},
				"12345": {
					"id-1": 9999,
					"id-2": 9999,
					"id-3": 9999
				}
			}`))
		})
	})

	Describe("authentication", func() {
		It("returns a 401 Unauthorized without basic auth credentials", func() {
			fakestoreTopN := func() map[int64]map[string]uint64 {
				return make(map[int64]map[string]uint64)
			}

			server := web.NewServer(0, "username", "password", fakestoreTopN)
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

		It("returns a 401 Unauthorized with invalid credentials", func() {
			fakestoreTopN := func() map[int64]map[string]uint64 {
				return make(map[int64]map[string]uint64)
			}

			server := web.NewServer(0, "username", "password", fakestoreTopN)
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

				req.SetBasicAuth("invalid-username", "invalid-password")

				resp, err = http.DefaultClient.Do(req)
				if err != nil {
					return err
				}

				return nil
			}).Should(Succeed())

			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
		})
	})
})
