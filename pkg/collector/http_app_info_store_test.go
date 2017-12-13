package collector_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/collector"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HTTPAppInfoStore", func() {
	It("issues GET requests to Cloud Controller for AppInfo", func() {
		client := &fakeHTTPClient{responses: happyPath()}
		auth := &spyAuthenticator{refreshToken: "valid-token"}
		store := collector.NewHTTPAppInfoStore("http://api.addr.com", client, auth)

		actual, err := store.Lookup([]string{"a", "b"})

		Expect(err).ToNot(HaveOccurred())
		expected := map[collector.AppGUID]collector.AppInfo{
			"a": collector.AppInfo{
				Name:  "app1",
				Space: "space1",
				Org:   "org1",
			},
			"b": collector.AppInfo{
				Name:  "app2",
				Space: "space2",
				Org:   "org2",
			},
		}
		Expect(actual).To(Equal(expected))
		Expect(client.validAuth).To(Equal([]bool{true, true, true}))
		Expect(client.requests).To(HaveLen(3))

		req := client.requests[0]
		Expect(req.URL.Host).To(Equal("api.addr.com"))
		Expect(req.URL.Path).To(Equal("/v3/apps"))
		Expect(req.URL.Query().Get("guids")).To(Or(
			Equal("a,b"),
			Equal("b,a"),
		))
		Expect(req.URL.Query().Get("per_page")).To(Equal("5000"))

		req = client.requests[1]
		Expect(req.URL.Host).To(Equal("api.addr.com"))
		Expect(req.URL.Path).To(Equal("/v2/organizations"))
		Expect(req.URL.Query().Get("q")).To(Or(
			Equal("space_guid IN e,f"),
			Equal("space_guid IN f,e"),
		))
		Expect(req.URL.Query().Get("results-per-page")).To(Equal("100"))

		req = client.requests[2]
		Expect(req.URL.Host).To(Equal("api.addr.com"))
		Expect(req.URL.Path).To(Equal("/v3/spaces"))
		Expect(req.URL.Query().Get("organization_guids")).To(Or(
			Equal("g,h"),
			Equal("h,g"),
		))
		Expect(req.URL.Query().Get("per_page")).To(Equal("5000"))
	})

	It("acquires a token from the authenticator", func() {
		client := &fakeHTTPClient{responses: happyPath()}
		auth := &spyAuthenticator{refreshToken: "valid-token"}
		store := collector.NewHTTPAppInfoStore("http://api.addr.com", client, auth)

		store.Lookup([]string{"a", "b"})

		Expect(auth.refreshCalled).To(BeTrue())
	})

	It("returns an error when the authenticator fails", func() {
		client := &fakeHTTPClient{responses: happyPath()}
		auth := &spyAuthenticator{refreshError: errors.New("request failed")}
		store := collector.NewHTTPAppInfoStore("http://api.addr.com", client, auth)

		_, err := store.Lookup([]string{"a", "b"})

		Expect(err).To(HaveOccurred())
	})

	It("returns an empty map when no GUIDInstances are passed in", func() {
		client := &fakeHTTPClient{responses: happyPath()}
		auth := &spyAuthenticator{}
		store := collector.NewHTTPAppInfoStore("http://api.addr.com", client, auth)

		data, _ := store.Lookup(nil)

		Expect(data).To(HaveLen(0))
		Expect(client.doCalled).To(BeFalse())
		Expect(auth.refreshCalled).To(BeFalse())
	})

	It("returns an error when getting apps fails", func() {
		client := &fakeHTTPClient{responses: appRequestFailed()}
		auth := &spyAuthenticator{refreshToken: "valid-token"}
		store := collector.NewHTTPAppInfoStore("http://api.addr.com", client, auth)

		_, err := store.Lookup([]string{"a", "b"})
		Expect(err).To(HaveOccurred())
	})

	It("returns an error when getting orgs fails", func() {
		client := &fakeHTTPClient{responses: orgRequestFailed()}
		auth := &spyAuthenticator{refreshToken: "valid-token"}
		store := collector.NewHTTPAppInfoStore("http://api.addr.com", client, auth)

		_, err := store.Lookup([]string{"a", "b"})
		Expect(err).To(HaveOccurred())
	})

	It("returns an error when getting spaces fails", func() {
		client := &fakeHTTPClient{responses: spaceRequestFailed()}
		auth := &spyAuthenticator{refreshToken: "valid-token"}
		store := collector.NewHTTPAppInfoStore("http://api.addr.com", client, auth)

		_, err := store.Lookup([]string{"a", "b"})
		Expect(err).To(HaveOccurred())
	})

	It("returns an error when the app request is not a 200 status", func() {
		client := &fakeHTTPClient{responses: appRequestNotOK()}
		auth := &spyAuthenticator{refreshToken: "valid-token"}
		store := collector.NewHTTPAppInfoStore("http://api.addr.com", client, auth)

		_, err := store.Lookup([]string{"a", "b"})
		Expect(err).To(HaveOccurred())
	})

	It("returns an error when the org request is not a 200 status", func() {
		client := &fakeHTTPClient{responses: orgRequestNotOK()}
		auth := &spyAuthenticator{refreshToken: "valid-token"}
		store := collector.NewHTTPAppInfoStore("http://api.addr.com", client, auth)

		_, err := store.Lookup([]string{"a", "b"})
		Expect(err).To(HaveOccurred())
	})

	It("returns an error when the spaces request is not a 200 status", func() {
		client := &fakeHTTPClient{responses: spaceRequestNotOK()}
		auth := &spyAuthenticator{refreshToken: "valid-token"}
		store := collector.NewHTTPAppInfoStore("http://api.addr.com", client, auth)

		_, err := store.Lookup([]string{"a", "b"})
		Expect(err).To(HaveOccurred())
	})

	It("returns an error when app json unmarshalling fails", func() {
		client := &fakeHTTPClient{responses: appRequestInvalidJSON()}
		auth := &spyAuthenticator{refreshToken: "valid-token"}
		store := collector.NewHTTPAppInfoStore("http://api.addr.com", client, auth)

		_, err := store.Lookup([]string{"a", "b"})
		Expect(err).To(HaveOccurred())
	})

	It("returns an error when org json unmarshalling fails", func() {
		client := &fakeHTTPClient{responses: orgRequestInvalidJSON()}
		auth := &spyAuthenticator{refreshToken: "valid-token"}
		store := collector.NewHTTPAppInfoStore("http://api.addr.com", client, auth)

		_, err := store.Lookup([]string{"a", "b"})
		Expect(err).To(HaveOccurred())
	})

	It("returns an error when space json unmarshalling fails", func() {
		client := &fakeHTTPClient{responses: spaceRequestInvalidJSON()}
		auth := &spyAuthenticator{refreshToken: "valid-token"}
		store := collector.NewHTTPAppInfoStore("http://api.addr.com", client, auth)

		_, err := store.Lookup([]string{"a", "b"})
		Expect(err).To(HaveOccurred())
	})
})

