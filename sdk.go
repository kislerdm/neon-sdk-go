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

// AddProjectJWKS Add a new JWKS URL to a project, such that it can be used for verifying JWTs used as the authentication mechanism for the specified project.
// The URL must be a valid HTTPS URL that returns a JSON Web Key Set.
// The `provider_name` field allows you to specify which authentication provider you're using (e.g., Clerk, Auth0, AWS Cognito, etc.).
// The `branch_id` can be used to specify on which branches the JWKS URL will be accepted. If not specified, then it will work on any branch.
// The `role_names` can be used to specify for which roles the JWKS URL will be accepted.
// The `jwt_audience` can be used to specify which "aud" values should be accepted by Neon in the JWTs that are used for authentication.
func (c Client) AddProjectJWKS(projectID string, cfg AddProjectJWKSRequest) (JWKSCreationOperation, error) {
	var v JWKSCreationOperation
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/jwks", "POST", cfg, &v); err != nil {
		return JWKSCreationOperation{}, err
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

// CreateOrgApiKey Creates an API key for the specified organization.
// The `key_name` is a user-specified name for the key.
// This method returns an `id` and `key`. The `key` is a randomly generated, 64-bit token required to access the Neon API.
// API keys can also be managed in the Neon Console.
// See [Manage API keys](https://neon.tech/docs/manage/api-keys/).
func (c Client) CreateOrgApiKey(orgID string, cfg OrgApiKeyCreateRequest) (OrgApiKeyCreateResponse, error) {
	var v OrgApiKeyCreateResponse
	if err := c.requestHandler(c.baseURL+"/organizations/"+orgID+"/api_keys", "POST", cfg, &v); err != nil {
		return OrgApiKeyCreateResponse{}, err
	}
	return v, nil
}

// CreateOrganizationInvitations Creates invitations for a specific organization.
// If the invited user has an existing account, they automatically join as a member.
// If they don't yet have an account, they are invited to create one, after which they become a member.
// Each invited user receives an email notification.
func (c Client) CreateOrganizationInvitations(orgID string, cfg OrganizationInvitesCreateRequest) (OrganizationInvitationsResponse, error) {
	var v OrganizationInvitationsResponse
	if err := c.requestHandler(c.baseURL+"/organizations/"+orgID+"/invitations", "POST", cfg, &v); err != nil {
		return OrganizationInvitationsResponse{}, err
	}
	return v, nil
}

// CreateProject Creates a Neon project.
// A project is the top-level object in the Neon object hierarchy.
// Plan limits define how many projects you can create.
// For more information, see [Manage projects](https://neon.tech/docs/manage/projects/).
// You can specify a region and Postgres version in the request body.
// Neon currently supports PostgreSQL 14, 15, 16, and 17.
// For supported regions and `region_id` values, see [Regions](https://neon.tech/docs/introduction/regions/).
func (c Client) CreateProject(cfg ProjectCreateRequest) (CreatedProject, error) {
	var v CreatedProject
	if err := c.requestHandler(c.baseURL+"/projects", "POST", cfg, &v); err != nil {
		return CreatedProject{}, err
	}
	return v, nil
}

// CreateProjectBranch Creates a branch in the specified project.
// You can obtain a `project_id` by listing the projects for your Neon account.
// This method does not require a request body, but you can specify one to create a compute endpoint for the branch or to select a non-default parent branch.
// The default behavior is to create a branch from the project's default branch with no compute endpoint, and the branch name is auto-generated.
// There is a maximum of one read-write endpoint per branch.
// A branch can have multiple read-only endpoints.
// For related information, see [Manage branches](https://neon.tech/docs/manage/branches/).
func (c Client) CreateProjectBranch(projectID string, cfg *CreateProjectBranchReqObj) (CreatedBranch, error) {
	var v CreatedBranch
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches", "POST", cfg, &v); err != nil {
		return CreatedBranch{}, err
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

// CreateProjectBranchRole Creates a Postgres role in the specified branch.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain the `branch_id` by listing the project's branches.
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

// CreateProjectEndpoint Creates a compute endpoint for the specified branch.
// An endpoint is a Neon compute instance.
// There is a maximum of one read-write compute endpoint per branch.
// If the specified branch already has a read-write compute endpoint, the operation fails.
// A branch can have multiple read-only compute endpoints.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain `branch_id` by listing the project's branches.
// A `branch_id` has a `br-` prefix.
// For supported regions and `region_id` values, see [Regions](https://neon.tech/docs/introduction/regions/).
// For more information about compute endpoints, see [Manage computes](https://neon.tech/docs/manage/endpoints/).
func (c Client) CreateProjectEndpoint(projectID string, cfg EndpointCreateRequest) (EndpointOperations, error) {
	var v EndpointOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints", "POST", cfg, &v); err != nil {
		return EndpointOperations{}, err
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

// DeleteProjectBranch Deletes the specified branch from a project, and places
// all compute endpoints into an idle state, breaking existing client connections.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain a `branch_id` by listing the project's branches.
// For related information, see [Manage branches](https://neon.tech/docs/manage/branches/).
// When a successful response status is received, the compute endpoints are still active,
// and the branch is not yet deleted from storage.
// The deletion occurs after all operations finish.
// You cannot delete a project's root or default branch, and you cannot delete a branch that has a child branch.
// A project must have at least one branch.
func (c Client) DeleteProjectBranch(projectID string, branchID string) (BranchOperations, error) {
	var v BranchOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID, "DELETE", nil, &v); err != nil {
		return BranchOperations{}, err
	}
	return v, nil
}

// DeleteProjectBranchDatabase Deletes the specified database from the branch.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain the `branch_id` and `database_name` by listing the branch's databases.
// For related information, see [Manage databases](https://neon.tech/docs/manage/databases/).
func (c Client) DeleteProjectBranchDatabase(projectID string, branchID string, databaseName string) (DatabaseOperations, error) {
	var v DatabaseOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/databases/"+databaseName, "DELETE", nil, &v); err != nil {
		return DatabaseOperations{}, err
	}
	return v, nil
}

// DeleteProjectBranchRole Deletes the specified Postgres role from the branch.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain the `branch_id` by listing the project's branches.
// You can obtain the `role_name` by listing the roles for a branch.
// For related information, see [Manage roles](https://neon.tech/docs/manage/roles/).
func (c Client) DeleteProjectBranchRole(projectID string, branchID string, roleName string) (RoleOperations, error) {
	var v RoleOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/roles/"+roleName, "DELETE", nil, &v); err != nil {
		return RoleOperations{}, err
	}
	return v, nil
}

// DeleteProjectEndpoint Delete the specified compute endpoint.
// A compute endpoint is a Neon compute instance.
// Deleting a compute endpoint drops existing network connections to the compute endpoint.
// The deletion is completed when last operation in the chain finishes successfully.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain an `endpoint_id` by listing your project's compute endpoints.
// An `endpoint_id` has an `ep-` prefix.
// For information about compute endpoints, see [Manage computes](https://neon.tech/docs/manage/endpoints/).
func (c Client) DeleteProjectEndpoint(projectID string, endpointID string) (EndpointOperations, error) {
	var v EndpointOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints/"+endpointID, "DELETE", nil, &v); err != nil {
		return EndpointOperations{}, err
	}
	return v, nil
}

// DeleteProjectJWKS Deletes a JWKS URL from the specified project
func (c Client) DeleteProjectJWKS(projectID string, jwksID string) (JWKS, error) {
	var v JWKS
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/jwks/"+jwksID, "DELETE", nil, &v); err != nil {
		return JWKS{}, err
	}
	return v, nil
}

// GetActiveRegions Retrieves the list of supported Neon regions
func (c Client) GetActiveRegions() (ActiveRegionsResponse, error) {
	var v ActiveRegionsResponse
	if err := c.requestHandler(c.baseURL+"/regions", "GET", nil, &v); err != nil {
		return ActiveRegionsResponse{}, err
	}
	return v, nil
}

