package sdk

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// NewClient initialised the Client to communicate to the Neon Platform.
func NewClient(cfg Config) (*Client, error) {
	if _, ok := (cfg.HTTPClient).(mockHTTPClient); !ok && cfg.Key == "" {
		return nil, errors.New(
			"authorization key must be provided: https://neon.tech/docs/reference/api-reference/#authentication",
		)
	}

	c := &Client{
		baseURL: baseURL,
		cfg:     cfg,
	}

	if c.cfg.HTTPClient == nil {
		c.cfg.HTTPClient = &http.Client{Timeout: defaultTimeout}
	}

	return c, nil
}

// Config defines the client's configuration.
type Config struct {
	// Key defines the access API key.
	Key string

	// HTTPClient HTTP client to communicate with the API.
	HTTPClient HTTPClient
}

const (
	baseURL        = "https://console.neon.tech/api/v2"
	defaultTimeout = 2 * time.Minute
)

// Client defines the Neon SDK client.
type Client struct {
	cfg Config

	baseURL string
}

// HTTPClient client to handle http requests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func setHeaders(req *http.Request, token string) {
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}
}

func (c Client) requestHandler(url string, t string, reqPayload interface{}, responsePayload interface{}) error {
	var body io.Reader
	var err error

	if reqPayload != nil {
		if v := reflect.ValueOf(reqPayload); v.Kind() == reflect.Struct || !v.IsNil() {
			b, err := json.Marshal(reqPayload)
			if err != nil {
				return err
			}
			body = bytes.NewReader(b)
		}
	}

	req, _ := http.NewRequest(t, url, body)
	setHeaders(req, c.cfg.Key)

	res, err := c.cfg.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode > 299 {
		return convertErrorResponse(res)
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

// ListApiKeys Retrieves the API keys for your Neon account.
// The response does not include API key tokens. A token is only provided when creating an API key.
// API keys can also be managed in the Neon Console.
// For more information, see [Manage API keys](https://neon.tech/docs/manage/api-keys/).
func (c Client) ListApiKeys() ([]ApiKeysListResponseItem, error) {
	var v []ApiKeysListResponseItem
	if err := c.requestHandler(c.baseURL+"/api_keys", "GET", nil, &v); err != nil {
		return nil, err
	}
	return v, nil
}

// CreateApiKey Creates an API key.
// The `key_name` is a user-specified name for the key.
// This method returns an `id` and `key`. The `key` is a randomly generated, 64-bit token required to access the Neon API.
// API keys can also be managed in the Neon Console.
// See [Manage API keys](https://neon.tech/docs/manage/api-keys/).
func (c Client) CreateApiKey(cfg ApiKeyCreateRequest) (ApiKeyCreateResponse, error) {
	var v ApiKeyCreateResponse
	if err := c.requestHandler(c.baseURL+"/api_keys", "POST", cfg, &v); err != nil {
		return ApiKeyCreateResponse{}, err
	}
	return v, nil
}

// RevokeApiKey Revokes the specified API key.
// An API key that is no longer needed can be revoked.
// This action cannot be reversed.
// You can obtain `key_id` values by listing the API keys for your Neon account.
// API keys can also be managed in the Neon Console.
// See [Manage API keys](https://neon.tech/docs/manage/api-keys/).
func (c Client) RevokeApiKey(keyID int64) (ApiKeyRevokeResponse, error) {
	var v ApiKeyRevokeResponse
	if err := c.requestHandler(c.baseURL+"/api_keys/"+strconv.FormatInt(keyID, 10), "DELETE", nil, &v); err != nil {
		return ApiKeyRevokeResponse{}, err
	}
	return v, nil
}

// GetProjectOperation Retrieves details for the specified operation.
// An operation is an action performed on a Neon project resource.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain a `operation_id` by listing operations for the project.
func (c Client) GetProjectOperation(projectID string, operationID string) (OperationResponse, error) {
	var v OperationResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/operations/"+operationID, "GET", nil, &v); err != nil {
		return OperationResponse{}, err
	}
	return v, nil
}

// ListProjects Retrieves a list of projects for the Neon account.
// A project is the top-level object in the Neon object hierarchy.
// For more information, see [Manage projects](https://neon.tech/docs/manage/projects/).
func (c Client) ListProjects(cursor *string, limit *int) (ListProjectsRespObj, error) {
	var queryElements []string
	if cursor != nil {
		queryElements = append(queryElements, "cursor="+*cursor)
	}
	if limit != nil {
		queryElements = append(queryElements, "limit="+strconv.FormatInt(int64(*limit), 10))
	}
	query := "?" + strings.Join(queryElements, "&")
	var v ListProjectsRespObj
	if err := c.requestHandler(c.baseURL+"/projects"+query, "GET", nil, &v); err != nil {
		return ListProjectsRespObj{}, err
	}
	return v, nil
}

// CreateProject Creates a Neon project.
// A project is the top-level object in the Neon object hierarchy.
// Plan limits define how many projects you can create.
// Neon's Free plan permits one project per Neon account.
// For more information, see [Manage projects](https://neon.tech/docs/manage/projects/).
// You can specify a region and PostgreSQL version in the request body.
// Neon currently supports PostgreSQL 14 and 15.
// For supported regions and `region_id` values, see [Regions](https://neon.tech/docs/introduction/regions/).
func (c Client) CreateProject(cfg ProjectCreateRequest) (CreatedProject, error) {
	var v CreatedProject
	if err := c.requestHandler(c.baseURL+"/projects", "POST", cfg, &v); err != nil {
		return CreatedProject{}, err
	}
	return v, nil
}

// ListSharedProjects Retrieves a list of shared projects for the Neon account.
// A project is the top-level object in the Neon object hierarchy.
// For more information, see [Manage projects](https://neon.tech/docs/manage/projects/).
func (c Client) ListSharedProjects(cursor *string, limit *int) (ListSharedProjectsRespObj, error) {
	var queryElements []string
	if cursor != nil {
		queryElements = append(queryElements, "cursor="+*cursor)
	}
	if limit != nil {
		queryElements = append(queryElements, "limit="+strconv.FormatInt(int64(*limit), 10))
	}
	query := "?" + strings.Join(queryElements, "&")
	var v ListSharedProjectsRespObj
	if err := c.requestHandler(c.baseURL+"/projects/shared"+query, "GET", nil, &v); err != nil {
		return ListSharedProjectsRespObj{}, err
	}
	return v, nil
}

// GetProject Retrieves information about the specified project.
// A project is the top-level object in the Neon object hierarchy.
// You can obtain a `project_id` by listing the projects for your Neon account.
func (c Client) GetProject(projectID string) (ProjectResponse, error) {
	var v ProjectResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID, "GET", nil, &v); err != nil {
		return ProjectResponse{}, err
	}
	return v, nil
}

// UpdateProject Updates the specified project.
// You can obtain a `project_id` by listing the projects for your Neon account.
// Neon permits updating the project name only.
func (c Client) UpdateProject(projectID string, cfg ProjectUpdateRequest) (UpdateProjectRespObj, error) {
	var v UpdateProjectRespObj
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID, "PATCH", cfg, &v); err != nil {
		return UpdateProjectRespObj{}, err
	}
	return v, nil
}

// DeleteProject Deletes the specified project.
// You can obtain a `project_id` by listing the projects for your Neon account.
// Deleting a project is a permanent action.
// Deleting a project also deletes endpoints, branches, databases, and users that belong to the project.
func (c Client) DeleteProject(projectID string) (ProjectResponse, error) {
	var v ProjectResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID, "DELETE", nil, &v); err != nil {
		return ProjectResponse{}, err
	}
	return v, nil
}

// ListProjectOperations Retrieves a list of operations for the specified Neon project.
// You can obtain a `project_id` by listing the projects for your Neon account.
// The number of operations returned can be large.
// To paginate the response, issue an initial request with a `limit` value.
// Then, add the `cursor` value that was returned in the response to the next request.
func (c Client) ListProjectOperations(projectID string, cursor *string, limit *int) (ListOperations, error) {
	var queryElements []string
	if cursor != nil {
		queryElements = append(queryElements, "cursor="+*cursor)
	}
	if limit != nil {
		queryElements = append(queryElements, "limit="+strconv.FormatInt(int64(*limit), 10))
	}
	query := "?" + strings.Join(queryElements, "&")
	var v ListOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/operations"+query, "GET", nil, &v); err != nil {
		return ListOperations{}, err
	}
	return v, nil
}

