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

	// ListProjects returns existing projects.
	ListProjects() ([]ProjectInfo, error)

	// ReadProject returns project info.
	ReadProject(projectID string) (ProjectInfo, error)

	// CreateProject creates new project.
	CreateProject(settings ProjectSettingsRequestCreate) (ProjectInfo, error)

	// UpdateProject updates existing project.
	UpdateProject(projectID string, settings ProjectSettingsRequestUpdate) (ProjectInfo, error)

	// DeleteProject deletes existing project.
	DeleteProject(projectID string) (ProjectStatus, error)

	// StartProject starts the project.
	StartProject(projectID string) (ProjectStatus, error)

	// StopProject starts running project.
	StopProject(projectID string) (ProjectStatus, error)

	// ListDatabases return existing databases.
	ListDatabases(projectID string) ([]DatabaseResponse, error)

	// CreateDatabase creates new database in the project.
	CreateDatabase(projectID string, cfg DatabaseRequest) (DatabaseResponse, error)

	// UpdateDatabase updates the database in the project.
	UpdateDatabase(projectID string, databaseID int, cfg DatabaseRequest) (DatabaseResponse, error)

	// DeleteDatabase deletes the database in the project.
	DeleteDatabase(projectID string, databaseID int) (DatabaseResponse, error)

	// ReadDatabase returns database info.
	ReadDatabase(projectID string, databaseID int) (DatabaseResponse, error)

	// ListRoles return existing roles.
	ListRoles(projectID string) ([]RoleResponse, error)

	// CreateRole creates new role in the project.
	CreateRole(projectID string, cfg RoleRequest) (RoleResponse, error)

	// ReadRole returns role info.
	ReadRole(projectID string, roleName string) (RoleResponse, error)

	// DeleteRole deletes the role in the project.
	DeleteRole(projectID string, roleName string) (RoleResponse, error)

	// ResetRolePassword resets the role's password.
	ResetRolePassword(projectID string, roleName string) (RolePasswordResponse, error)
}

// HTTPClient client to handle http requests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
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

type options struct {
	// APIKey API access key.
	APIKey string

	// HTTPClient client to communicate with the API over http.
	HTTPClient HTTPClient
}

type client struct {
	options options

	baseURL string
}