// GetConnectionURI Retrieves a connection URI for the specified database.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain the `database_name` by listing the databases for a branch.
// You can obtain a `role_name` by listing the roles for a branch.
func (c Client) GetConnectionURI(projectID string, branchID *string, endpointID *string, databaseName string, roleName string, pooled *bool) (ConnectionURIResponse, error) {
	var (
		queryElements []string
		query         string
	)
	queryElements = append(queryElements, "database_name="+databaseName)
	queryElements = append(queryElements, "role_name="+roleName)
	if branchID != nil {
		queryElements = append(queryElements, "branch_id="+*branchID)
	}
	if endpointID != nil {
		queryElements = append(queryElements, "endpoint_id="+*endpointID)
	}
	if pooled != nil {
		queryElements = append(queryElements, "pooled="+func(pooled bool) string {
			if pooled {
				return "true"
			}
			return "false"
		}(*pooled))
	}
	if len(queryElements) > 0 {
		query = "?" + strings.Join(queryElements, "&")
	}
	var v ConnectionURIResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/connection_uri"+query, "GET", nil, &v); err != nil {
		return ConnectionURIResponse{}, err
	}
	return v, nil
}

// GetConsumptionHistoryPerAccount Retrieves consumption metrics for Scale and Business plan accounts. History begins at the time of upgrade.
// Available for Scale and Business plan users only.
func (c Client) GetConsumptionHistoryPerAccount(from time.Time, to time.Time, granularity ConsumptionHistoryGranularity, orgID *string, includeV1Metrics *bool) (ConsumptionHistoryPerAccountResponse, error) {
	var (
		queryElements []string
		query         string
	)
	queryElements = append(queryElements, "from="+from.Format(time.RFC3339))
	queryElements = append(queryElements, "to="+to.Format(time.RFC3339))
	queryElements = append(queryElements, "granularity="+string(granularity))
	if orgID != nil {
		queryElements = append(queryElements, "org_id="+*orgID)
	}
	if includeV1Metrics != nil {
		queryElements = append(queryElements, "include_v1_metrics="+func(includeV1Metrics bool) string {
			if includeV1Metrics {
				return "true"
			}
			return "false"
		}(*includeV1Metrics))
	}
	if len(queryElements) > 0 {
		query = "?" + strings.Join(queryElements, "&")
	}
	var v ConsumptionHistoryPerAccountResponse
	if err := c.requestHandler(c.baseURL+"/consumption_history/account"+query, "GET", nil, &v); err != nil {
		return ConsumptionHistoryPerAccountResponse{}, err
	}
	return v, nil
}

// GetConsumptionHistoryPerProject Retrieves consumption metrics for Scale and Business plan projects. History begins at the time of upgrade.
// Available for Scale and Business plan users only.
// Issuing a call to this API does not wake a project's compute endpoint.
func (c Client) GetConsumptionHistoryPerProject(cursor *string, limit *int, projectIDs []string, from time.Time, to time.Time, granularity ConsumptionHistoryGranularity, orgID *string, includeV1Metrics *bool) (GetConsumptionHistoryPerProjectRespObj, error) {
	var (
		queryElements []string
		query         string
	)
	queryElements = append(queryElements, "from="+from.Format(time.RFC3339))
	queryElements = append(queryElements, "to="+to.Format(time.RFC3339))
	queryElements = append(queryElements, "granularity="+string(granularity))
	if cursor != nil {
		queryElements = append(queryElements, "cursor="+*cursor)
	}
	if limit != nil {
		queryElements = append(queryElements, "limit="+strconv.FormatInt(int64(*limit), 10))
	}
	if len(projectIDs) > 0 {
		queryElements = append(queryElements, "project_ids="+strings.Join(projectIDs, ","))
	}
	if orgID != nil {
		queryElements = append(queryElements, "org_id="+*orgID)
	}
	if includeV1Metrics != nil {
		queryElements = append(queryElements, "include_v1_metrics="+func(includeV1Metrics bool) string {
			if includeV1Metrics {
				return "true"
			}
			return "false"
		}(*includeV1Metrics))
	}
	if len(queryElements) > 0 {
		query = "?" + strings.Join(queryElements, "&")
	}
	var v GetConsumptionHistoryPerProjectRespObj
	if err := c.requestHandler(c.baseURL+"/consumption_history/projects"+query, "GET", nil, &v); err != nil {
		return GetConsumptionHistoryPerProjectRespObj{}, err
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

// GetCurrentUserOrganizations Retrieves information about the current Neon user's organizations
func (c Client) GetCurrentUserOrganizations() (OrganizationsResponse, error) {
	var v OrganizationsResponse
	if err := c.requestHandler(c.baseURL+"/users/me/organizations", "GET", nil, &v); err != nil {
		return OrganizationsResponse{}, err
	}
	return v, nil
}

// GetOrganization Retrieves information about the specified organization.
func (c Client) GetOrganization(orgID string) (Organization, error) {
	var v Organization
	if err := c.requestHandler(c.baseURL+"/organizations/"+orgID, "GET", nil, &v); err != nil {
		return Organization{}, err
	}
	return v, nil
}

// GetOrganizationInvitations Retrieves information about extended invitations for the specified organization
func (c Client) GetOrganizationInvitations(orgID string) (OrganizationInvitationsResponse, error) {
	var v OrganizationInvitationsResponse
	if err := c.requestHandler(c.baseURL+"/organizations/"+orgID+"/invitations", "GET", nil, &v); err != nil {
		return OrganizationInvitationsResponse{}, err
	}
	return v, nil
}

// GetOrganizationMember Retrieves information about the specified organization member.
func (c Client) GetOrganizationMember(orgID string, memberID string) (Member, error) {
	var v Member
	if err := c.requestHandler(c.baseURL+"/organizations/"+orgID+"/members/"+memberID, "GET", nil, &v); err != nil {
		return Member{}, err
	}
	return v, nil
}

// GetOrganizationMembers Retrieves information about the specified organization members.
func (c Client) GetOrganizationMembers(orgID string) (OrganizationMembersResponse, error) {
	var v OrganizationMembersResponse
	if err := c.requestHandler(c.baseURL+"/organizations/"+orgID+"/members", "GET", nil, &v); err != nil {
		return OrganizationMembersResponse{}, err
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

// GetProjectBranch Retrieves information about the specified branch.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain a `branch_id` by listing the project's branches.
// A `branch_id` value has a `br-` prefix.
// Each Neon project is initially created with a root and default branch named `main`.
// A project can contain one or more branches.
// A parent branch is identified by a `parent_id` value, which is the `id` of the parent branch.
// For related information, see [Manage branches](https://neon.tech/docs/manage/branches/).
func (c Client) GetProjectBranch(projectID string, branchID string) (GetProjectBranchRespObj, error) {
	var v GetProjectBranchRespObj
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID, "GET", nil, &v); err != nil {
		return GetProjectBranchRespObj{}, err
	}
	return v, nil
}

// GetProjectBranchDatabase Retrieves information about the specified database.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain the `branch_id` and `database_name` by listing the branch's databases.
// For related information, see [Manage databases](https://neon.tech/docs/manage/databases/).
func (c Client) GetProjectBranchDatabase(projectID string, branchID string, databaseName string) (DatabaseResponse, error) {
	var v DatabaseResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/databases/"+databaseName, "GET", nil, &v); err != nil {
		return DatabaseResponse{}, err
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

// GetProjectBranchRolePassword Retrieves the password for the specified Postgres role, if possible.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain the `branch_id` by listing the project's branches.
// You can obtain the `role_name` by listing the roles for a branch.
// For related information, see [Manage roles](https://neon.tech/docs/manage/roles/).
func (c Client) GetProjectBranchRolePassword(projectID string, branchID string, roleName string) (RolePasswordResponse, error) {
	var v RolePasswordResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/roles/"+roleName+"/reveal_password", "GET", nil, &v); err != nil {
		return RolePasswordResponse{}, err
	}
	return v, nil
}

// GetProjectBranchSchema Retrieves the schema from the specified database. The `lsn` and `timestamp` values cannot be specified at the same time. If both are omitted, the database schema is retrieved from database's head.
func (c Client) GetProjectBranchSchema(projectID string, branchID string, dbName string, lsn *string, timestamp *time.Time) (BranchSchemaResponse, error) {
	var (
		queryElements []string
		query         string
	)
	queryElements = append(queryElements, "db_name="+dbName)
	if lsn != nil {
		queryElements = append(queryElements, "lsn="+*lsn)
	}
	if timestamp != nil {
		queryElements = append(queryElements, "timestamp="+timestamp.Format(time.RFC3339))
	}
	if len(queryElements) > 0 {
		query = "?" + strings.Join(queryElements, "&")
	}
	var v BranchSchemaResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/schema"+query, "GET", nil, &v); err != nil {
		return BranchSchemaResponse{}, err
	}
	return v, nil
}

// GetProjectEndpoint Retrieves information about the specified compute endpoint.
// A compute endpoint is a Neon compute instance.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain an `endpoint_id` by listing your project's compute endpoints.
// An `endpoint_id` has an `ep-` prefix.
// For information about compute endpoints, see [Manage computes](https://neon.tech/docs/manage/endpoints/).
func (c Client) GetProjectEndpoint(projectID string, endpointID string) (EndpointResponse, error) {
	var v EndpointResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints/"+endpointID, "GET", nil, &v); err != nil {
		return EndpointResponse{}, err
	}
	return v, nil
}

// GetProjectJWKS Returns all the available JWKS URLs that can be used for verifying JWTs used as the authentication mechanism for the specified project.
func (c Client) GetProjectJWKS(projectID string) (ProjectJWKSResponse, error) {
	var v ProjectJWKSResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/jwks", "GET", nil, &v); err != nil {
		return ProjectJWKSResponse{}, err
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

// GrantPermissionToProject Grants project access to the account associated with the specified email address
func (c Client) GrantPermissionToProject(projectID string, cfg GrantPermissionToProjectRequest) (ProjectPermission, error) {
	var v ProjectPermission
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/permissions", "POST", cfg, &v); err != nil {
		return ProjectPermission{}, err
	}
	return v, nil
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

// ListOrgApiKeys Retrieves the API keys for the specified organization.
// The response does not include API key tokens. A token is only provided when creating an API key.
// API keys can also be managed in the Neon Console.
// For more information, see [Manage API keys](https://neon.tech/docs/manage/api-keys/).
func (c Client) ListOrgApiKeys(orgID string) ([]OrgApiKeysListResponseItem, error) {
	var v []OrgApiKeysListResponseItem
	if err := c.requestHandler(c.baseURL+"/organizations/"+orgID+"/api_keys", "GET", nil, &v); err != nil {
		return nil, err
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

// ListProjectBranchEndpoints Retrieves a list of compute endpoints for the specified branch.
// Neon permits only one read-write compute endpoint per branch.
// A branch can have multiple read-only compute endpoints.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain the `branch_id` by listing the project's branches.
func (c Client) ListProjectBranchEndpoints(projectID string, branchID string) (EndpointsResponse, error) {
	var v EndpointsResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/endpoints", "GET", nil, &v); err != nil {
		return EndpointsResponse{}, err
	}
	return v, nil
}

// ListProjectBranchRoles Retrieves a list of Postgres roles from the specified branch.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain the `branch_id` by listing the project's branches.
// For related information, see [Manage roles](https://neon.tech/docs/manage/roles/).
func (c Client) ListProjectBranchRoles(projectID string, branchID string) (RolesResponse, error) {
	var v RolesResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/roles", "GET", nil, &v); err != nil {
		return RolesResponse{}, err
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
func (c Client) ListProjectBranches(projectID string, search *string) (ListProjectBranchesRespObj, error) {
	var (
		queryElements []string
		query         string
	)
	if search != nil {
		queryElements = append(queryElements, "search="+*search)
	}
	if len(queryElements) > 0 {
		query = "?" + strings.Join(queryElements, "&")
	}
	var v ListProjectBranchesRespObj
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches"+query, "GET", nil, &v); err != nil {
		return ListProjectBranchesRespObj{}, err
	}
	return v, nil
}

// ListProjectEndpoints Retrieves a list of compute endpoints for the specified project.
// A compute endpoint is a Neon compute instance.
// You can obtain a `project_id` by listing the projects for your Neon account.
// For information about compute endpoints, see [Manage computes](https://neon.tech/docs/manage/endpoints/).
func (c Client) ListProjectEndpoints(projectID string) (EndpointsResponse, error) {
	var v EndpointsResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints", "GET", nil, &v); err != nil {
		return EndpointsResponse{}, err
	}
	return v, nil
}

// ListProjectOperations Retrieves a list of operations for the specified Neon project.
// You can obtain a `project_id` by listing the projects for your Neon account.
// The number of operations returned can be large.
// To paginate the response, issue an initial request with a `limit` value.
// Then, add the `cursor` value that was returned in the response to the next request.
func (c Client) ListProjectOperations(projectID string, cursor *string, limit *int) (ListOperations, error) {
	var (
		queryElements []string
		query         string
	)
	if cursor != nil {
		queryElements = append(queryElements, "cursor="+*cursor)
	}
	if limit != nil {
		queryElements = append(queryElements, "limit="+strconv.FormatInt(int64(*limit), 10))
	}
	if len(queryElements) > 0 {
		query = "?" + strings.Join(queryElements, "&")
	}
	var v ListOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/operations"+query, "GET", nil, &v); err != nil {
		return ListOperations{}, err
	}
	return v, nil
}

// ListProjectPermissions Retrieves details about users who have access to the project, including the permission `id`, the granted-to email address, and the date project access was granted.
func (c Client) ListProjectPermissions(projectID string) (ProjectPermissions, error) {
	var v ProjectPermissions
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/permissions", "GET", nil, &v); err != nil {
		return ProjectPermissions{}, err
	}
	return v, nil
}