// ListProjectPermissions Return project's permissions
func (c Client) ListProjectPermissions(projectID string) (ProjectPermissions, error) {
	var v ProjectPermissions
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/permissions", "GET", nil, &v); err != nil {
		return ProjectPermissions{}, err
	}
	return v, nil
}

// GrantPermissionToProject Grant project permission to the user
func (c Client) GrantPermissionToProject(projectID string, cfg GrantPermissionToProjectRequest) (ProjectPermission, error) {
	var v ProjectPermission
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/permissions", "POST", cfg, &v); err != nil {
		return ProjectPermission{}, err
	}
	return v, nil
}

// RevokePermissionFromProject Revoke permission from the user
func (c Client) RevokePermissionFromProject(projectID string, permissionID string) (ProjectPermission, error) {
	var v ProjectPermission
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/permissions/"+permissionID, "DELETE", nil, &v); err != nil {
		return ProjectPermission{}, err
	}
	return v, nil
}

// ListProjectBranches Retrieves a list of branches for the specified project.
// You can obtain a `project_id` by listing the projects for your Neon account.
// Each Neon project has a root branch named `main`.
// A `branch_id` value has a `br-` prefix.
// A project may contain child branches that were branched from `main` or from another branch.
// A parent branch is identified by the `parent_id` value, which is the `id` of the parent branch.
// For related information, see [Manage branches](https://neon.tech/docs/manage/branches/).
func (c Client) ListProjectBranches(projectID string) (BranchesResponse, error) {
	var v BranchesResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches", "GET", nil, &v); err != nil {
		return BranchesResponse{}, err
	}
	return v, nil
}

// CreateProjectBranch Creates a branch in the specified project.
// You can obtain a `project_id` by listing the projects for your Neon account.
// This method does not require a request body, but you can specify one to create an endpoint for the branch or to select a non-default parent branch.
// The default behavior is to create a branch from the project's root branch (`main`) with no endpoint, and the branch name is auto-generated.
// For related information, see [Manage branches](https://neon.tech/docs/manage/branches/).
func (c Client) CreateProjectBranch(projectID string, cfg *BranchCreateRequest) (CreatedBranch, error) {
	var v CreatedBranch
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches", "POST", cfg, &v); err != nil {
		return CreatedBranch{}, err
	}
	return v, nil
}

// GetProjectBranch Retrieves information about the specified branch.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain a `branch_id` by listing the project's branches.
// A `branch_id` value has a `br-` prefix.
// Each Neon project has a root branch named `main`.
// A project may contain child branches that were branched from `main` or from another branch.
// A parent branch is identified by a `parent_id` value, which is the `id` of the parent branch.
// For related information, see [Manage branches](https://neon.tech/docs/manage/branches/).
func (c Client) GetProjectBranch(projectID string, branchID string) (BranchResponse, error) {
	var v BranchResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID, "GET", nil, &v); err != nil {
		return BranchResponse{}, err
	}
	return v, nil
}

// UpdateProjectBranch Updates the specified branch. Only changing the branch name is supported.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain the `branch_id` by listing the project's branches.
// For more information, see [Manage branches](https://neon.tech/docs/manage/branches/).
func (c Client) UpdateProjectBranch(projectID string, branchID string, cfg BranchUpdateRequest) (BranchOperations, error) {
	var v BranchOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID, "PATCH", cfg, &v); err != nil {
		return BranchOperations{}, err
	}
	return v, nil
}

// DeleteProjectBranch Deletes the specified branch from a project, and places
// all endpoints into an idle state, breaking existing client connections.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain a `branch_id` by listing the project's branches.
// For related information, see [Manage branches](https://neon.tech/docs/manage/branches/).
// When a successful response status is received, the endpoints are still active,
// and the branch is not yet deleted from storage.
// The deletion occurs after all operations finish.
// You cannot delete a branch if it is the only remaining branch in the project.
// A project must have at least one branch.
func (c Client) DeleteProjectBranch(projectID string, branchID string) (BranchOperations, error) {
	var v BranchOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID, "DELETE", nil, &v); err != nil {
		return BranchOperations{}, err
	}
	return v, nil
}

// SetPrimaryProjectBranch The primary mark is automatically removed from the previous primary branch.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain the `branch_id` by listing the project's branches.
// For more information, see [Manage branches](https://neon.tech/docs/manage/branches/).
func (c Client) SetPrimaryProjectBranch(projectID string, branchID string) (BranchOperations, error) {
	var v BranchOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/set_as_primary", "POST", nil, &v); err != nil {
		return BranchOperations{}, err
	}
	return v, nil
}

// ListProjectBranchEndpoints Retrieves a list of endpoints for the specified branch.
// Currently, Neon permits only one endpoint per branch.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain the `branch_id` by listing the project's branches.
func (c Client) ListProjectBranchEndpoints(projectID string, branchID string) (EndpointsResponse, error) {
	var v EndpointsResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/endpoints", "GET", nil, &v); err != nil {
		return EndpointsResponse{}, err
	}
	return v, nil
}

// ListProjectBranchDatabases Retrieves a list of databases for the specified branch.
// A branch can have multiple databases.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain the `branch_id` by listing the project's branches.
// For related information, see [Manage databases](https://neon.tech/docs/manage/databases/).
func (c Client) ListProjectBranchDatabases(projectID string, branchID string) (DatabasesResponse, error) {
	var v DatabasesResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/databases", "GET", nil, &v); err != nil {
		return DatabasesResponse{}, err
	}
	return v, nil
}

// CreateProjectBranchDatabase Creates a database in the specified branch.
// A branch can have multiple databases.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain the `branch_id` by listing the project's branches.
// For related information, see [Manage databases](https://neon.tech/docs/manage/databases/).
func (c Client) CreateProjectBranchDatabase(projectID string, branchID string, cfg DatabaseCreateRequest) (DatabaseOperations, error) {
	var v DatabaseOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/databases", "POST", cfg, &v); err != nil {
		return DatabaseOperations{}, err
	}
	return v, nil
}

// GetProjectBranchDatabase Retrieves information about the specified database.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain the `branch_id` and `database_name` by listing branch's databases.
// For related information, see [Manage databases](https://neon.tech/docs/manage/databases/).
func (c Client) GetProjectBranchDatabase(projectID string, branchID string, databaseName string) (DatabaseResponse, error) {
	var v DatabaseResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/databases/"+databaseName, "GET", nil, &v); err != nil {
		return DatabaseResponse{}, err
	}
	return v, nil
}

// UpdateProjectBranchDatabase Updates the specified database in the branch.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain the `branch_id` and `database_name` by listing the branch's databases.
// For related information, see [Manage databases](https://neon.tech/docs/manage/databases/).
func (c Client) UpdateProjectBranchDatabase(projectID string, branchID string, databaseName string, cfg DatabaseUpdateRequest) (DatabaseOperations, error) {
	var v DatabaseOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/databases/"+databaseName, "PATCH", cfg, &v); err != nil {
		return DatabaseOperations{}, err
	}
	return v, nil
}

// DeleteProjectBranchDatabase Deletes the specified database from the branch.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain the `branch_id` and `database_name` by listing branch's databases.
// For related information, see [Manage databases](https://neon.tech/docs/manage/databases/).
func (c Client) DeleteProjectBranchDatabase(projectID string, branchID string, databaseName string) (DatabaseOperations, error) {
	var v DatabaseOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/databases/"+databaseName, "DELETE", nil, &v); err != nil {
		return DatabaseOperations{}, err
	}
	return v, nil
}

// ListProjectBranchRoles Retrieves a list of roles from the specified branch.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain the `branch_id` by listing the project's branches.
// In Neon, the terms "role" and "user" are synonymous.
// For related information, see [Manage roles](https://neon.tech/docs/manage/roles/).
func (c Client) ListProjectBranchRoles(projectID string, branchID string) (RolesResponse, error) {
	var v RolesResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/roles", "GET", nil, &v); err != nil {
		return RolesResponse{}, err
	}
	return v, nil
}

