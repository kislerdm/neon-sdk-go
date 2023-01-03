package sdk

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

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
	return "[HTTP Code: " + strconv.Itoa(e.HTTPCode) + "][Error Code: " + e.Code + "] " + e.Message
}

func (e Error) httpResp() *http.Response {
	o, _ := json.Marshal(e.errorResp)
	return &http.Response{
		Status:        e.Code,
		StatusCode:    e.HTTPCode,
		Body:          io.NopCloser(bytes.NewReader(o)),
		ContentLength: int64(len(o)),
	}
}

func convertErrorResponse(res *http.Response) error {
	var v errorResp
	buf, err := io.ReadAll(res.Body)
	defer func() { _ = res.Body.Close() }()
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

// Client defines the Neon SDK client.
type Client interface {
	// SuspendProjectEndpoint Suspend the specified endpoint
	SuspendProjectEndpoint(projectID string, endpointID string) (EndpointOperations, error)

	// RevokeApiKey Revokes the specified API key
	RevokeApiKey(keyID int64) (ApiKeyRevokeResponse, error)

	// GetProjectOperation Retrieves details for the specified operation
	GetProjectOperation(projectID string, operationID string) (OperationResponse, error)

	// DeleteProject Deletes the specified project
	DeleteProject(projectID string) (ProjectResponse, error)

	// GetProject Retrieves information about the specified project
	GetProject(projectID string) (ProjectResponse, error)

	// UpdateProject Updates the specified project
	UpdateProject(projectID string, cfg ProjectUpdateRequest) (ProjectOperations, error)

	// ListProjectOperations Retrieves a list of operations for the specified Neon project
	ListProjectOperations(projectID string) (ListOperations, error)

	// DeleteProjectBranch Deletes the specified branch from a project, and places
	// all endpoints into idle state, breaking the existing client connections.
	// When a successful response status is received, the endpoints are still active,
	// and the branch is not yet deleted from storage.
	// The deletion occurs after all operations finish.
	// You cannot delete a branch if it is only remaining branch in the project.
	// A project must have at least one branch.
	DeleteProjectBranch(projectID string, branchID string) (BranchOperations, error)

	// GetProjectBranch Retrieves information about the specified branch
	GetProjectBranch(projectID string, branchID string) (BranchResponse, error)

	// UpdateProjectBranch Updates the specified branch. Only changing the branch name is supported.
	UpdateProjectBranch(projectID string, branchID string, cfg BranchUpdateRequest) (BranchOperations, error)

	// CreateProjectBranch Creates a branch in the specified project
	CreateProjectBranch(projectID string, cfg *BranchCreateRequest) (CreatedBranch, error)

	// ListProjectBranches Retrieves a list of branches for the specified project
	ListProjectBranches(projectID string) (BranchesResponse, error)

	ListProjectBranchEndpoints(projectID string, branchID string) (EndpointsResponse, error)

	// CreateProjectBranchDatabase Creates a database in the specified branch
	CreateProjectBranchDatabase(projectID string, branchID string, cfg DatabaseCreateRequest) (DatabaseOperations, error)

	// ListProjectBranchDatabases Retrieves a list of databases for the specified branch
	ListProjectBranchDatabases(projectID string, branchID string) (DatabasesResponse, error)

	// ListProjectBranchRoles Retrieves a list of roles from the specified branch
	ListProjectBranchRoles(projectID string, branchID string) (RolesResponse, error)

	// CreateProjectBranchRole Creates a role in the specified branch.
	// Connections established to the active read-write endpoint will be dropped.
	// If the read-write endpoint is idle, the endpoint becomes active for a short period of time and suspended again afterward.
	CreateProjectBranchRole(projectID string, branchID string, cfg RoleCreateRequest) (RoleOperations, error)

	// StartProjectEndpoint Starts an endpoint. The endpoint is ready to use
	// after the last operation in chain finishes successfully.
	StartProjectEndpoint(projectID string, endpointID string) (EndpointOperations, error)

	// ListApiKeys Retrieves a list of API keys for the Neon account
	ListApiKeys() ([]ApiKeysListResponseItem, error)

	// CreateApiKey Creates an API key and returns its value
	CreateApiKey(cfg ApiKeyCreateRequest) (ApiKeyCreateResponse, error)

	// CreateProject Creates a Neon project
	CreateProject(cfg ProjectCreateRequest) (CreatedProject, error)

	// ListProjects Retrieves a list of projects for the Neon account
	ListProjects() (ProjectsResponse, error)

	// DeleteProjectBranchDatabase Deletes the specified database from the branch
	DeleteProjectBranchDatabase(projectID string, branchID string, databaseName string) (DatabaseOperations, error)

	// GetProjectBranchDatabase Retrieves information about the specified database
	GetProjectBranchDatabase(projectID string, branchID string, databaseName string) (DatabaseResponse, error)

	// UpdateProjectBranchDatabase Updates the specified database in the branch
	UpdateProjectBranchDatabase(projectID string, branchID string, databaseName string, cfg DatabaseUpdateRequest) (DatabaseOperations, error)

	// ResetProjectBranchRolePassword Resets the password for the specified role.
	// Returns a new password and operations. The new password is ready to use when the last operation finishes.
	// The old password remains valid until last operation finishes.
	// Connections to the read-write endpoint are dropped. If idle,
	// the read-write endpoint becomes active for a short period of time.
	ResetProjectBranchRolePassword(projectID string, branchID string, roleName string) (RoleOperations, error)

	// DeleteProjectEndpoint Delete the specified endpoint. Deleting an endpoint drops existing network connections to the endpoint.
	// The deletion is completed when last operation in the chain finishes successfully.
	DeleteProjectEndpoint(projectID string, endpointID string) (EndpointOperations, error)

	// GetProjectEndpoint Retrieves information about the specified endpoint
	GetProjectEndpoint(projectID string, endpointID string) (EndpointResponse, error)

	// UpdateProjectEndpoint Updates the specified endpoint. Currently, only changing the associated branch is supported.
	// The branch the you specify cannot have an existing endpoint.
	// If the returned list of operations is not empty, the endpoint is not ready to use.
	// The client must wait for the last operation to finish before using the endpoint.
	// If the endpoint was idle before the update, the endpoint becomes active for a short period of time,
	// and the control plane suspends it again after the update.
	UpdateProjectEndpoint(projectID string, endpointID string, cfg EndpointUpdateRequest) (EndpointOperations, error)

	// DeleteProjectBranchRole Deletes the specified role from the branch
	DeleteProjectBranchRole(projectID string, branchID string, roleName string) (RoleOperations, error)

	// GetProjectBranchRole Retrieves details about the specified role
	GetProjectBranchRole(projectID string, branchID string, roleName string) (RoleResponse, error)

	// ListProjectEndpoints Retrieves a list of endpoints for the specified project
	ListProjectEndpoints(projectID string) (EndpointsResponse, error)

	// CreateProjectEndpoint Creates an endpoint for the specified branch.
	// There is a maximum of one read-write endpoint per branch.
	// If the branch already has an endpoint, the operation will fail.
	// Currently, only read-write endpoints are supported.
	CreateProjectEndpoint(projectID string, cfg EndpointCreateRequest) (EndpointOperations, error)
}

// HTTPClient client to handle http requests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type options struct {
	// key API access key.
	key string

	// httpClient Client to communicate with the API over http.
	httpClient HTTPClient
}