// ListProjects Retrieves a list of projects for the Neon account.
// A project is the top-level object in the Neon object hierarchy.
// For more information, see [Manage projects](https://neon.tech/docs/manage/projects/).
func (c Client) ListProjects(cursor *string, limit *int, search *string, orgID *string) (ListProjectsRespObj, error) {
	var (
		queryElements []string
		query         string
	)
	if cursor != nil {
		queryElements = append(queryElements, "cursor="+*cursor)
	}
	if limit != nil {
		queryElements = append(queryElements, "limit="+strconv.FormatInt(int64(*limit), 10))
	}
	if search != nil {
		queryElements = append(queryElements, "search="+*search)
	}
	if orgID != nil {
		queryElements = append(queryElements, "org_id="+*orgID)
	}
	if len(queryElements) > 0 {
		query = "?" + strings.Join(queryElements, "&")
	}
	var v ListProjectsRespObj
	if err := c.requestHandler(c.baseURL+"/projects"+query, "GET", nil, &v); err != nil {
		return ListProjectsRespObj{}, err
	}
	return v, nil
}

// ListSharedProjects Retrieves a list of shared projects for the Neon account.
// A project is the top-level object in the Neon object hierarchy.
// For more information, see [Manage projects](https://neon.tech/docs/manage/projects/).
func (c Client) ListSharedProjects(cursor *string, limit *int, search *string) (ListSharedProjectsRespObj, error) {
	var (
		queryElements []string
		query         string
	)
	if cursor != nil {
		queryElements = append(queryElements, "cursor="+*cursor)
	}
	if limit != nil {
		queryElements = append(queryElements, "limit="+strconv.FormatInt(int64(*limit), 10))
	}
	if search != nil {
		queryElements = append(queryElements, "search="+*search)
	}
	if len(queryElements) > 0 {
		query = "?" + strings.Join(queryElements, "&")
	}
	var v ListSharedProjectsRespObj
	if err := c.requestHandler(c.baseURL+"/projects/shared"+query, "GET", nil, &v); err != nil {
		return ListSharedProjectsRespObj{}, err
	}
	return v, nil
}

// RemoveOrganizationMember Remove member from the organization.
// Only an admin of the organization can perform this action.
// If another admin is being removed, it will not be allows in case it is the only admin left in the organization.
func (c Client) RemoveOrganizationMember(orgID string, memberID string) (EmptyResponse, error) {
	var v EmptyResponse
	if err := c.requestHandler(c.baseURL+"/organizations/"+orgID+"/members/"+memberID, "DELETE", nil, &v); err != nil {
		return EmptyResponse{}, err
	}
	return v, nil
}