// CreateProjectBranchRole Creates a role in the specified branch.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain the `branch_id` by listing the project's branches.
// In Neon, the terms "role" and "user" are synonymous.
// For related information, see [Manage roles](https://neon.tech/docs/manage/roles/).
// Connections established to the active compute endpoint will be dropped.
// If the compute endpoint is idle, the endpoint becomes active for a short period of time and is suspended afterward.
func (c Client) CreateProjectBranchRole(projectID string, branchID string, cfg RoleCreateRequest) (RoleOperations, error) {
	var v RoleOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/roles", "POST", cfg, &v); err != nil {
		return RoleOperations{}, err
	}
	return v, nil
}

// GetProjectBranchRole Retrieves details about the specified role.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain the `branch_id` by listing the project's branches.
// You can obtain the `role_name` by listing the roles for a branch.
// In Neon, the terms "role" and "user" are synonymous.
// For related information, see [Manage roles](https://neon.tech/docs/manage/roles/).
func (c Client) GetProjectBranchRole(projectID string, branchID string, roleName string) (RoleResponse, error) {
	var v RoleResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/roles/"+roleName, "GET", nil, &v); err != nil {
		return RoleResponse{}, err
	}
	return v, nil
}

// DeleteProjectBranchRole Deletes the specified role from the branch.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain the `branch_id` by listing the project's branches.
// You can obtain the `role_name` by listing the roles for a branch.
// In Neon, the terms "role" and "user" are synonymous.
// For related information, see [Manage roles](https://neon.tech/docs/manage/roles/).
func (c Client) DeleteProjectBranchRole(projectID string, branchID string, roleName string) (RoleOperations, error) {
	var v RoleOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/roles/"+roleName, "DELETE", nil, &v); err != nil {
		return RoleOperations{}, err
	}
	return v, nil
}

// GetProjectBranchRolePassword Retrieves the password for the specified role, if possible.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain the `branch_id` by listing the project's branches.
// You can obtain the `role_name` by listing the roles for a branch.
// In Neon, the terms "role" and "user" are synonymous.
// For related information, see [Manage roles](https://neon.tech/docs/manage/roles/).
func (c Client) GetProjectBranchRolePassword(projectID string, branchID string, roleName string) (RolePasswordResponse, error) {
	var v RolePasswordResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/roles/"+roleName+"/reveal_password", "GET", nil, &v); err != nil {
		return RolePasswordResponse{}, err
	}
	return v, nil
}

// ResetProjectBranchRolePassword Resets the password for the specified role.
// Returns a new password and operations. The new password is ready to use when the last operation finishes.
// The old password remains valid until last operation finishes.
// Connections to the compute endpoint are dropped. If idle,
// the compute endpoint becomes active for a short period of time.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain the `branch_id` by listing the project's branches.
// You can obtain the `role_name` by listing the roles for a branch.
// In Neon, the terms "role" and "user" are synonymous.
// For related information, see [Manage roles](https://neon.tech/docs/manage/roles/).
func (c Client) ResetProjectBranchRolePassword(projectID string, branchID string, roleName string) (RoleOperations, error) {
	var v RoleOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/roles/"+roleName+"/reset_password", "POST", nil, &v); err != nil {
		return RoleOperations{}, err
	}
	return v, nil
}

// ListProjectEndpoints Retrieves a list of endpoints for the specified project.
// An endpoint is a Neon compute instance.
// You can obtain a `project_id` by listing the projects for your Neon account.
// For more information about endpoints, see [Manage endpoints](https://neon.tech/docs/manage/endpoints/).
func (c Client) ListProjectEndpoints(projectID string) (EndpointsResponse, error) {
	var v EndpointsResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints", "GET", nil, &v); err != nil {
		return EndpointsResponse{}, err
	}
	return v, nil
}

// CreateProjectEndpoint Creates an endpoint for the specified branch.
// An endpoint is a Neon compute instance.
// There is a maximum of one read-write endpoint per branch.
// If the specified branch already has a read-write endpoint, the operation fails.
// A branch can have multiple read-only endpoints.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain `branch_id` by listing the project's branches.
// A `branch_id` has a `br-` prefix.
// For supported regions and `region_id` values, see [Regions](https://neon.tech/docs/introduction/regions/).
// For more information about endpoints, see [Manage endpoints](https://neon.tech/docs/manage/endpoints/).
func (c Client) CreateProjectEndpoint(projectID string, cfg EndpointCreateRequest) (EndpointOperations, error) {
	var v EndpointOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints", "POST", cfg, &v); err != nil {
		return EndpointOperations{}, err
	}
	return v, nil
}

// GetProjectEndpoint Retrieves information about the specified endpoint.
// An endpoint is a Neon compute instance.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain an `endpoint_id` by listing your project's endpoints.
// An `endpoint_id` has an `ep-` prefix.
// For more information about endpoints, see [Manage endpoints](https://neon.tech/docs/manage/endpoints/).
func (c Client) GetProjectEndpoint(projectID string, endpointID string) (EndpointResponse, error) {
	var v EndpointResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints/"+endpointID, "GET", nil, &v); err != nil {
		return EndpointResponse{}, err
	}
	return v, nil
}

// UpdateProjectEndpoint Updates the specified endpoint. Currently, only changing the associated branch is supported.
// The branch that you specify cannot have an existing endpoint.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain an `endpoint_id` and `branch_id` by listing your project's endpoints.
// An `endpoint_id` has an `ep-` prefix. A `branch_id` has a `br-` prefix.
// For more information about endpoints, see [Manage endpoints](https://neon.tech/docs/manage/endpoints/).
// If the returned list of operations is not empty, the endpoint is not ready to use.
// The client must wait for the last operation to finish before using the endpoint.
// If the endpoint was idle before the update, the endpoint becomes active for a short period of time,
// and the control plane suspends it again after the update.
func (c Client) UpdateProjectEndpoint(projectID string, endpointID string, cfg EndpointUpdateRequest) (EndpointOperations, error) {
	var v EndpointOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints/"+endpointID, "PATCH", cfg, &v); err != nil {
		return EndpointOperations{}, err
	}
	return v, nil
}

// DeleteProjectEndpoint Delete the specified endpoint.
// An endpoint is a Neon compute instance.
// Deleting an endpoint drops existing network connections to the endpoint.
// The deletion is completed when last operation in the chain finishes successfully.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain an `endpoint_id` by listing your project's endpoints.
// An `endpoint_id` has an `ep-` prefix.
// For more information about endpoints, see [Manage endpoints](https://neon.tech/docs/manage/endpoints/).
func (c Client) DeleteProjectEndpoint(projectID string, endpointID string) (EndpointOperations, error) {
	var v EndpointOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints/"+endpointID, "DELETE", nil, &v); err != nil {
		return EndpointOperations{}, err
	}
	return v, nil
}

// StartProjectEndpoint Starts an endpoint. The endpoint is ready to use
// after the last operation in chain finishes successfully.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain an `endpoint_id` by listing your project's endpoints.
// An `endpoint_id` has an `ep-` prefix.
// For more information about endpoints, see [Manage endpoints](https://neon.tech/docs/manage/endpoints/).
func (c Client) StartProjectEndpoint(projectID string, endpointID string) (EndpointOperations, error) {
	var v EndpointOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints/"+endpointID+"/start", "POST", nil, &v); err != nil {
		return EndpointOperations{}, err
	}
	return v, nil
}

// SuspendProjectEndpoint Suspend the specified endpoint
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain an `endpoint_id` by listing your project's endpoints.
// An `endpoint_id` has an `ep-` prefix.
// For more information about endpoints, see [Manage endpoints](https://neon.tech/docs/manage/endpoints/).
func (c Client) SuspendProjectEndpoint(projectID string, endpointID string) (EndpointOperations, error) {
	var v EndpointOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints/"+endpointID+"/suspend", "POST", nil, &v); err != nil {
		return EndpointOperations{}, err
	}
	return v, nil
}

