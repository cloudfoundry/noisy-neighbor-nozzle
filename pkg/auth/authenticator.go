package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// Authenticator stores authentication information that can be used to get an
// auth token.
type Authenticator struct {
	clientID     string
	clientSecret string
	uaaAddr      string
	httpClient   HTTPClient
}

// NewAuthenticator returns an initialized Authenticator. The authenticator, by
// default is configured to use http.DefaultClient. It is recommended that the
// authenticator is configured with the `WithHTTPClient` AuthenticatorOption.
func NewAuthenticator(id, secret, uaaAddr string, opts ...AuthenticatorOption) *Authenticator {
	a := &Authenticator{
		clientID:     id,
		clientSecret: secret,
		uaaAddr:      uaaAddr,
		httpClient:   http.DefaultClient,
	}

	for _, o := range opts {
		o(a)
	}

	return a
}

// RefreshAuthToken will request a new auth token from UAA.
func (a *Authenticator) RefreshAuthToken() (string, error) {
	response, err := a.httpClient.PostForm(a.uaaAddr+"/oauth/token", url.Values{
		"response_type": {"token"},
		"grant_type":    {"client_credentials"},
		"client_id":     {a.clientID},
		"client_secret": {a.clientSecret},
	})
	if err != nil {
		return "", err
	}
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Expected 200 status code from /oauth/token, got %d", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return "", err
	}

	oauthResponse := make(map[string]interface{})
	err = json.Unmarshal(body, &oauthResponse)
	if err != nil {
		return "", err
	}

	accessTokenInterface, ok := oauthResponse["access_token"]
	if !ok {
		return "", errors.New("No access_token on UAA oauth response")
	}

	accessToken, ok := accessTokenInterface.(string)
	if !ok {
		return "", errors.New("access_token on UAA oauth response not a string")
	}

	return accessToken, nil
}

// CheckToken validates an auth token with the UAA. It also ensures that the
// given auth token has permissions to a given scope.
func (a *Authenticator) CheckToken(token, scope string) bool {
	if token == "" || scope == "" {
		return false
	}

	form := url.Values{
		"token":  {token},
		"scopes": {scope},
	}
	req, err := http.NewRequest(
		http.MethodPost,
		a.uaaAddr+"/check_token",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		log.Fatalf("failed to build request to UAA: %s", err)
	}
	req.SetBasicAuth(a.clientID, a.clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := a.httpClient.Do(req)
	if err != nil {
		log.Printf("failed to check token: %s", err)
		return false
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Printf("expected 200 status code from /check_token, got %d", response.StatusCode)
		return false
	}

	return true
}

// HTTPClient is an interface that http.Client conforms to.
type HTTPClient interface {
	PostForm(string, url.Values) (*http.Response, error)
	Do(*http.Request) (*http.Response, error)
}

// AuthenticatorOption is a type of function that can be passed into
// NewAuthenticator for optional configuration.
type AuthenticatorOption func(*Authenticator)

// WithHTTPClient is an AuthenticatorOption to configure the HTTPClient to be
// used by the authenticator.
func WithHTTPClient(c HTTPClient) AuthenticatorOption {
	return func(a *Authenticator) {
		a.httpClient = c
	}
}