type client struct {
	options options

	baseURL string
}

const (
	baseURL        = "https://console.neon.tech/api/v2"
	defaultTimeout = 2 * time.Minute
)

// NewClient initialised the Client to communicate to the Neon Platform.
func NewClient(optFns ...func(*options)) (Client, error) {
	o := options{
		key:        "",
		httpClient: nil,
	}

	for _, fn := range optFns {
		fn(&o)
	}

	resolveHTTPClient(&o)
	if err := resolveApiKey(&o); err != nil {
		return nil, err
	}

	return &client{
		baseURL: baseURL,
		options: o,
	}, nil
}

func resolveApiKey(o *options) error {
	if o.key == "" {
		o.key = os.Getenv("NEON_API_KEY")
	}

	if _, ok := (o.httpClient).(mockHTTPClient); !ok && o.key == "" {
		return errors.New(
			"authorization key must be provided: https://neon.tech/docs/reference/api-reference/#authentication",
		)
	}

	return nil
}

func resolveHTTPClient(o *options) {
	if o.httpClient == nil {
		o.httpClient = &http.Client{Timeout: defaultTimeout}
	}
}

// WithHTTPClient sets custom http Client.
func WithHTTPClient(client HTTPClient) func(*options) {
	return func(o *options) {
		o.httpClient = client
	}
}

