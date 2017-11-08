package app_test

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/internal/app"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Authenticator", func() {
	Describe("Token()", func() {
		It("returns an authentication token", func() {
			httpClient := &spyHTTPClient{
				statusCode: 200,
			}
			authenticator := app.NewAuthenticator(
				"id",
				"secret",
				"http://localhost",
				app.WithHTTPClient(httpClient),
			)

			token, err := authenticator.Token()
			Expect(err).ToNot(HaveOccurred())
			Expect(token).To(Equal("bearer my-token"))

			Expect(httpClient.url).To(Equal("http://localhost/oauth/token"))
			Expect(httpClient.body).To(Equal(url.Values{
				"response_type": {"token"},
				"grant_type":    {"client_credentials"},
				"client_id":     {"id"},
				"client_secret": {"secret"},
			}))
		})
	})

	Describe("CheckToken()", func() {
		It("returns true if UAA responds with 200", func() {
			httpClient := &spyHTTPClient{
				statusCode: 200,
			}
			a := app.NewAuthenticator("", "",
				"http://localhost",
				app.WithHTTPClient(httpClient),
			)

			Expect(a.CheckToken("token", "scope")).To(BeTrue())

			Expect(httpClient.url).To(Equal("http://localhost/check_token"))
			Expect(httpClient.body).To(Equal(url.Values{
				"token":  {"token"},
				"scopes": {"scope"},
			}))
		})

		It("returns false if UAA responds with non 200", func() {
			httpClient := &spyHTTPClient{
				statusCode: 401,
			}
			a := app.NewAuthenticator("", "",
				"http://localhost/api",
				app.WithHTTPClient(httpClient),
			)

			Expect(a.CheckToken("token", "scope")).To(BeFalse())
		})

		It("returns false if no token is given", func() {
			httpClient := &spyHTTPClient{
				statusCode: 200,
			}
			a := app.NewAuthenticator("", "",
				"http://localhost",
				app.WithHTTPClient(httpClient),
			)

			Expect(a.CheckToken("", "scope")).To(BeFalse())
		})

		It("returns false if no scope is given", func() {
			httpClient := &spyHTTPClient{
				statusCode: 200,
			}
			a := app.NewAuthenticator("", "",
				"http://localhost",
				app.WithHTTPClient(httpClient),
			)

			Expect(a.CheckToken("token", "")).To(BeFalse())
		})
	})
})

type spyHTTPClient struct {
	statusCode int

	url  string
	body url.Values
}

func (s *spyHTTPClient) PostForm(url string, data url.Values) (*http.Response, error) {
	s.url = url
	s.body = data

	reader := &spyReadCloser{
		strings.NewReader(`{"access_token": "my-token"}`),
	}

	return &http.Response{StatusCode: s.statusCode, Body: reader}, nil
}

func (s *spyHTTPClient) Do(r *http.Request) (*http.Response, error) {
	body, err := ioutil.ReadAll(r.Body)
	Expect(err).ToNot(HaveOccurred())
	defer r.Body.Close()

	s.url = r.URL.String()
	s.body, err = url.ParseQuery(string(body))
	Expect(err).ToNot(HaveOccurred())

	reader := &spyReadCloser{
		strings.NewReader(`{}`),
	}

	return &http.Response{StatusCode: s.statusCode, Body: reader}, nil
}

type spyReadCloser struct {
	*strings.Reader
}

func (s *spyReadCloser) Close() error {
	return nil
}