type fakeHTTPClient struct {
	doCalled  bool
	responses map[string]response
	validAuth []bool
	requests  []*http.Request
}

func (f *fakeHTTPClient) Do(req *http.Request) (*http.Response, error) {
	f.doCalled = true
	f.validAuth = append(f.validAuth, req.Header.Get("Authorization") == "bearer valid-token")
	f.requests = append(f.requests, req)
	resp, ok := f.responses[req.URL.Path]
	if !ok {
		return nil, nil
	}
	return resp.http, resp.err
}

type response struct {
	http *http.Response
	err  error
}

func happyPath() map[string]response {
	appsResp := response{
		http: &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(appsResponse())),
		},
		err: nil,
	}
	orgResp := response{
		http: &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(orgsResponse())),
		},
		err: nil,
	}
	spaceResp := response{
		http: &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(spacesResponse())),
		},
		err: nil,
	}
	return map[string]response{
		"/v3/apps":          appsResp,
		"/v2/organizations": orgResp,
		"/v3/spaces":        spaceResp,
	}
}

func appRequestFailed() map[string]response {
	appsResp := response{
		http: nil,
		err:  errors.New("request failed"),
	}
	return map[string]response{
		"/v3/apps": appsResp,
	}
}

func appRequestNotOK() map[string]response {
	appsResp := response{
		http: &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       ioutil.NopCloser(strings.NewReader("")),
		},
		err: nil,
	}

	return map[string]response{
		"/v3/apps": appsResp,
	}
}