func (c *client) ListRoles(projectID string) ([]RoleResponse, error) {
	var v []RoleResponse
	if err := c.requestHandler(c.baseURL+"projects/"+projectID+"/roles", "GET", nil, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func (c *client) CreateRole(projectID string, cfg RoleRequest) (RoleResponse, error) {
	var v RoleResponse
	if err := c.requestHandler(c.baseURL+"projects/"+projectID+"/roles", "POST", cfg, &v); err != nil {
		return RoleResponse{}, err
	}
	return v, nil
}

func (c *client) ReadRole(projectID string, roleName string) (RoleResponse, error) {
	var v RoleResponse
	if err := c.requestHandler(c.baseURL+"projects/"+projectID+"/roles/"+roleName, "GET", nil, &v); err != nil {
		return RoleResponse{}, err
	}
	return v, nil
}

func (c *client) DeleteRole(projectID string, roleName string) (RoleResponse, error) {
	var v RoleResponse
	if err := c.requestHandler(c.baseURL+"projects/"+projectID+"/roles/"+roleName, "DELETE", nil, &v); err != nil {
		return RoleResponse{}, err
	}
	return v, nil
}

func (c *client) ResetRolePassword(projectID string, roleName string) (RolePasswordResponse, error) {
	var v RolePasswordResponse
	if err := c.requestHandler(
		c.baseURL+"projects/"+projectID+"/roles/"+roleName+"/reset_password", "POST", nil, &v,
	); err != nil {
		return RolePasswordResponse{}, err
	}
	return v, nil
}

func (c *client) ListDatabases(projectID string) ([]DatabaseResponse, error) {
	var v []DatabaseResponse
	if err := c.requestHandler(c.baseURL+"projects/"+projectID+"/databases", "GET", nil, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func (c *client) ReadDatabase(projectID string, databaseID int) (DatabaseResponse, error) {
	var v DatabaseResponse
	if err := c.requestHandler(
		fmt.Sprintf("%sprojects/%s/databases/%d", c.baseURL, projectID, databaseID), "GET", nil, &v,
	); err != nil {
		return DatabaseResponse{}, err
	}
	return v, nil
}

func (c *client) CreateDatabase(projectID string, cfg DatabaseRequest) (DatabaseResponse, error) {
	var v DatabaseResponse
	if err := c.requestHandler(c.baseURL+"projects/"+projectID+"/databases", "POST", cfg, &v); err != nil {
		return DatabaseResponse{}, err
	}
	return v, nil
}

func (c *client) UpdateDatabase(projectID string, databaseID int, cfg DatabaseRequest) (DatabaseResponse, error) {
	var v DatabaseResponse
	if err := c.requestHandler(
		fmt.Sprintf("%sprojects/%s/databases/%d", c.baseURL, projectID, databaseID), "PUT", cfg, &v,
	); err != nil {
		return DatabaseResponse{}, err
	}
	return v, nil
}

func (c *client) DeleteDatabase(projectID string, databaseID int) (DatabaseResponse, error) {
	var v DatabaseResponse
	if err := c.requestHandler(
		fmt.Sprintf("%sprojects/%s/databases/%d", c.baseURL, projectID, databaseID), "DELETE", nil, &v,
	); err != nil {
		return DatabaseResponse{}, err
	}
	return v, nil
}

func (c *client) requestHandler(url string, t string, reqPayload interface{}, responsePayload interface{}) error {
	var body io.Reader
	var err error

	if reqPayload != nil {
		b, err := json.Marshal(reqPayload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}

	req, _ := http.NewRequest(t, url, body)
	setHeaders(req, c.options.APIKey)

	res, err := c.options.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode > 299 {
		return convertErrorResponse(res)
	}

	// cover non-existing object which will have 200+ status code
	// see the ticket https://github.com/neondatabase/neon/issues/2159
	if req.Method == "GET" && res.ContentLength < 10 {
		return Error{
			HTTPCode: 404,
			errorResp: errorResp{
				Code:    "",
				Message: "object not found",
			},
		}
	}

	if responsePayload != nil {
		buf, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		return json.Unmarshal(buf, responsePayload)
	}

	return nil
}

const baseURL = "https://console.neon.tech/api/v1/"

// NewClient initialised new client to communicate to Neon over http API.
// See details: https://neon.tech/docs/reference/about/
func NewClient(ctx context.Context, optFns ...func(*options)) (Client, error) {
	o := options{}
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

	return c.requestHandler(c.baseURL+"users/me", "GET", nil, nil)
}

func resolveHTTPClient(o *options) {
	if o.HTTPClient == nil {
		o.HTTPClient = &http.Client{Timeout: 5 * time.Minute}
	}
}

func resolveAPIKey(o *options) {
	if o.APIKey == "" {
		o.APIKey = os.Getenv("NEON_API_KEY")
	}
}

// WithAPIKey sets the API access key.
func WithAPIKey(apiKey string) func(*options) {
	return func(o *options) {
		o.APIKey = apiKey
	}
}

// WithHTTPClient sets custom http client.
func WithHTTPClient(client HTTPClient) func(*options) {
	return func(o *options) {
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
	var v []ProjectInfo
	if err := c.requestHandler(c.baseURL+"projects", "GET", nil, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func (c *client) ReadProject(projectID string) (ProjectInfo, error) {
	var v ProjectInfo
	if err := c.requestHandler(c.baseURL+"projects/"+projectID, "GET", nil, &v); err != nil {
		return ProjectInfo{}, err
	}
	return v, nil
}

func (c *client) UpdateProject(projectID string, settings ProjectSettingsRequestUpdate) (ProjectInfo, error) {
	var v ProjectInfo
	if err := c.requestHandler(c.baseURL+"projects/"+projectID, "PATCH", settings, &v); err != nil {
		return ProjectInfo{}, err
	}
	return v, nil
}

func (c *client) CreateProject(settings ProjectSettingsRequestCreate) (ProjectInfo, error) {
	var v ProjectInfo
	if err := c.requestHandler(c.baseURL+"projects", "POST", settings, &v); err != nil {
		return ProjectInfo{}, err
	}
	return v, nil
}

func (c *client) DeleteProject(projectID string) (ProjectStatus, error) {
	var v ProjectStatus
	if err := c.requestHandler(c.baseURL+"projects/"+projectID+"/delete", "POST", nil, &v); err != nil {
		return ProjectStatus{}, err
	}
	return v, nil
}

func (c *client) projectRunStatusUpdate(url string) (ProjectStatus, error) {
	var v ProjectStatus
	if err := c.requestHandler(url, "POST", nil, &v); err != nil {
		return ProjectStatus{}, err
	}
	return v, nil
}

func (c *client) StartProject(projectID string) (ProjectStatus, error) {
	return c.projectRunStatusUpdate(c.baseURL + "projects/" + projectID + "/start")
}

func (c *client) StopProject(projectID string) (ProjectStatus, error) {
	return c.projectRunStatusUpdate(c.baseURL + "projects/" + projectID + "/stop")
}