// ResetProjectBranchRolePassword Resets the password for the specified Postgres role.
// Returns a new password and operations. The new password is ready to use when the last operation finishes.
// The old password remains valid until last operation finishes.
// Connections to the compute endpoint are dropped. If idle,
// the compute endpoint becomes active for a short period of time.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain the `branch_id` by listing the project's branches.
// You can obtain the `role_name` by listing the roles for a branch.
// For related information, see [Manage roles](https://neon.tech/docs/manage/roles/).
func (c Client) ResetProjectBranchRolePassword(projectID string, branchID string, roleName string) (RoleOperations, error) {
	var v RoleOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/roles/"+roleName+"/reset_password", "POST", nil, &v); err != nil {
		return RoleOperations{}, err
	}
	return v, nil
}

// RestartProjectEndpoint Restart the specified compute endpoint: suspend immediately followed by start operations.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain an `endpoint_id` by listing your project's compute endpoints.
// An `endpoint_id` has an `ep-` prefix.
// For information about compute endpoints, see [Manage computes](https://neon.tech/docs/manage/endpoints/).
func (c Client) RestartProjectEndpoint(projectID string, endpointID string) (EndpointOperations, error) {
	var v EndpointOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints/"+endpointID+"/restart", "POST", nil, &v); err != nil {
		return EndpointOperations{}, err
	}
	return v, nil
}

// RestoreProjectBranch Restores a branch to an earlier state in its own or another branch's history
func (c Client) RestoreProjectBranch(projectID string, branchID string, cfg BranchRestoreRequest) (BranchOperations, error) {
	var v BranchOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/restore", "POST", cfg, &v); err != nil {
		return BranchOperations{}, err
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

// RevokeOrgApiKey Revokes the specified organization API key.
// An API key that is no longer needed can be revoked.
// This action cannot be reversed.
// You can obtain `key_id` values by listing the API keys for an organization.
// API keys can also be managed in the Neon Console.
// See [Manage API keys](https://neon.tech/docs/manage/api-keys/).
func (c Client) RevokeOrgApiKey(orgID string, keyID int64) (OrgApiKeyRevokeResponse, error) {
	var v OrgApiKeyRevokeResponse
	if err := c.requestHandler(c.baseURL+"/organizations/"+orgID+"/api_keys/"+strconv.FormatInt(keyID, 10), "DELETE", nil, &v); err != nil {
		return OrgApiKeyRevokeResponse{}, err
	}
	return v, nil
}

// RevokePermissionFromProject Revokes project access from the user associted with the specified permisison `id`. You can retrieve a user's permission `id` by listing project access.
func (c Client) RevokePermissionFromProject(projectID string, permissionID string) (ProjectPermission, error) {
	var v ProjectPermission
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/permissions/"+permissionID, "DELETE", nil, &v); err != nil {
		return ProjectPermission{}, err
	}
	return v, nil
}

// SetDefaultProjectBranch Sets the specified branch as the project's default branch.
// The default designation is automatically removed from the previous default branch.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain the `branch_id` by listing the project's branches.
// For more information, see [Manage branches](https://neon.tech/docs/manage/branches/).
func (c Client) SetDefaultProjectBranch(projectID string, branchID string) (BranchOperations, error) {
	var v BranchOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/set_as_default", "POST", nil, &v); err != nil {
		return BranchOperations{}, err
	}
	return v, nil
}

// StartProjectEndpoint Starts a compute endpoint. The compute endpoint is ready to use
// after the last operation in chain finishes successfully.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain an `endpoint_id` by listing your project's compute endpoints.
// An `endpoint_id` has an `ep-` prefix.
// For information about compute endpoints, see [Manage computes](https://neon.tech/docs/manage/endpoints/).
func (c Client) StartProjectEndpoint(projectID string, endpointID string) (EndpointOperations, error) {
	var v EndpointOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints/"+endpointID+"/start", "POST", nil, &v); err != nil {
		return EndpointOperations{}, err
	}
	return v, nil
}

// SuspendProjectEndpoint Suspend the specified compute endpoint
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain an `endpoint_id` by listing your project's compute endpoints.
// An `endpoint_id` has an `ep-` prefix.
// For information about compute endpoints, see [Manage computes](https://neon.tech/docs/manage/endpoints/).
func (c Client) SuspendProjectEndpoint(projectID string, endpointID string) (EndpointOperations, error) {
	var v EndpointOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints/"+endpointID+"/suspend", "POST", nil, &v); err != nil {
		return EndpointOperations{}, err
	}
	return v, nil
}

// TransferProjectsFromUserToOrg Transfers selected projects, identified by their IDs, from your personal account to a specified organization.
func (c Client) TransferProjectsFromUserToOrg(cfg TransferProjectsToOrganizationRequest) (EmptyResponse, error) {
	var v EmptyResponse
	if err := c.requestHandler(c.baseURL+"/users/me/projects/transfer", "POST", cfg, &v); err != nil {
		return EmptyResponse{}, err
	}
	return v, nil
}