// WithAPIKey sets the Neon API key.
func WithAPIKey(key string) func(*options) {
	return func(o *options) {
		o.key = key
	}
}

func setHeaders(req *http.Request, token string) {
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}
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
	setHeaders(req, c.options.key)

	res, err := c.options.httpClient.Do(req)
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
		defer func() { _ = res.Body.Close() }()
		if err != nil {
			return err
		}
		return json.Unmarshal(buf, responsePayload)
	}

	return nil
}

func (c *client) SuspendProjectEndpoint(projectID string, endpointID string) (EndpointOperations, error) {
	var v EndpointOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints/"+endpointID+"/suspend", "POST", nil, &v); err != nil {
		return EndpointOperations{}, err
	}
	return v, nil
}

func (c *client) RevokeApiKey(keyID int64) (ApiKeyRevokeResponse, error) {
	var v ApiKeyRevokeResponse
	if err := c.requestHandler(c.baseURL+"/api_keys/"+strconv.FormatInt(keyID, 10), "DELETE", nil, &v); err != nil {
		return ApiKeyRevokeResponse{}, err
	}
	return v, nil
}

func (c *client) GetProjectOperation(projectID string, operationID string) (OperationResponse, error) {
	var v OperationResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/operations/"+operationID, "GET", nil, &v); err != nil {
		return OperationResponse{}, err
	}
	return v, nil
}

func (c *client) DeleteProject(projectID string) (ProjectResponse, error) {
	var v ProjectResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID, "DELETE", nil, &v); err != nil {
		return ProjectResponse{}, err
	}
	return v, nil
}

func (c *client) GetProject(projectID string) (ProjectResponse, error) {
	var v ProjectResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID, "GET", nil, &v); err != nil {
		return ProjectResponse{}, err
	}
	return v, nil
}

func (c *client) UpdateProject(projectID string, cfg ProjectUpdateRequest) (ProjectOperations, error) {
	var v ProjectOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID, "PATCH", cfg, &v); err != nil {
		return ProjectOperations{}, err
	}
	return v, nil
}

func (c *client) ListProjectOperations(projectID string) (ListOperations, error) {
	var v ListOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/operations", "GET", nil, &v); err != nil {
		return ListOperations{}, err
	}
	return v, nil
}

func (c *client) DeleteProjectBranch(projectID string, branchID string) (BranchOperations, error) {
	var v BranchOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID, "DELETE", nil, &v); err != nil {
		return BranchOperations{}, err
	}
	return v, nil
}

func (c *client) GetProjectBranch(projectID string, branchID string) (BranchResponse, error) {
	var v BranchResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID, "GET", nil, &v); err != nil {
		return BranchResponse{}, err
	}
	return v, nil
}

func (c *client) UpdateProjectBranch(projectID string, branchID string, cfg BranchUpdateRequest) (BranchOperations, error) {
	var v BranchOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID, "PATCH", cfg, &v); err != nil {
		return BranchOperations{}, err
	}
	return v, nil
}

