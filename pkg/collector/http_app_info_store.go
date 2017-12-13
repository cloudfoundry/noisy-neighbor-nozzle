package collector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// AppGUID represets an application GUID.
type AppGUID string
type spaceGUID string
type orgGUID string

const (
	defaultV2PerPage = "100"
	defaultV3PerPage = "5000"
)

// HTTPClient supports HTTP requests.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// HTTPAppInfoStore provides a focused source of Cloud Controller API data.
type HTTPAppInfoStore struct {
	apiAddr string
	client  HTTPClient
	auth    Authenticator
}

// NewHTTPAppInfoStore initializes an APIStore and sends all HTTP requests to
// the API URL specified by apiAddr.
func NewHTTPAppInfoStore(
	apiAddr string,
	client HTTPClient,
	auth Authenticator,
) *HTTPAppInfoStore {
	return &HTTPAppInfoStore{
		apiAddr: apiAddr,
		client:  client,
		auth:    auth,
	}
}

// Lookup reads AppInfo from a remote API.
func (s *HTTPAppInfoStore) Lookup(guids []string) (map[AppGUID]AppInfo, error) {
	if len(guids) < 1 {
		return nil, nil
	}
	token, err := s.auth.RefreshAuthToken()
	if err != nil {
		return nil, err
	}
	token = fmt.Sprintf("bearer %s", token)

	appSpaces, err := s.lookupAppNames(guids, token)
	if err != nil {
		return nil, err
	}
	var spaceGUIDs []string
	for _, v := range appSpaces {
		spaceGUIDs = append(spaceGUIDs, v.spaceGUID)
	}

	orgs, err := s.lookupOrgs(spaceGUIDs, token)
	if err != nil {
		return nil, err
	}
	var orgGUIDs []string
	for _, v := range orgs {
		orgGUIDs = append(orgGUIDs, v.guid)
	}

	spaces, err := s.lookupSpaces(orgGUIDs, token)
	if err != nil {
		return nil, err
	}

	res := make(map[AppGUID]AppInfo)
	for k, v := range appSpaces {
		space := spaces[spaceGUID(v.spaceGUID)]
		org := orgs[orgGUID(space.orgGUID)]
		res[k] = AppInfo{
			Name:  v.name,
			Space: space.name,
			Org:   org.name,
		}
	}

	return res, nil
}

func (s *HTTPAppInfoStore) lookupAppNames(guids []string, authToken string) (map[AppGUID]app, error) {
	u, err := url.Parse(fmt.Sprintf("%s/v3/apps", s.apiAddr))
	if err != nil {
		return nil, err
	}

	query := url.Values{
		"guids":    {strings.Join(guids, ",")},
		"per_page": {defaultV3PerPage},
	}
	u.RawQuery = query.Encode()

	request, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", authToken)

	r, err := s.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		buf := bytes.NewBuffer(nil)
		_, _ = buf.ReadFrom(r.Body)
		err := fmt.Errorf("failed to get apps, expected 200, got %d: %s", r.StatusCode, buf.String())

		return nil, err
	}

	var resp V3Response
	err = json.NewDecoder(r.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}

	appSpaces := make(map[AppGUID]app)

	for _, r := range resp.Resources {
		appSpaces[AppGUID(r.GUID)] = app{
			name:      r.Name,
			guid:      r.GUID,
			spaceGUID: r.Relationships["space"].Data.GUID,
		}
	}
	return appSpaces, nil
}

func (s *HTTPAppInfoStore) lookupOrgs(spaceGUIDs []string, authToken string) (map[orgGUID]org, error) {
	u, err := url.Parse(fmt.Sprintf("%s/v2/organizations", s.apiAddr))
	if err != nil {
		return nil, err
	}

	query := url.Values{
		"q":                {fmt.Sprintf("space_guid IN %s", strings.Join(spaceGUIDs, ","))},
		"results-per-page": {defaultV2PerPage},
	}

	u.RawQuery = query.Encode()

	request, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", authToken)

	r, err := s.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		buf := bytes.NewBuffer(nil)
		_, _ = buf.ReadFrom(r.Body)
		err := fmt.Errorf("failed to get orgs, expected 200, got %d: %s", r.StatusCode, buf.String())

		return nil, err
	}

	var resp V2Response
	err = json.NewDecoder(r.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}

	orgs := make(map[orgGUID]org)

	for _, r := range resp.Resources {
		orgs[orgGUID(r.Metadata.GUID)] = org{
			name: r.Entity.Name,
			guid: r.Metadata.GUID,
		}
	}
	return orgs, nil
}

func (s *HTTPAppInfoStore) lookupSpaces(orgGUIDs []string, authToken string) (map[spaceGUID]space, error) {
	u, err := url.Parse(fmt.Sprintf("%s/v3/spaces", s.apiAddr))
	if err != nil {
		return nil, err
	}

	query := url.Values{
		"organization_guids": {strings.Join(orgGUIDs, ",")},
		"per_page":           {defaultV3PerPage},
	}
	u.RawQuery = query.Encode()

	request, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	request.Header.Add("Authorization", authToken)

	r, err := s.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		buf := bytes.NewBuffer(nil)
		_, _ = buf.ReadFrom(r.Body)
		err := fmt.Errorf("failed to get spaces, expected 200, got %d: %s", r.StatusCode, buf.String())

		return nil, err
	}

	var resp V3Response
	err = json.NewDecoder(r.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}

	spaces := make(map[spaceGUID]space)

	for _, r := range resp.Resources {
		spaces[spaceGUID(r.GUID)] = space{
			name:    r.Name,
			guid:    r.GUID,
			orgGUID: r.Relationships["organization"].Data.GUID,
		}
	}
	return spaces, nil
}

// AppInfo holds the names of an application, space, and organization.
type AppInfo struct {
	Name  string
	Space string
	Org   string
}

// String implements the Stringer interface.
func (a AppInfo) String() string {
	return fmt.Sprintf("%s.%s.%s", a.Org, a.Space, a.Name)
}

// V3Resource represents application data returned from the Cloud Controller
// API.
type V3Resource struct {
	GUID          string                    `json:"guid"`
	Name          string                    `json:"name"`
	Relationships map[string]V3Relationship `json:"relationships"`
}

// V3Response represents a list of V3 API resources and associated data.
type V3Response struct {
	Resources []V3Resource `json:"resources"`
}

// V3Relationship represents a V3 API resource relationship.
type V3Relationship struct {
	Data struct {
		GUID string `json:"guid"`
	} `json:"data"`
}

// V2Response represents a list of V2 API resources.
type V2Response struct {
	Resources []V2Resource `json:"resources"`
}

// V2Resource represents a single V2 API resource
type V2Resource struct {
	Metadata struct {
		GUID string `json:"guid"`
	} `json:"metadata"`

	Entity struct {
		Name string `json:"name"`
	} `json:"entity"`
}

type space struct {
	guid    string
	name    string
	orgGUID string
}

type org struct {
	guid string
	name string
}

type app struct {
	guid      string
	name      string
	spaceGUID string
}