// ListProjectsConsumption Retrieves a list consumption metrics for each project for the current billing period.
// **Important:** This is a preview API and may be subject to changes.
func (c Client) ListProjectsConsumption(cursor *string, limit *int, from *time.Time, to *time.Time) (ListProjectsConsumptionRespObj, error) {
	var queryElements []string
	if cursor != nil {
		queryElements = append(queryElements, "cursor="+*cursor)
	}
	if limit != nil {
		queryElements = append(queryElements, "limit="+strconv.FormatInt(int64(*limit), 10))
	}
	if from != nil {
		queryElements = append(queryElements, "from="+from.Format(time.RFC3339))
	}
	if to != nil {
		queryElements = append(queryElements, "to="+to.Format(time.RFC3339))
	}
	query := "?" + strings.Join(queryElements, "&")
	var v ListProjectsConsumptionRespObj
	if err := c.requestHandler(c.baseURL+"/consumption/projects"+query, "GET", nil, &v); err != nil {
		return ListProjectsConsumptionRespObj{}, err
	}
	return v, nil
}

// GetCurrentUserInfo Retrieves information about the current Neon user account.
func (c Client) GetCurrentUserInfo() (CurrentUserInfoResponse, error) {
	var v CurrentUserInfoResponse
	if err := c.requestHandler(c.baseURL+"/users/me", "GET", nil, &v); err != nil {
		return CurrentUserInfoResponse{}, err
	}
	return v, nil
}

// AllowedIps A list of IP addresses that are allowed to connect to the endpoint.
// If the list is empty or not set, all IP addresses are allowed.
// If primary_branch_only is true, the list will be applied only to the primary branch.
type AllowedIps struct {
	// Ips A list of IP addresses that are allowed to connect to the endpoint.
	Ips *[]string `json:"ips,omitempty"`
	// PrimaryBranchOnly If true, the list will be applied only to the primary branch.
	PrimaryBranchOnly bool `json:"primary_branch_only"`
}

type ApiKeyCreateRequest struct {
	// KeyName A user-specified API key name. This value is required when creating an API key.
	KeyName string `json:"key_name"`
}

type ApiKeyCreateResponse struct {
	// CreatedAt A timestamp indicating when the API key was created
	CreatedAt time.Time `json:"created_at"`
	// ID The API key ID
	ID int64 `json:"id"`
	// Key The generated 64-bit token required to access the Neon API
	Key string `json:"key"`
	// Name The user-specified API key name
	Name string `json:"name"`
}

type ApiKeyRevokeResponse struct {
	// ID The API key ID
	ID int64 `json:"id"`
	// LastUsedAt A timestamp indicating when the API was last used
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	// LastUsedFromAddr The IP address from which the API key was last used
	LastUsedFromAddr string `json:"last_used_from_addr"`
	// Name The user-specified API key name
	Name string `json:"name"`
	// Revoked A `true` or `false` value indicating whether the API key is revoked
	Revoked bool `json:"revoked"`
}

type ApiKeysListResponseItem struct {
	// CreatedAt A timestamp indicating when the API key was created
	CreatedAt time.Time `json:"created_at"`
	// ID The API key ID
	ID int64 `json:"id"`
	// LastUsedAt A timestamp indicating when the API was last used
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	// LastUsedFromAddr The IP address from which the API key was last used
	LastUsedFromAddr string `json:"last_used_from_addr"`
	// Name The user-specified API key name
	Name string `json:"name"`
}

type BillingAccount struct {
	// AddressCity Billing address city.
	AddressCity string `json:"address_city"`
	// AddressCountry Billing address country.
	AddressCountry string `json:"address_country"`
	// AddressLine1 Billing address line 1.
	AddressLine1 string `json:"address_line1"`
	// AddressLine2 Billing address line 2.
	AddressLine2 string `json:"address_line2"`
	// AddressPostalCode Billing address postal code.
	AddressPostalCode string `json:"address_postal_code"`
	// AddressState Billing address state or region.
	AddressState string `json:"address_state"`
	// Email Billing email, to receive emails related to invoices and subscriptions.
	Email string `json:"email"`
	// OrbPortalURL Orb user portal url
	OrbPortalURL  *string       `json:"orb_portal_url,omitempty"`
	PaymentSource PaymentSource `json:"payment_source"`
	// QuotaResetAtLast The last time the quota was reset. Defaults to the date-time the account is created.
	QuotaResetAtLast time.Time               `json:"quota_reset_at_last"`
	SubscriptionType BillingSubscriptionType `json:"subscription_type"`
}

// BillingSubscriptionType Type of subscription to Neon Cloud.
// Notice that for users without billing account this will be "UNKNOWN"
type BillingSubscriptionType string

type Branch struct {
	ActiveTimeSeconds  int64 `json:"active_time_seconds"`
	ComputeTimeSeconds int64 `json:"compute_time_seconds"`
	// CpuUsedSec CPU seconds used by all the endpoints of the branch, including deleted ones.
	// This value is reset at the beginning of each billing period.
	// Examples:
	// 1. A branch that uses 1 CPU for 1 second is equal to `cpu_used_sec=1`.
	// 2. A branch that uses 2 CPUs simultaneously for 1 second is equal to `cpu_used_sec=2`.
	CpuUsedSec int64 `json:"cpu_used_sec"`
	// CreatedAt A timestamp indicating when the branch was created
	CreatedAt time.Time `json:"created_at"`
	// CreationSource The branch creation source
	CreationSource    string      `json:"creation_source"`
	CurrentState      BranchState `json:"current_state"`
	DataTransferBytes int64       `json:"data_transfer_bytes"`
	// ID The branch ID. This value is generated when a branch is created. A `branch_id` value has a `br` prefix. For example: `br-small-term-683261`.
	ID string `json:"id"`
	// LastResetAt A timestamp indicating when the branch was last reset
	LastResetAt *time.Time `json:"last_reset_at,omitempty"`
	// LogicalSize The logical size of the branch, in bytes
	LogicalSize *int64 `json:"logical_size,omitempty"`
	// Name The branch name
	Name string `json:"name"`
	// ParentID The `branch_id` of the parent branch
	ParentID *string `json:"parent_id,omitempty"`
	// ParentLsn The Log Sequence Number (LSN) on the parent branch from which this branch was created
	ParentLsn *string `json:"parent_lsn,omitempty"`
	// ParentTimestamp The point in time on the parent branch from which this branch was created
	ParentTimestamp *time.Time   `json:"parent_timestamp,omitempty"`
	PendingState    *BranchState `json:"pending_state,omitempty"`
	// Primary Whether the branch is the project's primary branch
	Primary bool `json:"primary"`
	// ProjectID The ID of the project to which the branch belongs
	ProjectID string `json:"project_id"`
	// UpdatedAt A timestamp indicating when the branch was last updated
	UpdatedAt        time.Time `json:"updated_at"`
	WrittenDataBytes int64     `json:"written_data_bytes"`
}

type BranchCreateRequest struct {
	Branch    *BranchCreateRequestBranch            `json:"branch,omitempty"`
	Endpoints *[]BranchCreateRequestEndpointOptions `json:"endpoints,omitempty"`
}

type BranchCreateRequestBranch struct {
	// Name The branch name
	Name *string `json:"name,omitempty"`
	// ParentID The `branch_id` of the parent branch. If omitted or empty, the branch will be created from the project's primary branch.
	ParentID *string `json:"parent_id,omitempty"`
	// ParentLsn A Log Sequence Number (LSN) on the parent branch. The branch will be created with data from this LSN.
	ParentLsn *string `json:"parent_lsn,omitempty"`
	// ParentTimestamp A timestamp identifying a point in time on the parent branch. The branch will be created with data starting from this point in time.
	ParentTimestamp *time.Time `json:"parent_timestamp,omitempty"`
}

type BranchCreateRequestEndpointOptions struct {
	AutoscalingLimitMaxCu *ComputeUnit           `json:"autoscaling_limit_max_cu,omitempty"`
	AutoscalingLimitMinCu *ComputeUnit           `json:"autoscaling_limit_min_cu,omitempty"`
	Provisioner           *Provisioner           `json:"provisioner,omitempty"`
	SuspendTimeoutSeconds *SuspendTimeoutSeconds `json:"suspend_timeout_seconds,omitempty"`
	Type                  EndpointType           `json:"type"`
}