func (c *client) CreateProjectBranch(projectID string, cfg *BranchCreateRequest) (CreatedBranch, error) {
	var v CreatedBranch
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches", "POST", cfg, &v); err != nil {
		return CreatedBranch{}, err
	}
	return v, nil
}

func (c *client) ListProjectBranches(projectID string) (BranchesResponse, error) {
	var v BranchesResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches", "GET", nil, &v); err != nil {
		return BranchesResponse{}, err
	}
	return v, nil
}

func (c *client) ListProjectBranchEndpoints(projectID string, branchID string) (EndpointsResponse, error) {
	var v EndpointsResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/endpoints", "GET", nil, &v); err != nil {
		return EndpointsResponse{}, err
	}
	return v, nil
}

func (c *client) CreateProjectBranchDatabase(projectID string, branchID string, cfg DatabaseCreateRequest) (DatabaseOperations, error) {
	var v DatabaseOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/databases", "POST", cfg, &v); err != nil {
		return DatabaseOperations{}, err
	}
	return v, nil
}

func (c *client) ListProjectBranchDatabases(projectID string, branchID string) (DatabasesResponse, error) {
	var v DatabasesResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/databases", "GET", nil, &v); err != nil {
		return DatabasesResponse{}, err
	}
	return v, nil
}

func (c *client) ListProjectBranchRoles(projectID string, branchID string) (RolesResponse, error) {
	var v RolesResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/roles", "GET", nil, &v); err != nil {
		return RolesResponse{}, err
	}
	return v, nil
}

func (c *client) CreateProjectBranchRole(projectID string, branchID string, cfg RoleCreateRequest) (RoleOperations, error) {
	var v RoleOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/roles", "POST", cfg, &v); err != nil {
		return RoleOperations{}, err
	}
	return v, nil
}

func (c *client) StartProjectEndpoint(projectID string, endpointID string) (EndpointOperations, error) {
	var v EndpointOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints/"+endpointID+"/start", "POST", nil, &v); err != nil {
		return EndpointOperations{}, err
	}
	return v, nil
}

func (c *client) ListApiKeys() ([]ApiKeysListResponseItem, error) {
	var v []ApiKeysListResponseItem
	if err := c.requestHandler(c.baseURL+"/api_keys", "GET", nil, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func (c *client) CreateApiKey(cfg ApiKeyCreateRequest) (ApiKeyCreateResponse, error) {
	var v ApiKeyCreateResponse
	if err := c.requestHandler(c.baseURL+"/api_keys", "POST", cfg, &v); err != nil {
		return ApiKeyCreateResponse{}, err
	}
	return v, nil
}

func (c *client) CreateProject(cfg ProjectCreateRequest) (CreatedProject, error) {
	var v CreatedProject
	if err := c.requestHandler(c.baseURL+"/projects", "POST", cfg, &v); err != nil {
		return CreatedProject{}, err
	}
	return v, nil
}

func (c *client) ListProjects() (ProjectsResponse, error) {
	var v ProjectsResponse
	if err := c.requestHandler(c.baseURL+"/projects", "GET", nil, &v); err != nil {
		return ProjectsResponse{}, err
	}
	return v, nil
}

func (c *client) DeleteProjectBranchDatabase(projectID string, branchID string, databaseName string) (DatabaseOperations, error) {
	var v DatabaseOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/databases/"+databaseName, "DELETE", nil, &v); err != nil {
		return DatabaseOperations{}, err
	}
	return v, nil
}

func (c *client) GetProjectBranchDatabase(projectID string, branchID string, databaseName string) (DatabaseResponse, error) {
	var v DatabaseResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/databases/"+databaseName, "GET", nil, &v); err != nil {
		return DatabaseResponse{}, err
	}
	return v, nil
}

func (c *client) UpdateProjectBranchDatabase(projectID string, branchID string, databaseName string, cfg DatabaseUpdateRequest) (DatabaseOperations, error) {
	var v DatabaseOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/databases/"+databaseName, "PATCH", cfg, &v); err != nil {
		return DatabaseOperations{}, err
	}
	return v, nil
}

