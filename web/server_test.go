package web_test

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"code.cloudfoundry.org/noisyneighbor/cache"
	"code.cloudfoundry.org/noisyneighbor/web"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server", func() {
	Describe("/offenders", func() {
		It("returns a list of top offenders", func() {
			fakeCacheTopN := func(n int) []cache.Stat {
				return []cache.Stat{
					{ID: "id-1", Count: 1234567},
					{ID: "id-2", Count: 123456},
					{ID: "id-3", Count: 12345},
					{ID: "id-4", Count: 1234},
					{ID: "id-5", Count: 123},
				}
			}

			server := web.NewServer(0, "username", "password", fakeCacheTopN)
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

			Expect(body).To(MatchJSON(`[
				{"id": "id-1", "count": 1234567},	
				{"id": "id-2", "count": 123456},	
				{"id": "id-3", "count": 12345},	
				{"id": "id-4", "count": 1234},	
				{"id": "id-5", "count": 123}
			]`))
		})
	})

	Describe("authentication", func() {
		It("returns a 401 Unauthorized without basic auth credentials", func() {
			fakeCacheTopN := func(n int) []cache.Stat {
				return []cache.Stat{}
			}

			server := web.NewServer(0, "username", "password", fakeCacheTopN)
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
			fakeCacheTopN := func(n int) []cache.Stat {
				return []cache.Stat{}
			}

			server := web.NewServer(0, "username", "password", fakeCacheTopN)
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