// UpdateOrganizationMember Only an admin can perform this action.
func (c Client) UpdateOrganizationMember(orgID string, memberID string, cfg OrganizationMemberUpdateRequest) (Member, error) {
	var v Member
	if err := c.requestHandler(c.baseURL+"/organizations/"+orgID+"/members/"+memberID, "PATCH", cfg, &v); err != nil {
		return Member{}, err
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

// UpdateProjectBranch Updates the specified branch.
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

// UpdateProjectEndpoint Updates the specified compute endpoint.
// You can obtain a `project_id` by listing the projects for your Neon account.
// You can obtain an `endpoint_id` and `branch_id` by listing your project's compute endpoints.
// An `endpoint_id` has an `ep-` prefix. A `branch_id` has a `br-` prefix.
// For more information about compute endpoints, see [Manage computes](https://neon.tech/docs/manage/endpoints/).
// If the returned list of operations is not empty, the compute endpoint is not ready to use.
// The client must wait for the last operation to finish before using the compute endpoint.
// If the compute endpoint was idle before the update, it becomes active for a short period of time,
// and the control plane suspends it again after the update.
func (c Client) UpdateProjectEndpoint(projectID string, endpointID string, cfg EndpointUpdateRequest) (EndpointOperations, error) {
	var v EndpointOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints/"+endpointID, "PATCH", cfg, &v); err != nil {
		return EndpointOperations{}, err
	}
	return v, nil
}

type ActiveRegionsResponse struct {
	// Regions The list of active regions
	Regions []RegionResponse `json:"regions"`
}

// AddProjectJWKSRequest Add a new JWKS to a specific endpoint of a project
type AddProjectJWKSRequest struct {
	// BranchID Branch ID
	BranchID *string `json:"branch_id,omitempty"`
	// JwksURL The URL that lists the JWKS
	JwksURL string `json:"jwks_url"`
	// JwtAudience The name of the required JWT Audience to be used
	JwtAudience *string `json:"jwt_audience,omitempty"`
	// ProviderName The name of the authentication provider (e.g., Clerk, Stytch, Auth0)
	ProviderName string `json:"provider_name"`
	// RoleNames The roles the JWKS should be mapped to
	RoleNames *[]string `json:"role_names,omitempty"`
}

// AllowedIps A list of IP addresses that are allowed to connect to the compute endpoint.
// If the list is empty or not set, all IP addresses are allowed.
// If protected_branches_only is true, the list will be applied only to protected branches.
type AllowedIps struct {
	// Ips A list of IP addresses that are allowed to connect to the endpoint.
	Ips *[]string `json:"ips,omitempty"`
	// ProtectedBranchesOnly If true, the list will be applied only to protected branches.
	ProtectedBranchesOnly *bool `json:"protected_branches_only,omitempty"`
}

type AnnotationCreateValueRequest struct {
	AnnotationValue *AnnotationValueData `json:"annotation_value,omitempty"`
}

type AnnotationData struct {
	CreatedAt *time.Time           `json:"created_at,omitempty"`
	Object    AnnotationObjectData `json:"object"`
	UpdatedAt *time.Time           `json:"updated_at,omitempty"`
	Value     AnnotationValueData  `json:"value"`
}

type AnnotationObjectData struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type AnnotationResponse struct {
	Annotation AnnotationData `json:"annotation"`
}

// AnnotationValueData Annotation properties.
type AnnotationValueData map[string]interface{}

type AnnotationsMapResponse struct {
	Annotations AnnotationsMapResponseAnnotations `json:"annotations"`
}

type AnnotationsMapResponseAnnotations map[string]interface{}

type ApiKeyCreateRequest struct {
	// KeyName A user-specified API key name. This value is required when creating an API key.
	KeyName string `json:"key_name"`
}

type ApiKeyCreateResponse struct {
	// CreatedAt A timestamp indicating when the API key was created
	CreatedAt time.Time `json:"created_at"`
	// CreatedBy ID of the user who created this API key
	CreatedBy string `json:"created_by"`
	// ID The API key ID
	ID int64 `json:"id"`
	// Key The generated 64-bit token required to access the Neon API
	Key string `json:"key"`
	// Name The user-specified API key name
	Name string `json:"name"`
}

// ApiKeyCreatorData The user data of the user that created this API key.
type ApiKeyCreatorData struct {
	// ID of the user who created this API key
	ID string `json:"id"`
	// Image The URL to the user's avatar image.
	Image string `json:"image"`
	// Name The name of the user.
	Name string `json:"name"`
}

type ApiKeyRevokeResponse struct {
	// CreatedAt A timestamp indicating when the API key was created
	CreatedAt time.Time `json:"created_at"`
	// CreatedBy ID of the user who created this API key
	CreatedBy string `json:"created_by"`
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
	CreatedAt time.Time         `json:"created_at"`
	CreatedBy ApiKeyCreatorData `json:"created_by"`
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
	// AddressCountry Billing address country code defined by ISO 3166-1 alpha-2.
	AddressCountry string `json:"address_country"`
	// AddressCountryName Billing address country name.
	AddressCountryName *string `json:"address_country_name,omitempty"`
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
	// Name The full name of the individual or entity that owns the billing account. This name appears on invoices.
	Name string `json:"name"`
	// OrbPortalURL Orb user portal url
	OrbPortalURL  *string              `json:"orb_portal_url,omitempty"`
	PaymentMethod BillingPaymentMethod `json:"payment_method"`
	PaymentSource PaymentSource        `json:"payment_source"`
	// QuotaResetAtLast The last time the quota was reset. Defaults to the date-time the account is created.
	QuotaResetAtLast time.Time               `json:"quota_reset_at_last"`
	State            BillingAccountState     `json:"state"`
	SubscriptionType BillingSubscriptionType `json:"subscription_type"`
	// TaxID The tax identification number for the billing account, displayed on invoices.
	TaxID *string `json:"tax_id,omitempty"`
	// TaxIDType The type of the tax identification number based on the country.
	TaxIDType *string `json:"tax_id_type,omitempty"`
}

// BillingAccountState State of the billing account.
type BillingAccountState string

const (
	BillingAccountStateUNKNOWN     BillingAccountState = "UNKNOWN"
	BillingAccountStateActive      BillingAccountState = "active"
	BillingAccountStateDeactivated BillingAccountState = "deactivated"
	BillingAccountStateDeleted     BillingAccountState = "deleted"
	BillingAccountStateSuspended   BillingAccountState = "suspended"
)

// BillingPaymentMethod Indicates whether and how an account makes payments.
type BillingPaymentMethod string

const (
	BillingPaymentMethodUNKNOWN       BillingPaymentMethod = "UNKNOWN"
	BillingPaymentMethodAwsMp         BillingPaymentMethod = "aws_mp"
	BillingPaymentMethodAzureMp       BillingPaymentMethod = "azure_mp"
	BillingPaymentMethodDirectPayment BillingPaymentMethod = "direct_payment"
	BillingPaymentMethodNone          BillingPaymentMethod = "none"
	BillingPaymentMethodSponsorship   BillingPaymentMethod = "sponsorship"
	BillingPaymentMethodStaff         BillingPaymentMethod = "staff"
	BillingPaymentMethodStripe        BillingPaymentMethod = "stripe"
	BillingPaymentMethodTrial         BillingPaymentMethod = "trial"
	BillingPaymentMethodVercelMp      BillingPaymentMethod = "vercel_mp"
)

// BillingSubscriptionType Type of subscription to Neon Cloud.
// Notice that for users without billing account this will be "UNKNOWN"
type BillingSubscriptionType string

const (
	BillingSubscriptionTypeUNKNOWN        BillingSubscriptionType = "UNKNOWN"
	BillingSubscriptionTypeAwsMarketplace BillingSubscriptionType = "aws_marketplace"
	BillingSubscriptionTypeBusiness       BillingSubscriptionType = "business"
	BillingSubscriptionTypeDirectSales    BillingSubscriptionType = "direct_sales"
	BillingSubscriptionTypeFreeV2         BillingSubscriptionType = "free_v2"
	BillingSubscriptionTypeLaunch         BillingSubscriptionType = "launch"
	BillingSubscriptionTypeScale          BillingSubscriptionType = "scale"
	BillingSubscriptionTypeVercelPgLegacy BillingSubscriptionType = "vercel_pg_legacy"
)

type Branch struct {
	ActiveTimeSeconds  int64 `json:"active_time_seconds"`
	ComputeTimeSeconds int64 `json:"compute_time_seconds"`
	// CpuUsedSec CPU seconds used by all of the branch's compute endpoints, including deleted ones.
	// This value is reset at the beginning of each billing period.
	// Examples:
	// 1. A branch that uses 1 CPU for 1 second is equal to `cpu_used_sec=1`.
	// 2. A branch that uses 2 CPUs simultaneously for 1 second is equal to `cpu_used_sec=2`.
	CpuUsedSec int64 `json:"cpu_used_sec"`
	// CreatedAt A timestamp indicating when the branch was created
	CreatedAt time.Time        `json:"created_at"`
	CreatedBy *BranchCreatedBy `json:"created_by,omitempty"`
	// CreationSource The branch creation source
	CreationSource    string      `json:"creation_source"`
	CurrentState      BranchState `json:"current_state"`
	DataTransferBytes int64       `json:"data_transfer_bytes"`
	// Default Whether the branch is the project's default branch
	Default bool `json:"default"`
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
	// Primary DEPRECATED. Use `default` field.
	// Whether the branch is the project's primary branch
	Primary *bool `json:"primary,omitempty"`
	// ProjectID The ID of the project to which the branch belongs
	ProjectID string `json:"project_id"`
	// Protected Whether the branch is protected
	Protected bool `json:"protected"`
	// StateChangedAt A UTC timestamp indicating when the `current_state` began
	StateChangedAt time.Time `json:"state_changed_at"`
	// UpdatedAt A timestamp indicating when the branch was last updated
	UpdatedAt        time.Time `json:"updated_at"`
	WrittenDataBytes int64     `json:"written_data_bytes"`
}

type BranchCreateRequest struct {
	Branch    *BranchCreateRequestBranch            `json:"branch,omitempty"`
	Endpoints *[]BranchCreateRequestEndpointOptions `json:"endpoints,omitempty"`
}

type BranchCreateRequestBranch struct {
	// Archived Whether to create the branch as archived
	Archived *bool `json:"archived,omitempty"`
	// Name The branch name
	Name *string `json:"name,omitempty"`
	// ParentID The `branch_id` of the parent branch. If omitted or empty, the branch will be created from the project's default branch.
	ParentID *string `json:"parent_id,omitempty"`
	// ParentLsn A Log Sequence Number (LSN) on the parent branch. The branch will be created with data from this LSN.
	ParentLsn *string `json:"parent_lsn,omitempty"`
	// ParentTimestamp A timestamp identifying a point in time on the parent branch. The branch will be created with data starting from this point in time.
	// The timestamp must be provided in ISO 8601 format; for example: `2024-02-26T12:00:00Z`.
	ParentTimestamp *time.Time `json:"parent_timestamp,omitempty"`
	// Protected Whether the branch is protected
	Protected *bool `json:"protected,omitempty"`
	// SchemaInitializationType The type of schema initialization. Defines how the schema is initialized, currently only empty is supported. This parameter is under
	// active development and may change its semantics in the future.
	SchemaInitializationType *string `json:"schema_initialization_type,omitempty"`
}

type BranchCreateRequestEndpointOptions struct {
	AutoscalingLimitMaxCu *ComputeUnit           `json:"autoscaling_limit_max_cu,omitempty"`
	AutoscalingLimitMinCu *ComputeUnit           `json:"autoscaling_limit_min_cu,omitempty"`
	Provisioner           *Provisioner           `json:"provisioner,omitempty"`
	SuspendTimeoutSeconds *SuspendTimeoutSeconds `json:"suspend_timeout_seconds,omitempty"`
	Type                  EndpointType           `json:"type"`
}

// BranchCreatedBy The resolved user model that contains details of the user/org/integration/api_key used for branch creation. This field is filled only in listing/get/create/get/update/delete methods, if it is empty when calling other handlers, it does not mean that it is empty in the system.
type BranchCreatedBy struct {
	// Image The URL to the user's avatar image.
	Image *string `json:"image,omitempty"`
	// Name The name of the user.
	Name *string `json:"name,omitempty"`
}

type BranchOperations struct {
	BranchResponse
	OperationsResponse
}

type BranchResponse struct {
	Branch Branch `json:"branch"`
}

type BranchRestoreRequest struct {
	// PreserveUnderName If not empty, the previous state of the branch will be saved to a branch with this name.
	// If the branch has children or the `source_branch_id` is equal to the branch id, this field is required. All existing child branches will be moved to the newly created branch under the name `preserve_under_name`.
	PreserveUnderName *string `json:"preserve_under_name,omitempty"`
	// SourceBranchID The `branch_id` of the restore source branch.
	// If `source_timestamp` and `source_lsn` are omitted, the branch will be restored to head.
	// If `source_branch_id` is equal to the branch's id, `source_timestamp` or `source_lsn` is required.
	SourceBranchID string `json:"source_branch_id"`
	// SourceLsn A Log Sequence Number (LSN) on the source branch. The branch will be restored with data from this LSN.
	SourceLsn *string `json:"source_lsn,omitempty"`
	// SourceTimestamp A timestamp identifying a point in time on the source branch. The branch will be restored with data starting from this point in time.
	// The timestamp must be provided in ISO 8601 format; for example: `2024-02-26T12:00:00Z`.
	SourceTimestamp *time.Time `json:"source_timestamp,omitempty"`
}

type BranchSchemaResponse struct {
	Sql *string `json:"sql,omitempty"`
}

// BranchState The branch’s state, indicating if it is initializing, ready for use, or archived.
//   - 'init' - the branch is being created but is not available for querying.
//   - 'ready' - the branch is fully operational and ready for querying. Expect normal query response times.
//   - 'archived' - the branch is stored in cost-effective archival storage. Expect slow query response times.
type BranchState string

type BranchUpdateRequest struct {
	Branch BranchUpdateRequestBranch `json:"branch"`
}

type BranchUpdateRequestBranch struct {
	Name      *string `json:"name,omitempty"`
	Protected *bool   `json:"protected,omitempty"`
}

type BranchesResponse struct {
	Branches []Branch `json:"branches"`
}

type ComputeUnit float64

type ConnectionDetails struct {
	ConnectionParameters ConnectionParameters `json:"connection_parameters"`
	// ConnectionURI The connection URI is defined as specified here: [Connection URIs](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING-URIS)
	// The connection URI can be used to connect to a Postgres database with psql or defined in a DATABASE_URL environment variable.
	// When creating a branch from a parent with more than one role or database, the response body does not include a connection URI.
	ConnectionURI string `json:"connection_uri"`
}

type ConnectionParameters struct {
	// Database name
	Database string `json:"database"`
	// Host Hostname
	Host string `json:"host"`
	// Password for the role
	Password string `json:"password"`
	// PoolerHost Pooler hostname
	PoolerHost string `json:"pooler_host"`
	// Role name
	Role string `json:"role"`
}

type ConnectionURIResponse struct {
	// URI The connection URI.
	URI string `json:"uri"`
}

type ConnectionURIsOptionalResponse struct {
	ConnectionURIs *[]ConnectionDetails `json:"connection_uris,omitempty"`
}

type ConnectionURIsResponse struct {
	ConnectionURIs []ConnectionDetails `json:"connection_uris"`
}

type ConsumptionHistoryGranularity string

const (
	ConsumptionHistoryGranularityDaily   ConsumptionHistoryGranularity = "daily"
	ConsumptionHistoryGranularityHourly  ConsumptionHistoryGranularity = "hourly"
	ConsumptionHistoryGranularityMonthly ConsumptionHistoryGranularity = "monthly"
)

type ConsumptionHistoryPerAccountResponse struct {
	Periods []ConsumptionHistoryPerPeriod `json:"periods"`
}

type ConsumptionHistoryPerPeriod struct {
	Consumption []ConsumptionHistoryPerTimeframe `json:"consumption"`
	// PeriodEnd The end date-time of the billing period, available for the past periods only.
	PeriodEnd *time.Time `json:"period_end,omitempty"`
	// PeriodID The ID assigned to the specified billing period.
	PeriodID string `json:"period_id"`
	// PeriodPlan The billing plan applicable during the billing period.
	PeriodPlan string `json:"period_plan"`
	// PeriodStart The start date-time of the billing period.
	PeriodStart time.Time `json:"period_start"`
}

type ConsumptionHistoryPerProject struct {
	Periods []ConsumptionHistoryPerPeriod `json:"periods"`
	// ProjectID The project ID
	ProjectID string `json:"project_id"`
}

type ConsumptionHistoryPerProjectResponse struct {
	Projects []ConsumptionHistoryPerProject `json:"projects"`
}

type ConsumptionHistoryPerTimeframe struct {
	// ActiveTimeSeconds Seconds. The amount of time the compute endpoints have been active.
	ActiveTimeSeconds int `json:"active_time_seconds"`
	// ComputeTimeSeconds Seconds. The number of CPU seconds used by compute endpoints, including compute endpoints that have been deleted.
	ComputeTimeSeconds int `json:"compute_time_seconds"`
	// DataStorageBytesHour Bytes-Hour. The amount of storage consumed hourly.
	DataStorageBytesHour *int `json:"data_storage_bytes_hour,omitempty"`
	// SyntheticStorageSizeBytes Bytes. The space occupied in storage. Synthetic storage size combines the logical data size and Write-Ahead Log (WAL) size for all branches.
	SyntheticStorageSizeBytes int `json:"synthetic_storage_size_bytes"`
	// TimeframeEnd The specified end date-time for the reported consumption.
	TimeframeEnd time.Time `json:"timeframe_end"`
	// TimeframeStart The specified start date-time for the reported consumption.
	TimeframeStart time.Time `json:"timeframe_start"`
	// WrittenDataBytes Bytes. The amount of written data for all branches.
	WrittenDataBytes int `json:"written_data_bytes"`
}

type CreateProjectBranchReqObj struct {
	AnnotationCreateValueRequest
	BranchCreateRequest
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
	Email string `json:"email"`
	Image string `json:"image"`
	// Login DEPRECATED. Use `email` field.
	Login    string             `json:"login"`
	Name     string             `json:"name"`
	Provider IdentityProviderId `json:"provider"`
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
	// Login DEPRECATED. Use `email` field.
	Login               string      `json:"login"`
	MaxAutoscalingLimit ComputeUnit `json:"max_autoscaling_limit"`
	Name                string      `json:"name"`
	Plan                string      `json:"plan"`
	ProjectsLimit       int64       `json:"projects_limit"`
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

// EmptyResponse Empty response.
type EmptyResponse map[string]interface{}

type Endpoint struct {
	AutoscalingLimitMaxCu ComputeUnit `json:"autoscaling_limit_max_cu"`
	AutoscalingLimitMinCu ComputeUnit `json:"autoscaling_limit_min_cu"`
	// BranchID The ID of the branch that the compute endpoint is associated with
	BranchID string `json:"branch_id"`
	// ComputeReleaseVersion Attached compute's release version number.
	ComputeReleaseVersion *string `json:"compute_release_version,omitempty"`
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

const (
	EndpointPoolerModeTransaction EndpointPoolerMode = "transaction"
)

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

const (
	EndpointStateActive EndpointState = "active"
	EndpointStateIdle   EndpointState = "idle"
	EndpointStateInit   EndpointState = "init"
)

// EndpointType The compute endpoint type. Either `read_write` or `read_only`.
type EndpointType string

const (
	EndpointTypeReadOnly  EndpointType = "read_only"
	EndpointTypeReadWrite EndpointType = "read_write"
)

type EndpointUpdateRequest struct {
	Endpoint EndpointUpdateRequestEndpoint `json:"endpoint"`
}

type EndpointUpdateRequestEndpoint struct {
	AutoscalingLimitMaxCu *ComputeUnit `json:"autoscaling_limit_max_cu,omitempty"`
	AutoscalingLimitMinCu *ComputeUnit `json:"autoscaling_limit_min_cu,omitempty"`
	// BranchID DEPRECATED: This field will be removed in a future release.
	// The destination branch ID. The destination branch must not have an exsiting read-write endpoint.
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

type GetConsumptionHistoryPerProjectRespObj struct {
	ConsumptionHistoryPerProjectResponse
	PaginationResponse
}

type GetProjectBranchRespObj struct {
	AnnotationResponse
	BranchResponse
}

type GrantPermissionToProjectRequest struct {
	Email string `json:"email"`
}

// IdentityProviderId Identity provider id from keycloak
type IdentityProviderId string

const (
	IdentityProviderIdGithub    IdentityProviderId = "github"
	IdentityProviderIdGoogle    IdentityProviderId = "google"
	IdentityProviderIdHasura    IdentityProviderId = "hasura"
	IdentityProviderIdKeycloak  IdentityProviderId = "keycloak"
	IdentityProviderIdMicrosoft IdentityProviderId = "microsoft"
	IdentityProviderIdTest      IdentityProviderId = "test"
	IdentityProviderIdVercelmp  IdentityProviderId = "vercelmp"
)

type Invitation struct {
	// Email of the invited user
	Email string `json:"email"`
	ID    string `json:"id"`
	// InvitedAt Timestamp when the invitation was created
	InvitedAt time.Time `json:"invited_at"`
	// InvitedBy UUID for the user_id who extended the invitation
	InvitedBy string `json:"invited_by"`
	// OrgID Organization id as it is stored in Neon
	OrgID string     `json:"org_id"`
	Role  MemberRole `json:"role"`
}

type JWKS struct {
	// BranchID Branch ID
	BranchID *string `json:"branch_id,omitempty"`
	// CreatedAt The date and time when the JWKS was created
	CreatedAt time.Time `json:"created_at"`
	// ID JWKS ID
	ID string `json:"id"`
	// JwksURL The URL that lists the JWKS
	JwksURL string `json:"jwks_url"`
	// JwtAudience The name of the required JWT Audience to be used
	JwtAudience *string `json:"jwt_audience,omitempty"`
	// ProjectID Project ID
	ProjectID string `json:"project_id"`
	// ProviderName The name of the authentication provider (e.g., Clerk, Stytch, Auth0)
	ProviderName string `json:"provider_name"`
	// UpdatedAt The date and time when the JWKS was last modified
	UpdatedAt time.Time `json:"updated_at"`
}

type JWKSCreationOperation struct {
	JWKSResponse
	OperationsResponse
}

type JWKSResponse struct {
	Jwks JWKS `json:"jwks"`
}

type ListOperations struct {
	OperationsResponse
	PaginationResponse
}

type ListProjectBranchesRespObj struct {
	AnnotationsMapResponse
	BranchesResponse
}

type ListProjectsRespObj struct {
	PaginationResponse
	ProjectsApplicationsMapResponse
	ProjectsIntegrationsMapResponse
	ProjectsResponse
}

type ListSharedProjectsRespObj struct {
	PaginationResponse
	ProjectsResponse
}

// MaintenanceWindow A maintenance window is a time period during which Neon may perform maintenance on the project's infrastructure.
// During this time, the project's compute endpoints may be unavailable and existing connections can be
// interrupted.
type MaintenanceWindow struct {
	// EndTime End time of the maintenance window, in the format of "HH:MM". Uses UTC.
	EndTime string `json:"end_time"`
	// StartTime Start time of the maintenance window, in the format of "HH:MM". Uses UTC.
	StartTime string `json:"start_time"`
	// Weekdays A list of weekdays when the maintenance window is active.
	// Encoded as ints, where 1 - Monday, and 7 - Sunday.
	Weekdays []int `json:"weekdays"`
}

type Member struct {
	ID       string     `json:"id"`
	JoinedAt *time.Time `json:"joined_at,omitempty"`
	OrgID    string     `json:"org_id"`
	Role     MemberRole `json:"role"`
	UserID   string     `json:"user_id"`
}

// MemberRole The role of the organization member
type MemberRole string

const (
	MemberRoleAdmin  MemberRole = "admin"
	MemberRoleMember MemberRole = "member"
)

type MemberUserInfo struct {
	Email string `json:"email"`
}

type MemberWithUser struct {
	Member Member         `json:"member"`
	User   MemberUserInfo `json:"user"`
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

const (
	OperationActionApplyConfig                OperationAction = "apply_config"
	OperationActionApplyStorageConfig         OperationAction = "apply_storage_config"
	OperationActionCheckAvailability          OperationAction = "check_availability"
	OperationActionCreateBranch               OperationAction = "create_branch"
	OperationActionCreateCompute              OperationAction = "create_compute"
	OperationActionCreateTimeline             OperationAction = "create_timeline"
	OperationActionDeleteTimeline             OperationAction = "delete_timeline"
	OperationActionDetachParentBranch         OperationAction = "detach_parent_branch"
	OperationActionDisableMaintenance         OperationAction = "disable_maintenance"
	OperationActionPrepareSecondaryPageserver OperationAction = "prepare_secondary_pageserver"
	OperationActionReplaceSafekeeper          OperationAction = "replace_safekeeper"
	OperationActionStartCompute               OperationAction = "start_compute"
	OperationActionStartReservedCompute       OperationAction = "start_reserved_compute"
	OperationActionSuspendCompute             OperationAction = "suspend_compute"
	OperationActionSwitchPageserver           OperationAction = "switch_pageserver"
	OperationActionSyncDbsAndRolesFromCompute OperationAction = "sync_dbs_and_roles_from_compute"
	OperationActionTenantAttach               OperationAction = "tenant_attach"
	OperationActionTenantDetach               OperationAction = "tenant_detach"
	OperationActionTenantIgnore               OperationAction = "tenant_ignore"
	OperationActionTenantReattach             OperationAction = "tenant_reattach"
	OperationActionTimelineArchive            OperationAction = "timeline_archive"
	OperationActionTimelineUnarchive          OperationAction = "timeline_unarchive"
)

type OperationResponse struct {
	Operation Operation `json:"operation"`
}

// OperationStatus The status of the operation
type OperationStatus string

const (
	OperationStatusCancelled  OperationStatus = "cancelled"
	OperationStatusCancelling OperationStatus = "cancelling"
	OperationStatusError      OperationStatus = "error"
	OperationStatusFailed     OperationStatus = "failed"
	OperationStatusFinished   OperationStatus = "finished"
	OperationStatusRunning    OperationStatus = "running"
	OperationStatusScheduling OperationStatus = "scheduling"
	OperationStatusSkipped    OperationStatus = "skipped"
)

type OperationsResponse struct {
	Operations []Operation `json:"operations"`
}

type OrgApiKeyCreateRequest struct {
	ApiKeyCreateRequest
}

type OrgApiKeyCreateResponse struct {
	ApiKeyCreateResponse
}

type OrgApiKeyRevokeResponse struct {
	ApiKeyRevokeResponse
}

type OrgApiKeysListResponseItem struct {
	ApiKeysListResponseItem
}

type Organization struct {
	// CreatedAt A timestamp indicting when the organization was created
	CreatedAt time.Time `json:"created_at"`
	Handle    string    `json:"handle"`
	ID        string    `json:"id"`
	// ManagedBy Organizations created via the Console or the API are managed by `console`.
	// Organizations created by other methods can't be deleted via the Console or the API.
	ManagedBy string `json:"managed_by"`
	Name      string `json:"name"`
	Plan      string `json:"plan"`
	// UpdatedAt A timestamp indicating when the organization was updated
	UpdatedAt time.Time `json:"updated_at"`
}

type OrganizationInvitationsResponse struct {
	Invitations []Invitation `json:"invitations"`
}

type OrganizationInviteCreateRequest struct {
	Email string     `json:"email"`
	Role  MemberRole `json:"role"`
}

type OrganizationInvitesCreateRequest struct {
	Invitations []OrganizationInviteCreateRequest `json:"invitations"`
}

type OrganizationMemberUpdateRequest struct {
	Role MemberRole `json:"role"`
}

type OrganizationMembersResponse struct {
	Members []MemberWithUser `json:"members"`
}

type OrganizationsResponse struct {
	Organizations []Organization `json:"organizations"`
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

// PgSettingsData A raw representation of Postgres settings
type PgSettingsData map[string]interface{}

// PgVersion The major Postgres version number. Currently supported versions are `14`, `15`, `16`, and `17`.
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
	// HistoryRetentionSeconds The number of seconds to retain the shared history for all branches in this project. The default for all plans is 1 day (86400 seconds).
	HistoryRetentionSeconds int32 `json:"history_retention_seconds"`
	// ID The project ID
	ID string `json:"id"`
	// MaintenanceStartsAt A timestamp indicating when project maintenance begins. If set, the project is placed into maintenance mode at this time.
	MaintenanceStartsAt *time.Time `json:"maintenance_starts_at,omitempty"`
	// Name The project name
	Name      string            `json:"name"`
	OrgID     *string           `json:"org_id,omitempty"`
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

type ProjectCreateRequest struct {
	Project ProjectCreateRequestProject `json:"project"`
}

type ProjectCreateRequestProject struct {
	AutoscalingLimitMaxCu   *ComputeUnit                       `json:"autoscaling_limit_max_cu,omitempty"`
	AutoscalingLimitMinCu   *ComputeUnit                       `json:"autoscaling_limit_min_cu,omitempty"`
	Branch                  *ProjectCreateRequestProjectBranch `json:"branch,omitempty"`
	DefaultEndpointSettings *DefaultEndpointSettings           `json:"default_endpoint_settings,omitempty"`
	// HistoryRetentionSeconds The number of seconds to retain the shared history for all branches in this project.
	// The default is 1 day (86400 seconds).
	HistoryRetentionSeconds *int32 `json:"history_retention_seconds,omitempty"`
	// Name The project name
	Name *string `json:"name,omitempty"`
	// OrgID Organization id in case the project created belongs to an organization.
	// If not present, project is owned by a user and not by org.
	OrgID       *string      `json:"org_id,omitempty"`
	PgVersion   *PgVersion   `json:"pg_version,omitempty"`
	Provisioner *Provisioner `json:"provisioner,omitempty"`
	// RegionID The region identifier. Refer to our [Regions](https://neon.tech/docs/introduction/regions) documentation for supported regions. Values are specified in this format: `aws-us-east-1`
	RegionID *string              `json:"region_id,omitempty"`
	Settings *ProjectSettingsData `json:"settings,omitempty"`
	// StorePasswords Whether or not passwords are stored for roles in the Neon project. Storing passwords facilitates access to Neon features that require authorization.
	StorePasswords *bool `json:"store_passwords,omitempty"`
}

type ProjectCreateRequestProjectBranch struct {
	// DatabaseName The database name. If not specified, the default database name, `neondb`, will be used.
	DatabaseName *string `json:"database_name,omitempty"`
	// Name The default branch name. If not specified, the default branch name, `main`, will be used.
	Name *string `json:"name,omitempty"`
	// RoleName The role name. If not specified, the default role name, `{database_name}_owner`, will be used.
	RoleName *string `json:"role_name,omitempty"`
}

// ProjectJWKSResponse The list of configured JWKS definitions for a project
type ProjectJWKSResponse struct {
	Jwks []JWKS `json:"jwks"`
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
	Name string `json:"name"`
	// OrgID Organization id if a project belongs to organization.
	// Permissions for the project will be given to organization members as defined by the organization admins.
	// The permissions of the project do not depend on the user that created the project if a project belongs to an organization.
	OrgID     *string   `json:"org_id,omitempty"`
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
	Name             string                  `json:"name"`
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
	// BlockPublicConnections When set, connections from the public internet
	// are disallowed. This supersedes the AllowedIPs list.
	// (IN DEVELOPMENT - NOT AVAILABLE YET)
	BlockPublicConnections *bool `json:"block_public_connections,omitempty"`
	// BlockVpcConnections When set, connections using VPC endpoints
	// are disallowed.
	// (IN DEVELOPMENT - NOT AVAILABLE YET)
	BlockVpcConnections *bool `json:"block_vpc_connections,omitempty"`
	// EnableLogicalReplication Sets wal_level=logical for all compute endpoints in this project.
	// All active endpoints will be suspended.
	// Once enabled, logical replication cannot be disabled.
	EnableLogicalReplication *bool              `json:"enable_logical_replication,omitempty"`
	MaintenanceWindow        *MaintenanceWindow `json:"maintenance_window,omitempty"`
	Quota                    *ProjectQuota      `json:"quota,omitempty"`
}

type ProjectUpdateRequest struct {
	Project ProjectUpdateRequestProject `json:"project"`
}

type ProjectUpdateRequestProject struct {
	DefaultEndpointSettings *DefaultEndpointSettings `json:"default_endpoint_settings,omitempty"`
	// HistoryRetentionSeconds The number of seconds to retain the shared history for all branches in this project.
	// The default is 1 day (604800 seconds).
	HistoryRetentionSeconds *int32 `json:"history_retention_seconds,omitempty"`
	// Name The project name
	Name     *string              `json:"name,omitempty"`
	Settings *ProjectSettingsData `json:"settings,omitempty"`
}

// ProjectsApplicationsMapResponse A map where key is a project ID and a value is a list of installed applications.
type ProjectsApplicationsMapResponse struct {
	Applications ProjectsApplicationsMapResponseApplications `json:"applications"`
}

type ProjectsApplicationsMapResponseApplications map[string]interface{}

// ProjectsIntegrationsMapResponse A map where key is a project ID and a value is a list of installed integrations.
type ProjectsIntegrationsMapResponse struct {
	Integrations ProjectsIntegrationsMapResponseIntegrations `json:"integrations"`
}

type ProjectsIntegrationsMapResponseIntegrations map[string]interface{}

type ProjectsResponse struct {
	Projects []ProjectListItem `json:"projects"`
}

// Provisioner The Neon compute provisioner.
// Specify the `k8s-neonvm` provisioner to create a compute endpoint that supports Autoscaling.
//
// Provisioner can be one of the following values:
// * k8s-pod
// * k8s-neonvm
//
// Clients must expect, that any string value that is not documented in the description above should be treated as a error. UNKNOWN value if safe to treat as an error too.
type Provisioner string

type RegionResponse struct {
	// Default Whether this region is used by default in new projects.
	Default bool `json:"default"`
	// GeoLat The geographical latitude (approximate) for the region. Empty if unknown.
	GeoLat string `json:"geo_lat"`
	// GeoLong The geographical longitude (approximate) for the region. Empty if unknown.
	GeoLong string `json:"geo_long"`
	// Name A short description of the region.
	Name string `json:"name"`
	// RegionID The region ID as used in other API endpoints
	RegionID string `json:"region_id"`
}

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
// The minimum value is `60` seconds (1 minute).
// The maximum value is `604800` seconds (1 week). For more information, see
// [Auto-suspend configuration](https://neon.tech/docs/manage/endpoints#auto-suspend-configuration).
type SuspendTimeoutSeconds int64

type TransferProjectsToOrganizationRequest struct {
	OrgID string `json:"org_id"`
	// ProjectIDs The list of projects ids to transfer. Maximum of 400 project ids
	ProjectIDs []string `json:"project_ids"`
}

type UpdateProjectRespObj struct {
	OperationsResponse
	ProjectResponse
}