func (c *client) ResetProjectBranchRolePassword(projectID string, branchID string, roleName string) (RoleOperations, error) {
	var v RoleOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/roles/"+roleName+"/reset_password", "POST", nil, &v); err != nil {
		return RoleOperations{}, err
	}
	return v, nil
}

func (c *client) DeleteProjectEndpoint(projectID string, endpointID string) (EndpointOperations, error) {
	var v EndpointOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints/"+endpointID, "DELETE", nil, &v); err != nil {
		return EndpointOperations{}, err
	}
	return v, nil
}

func (c *client) GetProjectEndpoint(projectID string, endpointID string) (EndpointResponse, error) {
	var v EndpointResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints/"+endpointID, "GET", nil, &v); err != nil {
		return EndpointResponse{}, err
	}
	return v, nil
}

func (c *client) UpdateProjectEndpoint(projectID string, endpointID string, cfg EndpointUpdateRequest) (EndpointOperations, error) {
	var v EndpointOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints/"+endpointID, "PATCH", cfg, &v); err != nil {
		return EndpointOperations{}, err
	}
	return v, nil
}

func (c *client) DeleteProjectBranchRole(projectID string, branchID string, roleName string) (RoleOperations, error) {
	var v RoleOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/roles/"+roleName, "DELETE", nil, &v); err != nil {
		return RoleOperations{}, err
	}
	return v, nil
}

func (c *client) GetProjectBranchRole(projectID string, branchID string, roleName string) (RoleResponse, error) {
	var v RoleResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/roles/"+roleName, "GET", nil, &v); err != nil {
		return RoleResponse{}, err
	}
	return v, nil
}

func (c *client) ListProjectEndpoints(projectID string) (EndpointsResponse, error) {
	var v EndpointsResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints", "GET", nil, &v); err != nil {
		return EndpointsResponse{}, err
	}
	return v, nil
}

func (c *client) CreateProjectEndpoint(projectID string, cfg EndpointCreateRequest) (EndpointOperations, error) {
	var v EndpointOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints", "POST", cfg, &v); err != nil {
		return EndpointOperations{}, err
	}
	return v, nil
}

type ApiKeyCreateRequest struct {
	KeyName string `json:"key_name"`
}

type ApiKeyCreateResponse struct {
	ID  int64  `json:"id"`
	Key string `json:"key"`
}

type ApiKeyRevokeResponse struct {
	LastUsedAt       time.Time `json:"last_used_at,omitempty"`
	LastUsedFromAddr string    `json:"last_used_from_addr"`
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	Revoked          bool      `json:"revoked"`
}

type ApiKeysListResponseItem struct {
	CreatedAt        time.Time `json:"created_at"`
	ID               int64     `json:"id"`
	LastUsedAt       time.Time `json:"last_used_at,omitempty"`
	LastUsedFromAddr string    `json:"last_used_from_addr"`
	Name             string    `json:"name"`
}