type BranchOperations struct {
	BranchResponse
	OperationsResponse
}

type BranchResponse struct {
	Branch Branch `json:"branch"`
}

// BranchState The branch state
type BranchState string

type BranchUpdateRequest struct {
	Branch BranchUpdateRequestBranch `json:"branch"`
}

type BranchUpdateRequestBranch struct {
	Name *string `json:"name,omitempty"`
}

type BranchesResponse struct {
	Branches []Branch `json:"branches"`
}

type ComputeUnit float64

type ConnectionDetails struct {
	ConnectionParameters ConnectionParameters `json:"connection_parameters"`
	// ConnectionURI Connection URI is same as specified in https://www.postgresql.org/docs/current/libpq-connect.html#id-1.7.3.8.3.6
	// It is a ready to use string for psql or for DATABASE_URL environment variable.
	ConnectionURI string `json:"connection_uri"`
}

type ConnectionParameters struct {
	// Database name.
	Database string `json:"database"`
	// Host name.
	Host string `json:"host"`
	// Password for the role.
	Password string `json:"password"`
	// PoolerHost Pooler host name.
	PoolerHost string `json:"pooler_host"`
	// Role name.
	Role string `json:"role"`
}

type ConnectionURIsOptionalResponse struct {
	ConnectionUris *[]ConnectionDetails `json:"connection_uris,omitempty"`
}

type ConnectionURIsResponse struct {
	ConnectionUris []ConnectionDetails `json:"connection_uris"`
}

type CreatedBranch struct {
	BranchResponse
	ConnectionURIsOptionalResponse
	DatabasesResponse
	EndpointsResponse
	OperationsResponse
	RolesResponse
}

type CreatedProject struct {
	BranchResponse
	ConnectionURIsResponse
	DatabasesResponse
	EndpointsResponse
	OperationsResponse
	ProjectResponse
	RolesResponse
}

