package app_test

import (
	"net/http"
	"net/url"
	"strings"

	"code.cloudfoundry.org/noisyneighbor/internal/app"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Authenticator", func() {
	It("returns an authentication token", func() {
		httpClient := &spyHTTPClient{}
		authenticator := app.NewAuthenticator(
			"id",
			"secret",
			"http://localhost/api",
			app.WithHTTPClient(httpClient),
		)

		token, err := authenticator.Token()
		Expect(err).ToNot(HaveOccurred())
		Expect(token).To(Equal("bearer my-token"))

		Expect(httpClient.url()).To(Equal("http://localhost/api/oauth/token"))
		Expect(httpClient.body()).To(Equal(url.Values{
			"response_type": {"token"},
			"grant_type":    {"client_credentials"},
			"client_id":     {"id"},
			"client_secret": {"secret"},
		}))
	})
})

type spyHTTPClient struct {
	_url  string
	_body url.Values
}

func (s *spyHTTPClient) PostForm(url string, data url.Values) (*http.Response, error) {
	s._url = url
	s._body = data

	reader := &spyReadCloser{
		strings.NewReader(`{"access_token": "my-token"}`),
	}

	return &http.Response{StatusCode: 200, Body: reader}, nil
}

func (s *spyHTTPClient) url() string {
	return s._url
}

func (s *spyHTTPClient) body() url.Values {
	return s._body
}

type spyReadCloser struct {
	*strings.Reader
}

func (s *spyReadCloser) Close() error {
	return nil
}