type Branch struct {
	ParentID        string    `json:"parent_id,omitempty"`
	ParentTimestamp time.Time `json:"parent_timestamp,omitempty"`
	Name            string    `json:"name"`
	ParentLsn       string    `json:"parent_lsn,omitempty"`
	// LogicalSize Branch logical size in MB
	LogicalSize int64 `json:"logical_size,omitempty"`
	// PhysicalSize Branch physical size in MB
	PhysicalSize int64       `json:"physical_size,omitempty"`
	CurrentState BranchState `json:"current_state"`
	PendingState BranchState `json:"pending_state,omitempty"`
	// LogicalSizeLimit Branch logical size limit in MB
	LogicalSizeLimit int64     `json:"logical_size_limit,omitempty"`
	UpdatedAt        time.Time `json:"updated_at"`
	ID               string    `json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	ProjectID        string    `json:"project_id"`
}

type BranchCreateRequest struct {
	Endpoints []BranchCreateRequestEndpointOptions `json:"endpoints,omitempty"`
	Branch    BranchCreateRequestBranch            `json:"branch,omitempty"`
}

type BranchCreateRequestBranch struct {
	ParentLsn       string    `json:"parent_lsn,omitempty"`
	ParentTimestamp time.Time `json:"parent_timestamp,omitempty"`
	ParentID        string    `json:"parent_id,omitempty"`
	Name            string    `json:"name,omitempty"`
}

type BranchCreateRequestEndpointOptions struct {
	AutoscalingLimitMaxCu int32        `json:"autoscaling_limit_max_cu,omitempty"`
	Type                  EndpointType `json:"type"`
	AutoscalingLimitMinCu int32        `json:"autoscaling_limit_min_cu,omitempty"`
}

type BranchOperations struct {
	BranchResponse
	OperationsResponse
}

type BranchResponse struct {
	Branch Branch `json:"branch"`
}

type BranchState string

type BranchUpdateRequest struct {
	Branch BranchUpdateRequestBranch `json:"branch"`
}

type BranchUpdateRequestBranch struct {
	Name string `json:"name,omitempty"`
}

type BranchesResponse struct {
	Branches []Branch `json:"branches"`
}

type ConnectionURI struct {
	// ConnectionURI Connection URI is same as specified in https://www.postgresql.org/docs/current/libpq-connect.html#id-1.7.3.8.3.6
	// It is a ready to use string for psql or for DATABASE_URL environment variable.
	ConnectionURI string `json:"connection_uri"`
}

type ConnectionURIsResponse struct {
	ConnectionUris []ConnectionURI `json:"connection_uris"`
}

type CreatedBranch struct {
	EndpointsResponse
	OperationsResponse
	BranchResponse
}

type CreatedProject struct {
	ConnectionURIsResponse
	RolesResponse
	DatabasesResponse
	OperationsResponse
	BranchResponse
	EndpointsResponse
	ProjectResponse
}

type Database struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ID        int64     `json:"id"`
	BranchID  string    `json:"branch_id"`
	Name      string    `json:"name"`
	OwnerName string    `json:"owner_name"`
}

type DatabaseCreateRequest struct {
	Database DatabaseCreateRequestDatabase `json:"database"`
}

type DatabaseCreateRequestDatabase struct {
	Name      string `json:"name"`
	OwnerName string `json:"owner_name"`
}

type DatabaseOperations struct {
	DatabaseResponse
	OperationsResponse
}

type DatabaseResponse struct {
	Database Database `json:"database"`
}

type DatabaseUpdateRequest struct {
	Database DatabaseUpdateRequestDatabase `json:"database"`
}

type DatabaseUpdateRequestDatabase struct {
	Name      string `json:"name,omitempty"`
	OwnerName string `json:"owner_name,omitempty"`
}

type DatabasesResponse struct {
	Databases []Database `json:"databases"`
}

type Endpoint struct {
	// PasswordlessAccess Allow or restrict passwordless access to this endpoint.
	PasswordlessAccess    bool               `json:"passwordless_access"`
	CreatedAt             time.Time          `json:"created_at"`
	PoolerMode            EndpointPoolerMode `json:"pooler_mode"`
	AutoscalingLimitMinCu int32              `json:"autoscaling_limit_min_cu"`
	// ProxyHost DEPRECATED. You should use "host" property instead.
	ProxyHost    string        `json:"proxy_host"`
	PendingState EndpointState `json:"pending_state,omitempty"`
	UpdatedAt    time.Time     `json:"updated_at"`
	// PoolerEnabled Enable or disable connections pooling for this endpoint.
	PoolerEnabled bool          `json:"pooler_enabled"`
	RegionID      string        `json:"region_id"`
	ID            string        `json:"id"`
	CurrentState  EndpointState `json:"current_state"`
	// Host Hostname to connect to.
	Host      string              `json:"host"`
	Settings  ProjectSettingsData `json:"settings"`
	BranchID  string              `json:"branch_id"`
	ProjectID string              `json:"project_id"`
	// Disabled Restrict any connections to this endpoint.
	Disabled bool         `json:"disabled"`
	Type     EndpointType `json:"type"`
	// LastActive Timestamp of the last detected activity of the endpoint.
	LastActive            time.Time `json:"last_active,omitempty"`
	AutoscalingLimitMaxCu int32     `json:"autoscaling_limit_max_cu"`
}

type EndpointCreateRequest struct {
	Endpoint EndpointCreateRequestEndpoint `json:"endpoint"`
}

type EndpointCreateRequestEndpoint struct {
	AutoscalingLimitMinCu int32                `json:"autoscaling_limit_min_cu,omitempty"`
	Type                  EndpointType         `json:"type"`
	Settings              EndpointSettingsData `json:"settings,omitempty"`
	// Disabled Restrict any connections to this endpoint.
	Disabled   bool               `json:"disabled,omitempty"`
	BranchID   string             `json:"branch_id"`
	PoolerMode EndpointPoolerMode `json:"pooler_mode,omitempty"`
	// PasswordlessAccess NOT IMPLEMENTED YET
	PasswordlessAccess bool `json:"passwordless_access,omitempty"`
	// RegionID Only project region id is allowed for now
	RegionID              string `json:"region_id,omitempty"`
	PoolerEnabled         bool   `json:"pooler_enabled,omitempty"`
	AutoscalingLimitMaxCu int32  `json:"autoscaling_limit_max_cu,omitempty"`
}

type EndpointOperations struct {
	EndpointResponse
	OperationsResponse
}

type EndpointPoolerMode string

type EndpointResponse struct {
	Endpoint Endpoint `json:"endpoint"`
}

// EndpointSettingsData Endpoint settings is a collection of settings for an Endpoint
type EndpointSettingsData struct {
	PgSettings PgSettingsData `json:"pg_settings,omitempty"`
}

type EndpointState string

// EndpointType Endpoint type. Either "read_write" for read-write primary or "read_only" for read-only secondary.
// "read_only" endpoints are NOT yet implemented.
type EndpointType string

type EndpointUpdateRequest struct {
	Endpoint EndpointUpdateRequestEndpoint `json:"endpoint"`
}

type EndpointUpdateRequestEndpoint struct {
	PoolerMode EndpointPoolerMode `json:"pooler_mode,omitempty"`
	// Disabled Restrict any connections to this endpoint.
	Disabled bool `json:"disabled,omitempty"`
	// PasswordlessAccess NOT IMPLEMENTED YET
	PasswordlessAccess bool `json:"passwordless_access,omitempty"`
	// BranchID Destination branch identifier. The destination branch must not have RW endpoint.
	BranchID              string               `json:"branch_id,omitempty"`
	AutoscalingLimitMinCu int32                `json:"autoscaling_limit_min_cu,omitempty"`
	AutoscalingLimitMaxCu int32                `json:"autoscaling_limit_max_cu,omitempty"`
	Settings              EndpointSettingsData `json:"settings,omitempty"`
	PoolerEnabled         bool                 `json:"pooler_enabled,omitempty"`
}

type EndpointsResponse struct {
	Endpoints []Endpoint `json:"endpoints"`
}

type ListOperations struct {
	OperationsResponse
	PaginationResponse
}

type Operation struct {
	Error         string          `json:"error,omitempty"`
	EndpointID    string          `json:"endpoint_id,omitempty"`
	FailuresCount int32           `json:"failures_count"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	ID            string          `json:"id"`
	Status        OperationStatus `json:"status"`
	BranchID      string          `json:"branch_id,omitempty"`
	RetryAt       time.Time       `json:"retry_at,omitempty"`
	ProjectID     string          `json:"project_id"`
	Action        OperationAction `json:"action"`
}