func appRequestInvalidJSON() map[string]response {
	appsResp := response{
		http: &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader("{")),
		},
		err: nil,
	}

	return map[string]response{
		"/v3/apps": appsResp,
	}
}

func orgRequestFailed() map[string]response {
	appsResp := response{
		http: &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(appsResponse())),
		},
		err: nil,
	}
	orgResp := response{
		http: nil,
		err:  errors.New("request failed"),
	}
	return map[string]response{
		"/v3/apps":          appsResp,
		"/v2/organizations": orgResp,
	}
}

func orgRequestInvalidJSON() map[string]response {
	appsResp := response{
		http: &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(appsResponse())),
		},
		err: nil,
	}
	orgResp := response{
		http: &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader("{")),
		},
		err: nil,
	}
	return map[string]response{
		"/v3/apps":          appsResp,
		"/v2/organizations": orgResp,
	}
}

func orgRequestNotOK() map[string]response {
	appsResp := response{
		http: &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(appsResponse())),
		},
		err: nil,
	}
	orgResp := response{
		http: &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       ioutil.NopCloser(strings.NewReader("")),
		},
		err: nil,
	}

	return map[string]response{
		"/v3/apps":          appsResp,
		"/v2/organizations": orgResp,
	}
}

func spaceRequestFailed() map[string]response {
	appsResp := response{
		http: &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(appsResponse())),
		},
		err: nil,
	}
	orgResp := response{
		http: &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(orgsResponse())),
		},
		err: nil,
	}
	spaceResp := response{
		http: nil,
		err:  errors.New("request failed"),
	}
	return map[string]response{
		"/v3/apps":          appsResp,
		"/v2/organizations": orgResp,
		"/v3/spaces":        spaceResp,
	}
}

func spaceRequestNotOK() map[string]response {
	appsResp := response{
		http: &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(appsResponse())),
		},
		err: nil,
	}
	orgResp := response{
		http: &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(orgsResponse())),
		},
		err: nil,
	}
	spaceResp := response{
		http: &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       ioutil.NopCloser(strings.NewReader("")),
		},
		err: nil,
	}
	return map[string]response{
		"/v3/apps":          appsResp,
		"/v2/organizations": orgResp,
		"/v3/spaces":        spaceResp,
	}
}

func spaceRequestInvalidJSON() map[string]response {
	appsResp := response{
		http: &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(appsResponse())),
		},
		err: nil,
	}
	orgResp := response{
		http: &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(orgsResponse())),
		},
		err: nil,
	}
	spaceResp := response{
		http: &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader("{")),
		},
		err: nil,
	}
	return map[string]response{
		"/v3/apps":          appsResp,
		"/v2/organizations": orgResp,
		"/v3/spaces":        spaceResp,
	}
}

func appsResponse() string {
	return `
{
  "resources": [
    {
      "guid": "a",
      "name": "app1",
      "relationships": {
        "space": {
          "data": {
            "guid": "e"
          }
        }
      }
    },
    {
      "guid": "b",
      "name": "app2",
      "relationships": {
        "space": {
          "data": {
            "guid": "f"
          }
        }
      }
    }
  ]
}
`
}

func orgsResponse() string {
	return `
{
  "resources": [
    {
      "metadata": {
        "guid": "g"
      },
      "entity": {
        "name": "org1"
      }
    },
    {
      "metadata": {
        "guid": "h"
      },
      "entity": {
        "name": "org2"
      }
    }
  ]
}
`
}

func spacesResponse() string {
	return `
{
  "resources": [
    {
      "guid": "e",
      "name": "space1",
      "relationships": {
        "organization": {
          "data": {
            "guid": "g"
          }
        }
      }
    },
    {
      "guid": "f",
      "name": "space2",
      "relationships": {
        "organization": {
          "data": {
            "guid": "h"
          }
        }
      }
    }
  ]
}
`
}
