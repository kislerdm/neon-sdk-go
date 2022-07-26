package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Client to interact with Neon API.
type Client interface {
	// ValidateAPIKey makes a call to validate API access key.
	ValidateAPIKey() error

	// ListProjects returns existing Projects.
	ListProjects() ([]ProjectInfo, error)

	// ReadInfoProject returns Project info.
	ReadInfoProject(projectID string) (ProjectInfo, error)

	// CreateProject creates new Project.
	CreateProject(settings ProjectSettingsRequestCreate) (ProjectInfo, error)

	// UpdateProject updates existing Project.
	UpdateProject(projectID string, settings ProjectSettingsRequestUpdate) (ProjectInfo, error)

	// DeleteProject deletes existing Project.
	DeleteProject(projectID string) (ProjectDeleteResponse, error)
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Options client options.
type Options struct {
	// APIKey API access key.
	APIKey string

	// HTTPClient client to communicate with the API over http.
	HTTPClient httpClient
}

type client struct {
	options Options

	baseURL string
}

type reqType string

const (
	get    reqType = "GET"
	post   reqType = "POST"
	patch  reqType = "PATCH"
	put    reqType = "PUT"
	delete reqType = "DELETE"
)

func (r reqType) String() string {
	return string(r)
}

type errorResp struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error API error.
type Error struct {
	HTTPCode int
	errorResp
}

func (e Error) Error() string {
	return fmt.Sprintf("[HTTP Code: %d][Error Code: %s] %s", e.HTTPCode, e.Code, e.Message)
}

func (c *client) marshal(v interface{}) (io.Reader, error) {
	body, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(body), nil
}

func (c *client) unmarshal(body io.ReadCloser, v interface{}) error {
	buf, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	return json.Unmarshal(buf, &v)
}

func (c *client) send(url string, t reqType, v interface{}) (*http.Response, error) {
	var body io.Reader
	var err error
	if v != nil {
		body, err = c.marshal(v)
		if err != nil {
			return nil, err
		}
	}

	req, _ := http.NewRequest(t.String(), url, body)
	setHeaders(req, c.options.APIKey)

	res, err := c.options.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode > 299 {
		return res, convertErrorResponse(res)
	}

	// cover non-existing object which will have 200+ status code
	// see the ticket https://github.com/neondatabase/neon/issues/2159
	if req.Method == get.String() && res.ContentLength < 10 {
		return nil, Error{
			HTTPCode: 404,
			errorResp: errorResp{
				Code:    "",
				Message: "object not found",
			},
		}
	}

	return res, nil
}

const baseURL = "https://console.neon.tech/api/v1/"

// NewClient initialised new client to communicate to Neon over http API.
// See details: https://neon.tech/docs/reference/about/
func NewClient(ctx context.Context, optFns ...func(*Options)) (Client, error) {
	o := Options{}
	for _, fn := range optFns {
		fn(&o)
	}

	resolveAPIKey(&o)
	resolveHTTPClient(&o)

	c := client{
		options: o,
		baseURL: baseURL,
	}

	if err := c.ValidateAPIKey(); err != nil {
		return nil, fmt.Errorf(
			"invalid API access key. find details: https://neon.tech/docs/get-started-with-neon/using-api-keys/",
		)
	}

	return &c, nil
}

func (c *client) ValidateAPIKey() error {
	if c.options.APIKey == "" {
		return fmt.Errorf("API key is not set")
	}

	_, err := c.send(c.baseURL+"users/me", get, nil)
	return err
}

func resolveHTTPClient(o *Options) {
	if o.HTTPClient == nil {
		o.HTTPClient = &http.Client{Timeout: 5 * time.Minute}
	}
}

func resolveAPIKey(o *Options) {
	if o.APIKey == "" {
		o.APIKey = os.Getenv("NEON_API_KEY")
	}
}

// WithAPIKey sets the API access key.
func WithAPIKey(apiKey string) func(*Options) {
	return func(o *Options) {
		o.APIKey = apiKey
	}
}

// WithHTTPClient sets custom http client.
func WithHTTPClient(client httpClient) func(*Options) {
	return func(o *Options) {
		o.HTTPClient = client
	}
}

func setHeaders(req *http.Request, token string) {
	req.Header.Add("accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}
}

func convertErrorResponse(res *http.Response) error {
	var v errorResp
	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return Error{
			HTTPCode: res.StatusCode,
			errorResp: errorResp{
				Message: "cannot read response bytes",
			},
		}
	}
	if err := json.Unmarshal(buf, &v); err != nil {
		return Error{
			HTTPCode: res.StatusCode,
			errorResp: errorResp{
				Message: err.Error(),
			},
		}
	}
	return Error{
		HTTPCode:  res.StatusCode,
		errorResp: v,
	}
}

func (c *client) ListProjects() ([]ProjectInfo, error) {
	res, err := c.send(c.baseURL+"projects", get, nil)
	defer res.Body.Close()

	if err != nil {
		return nil, err
	}

	var v []ProjectInfo
	if err := c.unmarshal(res.Body, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func (c *client) ReadInfoProject(projectID string) (ProjectInfo, error) {
	res, err := c.send(c.baseURL+"projects/"+projectID, get, nil)
	defer res.Body.Close()

	if err != nil {
		return ProjectInfo{}, err
	}

	var v ProjectInfo
	if err := c.unmarshal(res.Body, &v); err != nil {
		return ProjectInfo{}, err
	}
	return v, nil
}

func (c *client) CreateProject(settings ProjectSettingsRequestCreate) (ProjectInfo, error) {
	res, err := c.send(c.baseURL+"projects", post, settings)
	defer res.Body.Close()

	if err != nil {
		return ProjectInfo{}, err
	}

	var v ProjectInfo
	if err := c.unmarshal(res.Body, &v); err != nil {
		return ProjectInfo{}, err
	}
	return v, nil
}

func (c *client) UpdateProject(projectID string, settings ProjectSettingsRequestUpdate) (ProjectInfo, error) {
	res, err := c.send(c.baseURL+"projects/"+projectID, patch, settings)
	defer res.Body.Close()

	if err != nil {
		return ProjectInfo{}, err
	}

	var v ProjectInfo
	if err := c.unmarshal(res.Body, &v); err != nil {
		return ProjectInfo{}, err
	}
	return v, nil
}

func (c *client) DeleteProject(projectID string) (ProjectDeleteResponse, error) {
	res, err := c.send(c.baseURL+"projects/"+projectID+"/delete", post, nil)
	defer res.Body.Close()

	if err != nil {
		return ProjectDeleteResponse{}, err
	}

	var v ProjectDeleteResponse
	if err := c.unmarshal(res.Body, &v); err != nil {
		return ProjectDeleteResponse{}, err
	}
	return v, nil
}