type OperationAction string

type OperationResponse struct {
	Operation Operation `json:"operation"`
}

type OperationStatus string

type OperationsResponse struct {
	Operations []Operation `json:"operations"`
}

// Pagination Cursor based pagination is used. The user must pass the cursor as is to the backend.
// For more information about cursor based pagination, see
// https://learn.microsoft.com/en-us/ef/core/querying/pagination#keyset-pagination
type Pagination struct {
	Cursor string `json:"cursor"`
}

type PaginationResponse struct {
	Pagination Pagination `json:"pagination,omitempty"`
}

// PgSettingsData is a raw representation of Postgres settings
type PgSettingsData struct{}

// PgVersion Major version of the Postgres
type PgVersion int

type Project struct {
	CreatedAt  time.Time `json:"created_at"`
	ID         string    `json:"id"`
	PlatformID string    `json:"platform_id"`
	Name       string    `json:"name"`
	PgVersion  PgVersion `json:"pg_version"`
	// Locked Currently, a project may not have more than one running operations chain.
	// If there are any running operations, 'locked' will be set to 'true'.
	// This attributed is considered to be temporary, and could be gone soon.
	Locked      bool   `json:"locked"`
	Provisioner string `json:"provisioner,omitempty"`
	ProxyHost   string `json:"proxy_host"`
	// MaintenanceStartsAt If set, means project will be in maintenance since that time.
	MaintenanceStartsAt     time.Time           `json:"maintenance_starts_at,omitempty"`
	RegionID                string              `json:"region_id"`
	UpdatedAt               time.Time           `json:"updated_at"`
	DefaultEndpointSettings ProjectSettingsData `json:"default_endpoint_settings,omitempty"`
}