type CurrentUserAuthAccount struct {
	Email    string `json:"email"`
	Image    string `json:"image"`
	Login    string `json:"login"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
}

type CurrentUserInfoResponse struct {
	// ActiveSecondsLimit Control plane observes active endpoints of a user this amount of wall-clock time.
	ActiveSecondsLimit  int64                    `json:"active_seconds_limit"`
	AuthAccounts        []CurrentUserAuthAccount `json:"auth_accounts"`
	BillingAccount      BillingAccount           `json:"billing_account"`
	BranchesLimit       int64                    `json:"branches_limit"`
	ComputeSecondsLimit *int64                   `json:"compute_seconds_limit,omitempty"`
	Email               string                   `json:"email"`
	ID                  string                   `json:"id"`
	Image               string                   `json:"image"`
	LastName            string                   `json:"last_name"`
	Login               string                   `json:"login"`
	MaxAutoscalingLimit ComputeUnit              `json:"max_autoscaling_limit"`
	Name                string                   `json:"name"`
	Plan                string                   `json:"plan"`
	ProjectsLimit       int64                    `json:"projects_limit"`
}

type Database struct {
	// BranchID The ID of the branch to which the database belongs
	BranchID string `json:"branch_id"`
	// CreatedAt A timestamp indicating when the database was created
	CreatedAt time.Time `json:"created_at"`
	// ID The database ID
	ID int64 `json:"id"`
	// Name The database name
	Name string `json:"name"`
	// OwnerName The name of role that owns the database
	OwnerName string `json:"owner_name"`
	// UpdatedAt A timestamp indicating when the database was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

type DatabaseCreateRequest struct {
	Database DatabaseCreateRequestDatabase `json:"database"`
}

type DatabaseCreateRequestDatabase struct {
	// Name The name of the datbase
	Name string `json:"name"`
	// OwnerName The name of the role that owns the database
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
	// Name The name of the database
	Name *string `json:"name,omitempty"`
	// OwnerName The name of the role that owns the database
	OwnerName *string `json:"owner_name,omitempty"`
}

type DatabasesResponse struct {
	Databases []Database `json:"databases"`
}

// DefaultEndpointSettings A collection of settings for a Neon endpoint
type DefaultEndpointSettings struct {
	AutoscalingLimitMaxCu *ComputeUnit           `json:"autoscaling_limit_max_cu,omitempty"`
	AutoscalingLimitMinCu *ComputeUnit           `json:"autoscaling_limit_min_cu,omitempty"`
	PgSettings            *PgSettingsData        `json:"pg_settings,omitempty"`
	PgbouncerSettings     *PgbouncerSettingsData `json:"pgbouncer_settings,omitempty"`
	SuspendTimeoutSeconds *SuspendTimeoutSeconds `json:"suspend_timeout_seconds,omitempty"`
}

type Endpoint struct {
	AutoscalingLimitMaxCu ComputeUnit `json:"autoscaling_limit_max_cu"`
	AutoscalingLimitMinCu ComputeUnit `json:"autoscaling_limit_min_cu"`
	// BranchID The ID of the branch that the compute endpoint is associated with
	BranchID string `json:"branch_id"`
	// CreatedAt A timestamp indicating when the compute endpoint was created
	CreatedAt time.Time `json:"created_at"`
	// CreationSource The compute endpoint creation source
	CreationSource string        `json:"creation_source"`
	CurrentState   EndpointState `json:"current_state"`
	// Disabled Whether to restrict connections to the compute endpoint.
	// Enabling this option schedules a suspend compute operation.
	// A disabled compute endpoint cannot be enabled by a connection or
	// console action. However, the compute endpoint is periodically
	// enabled by check_availability operations.
	Disabled bool `json:"disabled"`
	// Host The hostname of the compute endpoint. This is the hostname specified when connecting to a Neon database.
	Host string `json:"host"`
	// ID The compute endpoint ID. Compute endpoint IDs have an `ep-` prefix. For example: `ep-little-smoke-851426`
	ID string `json:"id"`
	// LastActive A timestamp indicating when the compute endpoint was last active
	LastActive *time.Time `json:"last_active,omitempty"`
	// PasswordlessAccess Whether to permit passwordless access to the compute endpoint
	PasswordlessAccess bool           `json:"passwordless_access"`
	PendingState       *EndpointState `json:"pending_state,omitempty"`
	// PoolerEnabled Whether connection pooling is enabled for the compute endpoint
	PoolerEnabled bool               `json:"pooler_enabled"`
	PoolerMode    EndpointPoolerMode `json:"pooler_mode"`
	// ProjectID The ID of the project to which the compute endpoint belongs
	ProjectID   string      `json:"project_id"`
	Provisioner Provisioner `json:"provisioner"`
	// ProxyHost DEPRECATED. Use the "host" property instead.
	ProxyHost string `json:"proxy_host"`
	// RegionID The region identifier
	RegionID              string                `json:"region_id"`
	Settings              EndpointSettingsData  `json:"settings"`
	SuspendTimeoutSeconds SuspendTimeoutSeconds `json:"suspend_timeout_seconds"`
	Type                  EndpointType          `json:"type"`
	// UpdatedAt A timestamp indicating when the compute endpoint was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

type EndpointCreateRequest struct {
	Endpoint EndpointCreateRequestEndpoint `json:"endpoint"`
}

type EndpointCreateRequestEndpoint struct {
	AutoscalingLimitMaxCu *ComputeUnit `json:"autoscaling_limit_max_cu,omitempty"`
	AutoscalingLimitMinCu *ComputeUnit `json:"autoscaling_limit_min_cu,omitempty"`
	// BranchID The ID of the branch the compute endpoint will be associated with
	BranchID string `json:"branch_id"`
	// Disabled Whether to restrict connections to the compute endpoint.
	// Enabling this option schedules a suspend compute operation.
	// A disabled compute endpoint cannot be enabled by a connection or
	// console action. However, the compute endpoint is periodically
	// enabled by check_availability operations.
	Disabled *bool `json:"disabled,omitempty"`
	// PasswordlessAccess NOT YET IMPLEMENTED. Whether to permit passwordless access to the compute endpoint.
	PasswordlessAccess *bool `json:"passwordless_access,omitempty"`
	// PoolerEnabled Whether to enable connection pooling for the compute endpoint
	PoolerEnabled *bool               `json:"pooler_enabled,omitempty"`
	PoolerMode    *EndpointPoolerMode `json:"pooler_mode,omitempty"`
	Provisioner   *Provisioner        `json:"provisioner,omitempty"`
	// RegionID The region where the compute endpoint will be created. Only the project's `region_id` is permitted.
	RegionID              *string                `json:"region_id,omitempty"`
	Settings              *EndpointSettingsData  `json:"settings,omitempty"`
	SuspendTimeoutSeconds *SuspendTimeoutSeconds `json:"suspend_timeout_seconds,omitempty"`
	Type                  EndpointType           `json:"type"`
}

type EndpointOperations struct {
	EndpointResponse
	OperationsResponse
}

// EndpointPoolerMode The connection pooler mode. Neon supports PgBouncer in `transaction` mode only.
type EndpointPoolerMode string

type EndpointResponse struct {
	Endpoint Endpoint `json:"endpoint"`
}

// EndpointSettingsData A collection of settings for a compute endpoint
type EndpointSettingsData struct {
	PgSettings        *PgSettingsData        `json:"pg_settings,omitempty"`
	PgbouncerSettings *PgbouncerSettingsData `json:"pgbouncer_settings,omitempty"`
}

// EndpointState The state of the compute endpoint
type EndpointState string

// EndpointType The compute endpoint type. Either `read_write` or `read_only`.
// The `read_only` compute endpoint type is not yet supported.
type EndpointType string

type EndpointUpdateRequest struct {
	Endpoint EndpointUpdateRequestEndpoint `json:"endpoint"`
}

type EndpointUpdateRequestEndpoint struct {
	AutoscalingLimitMaxCu *ComputeUnit `json:"autoscaling_limit_max_cu,omitempty"`
	AutoscalingLimitMinCu *ComputeUnit `json:"autoscaling_limit_min_cu,omitempty"`
	// BranchID The destination branch ID. The destination branch must not have an exsiting read-write endpoint.
	BranchID *string `json:"branch_id,omitempty"`
	// Disabled Whether to restrict connections to the compute endpoint.
	// Enabling this option schedules a suspend compute operation.
	// A disabled compute endpoint cannot be enabled by a connection or
	// console action. However, the compute endpoint is periodically
	// enabled by check_availability operations.
	Disabled *bool `json:"disabled,omitempty"`
	// PasswordlessAccess NOT YET IMPLEMENTED. Whether to permit passwordless access to the compute endpoint.
	PasswordlessAccess *bool `json:"passwordless_access,omitempty"`
	// PoolerEnabled Whether to enable connection pooling for the compute endpoint
	PoolerEnabled         *bool                  `json:"pooler_enabled,omitempty"`
	PoolerMode            *EndpointPoolerMode    `json:"pooler_mode,omitempty"`
	Provisioner           *Provisioner           `json:"provisioner,omitempty"`
	Settings              *EndpointSettingsData  `json:"settings,omitempty"`
	SuspendTimeoutSeconds *SuspendTimeoutSeconds `json:"suspend_timeout_seconds,omitempty"`
}

type EndpointsResponse struct {
	Endpoints []Endpoint `json:"endpoints"`
}

type GrantPermissionToProjectRequest struct {
	Email string `json:"email"`
}

type ListOperations struct {
	OperationsResponse
	PaginationResponse
}

type ListProjectsConsumptionRespObj struct {
	PaginationResponse
	ProjectsConsumptionResponse
}

type ListProjectsRespObj struct {
	PaginationResponse
	ProjectsResponse
}

type ListSharedProjectsRespObj struct {
	PaginationResponse
	ProjectsResponse
}

type Operation struct {
	Action OperationAction `json:"action"`
	// BranchID The branch ID
	BranchID *string `json:"branch_id,omitempty"`
	// CreatedAt A timestamp indicating when the operation was created
	CreatedAt time.Time `json:"created_at"`
	// EndpointID The endpoint ID
	EndpointID *string `json:"endpoint_id,omitempty"`
	// Error The error that occured
	Error *string `json:"error,omitempty"`
	// FailuresCount The number of times the operation failed
	FailuresCount int32 `json:"failures_count"`
	// ID The operation ID
	ID string `json:"id"`
	// ProjectID The Neon project ID
	ProjectID string `json:"project_id"`
	// RetryAt A timestamp indicating when the operation was last retried
	RetryAt *time.Time      `json:"retry_at,omitempty"`
	Status  OperationStatus `json:"status"`
	// TotalDurationMs The total duration of the operation in milliseconds
	TotalDurationMs int32 `json:"total_duration_ms"`
	// UpdatedAt A timestamp indicating when the operation status was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// OperationAction The action performed by the operation
type OperationAction string

type OperationResponse struct {
	Operation Operation `json:"operation"`
}

// OperationStatus The status of the operation
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
	Pagination *Pagination `json:"pagination,omitempty"`
}

type PaymentSource struct {
	Card *PaymentSourceBankCard `json:"card,omitempty"`
	// Type of payment source. E.g. "card".
	Type string `json:"type"`
}

type PaymentSourceBankCard struct {
	// Brand of credit card.
	Brand *string `json:"brand,omitempty"`
	// ExpMonth Credit card expiration month
	ExpMonth *int64 `json:"exp_month,omitempty"`
	// ExpYear Credit card expiration year
	ExpYear *int64 `json:"exp_year,omitempty"`
	// Last4 Last 4 digits of the card.
	Last4 string `json:"last4"`
}

// PgSettingsData A raw representation of PostgreSQL settings
type PgSettingsData map[string]interface{}

// PgVersion The major PostgreSQL version number. Currently supported versions are `14`, `15` and `16`.
type PgVersion int

// PgbouncerSettingsData A raw representation of PgBouncer settings
type PgbouncerSettingsData map[string]interface{}

type Project struct {
	// ActiveTimeSeconds Seconds. Control plane observed endpoints of this project being active this amount of wall-clock time.
	// The value has some lag.
	// The value is reset at the beginning of each billing period.
	ActiveTimeSeconds int64 `json:"active_time_seconds"`
	// BranchLogicalSizeLimit The logical size limit for a branch. The value is in MiB.
	BranchLogicalSizeLimit int64 `json:"branch_logical_size_limit"`
	// BranchLogicalSizeLimitBytes The logical size limit for a branch. The value is in B.
	BranchLogicalSizeLimitBytes int64 `json:"branch_logical_size_limit_bytes"`
	// ComputeLastActiveAt The most recent time when any endpoint of this project was active.
	//
	// Omitted when observed no actitivy for endpoints of this project.
	ComputeLastActiveAt *time.Time `json:"compute_last_active_at,omitempty"`
	// ComputeTimeSeconds Seconds. The number of CPU seconds used by the project's compute endpoints, including compute endpoints that have been deleted.
	// The value has some lag. The value is reset at the beginning of each billing period.
	// Examples:
	// 1. An endpoint that uses 1 CPU for 1 second is equal to `compute_time=1`.
	// 2. An endpoint that uses 2 CPUs simultaneously for 1 second is equal to `compute_time=2`.
	ComputeTimeSeconds int64 `json:"compute_time_seconds"`
	// ConsumptionPeriodEnd A date-time indicating when Neon Cloud plans to stop measuring consumption for current consumption period.
	ConsumptionPeriodEnd time.Time `json:"consumption_period_end"`
	// ConsumptionPeriodStart A date-time indicating when Neon Cloud started measuring consumption for current consumption period.
	ConsumptionPeriodStart time.Time `json:"consumption_period_start"`
	// CpuUsedSec DEPRECATED, use compute_time instead.
	CpuUsedSec int64 `json:"cpu_used_sec"`
	// CreatedAt A timestamp indicating when the project was created
	CreatedAt time.Time `json:"created_at"`
	// CreationSource The project creation source
	CreationSource string `json:"creation_source"`
	// DataStorageBytesHour Bytes-Hour. Project consumed that much storage hourly during the billing period. The value has some lag.
	// The value is reset at the beginning of each billing period.
	DataStorageBytesHour int64 `json:"data_storage_bytes_hour"`
	// DataTransferBytes Bytes. Egress traffic from the Neon cloud to the client for given project over the billing period.
	// Includes deleted endpoints. The value has some lag. The value is reset at the beginning of each billing period.
	DataTransferBytes       int64                    `json:"data_transfer_bytes"`
	DefaultEndpointSettings *DefaultEndpointSettings `json:"default_endpoint_settings,omitempty"`
	// HistoryRetentionSeconds The number of seconds to retain point-in-time restore (PITR) backup history for this project.
	HistoryRetentionSeconds int64 `json:"history_retention_seconds"`
	// ID The project ID
	ID string `json:"id"`
	// MaintenanceStartsAt A timestamp indicating when project maintenance begins. If set, the project is placed into maintenance mode at this time.
	MaintenanceStartsAt *time.Time `json:"maintenance_starts_at,omitempty"`
	// Name The project name
	Name      string            `json:"name"`
	Owner     *ProjectOwnerData `json:"owner,omitempty"`
	OwnerID   string            `json:"owner_id"`
	PgVersion PgVersion         `json:"pg_version"`
	// PlatformID The cloud platform identifier. Currently, only AWS is supported, for which the identifier is `aws`.
	PlatformID  string      `json:"platform_id"`
	Provisioner Provisioner `json:"provisioner"`
	// ProxyHost The proxy host for the project. This value combines the `region_id`, the `platform_id`, and the Neon domain (`neon.tech`).
	ProxyHost string `json:"proxy_host"`
	// QuotaResetAt DEPRECATED. Use `consumption_period_end` from the getProject endpoint instead.
	// A timestamp indicating when the project quota resets.
	QuotaResetAt *time.Time `json:"quota_reset_at,omitempty"`
	// RegionID The region identifier
	RegionID string               `json:"region_id"`
	Settings *ProjectSettingsData `json:"settings,omitempty"`
	// StorePasswords Whether or not passwords are stored for roles in the Neon project. Storing passwords facilitates access to Neon features that require authorization.
	StorePasswords bool `json:"store_passwords"`
	// SyntheticStorageSize The current space occupied by the project in storage, in bytes. Synthetic storage size combines the logical data size and Write-Ahead Log (WAL) size for all branches in a project.
	SyntheticStorageSize *int64 `json:"synthetic_storage_size,omitempty"`
	// UpdatedAt A timestamp indicating when the project was last updated
	UpdatedAt time.Time `json:"updated_at"`
	// WrittenDataBytes Bytes. Amount of WAL that travelled through storage for given project across all branches.
	// The value has some lag. The value is reset at the beginning of each billing period.
	WrittenDataBytes int64 `json:"written_data_bytes"`
}

type ProjectConsumption struct {
	// ActiveTimeSeconds Seconds. The amount of time that compute endpoints in this project have been active.
	// Expect some lag in the reported value.
	//
	// The value is reset at the beginning of each billing period.
	ActiveTimeSeconds int64 `json:"active_time_seconds"`
	// ActiveTimeSecondsUpdatedAt The timestamp of the last update of the `active_time_seconds` field.
	ActiveTimeSecondsUpdatedAt *time.Time `json:"active_time_seconds_updated_at,omitempty"`
	// ComputeTimeSeconds Seconds. The number of CPU seconds used by the project's compute endpoints, including compute endpoints that have been deleted.
	// Expect some lag in the reported value. The value is reset at the beginning of each billing period.
	// Examples:
	// 1. An endpoint that uses 1 CPU for 1 second is equal to `compute_time=1`.
	// 2. An endpoint that uses 2 CPUs simultaneously for 1 second is equal to `compute_time=2`.
	ComputeTimeSeconds int64 `json:"compute_time_seconds"`
	// ComputeTimeSecondsUpdatedAt The timestamp of the last update of `compute_time_seconds` field.
	ComputeTimeSecondsUpdatedAt *time.Time `json:"compute_time_seconds_updated_at,omitempty"`
	// DataStorageBytesHour Bytes-Hour. The amount of storage the project consumed during the billing period. Expect some lag in the reported value.
	// The value is reset at the beginning of each billing period.
	DataStorageBytesHour int64 `json:"data_storage_bytes_hour"`
	// DataStorageBytesHourUpdatedAt The timestamp of the last update of the `data_storage_bytes_hour` field.
	DataStorageBytesHourUpdatedAt *time.Time `json:"data_storage_bytes_hour_updated_at,omitempty"`
	// DataTransferBytes Bytes. The egress traffic from the Neon cloud to the client for the project over the billing period.
	// Includes egress traffic for deleted endpoints. Expect some lag in the reported value. The value is reset at the beginning of each billing period.
	DataTransferBytes int64 `json:"data_transfer_bytes"`
	// DataTransferBytesUpdatedAt Timestamp of the last update of `data_transfer_bytes` field
	DataTransferBytesUpdatedAt *time.Time `json:"data_transfer_bytes_updated_at,omitempty"`
	// PeriodEnd The end of the consumption period.
	PeriodEnd time.Time `json:"period_end"`
	// PeriodID The Id of the consumption period, used to reference the `previous_period_id` field.
	PeriodID string `json:"period_id"`
	// PeriodStart The start of the consumption period.
	PeriodStart time.Time `json:"period_start"`
	// PreviousPeriodID The `period_id` of the previous consumption period.
	PreviousPeriodID string `json:"previous_period_id"`
	// ProjectID The project ID
	ProjectID string `json:"project_id"`
	// SyntheticStorageSize Bytes. The current space occupied by project in storage. Expect some lag in the reported value.
	SyntheticStorageSize int64 `json:"synthetic_storage_size"`
	// SyntheticStorageSizeUpdatedAt The timestamp of the last update of the `synthetic_storage_size` field.
	SyntheticStorageSizeUpdatedAt *time.Time `json:"synthetic_storage_size_updated_at,omitempty"`
	// UpdatedAt A timestamp indicating when the period was last updated.
	UpdatedAt time.Time `json:"updated_at"`
	// WrittenDataBytes Bytes. The Amount of WAL that travelled through storage for given project for all branches.
	// Expect some lag in the reported value. The value is reset at the beginning of each billing period.
	WrittenDataBytes int64 `json:"written_data_bytes"`
	// WrittenDataBytesUpdatedAt The timestamp of the last update of `written_data_bytes` field.
	WrittenDataBytesUpdatedAt *time.Time `json:"written_data_bytes_updated_at,omitempty"`
}

type ProjectCreateRequest struct {
	Project ProjectCreateRequestProject `json:"project"`
}

type ProjectCreateRequestProject struct {
	AutoscalingLimitMaxCu   *ComputeUnit                       `json:"autoscaling_limit_max_cu,omitempty"`
	AutoscalingLimitMinCu   *ComputeUnit                       `json:"autoscaling_limit_min_cu,omitempty"`
	Branch                  *ProjectCreateRequestProjectBranch `json:"branch,omitempty"`
	DefaultEndpointSettings *DefaultEndpointSettings           `json:"default_endpoint_settings,omitempty"`
	// HistoryRetentionSeconds The number of seconds to retain the point-in-time restore (PITR) backup history for this project.
	// The default is 604800 seconds (7 days).
	HistoryRetentionSeconds *int64 `json:"history_retention_seconds,omitempty"`
	// Name The project name
	Name        *string      `json:"name,omitempty"`
	PgVersion   *PgVersion   `json:"pg_version,omitempty"`
	Provisioner *Provisioner `json:"provisioner,omitempty"`
	// RegionID The region identifier. Refer to our [Regions](https://neon.tech/docs/introduction/regions) documentation for supported regions. Values are specified in this format: `aws-us-east-1`
	RegionID *string              `json:"region_id,omitempty"`
	Settings *ProjectSettingsData `json:"settings,omitempty"`
	// StorePasswords Whether or not passwords are stored for roles in the Neon project. Storing passwords facilitates access to Neon features that require authorization.
	StorePasswords *bool `json:"store_passwords,omitempty"`
}

type ProjectCreateRequestProjectBranch struct {
	// DatabaseName The database name. If not specified, the default database name will be used.
	DatabaseName *string `json:"database_name,omitempty"`
	// Name The branch name. If not specified, the default branch name will be used.
	Name *string `json:"name,omitempty"`
	// RoleName The role name. If not specified, the default role name will be used.
	RoleName *string `json:"role_name,omitempty"`
}

// ProjectListItem Essential data about the project. Full data is available at the getProject endpoint.
type ProjectListItem struct {
	// ActiveTime Control plane observed endpoints of this project being active this amount of wall-clock time.
	ActiveTime int64 `json:"active_time"`
	// BranchLogicalSizeLimit The logical size limit for a branch. The value is in MiB.
	BranchLogicalSizeLimit int64 `json:"branch_logical_size_limit"`
	// BranchLogicalSizeLimitBytes The logical size limit for a branch. The value is in B.
	BranchLogicalSizeLimitBytes int64 `json:"branch_logical_size_limit_bytes"`
	// ComputeLastActiveAt The most recent time when any endpoint of this project was active.
	//
	// Omitted when observed no actitivy for endpoints of this project.
	ComputeLastActiveAt *time.Time `json:"compute_last_active_at,omitempty"`
	// CpuUsedSec DEPRECATED. Use data from the getProject endpoint instead.
	CpuUsedSec int64 `json:"cpu_used_sec"`
	// CreatedAt A timestamp indicating when the project was created
	CreatedAt time.Time `json:"created_at"`
	// CreationSource The project creation source
	CreationSource          string                   `json:"creation_source"`
	DefaultEndpointSettings *DefaultEndpointSettings `json:"default_endpoint_settings,omitempty"`
	// ID The project ID
	ID string `json:"id"`
	// MaintenanceStartsAt A timestamp indicating when project maintenance begins. If set, the project is placed into maintenance mode at this time.
	MaintenanceStartsAt *time.Time `json:"maintenance_starts_at,omitempty"`
	// Name The project name
	Name      string    `json:"name"`
	OwnerID   string    `json:"owner_id"`
	PgVersion PgVersion `json:"pg_version"`
	// PlatformID The cloud platform identifier. Currently, only AWS is supported, for which the identifier is `aws`.
	PlatformID  string      `json:"platform_id"`
	Provisioner Provisioner `json:"provisioner"`
	// ProxyHost The proxy host for the project. This value combines the `region_id`, the `platform_id`, and the Neon domain (`neon.tech`).
	ProxyHost string `json:"proxy_host"`
	// QuotaResetAt DEPRECATED. Use `consumption_period_end` from the getProject endpoint instead.
	// A timestamp indicating when the project quota resets
	QuotaResetAt *time.Time `json:"quota_reset_at,omitempty"`
	// RegionID The region identifier
	RegionID string               `json:"region_id"`
	Settings *ProjectSettingsData `json:"settings,omitempty"`
	// StorePasswords Whether or not passwords are stored for roles in the Neon project. Storing passwords facilitates access to Neon features that require authorization.
	StorePasswords bool `json:"store_passwords"`
	// SyntheticStorageSize The current space occupied by the project in storage, in bytes. Synthetic storage size combines the logical data size and Write-Ahead Log (WAL) size for all branches in a project.
	SyntheticStorageSize *int64 `json:"synthetic_storage_size,omitempty"`
	// UpdatedAt A timestamp indicating when the project was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

type ProjectOwnerData struct {
	BranchesLimit    int                     `json:"branches_limit"`
	Email            string                  `json:"email"`
	SubscriptionType BillingSubscriptionType `json:"subscription_type"`
}

type ProjectPermission struct {
	GrantedAt      time.Time  `json:"granted_at"`
	GrantedToEmail string     `json:"granted_to_email"`
	ID             string     `json:"id"`
	RevokedAt      *time.Time `json:"revoked_at,omitempty"`
}

type ProjectPermissions struct {
	ProjectPermissions []ProjectPermission `json:"project_permissions"`
}

// ProjectQuota Per-project consumption quota. If the quota is exceeded, all active computes
// are automatically suspended and it will not be possible to start them with
// an API method call or incoming proxy connections. The only exception is
// `logical_size_bytes`, which is applied on per-branch basis, i.e., only the
// compute on the branch that exceeds the `logical_size` quota will be suspended.
//
// Quotas are enforced based on per-project consumption metrics with the same names,
// which are reset at the end of each billing period (the first day of the month).
// Logical size is also an exception in this case, as it represents the total size
// of data stored in a branch, so it is not reset.
//
// A zero or empty quota value means 'unlimited'.
type ProjectQuota struct {
	// ActiveTimeSeconds The total amount of wall-clock time allowed to be spent by the project's compute endpoints.
	ActiveTimeSeconds *int64 `json:"active_time_seconds,omitempty"`
	// ComputeTimeSeconds The total amount of CPU seconds allowed to be spent by the project's compute endpoints.
	ComputeTimeSeconds *int64 `json:"compute_time_seconds,omitempty"`
	// DataTransferBytes Total amount of data transferred from all of a project's branches using the proxy.
	DataTransferBytes *int64 `json:"data_transfer_bytes,omitempty"`
	// LogicalSizeBytes Limit on the logical size of every project's branch.
	LogicalSizeBytes *int64 `json:"logical_size_bytes,omitempty"`
	// WrittenDataBytes Total amount of data written to all of a project's branches.
	WrittenDataBytes *int64 `json:"written_data_bytes,omitempty"`
}

type ProjectResponse struct {
	Project Project `json:"project"`
}

type ProjectSettingsData struct {
	AllowedIps *AllowedIps `json:"allowed_ips,omitempty"`
	// EnableLogicalReplication Sets wal_level=logical for all compute endpoints in this project.
	// All active endpoints will be suspended.
	// Once enabled, logical replication cannot be disabled.
	EnableLogicalReplication *bool         `json:"enable_logical_replication,omitempty"`
	Quota                    *ProjectQuota `json:"quota,omitempty"`
}

type ProjectUpdateRequest struct {
	Project ProjectUpdateRequestProject `json:"project"`
}

type ProjectUpdateRequestProject struct {
	DefaultEndpointSettings *DefaultEndpointSettings `json:"default_endpoint_settings,omitempty"`
	// HistoryRetentionSeconds The number of seconds to retain the point-in-time restore (PITR) backup history for this project.
	// The default is 604800 seconds (7 days).
	HistoryRetentionSeconds *int64 `json:"history_retention_seconds,omitempty"`
	// Name The project name
	Name     *string              `json:"name,omitempty"`
	Settings *ProjectSettingsData `json:"settings,omitempty"`
}

type ProjectsConsumptionResponse struct {
	PeriodsInResponse int64                `json:"periods_in_response"`
	Projects          []ProjectConsumption `json:"projects"`
}

type ProjectsResponse struct {
	Projects []ProjectListItem `json:"projects"`
}

// Provisioner The Neon compute provisioner.
// Specify the `k8s-neonvm` provisioner to create a compute endpoint that supports Autoscaling.
type Provisioner string

type Role struct {
	// BranchID The ID of the branch to which the role belongs
	BranchID string `json:"branch_id"`
	// CreatedAt A timestamp indicating when the role was created
	CreatedAt time.Time `json:"created_at"`
	// Name The role name
	Name string `json:"name"`
	// Password The role password
	Password *string `json:"password,omitempty"`
	// Protected Whether or not the role is system-protected
	Protected *bool `json:"protected,omitempty"`
	// UpdatedAt A timestamp indicating when the role was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

type RoleCreateRequest struct {
	Role RoleCreateRequestRole `json:"role"`
}

type RoleCreateRequestRole struct {
	// Name The role name. Cannot exceed 63 bytes in length.
	Name string `json:"name"`
}

type RoleOperations struct {
	OperationsResponse
	RoleResponse
}

type RolePasswordResponse struct {
	// Password The role password
	Password string `json:"password"`
}

type RoleResponse struct {
	Role Role `json:"role"`
}

type RolesResponse struct {
	Roles []Role `json:"roles"`
}

// SuspendTimeoutSeconds Duration of inactivity in seconds after which the compute endpoint is
// automatically suspended. The value `0` means use the global default.
// The value `-1` means never suspend. The default value is `300` seconds (5 minutes).
// The maximum value is `604800` seconds (1 week). For more information, see
// [Auto-suspend configuration](https://neon.tech/docs/manage/endpoints#auto-suspend-configuration).
type SuspendTimeoutSeconds int64

type UpdateProjectRespObj struct {
	OperationsResponse
	ProjectResponse
}