type ProjectCreateRequest struct {
	Project ProjectCreateRequestProject `json:"project"`
}

type ProjectCreateRequestProject struct {
	Provisioner             string         `json:"provisioner,omitempty"`
	RegionID                string         `json:"region_id,omitempty"`
	DefaultEndpointSettings PgSettingsData `json:"default_endpoint_settings,omitempty"`
	PgVersion               PgVersion      `json:"pg_version,omitempty"`
	Name                    string         `json:"name,omitempty"`
	AutoscalingLimitMinCu   int32          `json:"autoscaling_limit_min_cu,omitempty"`
	AutoscalingLimitMaxCu   int32          `json:"autoscaling_limit_max_cu,omitempty"`
}

type ProjectOperations struct {
	ProjectResponse
	OperationsResponse
}

type ProjectResponse struct {
	Project Project `json:"project"`
}

// ProjectSettingsData is a collection of settings for a Project
type ProjectSettingsData struct {
	PgSettings PgSettingsData `json:"pg_settings,omitempty"`
}

type ProjectUpdateRequest struct {
	Project ProjectUpdateRequestProject `json:"project"`
}

type ProjectUpdateRequestProject struct {
	AutoscalingLimitMaxCu   int32          `json:"autoscaling_limit_max_cu,omitempty"`
	Name                    string         `json:"name,omitempty"`
	DefaultEndpointSettings PgSettingsData `json:"default_endpoint_settings,omitempty"`
	AutoscalingLimitMinCu   int32          `json:"autoscaling_limit_min_cu,omitempty"`
}

type ProjectsResponse struct {
	Projects []Project `json:"projects"`
}

type Role struct {
	Protected bool      `json:"protected,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	BranchID  string    `json:"branch_id"`
	Name      string    `json:"name"`
	Password  string    `json:"password,omitempty"`
}

type RoleCreateRequest struct {
	Role RoleCreateRequestRole `json:"role"`
}

type RoleCreateRequestRole struct {
	Name string `json:"name"`
}

type RoleOperations struct {
	RoleResponse
	OperationsResponse
}

type RoleResponse struct {
	Role Role `json:"role"`
}

type RolesResponse struct {
	Roles []Role `json:"roles"`
}
