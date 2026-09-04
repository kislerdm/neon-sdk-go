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
    if _, ok := (cfg.HTTPClient).(MockHTTPClient); !ok && cfg.Key == "" {
		return nil, errors.New(
			"authorization key must be provided: https://neon.tech/docs/reference/api-reference/#authentication",
		)
	}

	c := &Client{
        baseURL: baseURL,
        cfg: cfg,
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


// AcceptProjectTransferRequest Accepts a transfer request for the specified project, transferring it to the specified organization
// or user. If org_id is not passed, the project will be transferred to the current user or organization account.
func (c Client) AcceptProjectTransferRequest(projectID string, requestID string, cfg *AcceptProjectTransferRequestReqObj) error {
return c.requestHandler(c.baseURL+"/projects/"+projectID+"/transfer_requests/"+requestID, "PUT", cfg, nil)
}

// AddBranchNeonAuthOauthProvider Adds an OAuth provider configuration to the specified branch's Neon Auth integration.
// After adding, users can authenticate using the configured provider.
func (c Client) AddBranchNeonAuthOauthProvider(projectID string, branchID string, cfg NeonAuthAddOAuthProviderRequest) (NeonAuthOauthProvider, error) {
	var v NeonAuthOauthProvider
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/auth/oauth_providers", "POST", cfg, &v); err != nil {
		return NeonAuthOauthProvider{}, err
	}
	return v, nil
}

// AddBranchNeonAuthTrustedDomain Adds a domain to the redirect URI whitelist for the specified branch.
// Only domains in this list are permitted as redirect targets after authentication.
func (c Client) AddBranchNeonAuthTrustedDomain(projectID string, branchID string, cfg NeonAuthAddDomainToRedirectURIWhitelistRequest) error {
return c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/auth/domains", "POST", cfg, nil)
}

// AddNeonAuthDomainToRedirectURIWhitelist DEPRECATED, use `/projects/{project_id}/branches/{branch_id}/auth/domains` instead. Adds a domain to the redirect_uri whitelist for the specified project.
func (c Client) AddNeonAuthDomainToRedirectURIWhitelist(projectID string, cfg NeonAuthAddDomainToRedirectURIWhitelistRequest) error {
return c.requestHandler(c.baseURL+"/projects/"+projectID+"/auth/domains", "POST", cfg, nil)
}

// AddNeonAuthOauthProvider DEPRECATED, use `/projects/{project_id}/branches/{branch_id}/auth/oauth_providers` instead.
// Adds an OAuth provider to the specified project.
func (c Client) AddNeonAuthOauthProvider(projectID string, cfg NeonAuthAddOAuthProviderRequest) (NeonAuthOauthProvider, error) {
	var v NeonAuthOauthProvider
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/auth/oauth_providers", "POST", cfg, &v); err != nil {
		return NeonAuthOauthProvider{}, err
	}
	return v, nil
}

// AddProjectJWKS Adds a JWKS URL to the specified project for verifying JWTs used as the authentication mechanism.
// The URL must be a valid HTTPS URL that returns a JSON Web Key Set.
// The `provider_name` field allows you to specify which authentication provider you're using (e.g., Clerk, Auth0, AWS Cognito).
// The `branch_id` scopes the JWKS URL to specific branches; if not specified, it applies to all branches.
// The `role_names` scopes the URL to specific roles; if not specified, default roles are used (`authenticator`, `authenticated`, `anonymous`).
// The `jwt_audience` specifies which `aud` values are accepted in JWTs.
func (c Client) AddProjectJWKS(projectID string, cfg AddProjectJWKSRequest) (JWKSCreationOperation, error) {
	var
	v JWKSCreationOperation
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/jwks", "POST", cfg, &v); err != nil {
		return JWKSCreationOperation{}, err
	}
	return v, nil
}

// AssignOrganizationVPCEndpoint Assigns a VPC endpoint to a Neon organization or updates its existing assignment.
func (c Client) AssignOrganizationVPCEndpoint(orgID string, regionID string, vpcEndpointID string, cfg VPCEndpointAssignment) error {
return c.requestHandler(c.baseURL+"/organizations/"+orgID+"/vpc/region/"+regionID+"/vpc_endpoints/"+vpcEndpointID, "POST", cfg, nil)
}

// AssignProjectVPCEndpoint Sets or updates a VPC endpoint restriction for a Neon project.
// When a VPC endpoint restriction is set, the project only accepts connections
// from the specified VPC.
// A VPC endpoint can be set as a restriction only after it is assigned to the
// parent organization of the Neon project.
func (c Client) AssignProjectVPCEndpoint(projectID string, vpcEndpointID string, cfg VPCEndpointAssignment) error {
return c.requestHandler(c.baseURL+"/projects/"+projectID+"/vpc_endpoints/"+vpcEndpointID, "POST", cfg, nil)
}

// CountProjectBranches Retrieves the total number of branches in the specified project.
// Supports an optional `search` parameter to count branches matching a name filter.
func (c Client) CountProjectBranches(projectID string, search *string) (CountProjectBranchesRespObj, error) {
	var (
		queryElements []string
		query string
	)
	if search != nil {
		queryElements = append(queryElements, "search="+*search)
	}
	if len(queryElements) > 0 {
		query = "?" + strings.Join(queryElements, "&")
	}
	var v CountProjectBranchesRespObj
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/count" + query, "GET", nil, &v); err != nil {
		return CountProjectBranchesRespObj{}, err
	}
	return v, nil
}

// CreateApiKey Creates an API key.
// The `key_name` is a user-specified name for the key.
// Returns an `id` and `key`; the `key` is a randomly generated, 64-bit token required to access the Neon API.
// Store the key securely — it is only returned once.
// API keys can also be managed in the Neon Console.
// See [Manage API keys](https://neon.com/docs/manage/api-keys/).
func (c Client) CreateApiKey(cfg ApiKeyCreateRequest) (ApiKeyCreateResponse, error) {
	var v ApiKeyCreateResponse
	if err := c.requestHandler(c.baseURL+"/api_keys", "POST", cfg, &v); err != nil {
		return ApiKeyCreateResponse{}, err
	}
	return v, nil
}

// CreateBranchNeonAuthNewUser Creates a new user in the Neon Auth user directory for the specified branch.
// The user is created in the `neon_auth.users_sync` table and can immediately authenticate
// using the branch's configured auth providers.
func (c Client) CreateBranchNeonAuthNewUser(projectID string, branchID string, cfg CreateBranchNeonAuthNewUserRequest) (NeonAuthCreateNewUserResponse, error) {
	var v NeonAuthCreateNewUserResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/auth/users", "POST", cfg, &v); err != nil {
		return NeonAuthCreateNewUserResponse{}, err
	}
	return v, nil
}

// CreateCredential Issues a new scoped service credential anchored to the specified
// branch. The response carries `api_token` and `s3_secret_access_key`
// exactly once — they are not stored server-side.
// **Note**: This endpoint is currently in Private Beta.
func (c Client) CreateCredential(projectID string, branchID string, cfg CreateCredentialRequest) (CreateCredentialResponse, error) {
	var v CreateCredentialResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/credentials", "POST", cfg, &v); err != nil {
		return CreateCredentialResponse{}, err
	}
	return v, nil
}

// CreateNeonAuth Enables Neon Auth for the specified branch by connecting it to an authentication provider.
// Creating the integration provisions the `neon_auth` schema in the branch database, which stores user identity data synchronized from the provider.
func (c Client) CreateNeonAuth(projectID string, branchID string, cfg EnableNeonAuthIntegrationRequest) (NeonAuthCreateIntegrationResponse, error) {
	var v NeonAuthCreateIntegrationResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/auth", "POST", cfg, &v); err != nil {
		return NeonAuthCreateIntegrationResponse{}, err
	}
	return v, nil
}

// CreateNeonAuthIntegration DEPRECATED, use `/projects/{project_id}/branches/{branch_id}/auth` instead. Creates a project on a third-party authentication provider's platform for use with Neon Auth.
// Use this endpoint if the frontend integration flow can't be used.
func (c Client) CreateNeonAuthIntegration(cfg NeonAuthCreateIntegrationRequest) (NeonAuthCreateIntegrationResponse, error) {
	var v NeonAuthCreateIntegrationResponse
	if err := c.requestHandler(c.baseURL+"/projects/auth/create", "POST", cfg, &v); err != nil {
		return NeonAuthCreateIntegrationResponse{}, err
	}
	return v, nil
}

// CreateNeonAuthNewUser DEPRECATED, use `/projects/{project_id}/branches/{branch_id}/auth/users` instead. Creates a new user in Neon Auth.
// The user will be created in your neon_auth.users_sync table and automatically propagated to your auth project, whether Neon-managed or provider-owned.
func (c Client) CreateNeonAuthNewUser(cfg NeonAuthCreateNewUserRequest) (NeonAuthCreateNewUserResponse, error) {
	var v NeonAuthCreateNewUserResponse
	if err := c.requestHandler(c.baseURL+"/projects/auth/user", "POST", cfg, &v); err != nil {
		return NeonAuthCreateNewUserResponse{}, err
	}
	return v, nil
}

// CreateNeonAuthProviderSDKKeys Generates SDK or API Keys for the auth provider. These might be called different things depending
// on the auth provider you're using, but are generally used for setting up the frontend and backend SDKs.
func (c Client) CreateNeonAuthProviderSDKKeys(cfg NeonAuthCreateAuthProviderSDKKeysRequest) (NeonAuthCreateIntegrationResponse, error) {
	var v NeonAuthCreateIntegrationResponse
	if err := c.requestHandler(c.baseURL+"/projects/auth/keys", "POST", cfg, &v); err != nil {
		return NeonAuthCreateIntegrationResponse{}, err
	}
	return v, nil
}

// CreateOrgApiKey Creates an API key for the specified organization.
// The `key_name` is a user-specified name for the key.
// Returns an `id` and `key`; the `key` is a randomly generated, 64-bit token required to access the Neon API.
// Store the key securely — it is only returned once.
// API keys can also be managed in the Neon Console.
// See [Manage API keys](https://neon.com/docs/manage/api-keys/).
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

// CreateProject Creates a Neon project within an organization.
// If using a personal API key, include the `org_id` parameter to specify which organization to create the project in.
// If using an org API key, `org_id` is automatically inferred from the key.
// Plan limits define how many projects you can create.
// For more information, see [Manage projects](https://neon.com/docs/manage/projects/).
// You can specify a region and Postgres version in the request body.
// Neon currently supports PostgreSQL 14, 15, 16, 17, and 18.
// For supported regions and `region_id` values, see [Regions](https://neon.com/docs/introduction/regions/).
func (c Client) CreateProject(cfg ProjectCreateRequest) (CreatedProject, error) {
	var v CreatedProject
	if err := c.requestHandler(c.baseURL+"/projects", "POST", cfg, &v); err != nil {
		return CreatedProject{}, err
	}
	return v, nil
}

// CreateProjectBranch Creates a branch in the specified project.
// No request body is required, but you can specify one to create a compute endpoint or select a non-default parent branch.
// By default, the branch is created from the project's default branch with no compute endpoint, and the branch name is auto-generated.
// To access the branch, add a `read_write` endpoint.
// Each branch supports one read-write endpoint and multiple read-only endpoints.
// For related information, see [Manage branches](https://neon.com/docs/manage/branches/).
func (c Client) CreateProjectBranch(projectID string, cfg *CreateProjectBranchReqObj) (CreatedBranch, error) {
	var v CreatedBranch
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches", "POST", cfg, &v); err != nil {
		return CreatedBranch{}, err
	}
	return v, nil
}

// CreateProjectBranchAnonymized Creates a new branch with anonymized data using PostgreSQL Anonymizer for static masking.
// This allows developers to work with masked production data.
// Optionally, provide `masking_rules` to set initial masking rules for the branch
// and `start_anonymization` to automatically start anonymization after creation. This
// combines functionality of updating masking rules and starting anonymization into the
// branch creation request.
// **Note**: This endpoint is currently in Beta.
func (c Client) CreateProjectBranchAnonymized(projectID string, cfg BranchAnonymizedCreateRequest) (CreatedBranch, error) {
	var v CreatedBranch
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branch_anonymized", "POST", cfg, &v); err != nil {
		return CreatedBranch{}, err
	}
	return v, nil
}

// CreateProjectBranchBucket Creates a new branchable object-storage bucket on the specified branch.
// Buckets are managed by the Neon Platform branchable-storage service.
// **Note**: This endpoint is currently in Private Beta.
func (c Client) CreateProjectBranchBucket(projectID string, branchID string, cfg BucketCreateRequest) (BucketResponse, error) {
	var v BucketResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/buckets", "POST", cfg, &v); err != nil {
		return BucketResponse{}, err
	}
	return v, nil
}

// CreateProjectBranchDataAPI Creates a new instance of Neon Data API in the specified branch.
// The Data API exposes a REST interface over the branch database. The `database_name` path parameter determines which database the API serves.
func (c Client) CreateProjectBranchDataAPI(projectID string, branchID string, databaseName string, cfg *DataAPICreateRequest) (DataAPICreateResponse, error) {
	var v DataAPICreateResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/data-api/"+databaseName, "POST", cfg, &v); err != nil {
		return DataAPICreateResponse{}, err
	}
	return v, nil
}

// CreateProjectBranchDatabase Creates a database in the specified branch.
// A branch can have multiple databases.
// For related information, see [Manage databases](https://neon.com/docs/manage/databases/).
func (c Client) CreateProjectBranchDatabase(projectID string, branchID string, cfg DatabaseCreateRequest) (DatabaseOperations, error) {
	var v DatabaseOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/databases", "POST", cfg, &v); err != nil {
		return DatabaseOperations{}, err
	}
	return v, nil
}

// CreateProjectBranchFunctionDeployment Creates a deployment for the function. Supply any subset of zip,
// environment, and runtime; omitted fields inherit the
// function's latest version. At least one field must be supplied. The
// first deployment of a function must include zip. The newest deployment
// becomes active.
// **Note**: This endpoint is currently in Private Beta.
func (c Client) CreateProjectBranchFunctionDeployment(projectID string, branchID string, slug string, cfg *) (NeonFunctionDeploymentResponse, error) {
	var v NeonFunctionDeploymentResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/functions/"+slug+"/deployments", "POST", cfg, &v); err != nil {
		return NeonFunctionDeploymentResponse{}, err
	}
	return v, nil
}

// CreateProjectBranchRole Creates a Postgres role in the specified branch.
// For related information, see [Manage roles](https://neon.com/docs/manage/roles/).
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
// A compute endpoint is a Neon compute instance.
// There is a maximum of one read-write compute endpoint per branch.
// If the specified branch already has a read-write compute endpoint, the operation fails.
// A branch can have multiple read-only compute endpoints.
// For more information about compute endpoints, see [Manage computes](https://neon.com/docs/manage/endpoints/).
func (c Client) CreateProjectEndpoint(projectID string, cfg EndpointCreateRequest) (EndpointOperations, error) {
	var v EndpointOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints", "POST", cfg, &v); err != nil {
		return EndpointOperations{}, err
	}
	return v, nil
}

// CreateProjectTransferRequest Creates a transfer request for the specified project. The request expires after a set period.
// To accept the request, the recipient calls `PUT /projects/{project_id}/transfer_requests/{request_id}`
// or uses the Neon Console claim link.
// The optional `ru` parameter redirects the recipient after acceptance.
func (c Client) CreateProjectTransferRequest(projectID string, cfg *CreateProjectTransferRequestReqObj) (ProjectTransferRequestResponse, error) {
	var v ProjectTransferRequestResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/transfer_requests", "POST", cfg, &v); err != nil {
		return ProjectTransferRequestResponse{}, err
	}
	return v, nil
}

// CreateSnapshot Creates a snapshot from the specified branch.
// This operation may initiate an asynchronous process.
// **Note**: This endpoint is currently in Beta.
func (c Client) CreateSnapshot(projectID string, branchID string, lsn *string, timestamp *string, name *string, expiresAt *string) (CreateSnapshotRespObj, error) {
	var (
		queryElements []string
		query string
	)
	if lsn != nil {
		queryElements = append(queryElements, "lsn="+*lsn)
	}
	if timestamp != nil {
		queryElements = append(queryElements, "timestamp="+*timestamp)
	}
	if name != nil {
		queryElements = append(queryElements, "name="+*name)
	}
	if expiresAt != nil {
		queryElements = append(queryElements, "expires_at="+*expiresAt)
	}
	if len(queryElements) > 0 {
		query = "?" + strings.Join(queryElements, "&")
	}
	var v CreateSnapshotRespObj
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/snapshot" + query, "POST", nil, &v); err != nil {
		return CreateSnapshotRespObj{}, err
	}
	return v, nil
}

// DeleteBranchNeonAuthOauthProvider Deletes a OAuth provider from the specified project.
func (c Client) DeleteBranchNeonAuthOauthProvider(projectID string, branchID string, oauthProviderID NeonAuthOauthProviderId) error {
return c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/auth/oauth_providers/"+string(oauthProviderID), "DELETE", nil, nil)
}

// DeleteBranchNeonAuthTrustedDomain Removes a domain from the redirect URI whitelist for the specified branch.
// After removal, the domain can no longer be used as a redirect target after authentication.
func (c Client) DeleteBranchNeonAuthTrustedDomain(projectID string, branchID string, cfg NeonAuthDeleteDomainFromRedirectURIWhitelistRequest) error {
return c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/auth/domains", "DELETE", cfg, nil)
}

// DeleteBranchNeonAuthUser Deletes the specified user from the Neon Auth user directory for the specified branch.
// Removes the user record from `neon_auth.users_sync`. This action cannot be undone.
func (c Client) DeleteBranchNeonAuthUser(projectID string, branchID string, authUserID string) error {
return c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/auth/users/"+authUserID, "DELETE", nil, nil)
}

// DeleteNeonAuthDomainFromRedirectURIWhitelist DEPRECATED, use `/projects/{project_id}/branches/{branch_id}/auth/domains` instead. Deletes a domain from the redirect_uri whitelist for the specified project.
func (c Client) DeleteNeonAuthDomainFromRedirectURIWhitelist(projectID string, cfg NeonAuthDeleteDomainFromRedirectURIWhitelistRequest) error {
return c.requestHandler(c.baseURL+"/projects/"+projectID+"/auth/domains", "DELETE", cfg, nil)
}

// DeleteNeonAuthIntegration DEPRECATED, use `/projects/{project_id}/branches/{branch_id}/auth` instead.
func (c Client) DeleteNeonAuthIntegration(projectID string, authProvider NeonAuthSupportedAuthProvider, cfg *DeleteNeonAuthIntegrationReqObj) error {
return c.requestHandler(c.baseURL+"/projects/"+projectID+"/auth/integration/"+string(authProvider), "DELETE", cfg, nil)
}

// DeleteNeonAuthOauthProvider DEPRECATED, use `/projects/{project_id}/branches/{branch_id}/auth/oauth_providers/{oauth_provider_id}` instead. Deletes a OAuth provider from the specified project.
func (c Client) DeleteNeonAuthOauthProvider(projectID string, oauthProviderID NeonAuthOauthProviderId) error {
return c.requestHandler(c.baseURL+"/projects/"+projectID+"/auth/oauth_providers/"+string(oauthProviderID), "DELETE", nil, nil)
}

// DeleteNeonAuthUser DEPRECATED, use `/projects/{project_id}/branches/{branch_id}/auth/users/{auth_user_id}` instead. Deletes the auth user for the specified project.
func (c Client) DeleteNeonAuthUser(projectID string, authUserID string) error {
return c.requestHandler(c.baseURL+"/projects/"+projectID+"/auth/users/"+authUserID, "DELETE", nil, nil)
}

// DeleteOrganizationSpendingLimit Removes the configured monthly spending limit for the specified organization.
// Idempotent — removing an already-unset limit still succeeds.
// Available to organization admins on Launch and Scale plans only.
func (c Client) DeleteOrganizationSpendingLimit(orgID string) (EmptyResponse, error) {
	var v EmptyResponse
	if err := c.requestHandler(c.baseURL+"/organizations/"+orgID+"/billing/spending_limit", "DELETE", nil, &v); err != nil {
		return EmptyResponse{}, err
	}
	return v, nil
}

// DeleteOrganizationVPCEndpoint Deletes the VPC endpoint from the specified Neon organization.
// If you delete a VPC endpoint from a Neon organization, that VPC endpoint cannot
// be added back to the Neon organization.
func (c Client) DeleteOrganizationVPCEndpoint(orgID string, regionID string, vpcEndpointID string) error {
return c.requestHandler(c.baseURL+"/organizations/"+orgID+"/vpc/region/"+regionID+"/vpc_endpoints/"+vpcEndpointID, "DELETE", nil, nil)
}

// DeleteProject Deletes the specified project and all its endpoints, branches, databases, and users.
// Deleted projects can be recovered within 7 days using `POST /projects/{project_id}/recover`.
// To list recoverable projects, use `GET /projects?recoverable=true`.
func (c Client) DeleteProject(projectID string) (ProjectResponse, error) {
	var v ProjectResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID, "DELETE", nil, &v); err != nil {
		return ProjectResponse{}, err
	}
	return v, nil
}

// DeleteProjectBranch Deletes the specified branch from a project and places all compute endpoints into an idle state, breaking existing client connections.
// The deletion completes after all operations finish.
// You cannot delete a project's root or default branch, or a branch that has a child branch.
// A project must have at least one branch.
// By default, deleted branches can be recovered within a 7-day grace period.
// Use the `hard_delete` parameter to permanently delete the branch immediately.
// For related information, see [Manage branches](https://neon.com/docs/manage/branches/).
func (c Client) DeleteProjectBranch(projectID string, branchID string, hardDelete *bool) (BranchOperations, error) {
	var (
		queryElements []string
		query string
	)
	if hardDelete != nil {
		queryElements = append(queryElements, "hard_delete="+func (hardDelete bool) string { if hardDelete { return "true" }; return "false" } (*hardDelete))
	}
	if len(queryElements) > 0 {
		query = "?" + strings.Join(queryElements, "&")
	}
	var v BranchOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID + query, "DELETE", nil, &v); err != nil {
		return BranchOperations{}, err
	}
	return v, nil
}

// DeleteProjectBranchBucket Deletes the named bucket from the specified branch.
// **Note**: This endpoint is currently in Private Beta.
func (c Client) DeleteProjectBranchBucket(projectID string, branchID string, bucketName string) error {
return c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/buckets/"+bucketName, "DELETE", nil, nil)
}

// DeleteProjectBranchBucketObject Deletes the named object from the bucket on the specified branch.
// Served by the user's session (no customer S3 credentials required).
// **Note**: This endpoint is currently in Private Beta.
func (c Client) DeleteProjectBranchBucketObject(projectID string, branchID string, bucketName string, objectKey string) error {
return c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/buckets/"+bucketName+"/objects/"+objectKey, "DELETE", nil, nil)
}

// DeleteProjectBranchBucketObjectsByPrefix Soft-deletes every object on the specified branch whose key starts with
// `prefix`, in a single call. Intended to back a "delete folder" action in
// an object browser: a `prefix` of `app/avatars/` removes every object
// beneath that folder. Served by the user's session (no customer S3
// credentials required).
// `prefix` must be non-empty, end with `/`, be at most 1024 bytes, and
// contain no control characters - a partial-segment prefix cannot
// accidentally delete sibling keys. Returns the number of objects
// soft-deleted (`deleted`), which may be 0 when no live object matched the
// prefix on this branch.
// Only objects physically present on this branch are tombstoned; objects
// inherited from an ancestor branch via copy-on-write (not materialized on
// this branch) are out of scope.
// **Note**: This endpoint is currently in Private Beta.
func (c Client) DeleteProjectBranchBucketObjectsByPrefix(projectID string, branchID string, bucketName string, prefix string) (BucketObjectsDeletePrefixResponse, error) {
	var (
		queryElements []string
		query string
	)
	queryElements = append(queryElements, "prefix="+prefix)
	if len(queryElements) > 0 {
		query = "?" + strings.Join(queryElements, "&")
	}
	var v BucketObjectsDeletePrefixResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/buckets/"+bucketName+"/objects-by-prefix" + query, "DELETE", nil, &v); err != nil {
		return BucketObjectsDeletePrefixResponse{}, err
	}
	return v, nil
}

// DeleteProjectBranchDataAPI Deletes the Neon Data API for the specified branch.
// Existing connections using the Data API endpoint will fail after deletion.
func (c Client) DeleteProjectBranchDataAPI(projectID string, branchID string, databaseName string) (EmptyResponse, error) {
	var v EmptyResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/data-api/"+databaseName, "DELETE", nil, &v); err != nil {
		return EmptyResponse{}, err
	}
	return v, nil
}

// DeleteProjectBranchDatabase Deletes the specified database from the branch.
// For related information, see [Manage databases](https://neon.com/docs/manage/databases/).
func (c Client) DeleteProjectBranchDatabase(projectID string, branchID string, databaseName string) (DatabaseOperations, error) {
	var v DatabaseOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/databases/"+databaseName, "DELETE", nil, &v); err != nil {
		return DatabaseOperations{}, err
	}
	return v, nil
}

// DeleteProjectBranchFunction Deletes the function identified by its slug.
// **Note**: This endpoint is currently in Private Beta.
func (c Client) DeleteProjectBranchFunction(projectID string, branchID string, slug string) error {
return c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/functions/"+slug, "DELETE", nil, nil)
}

// DeleteProjectBranchRole Deletes the specified Postgres role from the branch.
// For related information, see [Manage roles](https://neon.com/docs/manage/roles/).
func (c Client) DeleteProjectBranchRole(projectID string, branchID string, roleName string) (RoleOperations, error) {
	var v RoleOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/roles/"+roleName, "DELETE", nil, &v); err != nil {
		return RoleOperations{}, err
	}
	return v, nil
}

// DeleteProjectEndpoint Deletes the specified compute endpoint.
// A compute endpoint is a Neon compute instance.
// Deleting a compute endpoint drops existing network connections to the compute endpoint.
// The deletion is completed when the last operation in the chain finishes successfully.
// An `endpoint_id` has an `ep-` prefix.
// For information about compute endpoints, see [Manage computes](https://neon.com/docs/manage/endpoints/).
func (c Client) DeleteProjectEndpoint(projectID string, endpointID string) (EndpointOperations, error) {
	var v EndpointOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints/"+endpointID, "DELETE", nil, &v); err != nil {
		return EndpointOperations{}, err
	}
	return v, nil
}

// DeleteProjectJWKS Removes the specified JWKS URL from the project.
// JWTs signed by keys from the removed URL can no longer authenticate to the project's endpoints.
func (c Client) DeleteProjectJWKS(projectID string, jwksID string) (JWKS, error) {
	var v JWKS
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/jwks/"+jwksID, "DELETE", nil, &v); err != nil {
		return JWKS{}, err
	}
	return v, nil
}

// DeleteProjectVPCEndpoint Removes the specified VPC endpoint restriction from a Neon project.
func (c Client) DeleteProjectVPCEndpoint(projectID string, vpcEndpointID string) error {
return c.requestHandler(c.baseURL+"/projects/"+projectID+"/vpc_endpoints/"+vpcEndpointID, "DELETE", nil, nil)
}

// DeleteSnapshot Deletes the specified snapshot.
// **Note**: This endpoint is currently in Beta.
func (c Client) DeleteSnapshot(projectID string, snapshotID string) error {
return c.requestHandler(c.baseURL+"/projects/"+projectID+"/snapshots/"+snapshotID, "DELETE", nil, nil)
}

// DisableNeonAuth Disables the Neon Auth integration for the specified branch, removing the connection
// to the authentication provider.
// If `delete_data` is `true`, also deletes the `neon_auth` schema and all associated tables
// from the branch database.
// The integration can be re-enabled by calling `POST /projects/{project_id}/branches/{branch_id}/auth`.
func (c Client) DisableNeonAuth(projectID string, branchID string, cfg *DisableNeonAuthReqObj) error {
return c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/auth", "DELETE", cfg, nil)
}

// FinalizeRestoreBranch Finalize the restore operation for a branch created from a snapshot.
// This operation updates the branch so it functions as the original branch it replaced.
// This includes:
// - Reassigning any computes from the original branch to the restored branch (this will restart the computes)
// - Renaming the restored branch to the original branch's name
// - Renaming the original branch so it no longer uses the original name
// This operation only applies to branches created using the `restoreSnapshot` endpoint with `finalize_restore: false`.
// **Note**: This endpoint is currently in Beta.
func (c Client) FinalizeRestoreBranch(projectID string, branchID string, cfg *FinalizeRestoreBranchReqObj) (OperationsResponse, error) {
	var v OperationsResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/finalize_restore", "POST", cfg, &v); err != nil {
		return OperationsResponse{}, err
	}
	return v, nil
}

// GetActiveRegions Lists supported Neon regions.
// **Note:** Not all regions are available to all organizations. Pass the `org_id`
// parameter to get an accurate list of regions available to your organization.
func (c Client) GetActiveRegions(orgID *string) (ActiveRegionsResponse, error) {
	var (
		queryElements []string
		query string
	)
	if orgID != nil {
		queryElements = append(queryElements, "org_id="+*orgID)
	}
	if len(queryElements) > 0 {
		query = "?" + strings.Join(queryElements, "&")
	}
	var v ActiveRegionsResponse
	if err := c.requestHandler(c.baseURL+"/regions" + query, "GET", nil, &v); err != nil {
		return ActiveRegionsResponse{}, err
	}
	return v, nil
}

// GetAnonymizedBranchStatus Retrieves the current status of an anonymized branch, including its state and progress information.
// This endpoint allows you to monitor the anonymization process from initialization through completion.
// Only anonymized branches will have status information available.
// **Note**: This endpoint is currently in Beta.
func (c Client) GetAnonymizedBranchStatus(projectID string, branchID string) (AnonymizedBranchStatusResponse, error) {
	var v AnonymizedBranchStatusResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/anonymized_status", "GET", nil, &v); err != nil {
		return AnonymizedBranchStatusResponse{}, err
	}
	return v, nil
}

// GetAuthDetails Returns authentication details for the credentials used in the request,
// including the credential type (API key, Bearer token, or OAuth session)
// and the associated identity.
func (c Client) GetAuthDetails() (AuthDetailsResponse, error) {
	var v AuthDetailsResponse
	if err := c.requestHandler(c.baseURL+"/auth", "GET", nil, &v); err != nil {
		return AuthDetailsResponse{}, err
	}
	return v, nil
}

// GetAvailablePreloadLibraries Returns the shared preload libraries available for the specified project's Postgres version.
// Shared preload libraries are Postgres extensions that require the `shared_preload_libraries`
// setting and a compute restart to activate.
// Use this list to determine which libraries can be enabled in the project's
// `settings.preload_libraries` configuration.
func (c Client) GetAvailablePreloadLibraries(projectID string) (AvailablePreloadLibraries, error) {
	var v AvailablePreloadLibraries
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/available_preload_libraries", "GET", nil, &v); err != nil {
		return AvailablePreloadLibraries{}, err
	}
	return v, nil
}

// GetConnectionURI Retrieves a connection URI for the specified database.
// The URI uses the standard PostgreSQL connection string format. Set `pooled=true` to include the `-pooler` suffix for a connection pooler URI.
func (c Client) GetConnectionURI(projectID string, branchID *string, endpointID *string, databaseName string, roleName string, pooled *bool) (ConnectionURIResponse, error) {
	var (
		queryElements []string
		query string
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
		queryElements = append(queryElements, "pooled="+func (pooled bool) string { if pooled { return "true" }; return "false" } (*pooled))
	}
	if len(queryElements) > 0 {
		query = "?" + strings.Join(queryElements, "&")
	}
	var v ConnectionURIResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/connection_uri" + query, "GET", nil, &v); err != nil {
		return ConnectionURIResponse{}, err
	}
	return v, nil
}

// GetConsumptionHistoryPerBranchV2 Returns consumption metrics for each branch across one or more projects listed in
// `project_ids` (1 to 100 projects). Available for accounts on paid usage-based Launch, Scale,
// Agent, and Enterprise plans.
// History starts when the account first ingests branch-level consumption data.
// The `metrics` query parameter is required. Only these six values are supported on this
// endpoint:
// `compute_unit_seconds`, `root_branch_bytes_month`, `child_branch_bytes_month`,
// `instant_restore_bytes_month`, `public_network_transfer_bytes`, `private_network_transfer_bytes`.
// This endpoint does not support `extra_branches_month` or `snapshot_storage_bytes_month`.
// Use `GET /consumption_history/v2/projects` for those.
// Consumption metrics within each branch are returned in ascending time order (oldest first).
// This request does not wake project computes.
func (c Client) GetConsumptionHistoryPerBranchV2(cursor *string, limit *int, projectIDs []string, branchIDs []string, from time.Time, to time.Time, granularity ConsumptionHistoryGranularity, orgID string, metrics ConsumptionHistoryQueryMetrics) (GetConsumptionHistoryPerBranchV2RespObj, error) {
	var (
		queryElements []string
		query string
	)
	queryElements = append(queryElements, "project_ids="+projectIDs)
	queryElements = append(queryElements, "from="+from.Format(time.RFC3339))
	queryElements = append(queryElements, "to="+to.Format(time.RFC3339))
	queryElements = append(queryElements, "granularity="+string(granularity))
	queryElements = append(queryElements, "org_id="+orgID)
	queryElements = append(queryElements, "metrics="+string(metrics))
	if cursor != nil {
		queryElements = append(queryElements, "cursor="+*cursor)
	}
	if limit != nil {
		queryElements = append(queryElements, "limit="+strconv.FormatInt(int64(*limit), 10))
	}
	if len(branchIDs) > 0 {
		queryElements = append(queryElements, "branch_ids="+strings.Join(branchIDs, ","))
	}
	if len(queryElements) > 0 {
		query = "?" + strings.Join(queryElements, "&")
	}
	var v GetConsumptionHistoryPerBranchV2RespObj
	if err := c.requestHandler(c.baseURL+"/consumption_history/v2/branches" + query, "GET", nil, &v); err != nil {
		return GetConsumptionHistoryPerBranchV2RespObj{}, err
	}
	return v, nil
}

// GetConsumptionHistoryPerProject Retrieves consumption metrics for Scale, Business, and Enterprise plan projects. History begins at the time of upgrade.
// Results are ordered by time in ascending order (oldest to newest).
// Issuing a call to this API does not wake a project's compute endpoint.
func (c Client) GetConsumptionHistoryPerProject(cursor *string, limit *int, projectIDs []string, from time.Time, to time.Time, granularity ConsumptionHistoryGranularity, orgID *string, includeV1Metrics *bool, metrics *ConsumptionHistoryQueryMetrics) (GetConsumptionHistoryPerProjectRespObj, error) {
	var (
		queryElements []string
		query string
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
		queryElements = append(queryElements, "include_v1_metrics="+func (includeV1Metrics bool) string { if includeV1Metrics { return "true" }; return "false" } (*includeV1Metrics))
	}
	if metrics != nil {
		queryElements = append(queryElements, "metrics="+string(*metrics))
	}
	if len(queryElements) > 0 {
		query = "?" + strings.Join(queryElements, "&")
	}
	var v GetConsumptionHistoryPerProjectRespObj
	if err := c.requestHandler(c.baseURL+"/consumption_history/projects" + query, "GET", nil, &v); err != nil {
		return GetConsumptionHistoryPerProjectRespObj{}, err
	}
	return v, nil
}

// GetConsumptionHistoryPerProjectV2 Returns consumption metrics for up to `limit` projects per page. If `project_ids` is omitted,
// projects in the organization are included across pages (use `cursor`). If `project_ids` is
// provided, the response is limited to those projects (up to 100). Available for accounts on
// Launch, Scale, Agent, Business, and Enterprise plans.
// History starts when the account upgrades to an eligible plan.
// The `metrics` query parameter is required. Supported values:
// `compute_unit_seconds`, `root_branch_bytes_month`, `child_branch_bytes_month`,
// `instant_restore_bytes_month`, `public_network_transfer_bytes`, `private_network_transfer_bytes`,
// `extra_branches_month`, `snapshot_storage_bytes_month`.
// Consumption metrics within each project are returned in ascending time order (oldest first).
// This request does not wake project computes.
func (c Client) GetConsumptionHistoryPerProjectV2(cursor *string, limit *int, projectIDs []string, from time.Time, to time.Time, granularity ConsumptionHistoryGranularity, orgID string, metrics ConsumptionHistoryQueryMetrics) (GetConsumptionHistoryPerProjectV2RespObj, error) {
	var (
		queryElements []string
		query string
	)
	queryElements = append(queryElements, "from="+from.Format(time.RFC3339))
	queryElements = append(queryElements, "to="+to.Format(time.RFC3339))
	queryElements = append(queryElements, "granularity="+string(granularity))
	queryElements = append(queryElements, "org_id="+orgID)
	queryElements = append(queryElements, "metrics="+string(metrics))
	if cursor != nil {
		queryElements = append(queryElements, "cursor="+*cursor)
	}
	if limit != nil {
		queryElements = append(queryElements, "limit="+strconv.FormatInt(int64(*limit), 10))
	}
	if len(projectIDs) > 0 {
		queryElements = append(queryElements, "project_ids="+strings.Join(projectIDs, ","))
	}
	if len(queryElements) > 0 {
		query = "?" + strings.Join(queryElements, "&")
	}
	var v GetConsumptionHistoryPerProjectV2RespObj
	if err := c.requestHandler(c.baseURL+"/consumption_history/v2/projects" + query, "GET", nil, &v); err != nil {
		return GetConsumptionHistoryPerProjectV2RespObj{}, err
	}
	return v, nil
}

// GetCurrentUserInfo Retrieves information about the currently authenticated Neon user,
// including account identifiers, plan details, and linked auth accounts.
func (c Client) GetCurrentUserInfo() (CurrentUserInfoResponse, error) {
	var v CurrentUserInfoResponse
	if err := c.requestHandler(c.baseURL+"/users/me", "GET", nil, &v); err != nil {
		return CurrentUserInfoResponse{}, err
	}
	return v, nil
}

// GetCurrentUserOrganizations Retrieves the organizations that the currently authenticated user belongs to.
// When called with an organization- or project-scoped API key (which is not
// tied to a user), this returns the single organization that owns the key.
func (c Client) GetCurrentUserOrganizations() (OrganizationsResponse, error) {
	var v OrganizationsResponse
	if err := c.requestHandler(c.baseURL+"/users/me/organizations", "GET", nil, &v); err != nil {
		return OrganizationsResponse{}, err
	}
	return v, nil
}

// GetMaskingRules Retrieves the masking rules for the specified anonymized branch.
// Masking rules define how sensitive data should be anonymized using PostgreSQL Anonymizer.
// **Note**: This endpoint is currently in Beta.
func (c Client) GetMaskingRules(projectID string, branchID string) (MaskingRulesResponse, error) {
	var v MaskingRulesResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/masking_rules", "GET", nil, &v); err != nil {
		return MaskingRulesResponse{}, err
	}
	return v, nil
}

// GetNeonAuth Retrieves the Neon Auth integration details for the specified branch,
// including the auth provider type and integration status.
func (c Client) GetNeonAuth(projectID string, branchID string) (NeonAuthIntegration, error) {
	var v NeonAuthIntegration
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/auth", "GET", nil, &v); err != nil {
		return NeonAuthIntegration{}, err
	}
	return v, nil
}

// GetNeonAuthAllowLocalhost Retrieves the localhost allow setting for the specified branch's Neon Auth integration.
// When enabled, authentication flows work from `localhost` without adding it to the redirect URI whitelist.
func (c Client) GetNeonAuthAllowLocalhost(projectID string, branchID string) (NeonAuthAllowLocalhostResponse, error) {
	var v NeonAuthAllowLocalhostResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/auth/allow_localhost", "GET", nil, &v); err != nil {
		return NeonAuthAllowLocalhostResponse{}, err
	}
	return v, nil
}

// GetNeonAuthEmailAndPasswordConfig Retrieves the email and password authentication configuration for the specified branch's Neon Auth integration,
// including whether it is enabled and the email verification method.
func (c Client) GetNeonAuthEmailAndPasswordConfig(projectID string, branchID string) (NeonAuthEmailAndPasswordConfig, error) {
	var v NeonAuthEmailAndPasswordConfig
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/auth/email_and_password", "GET", nil, &v); err != nil {
		return NeonAuthEmailAndPasswordConfig{}, err
	}
	return v, nil
}

// GetNeonAuthEmailProvider Retrieves the email provider configuration for the specified branch's Neon Auth integration,
// including the provider type and server settings.
func (c Client) GetNeonAuthEmailProvider(projectID string, branchID string) (NeonAuthEmailServerConfig, error) {
	var v NeonAuthEmailServerConfig
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/auth/email_provider", "GET", nil, &v); err != nil {
		return NeonAuthEmailServerConfig{}, err
	}
	return v, nil
}

// GetNeonAuthEmailServer DEPRECATED, use `/projects/{project_id}/branches/{branch_id}/auth/email_provider` instead. Gets the email server configuration for the specified project.
func (c Client) GetNeonAuthEmailServer(projectID string) (NeonAuthEmailServerConfig, error) {
	var v NeonAuthEmailServerConfig
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/auth/email_server", "GET", nil, &v); err != nil {
		return NeonAuthEmailServerConfig{}, err
	}
	return v, nil
}

// GetNeonAuthPhoneNumberPlugin Returns the phone number plugin configuration for Neon Auth.
// The phone number plugin enables phone-based OTP authentication.
func (c Client) GetNeonAuthPhoneNumberPlugin(projectID string, branchID string) (NeonAuthPhoneNumberConfig, error) {
	var v NeonAuthPhoneNumberConfig
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/auth/plugins/phone-number", "GET", nil, &v); err != nil {
		return NeonAuthPhoneNumberConfig{}, err
	}
	return v, nil
}

// GetNeonAuthPluginConfigs Returns all plugin configurations for Neon Auth in a single response.
// This endpoint aggregates organization, email provider, email and password,
// OAuth providers, and localhost settings.
func (c Client) GetNeonAuthPluginConfigs(projectID string, branchID string) (NeonAuthPluginConfigs, error) {
	var v NeonAuthPluginConfigs
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/auth/plugins", "GET", nil, &v); err != nil {
		return NeonAuthPluginConfigs{}, err
	}
	return v, nil
}

// GetNeonAuthWebhookConfig Returns the webhook configuration for the specified branch's Neon Auth integration,
// including the endpoint URL and the events that trigger it.
func (c Client) GetNeonAuthWebhookConfig(projectID string, branchID string) (NeonAuthWebhookConfig, error) {
	var v NeonAuthWebhookConfig
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/auth/webhooks", "GET", nil, &v); err != nil {
		return NeonAuthWebhookConfig{}, err
	}
	return v, nil
}

// GetOrganization Retrieves details for the specified organization, including its name, plan, and configuration.
func (c Client) GetOrganization(orgID string) (Organization, error) {
	var v Organization
	if err := c.requestHandler(c.baseURL+"/organizations/"+orgID, "GET", nil, &v); err != nil {
		return Organization{}, err
	}
	return v, nil
}

// GetOrganizationInvitations Retrieves pending and accepted invitations for the specified organization.
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

// GetOrganizationMembers Retrieves a paginated list of members for the specified organization.
func (c Client) GetOrganizationMembers(orgID string, sortBy *string, cursor *string, sortOrder *string, limit *int) (GetOrganizationMembersRespObj, error) {
	var (
		queryElements []string
		query string
	)
	if sortBy != nil {
		queryElements = append(queryElements, "sort_by="+*sortBy)
	}
	if cursor != nil {
		queryElements = append(queryElements, "cursor="+*cursor)
	}
	if sortOrder != nil {
		queryElements = append(queryElements, "sort_order="+*sortOrder)
	}
	if limit != nil {
		queryElements = append(queryElements, "limit="+strconv.FormatInt(int64(*limit), 10))
	}
	if len(queryElements) > 0 {
		query = "?" + strings.Join(queryElements, "&")
	}
	var v GetOrganizationMembersRespObj
	if err := c.requestHandler(c.baseURL+"/organizations/"+orgID+"/members" + query, "GET", nil, &v); err != nil {
		return GetOrganizationMembersRespObj{}, err
	}
	return v, nil
}

// GetOrganizationSpendingLimit Returns the configured monthly spending limit for the specified organization.
// `spending_limit_cents: null` indicates that no limit is currently set.
// Available to organization members with read access on Launch and Scale plans only.
func (c Client) GetOrganizationSpendingLimit(orgID string) (SpendingLimitResponse, error) {
	var v SpendingLimitResponse
	if err := c.requestHandler(c.baseURL+"/organizations/"+orgID+"/billing/spending_limit", "GET", nil, &v); err != nil {
		return SpendingLimitResponse{}, err
	}
	return v, nil
}

// GetOrganizationVPCEndpointDetails Retrieves the current state and configuration details of a specified VPC endpoint.
func (c Client) GetOrganizationVPCEndpointDetails(orgID string, regionID string, vpcEndpointID string) (VPCEndpointDetails, error) {
	var v VPCEndpointDetails
	if err := c.requestHandler(c.baseURL+"/organizations/"+orgID+"/vpc/region/"+regionID+"/vpc_endpoints/"+vpcEndpointID, "GET", nil, &v); err != nil {
		return VPCEndpointDetails{}, err
	}
	return v, nil
}

// GetProject Retrieves information about the specified project.
// Returned details include the project settings, compute configuration, history retention, owner information, and current usage metrics.
func (c Client) GetProject(projectID string) (ProjectResponse, error) {
	var v ProjectResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID, "GET", nil, &v); err != nil {
		return ProjectResponse{}, err
	}
	return v, nil
}

// GetProjectAdvisorSecurityIssues Analyzes the database for security and performance issues.
// Returns a list of issues categorized by severity (ERROR, WARN, INFO).
// Requires read access to the project and Data API enabled.
func (c Client) GetProjectAdvisorSecurityIssues(projectID string, branchID *string, databaseName *string, category *AdvisorCategory, minSeverity *string) (GetProjectAdvisorSecurityIssuesRespObj, error) {
	var (
		queryElements []string
		query string
	)
	if branchID != nil {
		queryElements = append(queryElements, "branch_id="+*branchID)
	}
	if databaseName != nil {
		queryElements = append(queryElements, "database_name="+*databaseName)
	}
	if category != nil {
		queryElements = append(queryElements, "category="+string(*category))
	}
	if minSeverity != nil {
		queryElements = append(queryElements, "min_severity="+*minSeverity)
	}
	if len(queryElements) > 0 {
		query = "?" + strings.Join(queryElements, "&")
	}
	var v GetProjectAdvisorSecurityIssuesRespObj
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/advisors" + query, "GET", nil, &v); err != nil {
		return GetProjectAdvisorSecurityIssuesRespObj{}, err
	}
	return v, nil
}

// GetProjectBranch Retrieves information about the specified branch.
// A `branch_id` value has a `br-` prefix.
// Each Neon project is initially created with a root and default branch named `main`.
// A project can contain one or more branches.
// A parent branch is identified by a `parent_id` value, which is the `id` of the parent branch.
// For related information, see [Manage branches](https://neon.com/docs/manage/branches/).
func (c Client) GetProjectBranch(projectID string, branchID string) (GetProjectBranchRespObj, error) {
	var v GetProjectBranchRespObj
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID, "GET", nil, &v); err != nil {
		return GetProjectBranchRespObj{}, err
	}
	return v, nil
}

// GetProjectBranchAiGateway Returns the AI Gateway endpoint host for the specified branch, used to
// render code-snippet base URLs. A 200 response means the branch is
// registered and this region serves the AI gateway. A 404 response
// includes a `reason` field indicating why the gateway is unavailable.
// **Note**: This endpoint is currently in Private Beta.
func (c Client) GetProjectBranchAiGateway(projectID string, branchID string) (BranchAiGateway, error) {
	var v BranchAiGateway
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/ai_gateway", "GET", nil, &v); err != nil {
		return BranchAiGateway{}, err
	}
	return v, nil
}

// GetProjectBranchBucketObject Streams the raw bytes of the named object from the bucket on the
// specified branch, including objects inherited from ancestor branches.
// Served by the user's session (no customer S3 credentials required).
// The body is returned as `application/octet-stream` so a browser treats
// it as a download; the `Content-Length` and `ETag` response headers echo
// the stored object metadata.
// BINARY-STREAM EXCEPTION TO THE BUILD-GENERATED-TYPES RULE (#7029): the
// successful 200 body is the raw object stream, proxied verbatim from the
// platform storage admin endpoint. It is modeled as an
// `application/octet-stream` binary body (not a JSON response schema) and
// is streamed without buffering the whole object in memory. Error
// responses still use the generated `GeneralError` shape.
// **Note**: This endpoint is currently in Private Beta.
func (c Client) GetProjectBranchBucketObject(projectID string, branchID string, bucketName string, objectKey string) error {
return c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/buckets/"+bucketName+"/objects/"+objectKey+"/download", "GET", nil, nil)
}

// GetProjectBranchDataAPI Retrieves the Neon Data API configuration for the specified branch,
// including endpoint URL, enabled state, and database settings.
func (c Client) GetProjectBranchDataAPI(projectID string, branchID string, databaseName string) (DataAPIReponse, error) {
	var v DataAPIReponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/data-api/"+databaseName, "GET", nil, &v); err != nil {
		return DataAPIReponse{}, err
	}
	return v, nil
}

// GetProjectBranchDatabase Retrieves information about the specified database.
// For related information, see [Manage databases](https://neon.com/docs/manage/databases/).
func (c Client) GetProjectBranchDatabase(projectID string, branchID string, databaseName string) (DatabaseResponse, error) {
	var v DatabaseResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/databases/"+databaseName, "GET", nil, &v); err != nil {
		return DatabaseResponse{}, err
	}
	return v, nil
}

// GetProjectBranchFunction Returns the function identified by its slug.
// **Note**: This endpoint is currently in Private Beta.
func (c Client) GetProjectBranchFunction(projectID string, branchID string, slug string) (NeonFunctionResponse, error) {
	var v NeonFunctionResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/functions/"+slug, "GET", nil, &v); err != nil {
		return NeonFunctionResponse{}, err
	}
	return v, nil
}

// GetProjectBranchRole Retrieves details about the specified role.
// In Neon, the terms "role" and "user" are synonymous.
// For related information, see [Manage roles](https://neon.com/docs/manage/roles/).
func (c Client) GetProjectBranchRole(projectID string, branchID string, roleName string) (RoleResponse, error) {
	var v RoleResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/roles/"+roleName, "GET", nil, &v); err != nil {
		return RoleResponse{}, err
	}
	return v, nil
}

// GetProjectBranchRolePassword Retrieves the password for the specified Postgres role, if possible.
// For related information, see [Manage roles](https://neon.com/docs/manage/roles/).
func (c Client) GetProjectBranchRolePassword(projectID string, branchID string, roleName string) (RolePasswordResponse, error) {
	var v RolePasswordResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/roles/"+roleName+"/reveal_password", "GET", nil, &v); err != nil {
		return RolePasswordResponse{}, err
	}
	return v, nil
}

// GetProjectBranchSchema Retrieves the schema from the specified database. The `lsn` and `timestamp` values cannot be specified at the same time. If both are omitted, the database schema is retrieved from database's head.
func (c Client) GetProjectBranchSchema(projectID string, branchID string, dbName string, lsn *string, timestamp *time.Time, format *string) (BranchSchemaResponse, error) {
	var (
		queryElements []string
		query string
	)
	queryElements = append(queryElements, "db_name="+dbName)
	if lsn != nil {
		queryElements = append(queryElements, "lsn="+*lsn)
	}
	if timestamp != nil {
		queryElements = append(queryElements, "timestamp="+timestamp.Format(time.RFC3339))
	}
	if format != nil {
		queryElements = append(queryElements, "format="+*format)
	}
	if len(queryElements) > 0 {
		query = "?" + strings.Join(queryElements, "&")
	}
	var v BranchSchemaResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/schema" + query, "GET", nil, &v); err != nil {
		return BranchSchemaResponse{}, err
	}
	return v, nil
}

// GetProjectBranchSchemaComparison Compares the schema from the specified database with another branch's schema.
func (c Client) GetProjectBranchSchemaComparison(projectID string, branchID string, baseBranchID *string, dbName string, lsn *string, timestamp *time.Time, baseLsn *string, baseTimestamp *time.Time) (BranchSchemaCompareResponse, error) {
	var (
		queryElements []string
		query string
	)
	queryElements = append(queryElements, "db_name="+dbName)
	if baseBranchID != nil {
		queryElements = append(queryElements, "base_branch_id="+*baseBranchID)
	}
	if lsn != nil {
		queryElements = append(queryElements, "lsn="+*lsn)
	}
	if timestamp != nil {
		queryElements = append(queryElements, "timestamp="+timestamp.Format(time.RFC3339))
	}
	if baseLsn != nil {
		queryElements = append(queryElements, "base_lsn="+*baseLsn)
	}
	if baseTimestamp != nil {
		queryElements = append(queryElements, "base_timestamp="+baseTimestamp.Format(time.RFC3339))
	}
	if len(queryElements) > 0 {
		query = "?" + strings.Join(queryElements, "&")
	}
	var v BranchSchemaCompareResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/compare_schema" + query, "GET", nil, &v); err != nil {
		return BranchSchemaCompareResponse{}, err
	}
	return v, nil
}

// GetProjectBranchStorage Returns whether branchable object-storage is usable for the specified
// branch. A 200 response means the branch is registered in the storage
// service and the S3 data plane will accept requests for it. A 404
// response includes a `reason` field indicating why storage is unavailable.
// **Note**: This endpoint is currently in Private Beta.
func (c Client) GetProjectBranchStorage(projectID string, branchID string) (BranchStorage, error) {
	var v BranchStorage
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/storage", "GET", nil, &v); err != nil {
		return BranchStorage{}, err
	}
	return v, nil
}

// GetProjectEndpoint Retrieves information about the specified compute endpoint.
// A compute endpoint is a Neon compute instance.
// An `endpoint_id` has an `ep-` prefix.
// For information about compute endpoints, see [Manage computes](https://neon.com/docs/manage/endpoints/).
func (c Client) GetProjectEndpoint(projectID string, endpointID string) (EndpointResponse, error) {
	var v EndpointResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints/"+endpointID, "GET", nil, &v); err != nil {
		return EndpointResponse{}, err
	}
	return v, nil
}

// GetProjectJWKS Returns the JWKS URLs available for verifying JWTs used as the authentication mechanism for the specified project.
func (c Client) GetProjectJWKS(projectID string) (ProjectJWKSResponse, error) {
	var v ProjectJWKSResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/jwks", "GET", nil, &v); err != nil {
		return ProjectJWKSResponse{}, err
	}
	return v, nil
}

// GetProjectOperation Retrieves details for the specified operation.
// An operation is an action performed on a Neon project resource.
func (c Client) GetProjectOperation(projectID string, operationID string) (OperationResponse, error) {
	var v OperationResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/operations/"+operationID, "GET", nil, &v); err != nil {
		return OperationResponse{}, err
	}
	return v, nil
}

// GetSnapshotSchedule Returns the backup schedule for the specified branch, including the configured snapshot frequencies.
// **Note**: This endpoint is currently in Beta.
func (c Client) GetSnapshotSchedule(projectID string, branchID string) (SnapshotSchedule, error) {
	var v SnapshotSchedule
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/backup_schedule", "GET", nil, &v); err != nil {
		return SnapshotSchedule{}, err
	}
	return v, nil
}

// GrantPermissionToProject Grants project access to the account associated with the specified email address.
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
// For more information, see [Manage API keys](https://neon.com/docs/manage/api-keys/).
func (c Client) ListApiKeys() ([]ApiKeysListResponseItem, error) {
	var v []ApiKeysListResponseItem
	if err := c.requestHandler(c.baseURL+"/api_keys", "GET", nil, &v); err != nil {
		return nil, err
	}
	return v, nil
}

// ListBranchNeonAuthOauthProviders Lists the OAuth providers configured for the specified branch's Neon Auth integration.
func (c Client) ListBranchNeonAuthOauthProviders(projectID string, branchID string) (ListNeonAuthOauthProvidersResponse, error) {
	var v ListNeonAuthOauthProvidersResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/auth/oauth_providers", "GET", nil, &v); err != nil {
		return ListNeonAuthOauthProvidersResponse{}, err
	}
	return v, nil
}

// ListBranchNeonAuthTrustedDomains Lists the trusted domains in the redirect URI whitelist for the specified branch.
// Only domains in this list are permitted as redirect targets after authentication.
func (c Client) ListBranchNeonAuthTrustedDomains(projectID string, branchID string) (NeonAuthRedirectURIWhitelistResponse, error) {
	var v NeonAuthRedirectURIWhitelistResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/auth/domains", "GET", nil, &v); err != nil {
		return NeonAuthRedirectURIWhitelistResponse{}, err
	}
	return v, nil
}

// ListCredentials Returns metadata for customer-issued credentials on the branch.
// Secrets are never included.
// **Note**: This endpoint is currently in Private Beta.
func (c Client) ListCredentials(projectID string, branchID string) (ListCredentialsResponse, error) {
	var v ListCredentialsResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/credentials", "GET", nil, &v); err != nil {
		return ListCredentialsResponse{}, err
	}
	return v, nil
}

// ListNeonAuthIntegrations DEPRECATED, use `/projects/{project_id}/branches/{branch_id}/auth` instead.
func (c Client) ListNeonAuthIntegrations(projectID string) (ListNeonAuthIntegrationsResponse, error) {
	var v ListNeonAuthIntegrationsResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/auth/integrations", "GET", nil, &v); err != nil {
		return ListNeonAuthIntegrationsResponse{}, err
	}
	return v, nil
}

// ListNeonAuthOauthProviders DEPRECATED, use `/projects/{project_id}/branches/{branch_id}/auth/oauth_providers` instead. Lists the OAuth providers for the specified project.
func (c Client) ListNeonAuthOauthProviders(projectID string) (ListNeonAuthOauthProvidersResponse, error) {
	var v ListNeonAuthOauthProvidersResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/auth/oauth_providers", "GET", nil, &v); err != nil {
		return ListNeonAuthOauthProvidersResponse{}, err
	}
	return v, nil
}

// ListNeonAuthRedirectURIWhitelistDomains DEPRECATED, use `/projects/{project_id}/branches/{branch_id}/auth/domains` instead. Lists the domains in the redirect_uri whitelist for the specified project.
func (c Client) ListNeonAuthRedirectURIWhitelistDomains(projectID string) (NeonAuthRedirectURIWhitelistResponse, error) {
	var v NeonAuthRedirectURIWhitelistResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/auth/domains", "GET", nil, &v); err != nil {
		return NeonAuthRedirectURIWhitelistResponse{}, err
	}
	return v, nil
}

// ListOrgApiKeys Retrieves the API keys for the specified organization.
// The response does not include API key tokens. A token is only provided when creating an API key.
// API keys can also be managed in the Neon Console.
// For more information, see [Manage API keys](https://neon.com/docs/manage/api-keys/).
func (c Client) ListOrgApiKeys(orgID string) ([]OrgApiKeysListResponseItem, error) {
	var v []OrgApiKeysListResponseItem
	if err := c.requestHandler(c.baseURL+"/organizations/"+orgID+"/api_keys", "GET", nil, &v); err != nil {
		return nil, err
	}
	return v, nil
}

// ListOrganizationVPCEndpoints Retrieves the list of VPC endpoints for the specified Neon organization.
func (c Client) ListOrganizationVPCEndpoints(orgID string, regionID string) (VPCEndpointsResponse, error) {
	var v VPCEndpointsResponse
	if err := c.requestHandler(c.baseURL+"/organizations/"+orgID+"/vpc/region/"+regionID+"/vpc_endpoints", "GET", nil, &v); err != nil {
		return VPCEndpointsResponse{}, err
	}
	return v, nil
}

// ListOrganizationVPCEndpointsAllRegions Retrieves the list of VPC endpoints for the specified Neon organization across all regions.
func (c Client) ListOrganizationVPCEndpointsAllRegions(orgID string) (VPCEndpointsWithRegionResponse, error) {
	var v VPCEndpointsWithRegionResponse
	if err := c.requestHandler(c.baseURL+"/organizations/"+orgID+"/vpc/vpc_endpoints", "GET", nil, &v); err != nil {
		return VPCEndpointsWithRegionResponse{}, err
	}
	return v, nil
}

// ListProjectBranchBucketObjects Lists objects visible in the named bucket on the specified branch,
// including those inherited from ancestor branches. Listing is served by
// the user's session (no customer S3 credentials required).
// When `delimiter` is supplied (typically `/`), keys are collapsed into
// common prefixes (`folders`) so callers can render a folder-style
// browser; keys that do not contain the delimiter after `prefix` are
// returned as `objects`.
// **Note**: This endpoint is currently in Private Beta.
func (c Client) ListProjectBranchBucketObjects(projectID string, branchID string, bucketName string, prefix *string, delimiter *string, cursor *string, limit *int32) (BucketObjectsListResponse, error) {
	var (
		queryElements []string
		query string
	)
	if prefix != nil {
		queryElements = append(queryElements, "prefix="+*prefix)
	}
	if delimiter != nil {
		queryElements = append(queryElements, "delimiter="+*delimiter)
	}
	if cursor != nil {
		queryElements = append(queryElements, "cursor="+*cursor)
	}
	if limit != nil {
		queryElements = append(queryElements, "limit="+strconv.FormatInt(int64(*limit), 10))
	}
	if len(queryElements) > 0 {
		query = "?" + strings.Join(queryElements, "&")
	}
	var v BucketObjectsListResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/buckets/"+bucketName+"/objects" + query, "GET", nil, &v); err != nil {
		return BucketObjectsListResponse{}, err
	}
	return v, nil
}

// ListProjectBranchBuckets Lists branchable object-storage buckets visible on the specified branch,
// including those inherited from ancestor branches.
// **Note**: This endpoint is currently in Private Beta.
func (c Client) ListProjectBranchBuckets(projectID string, branchID string) (BucketsListResponse, error) {
	var v BucketsListResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/buckets", "GET", nil, &v); err != nil {
		return BucketsListResponse{}, err
	}
	return v, nil
}

// ListProjectBranchDatabases Retrieves a list of databases for the specified branch.
// A branch can have multiple databases.
// For related information, see [Manage databases](https://neon.com/docs/manage/databases/).
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
func (c Client) ListProjectBranchEndpoints(projectID string, branchID string) (EndpointsResponse, error) {
	var v EndpointsResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/endpoints", "GET", nil, &v); err != nil {
		return EndpointsResponse{}, err
	}
	return v, nil
}

// ListProjectBranchFunctions Lists functions on the specified branch.
// **Note**: This endpoint is currently in Private Beta.
func (c Client) ListProjectBranchFunctions(projectID string, branchID string, cursor *string, limit *int) (ListProjectBranchFunctionsRespObj, error) {
	var (
		queryElements []string
		query string
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
	var v ListProjectBranchFunctionsRespObj
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/functions" + query, "GET", nil, &v); err != nil {
		return ListProjectBranchFunctionsRespObj{}, err
	}
	return v, nil
}

// ListProjectBranchRoles Retrieves a list of Postgres roles from the specified branch.
// For related information, see [Manage roles](https://neon.com/docs/manage/roles/).
func (c Client) ListProjectBranchRoles(projectID string, branchID string) (RolesResponse, error) {
	var v RolesResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/roles", "GET", nil, &v); err != nil {
		return RolesResponse{}, err
	}
	return v, nil
}

// ListProjectBranches Retrieves a list of branches for the specified project.
// Each Neon project has a root branch named `main`.
// A `branch_id` value has a `br-` prefix.
// A project may contain child branches that were branched from `main` or from another branch.
// A parent branch is identified by the `parent_id` value, which is the `id` of the parent branch.
// For related information, see [Manage branches](https://neon.com/docs/manage/branches/).
func (c Client) ListProjectBranches(projectID string, search *string, sortBy *string, cursor *string, sortOrder *string, limit *int, includeDeleted *bool) (ListProjectBranchesRespObj, error) {
	var (
		queryElements []string
		query string
	)
	if search != nil {
		queryElements = append(queryElements, "search="+*search)
	}
	if sortBy != nil {
		queryElements = append(queryElements, "sort_by="+*sortBy)
	}
	if cursor != nil {
		queryElements = append(queryElements, "cursor="+*cursor)
	}
	if sortOrder != nil {
		queryElements = append(queryElements, "sort_order="+*sortOrder)
	}
	if limit != nil {
		queryElements = append(queryElements, "limit="+strconv.FormatInt(int64(*limit), 10))
	}
	if includeDeleted != nil {
		queryElements = append(queryElements, "include_deleted="+func (includeDeleted bool) string { if includeDeleted { return "true" }; return "false" } (*includeDeleted))
	}
	if len(queryElements) > 0 {
		query = "?" + strings.Join(queryElements, "&")
	}
	var v ListProjectBranchesRespObj
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches" + query, "GET", nil, &v); err != nil {
		return ListProjectBranchesRespObj{}, err
	}
	return v, nil
}

// ListProjectEndpoints Retrieves a list of compute endpoints for the specified project.
// A compute endpoint is a Neon compute instance.
// For information about compute endpoints, see [Manage computes](https://neon.com/docs/manage/endpoints/).
func (c Client) ListProjectEndpoints(projectID string) (EndpointsResponse, error) {
	var v EndpointsResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints", "GET", nil, &v); err != nil {
		return EndpointsResponse{}, err
	}
	return v, nil
}

// ListProjectOperations Retrieves a list of operations for the specified Neon project.
// The number of operations returned can be large.
// To paginate the response, issue an initial request with a `limit` value.
// Then, add the `cursor` value that was returned in the response to the next request.
// Operations older than 6 months may be deleted from our systems.
// If you need more history than that, you should store your own history.
func (c Client) ListProjectOperations(projectID string, cursor *string, limit *int) (ListOperations, error) {
	var (
		queryElements []string
		query string
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
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/operations" + query, "GET", nil, &v); err != nil {
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

// ListProjectVPCEndpoints Lists VPC endpoint restrictions for the specified Neon project.
func (c Client) ListProjectVPCEndpoints(projectID string) (VPCEndpointsResponse, error) {
	var v VPCEndpointsResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/vpc_endpoints", "GET", nil, &v); err != nil {
		return VPCEndpointsResponse{}, err
	}
	return v, nil
}

// ListProjects Retrieves a list of projects for the specified organization.
// If using a personal API key, include the `org_id` parameter to specify which organization to work with.
// If using an org API key, `org_id` is automatically inferred from the key.
// For more information, see [Manage organizations using the Neon API](https://neon.com/docs/manage/orgs-api)
// and [Manage projects](https://neon.com/docs/manage/projects/).
func (c Client) ListProjects(cursor *string, limit *int, search *string, orgID *string, timeout *int, recoverable *bool) (ListProjectsRespObj, error) {
	var (
		queryElements []string
		query string
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
	if timeout != nil {
		queryElements = append(queryElements, "timeout="+strconv.FormatInt(int64(*timeout), 10))
	}
	if recoverable != nil {
		queryElements = append(queryElements, "recoverable="+func (recoverable bool) string { if recoverable { return "true" }; return "false" } (*recoverable))
	}
	if len(queryElements) > 0 {
		query = "?" + strings.Join(queryElements, "&")
	}
	var v ListProjectsRespObj
	if err := c.requestHandler(c.baseURL+"/projects" + query, "GET", nil, &v); err != nil {
		return ListProjectsRespObj{}, err
	}
	return v, nil
}

// ListSharedProjects Retrieves a list of projects shared with your Neon account.
// For more information, see [Manage projects](https://neon.com/docs/manage/projects/).
func (c Client) ListSharedProjects(cursor *string, limit *int, search *string, timeout *int) (ListSharedProjectsRespObj, error) {
	var (
		queryElements []string
		query string
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
	if timeout != nil {
		queryElements = append(queryElements, "timeout="+strconv.FormatInt(int64(*timeout), 10))
	}
	if len(queryElements) > 0 {
		query = "?" + strings.Join(queryElements, "&")
	}
	var v ListSharedProjectsRespObj
	if err := c.requestHandler(c.baseURL+"/projects/shared" + query, "GET", nil, &v); err != nil {
		return ListSharedProjectsRespObj{}, err
	}
	return v, nil
}

// ListSnapshots Lists the snapshots for the specified project.
// Each snapshot represents a point-in-time backup of the project data.
// **Note**: This endpoint is currently in Beta.
func (c Client) ListSnapshots(projectID string) (ListSnapshotsRespObj, error) {
	var v ListSnapshotsRespObj
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/snapshots", "GET", nil, &v); err != nil {
		return ListSnapshotsRespObj{}, err
	}
	return v, nil
}

// PresignProjectBranchBucketObject Returns a presigned URL that transfers bytes directly to or from the
// object's bucket on the specified branch, without the caller ever
// handling S3 credentials. The `operation` field selects the direction:
// - `upload` returns a presigned `PUT` URL (the caller `PUT`s the file
// bytes straight to `url` with the returned `headers`). Authorized with
// project write access.
// - `download` returns a presigned `GET` URL (the caller `GET`s the
// bytes straight from `url`). Authorized with project read access.
// The platform mints a short-lived credential and builds the SigV4-signed
// URL against the branch's S3 data-plane host, returning it together with
// the HTTP method, any headers the caller must echo, and the URL's expiry.
// Served by the user's session (no customer S3 credentials required).
// **Note**: This endpoint is currently in Private Beta.
func (c Client) PresignProjectBranchBucketObject(projectID string, branchID string, bucketName string, objectKey string, cfg PresignRequest) (PresignResponse, error) {
	var v PresignResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/buckets/"+bucketName+"/objects/"+objectKey+"/presign", "POST", cfg, &v); err != nil {
		return PresignResponse{}, err
	}
	return v, nil
}

// RecoverProject Recovers a deleted project within the 7-day deletion recovery period.
// Restores branches, endpoints, settings, and connection strings.
// Some integrations require manual reconfiguration after recovery.
// To list recoverable projects, use `GET /projects?recoverable=true`.
func (c Client) RecoverProject(projectID string) (ProjectRecoverResponse, error) {
	var v ProjectRecoverResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/recover", "POST", nil, &v); err != nil {
		return ProjectRecoverResponse{}, err
	}
	return v, nil
}

// RecoverProjectBranch Recovers a deleted branch within the 7-day deletion recovery period.
// The branch must have been soft deleted and not yet permanently deleted.
// Recovery restores the branch and its endpoints to an idle state.
// Connection strings remain valid after recovery.
// TTL branches become non-TTL branches after recovery.
// To list deleted branches available for recovery, use `GET /projects/{project_id}/branches?include_deleted=true`.
func (c Client) RecoverProjectBranch(projectID string, branchID string) (BranchRecoverResponse, error) {
	var v BranchRecoverResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/recover", "POST", nil, &v); err != nil {
		return BranchRecoverResponse{}, err
	}
	return v, nil
}

// RemoveOrganizationMember Removes the specified member from the organization.
// Only organization admins can perform this action.
// The last admin in an organization cannot be removed.
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
// For related information, see [Manage roles](https://neon.com/docs/manage/roles/).
func (c Client) ResetProjectBranchRolePassword(projectID string, branchID string, roleName string) (RoleOperations, error) {
	var v RoleOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/roles/"+roleName+"/reset_password", "POST", nil, &v); err != nil {
		return RoleOperations{}, err
	}
	return v, nil
}

// RestartProjectEndpoint Restarts the specified compute endpoint by immediately suspending it and then starting it again.
// An `endpoint_id` has an `ep-` prefix.
// For information about compute endpoints, see [Manage computes](https://neon.com/docs/manage/endpoints/).
func (c Client) RestartProjectEndpoint(projectID string, endpointID string) (EndpointOperations, error) {
	var v EndpointOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints/"+endpointID+"/restart", "POST", nil, &v); err != nil {
		return EndpointOperations{}, err
	}
	return v, nil
}

// RestoreProjectBranch Restores a branch to an earlier state in its own or another branch's history
// by specifying an LSN or timestamp.
// Creates a new branch from the historical state.
func (c Client) RestoreProjectBranch(projectID string, branchID string, cfg BranchRestoreRequest) (BranchOperations, error) {
	var v BranchOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/restore", "POST", cfg, &v); err != nil {
		return BranchOperations{}, err
	}
	return v, nil
}

// RestoreSnapshot Restores the specified snapshot to a new branch,
// and optionally finalizes the restore operation to replace the original branch.
// **Note**: This endpoint is currently in Beta.
func (c Client) RestoreSnapshot(projectID string, snapshotID string, name *string, cfg *RestoreSnapshotReqObj) (RestoredSnapshot, error) {
	var (
		queryElements []string
		query string
	)
	if name != nil {
		queryElements = append(queryElements, "name="+*name)
	}
	if len(queryElements) > 0 {
		query = "?" + strings.Join(queryElements, "&")
	}
	var v RestoredSnapshot
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/snapshots/"+snapshotID+"/restore" + query, "POST", cfg, &v); err != nil {
		return RestoredSnapshot{}, err
	}
	return v, nil
}

// RevokeApiKey Revokes the specified API key.
// An API key that is no longer needed can be revoked.
// This action cannot be reversed.
// API keys can also be managed in the Neon Console.
// See [Manage API keys](https://neon.com/docs/manage/api-keys/).
func (c Client) RevokeApiKey(keyID int64) (ApiKeyRevokeResponse, error) {
	var v ApiKeyRevokeResponse
	if err := c.requestHandler(c.baseURL+"/api_keys/"+strconv.FormatInt(keyID, 10), "DELETE", nil, &v); err != nil {
		return ApiKeyRevokeResponse{}, err
	}
	return v, nil
}

// RevokeCredential Soft-deletes the credential.  Idempotent.
// **Note**: This endpoint is currently in Private Beta.
func (c Client) RevokeCredential(projectID string, branchID string, tokenID string) error {
return c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/credentials/"+tokenID, "DELETE", nil, nil)
}

// RevokeOrgApiKey Revokes the specified organization API key.
// An API key that is no longer needed can be revoked.
// This action cannot be reversed.
// API keys can also be managed in the Neon Console.
// See [Manage API keys](https://neon.com/docs/manage/api-keys/).
func (c Client) RevokeOrgApiKey(orgID string, keyID int64) (OrgApiKeyRevokeResponse, error) {
	var v OrgApiKeyRevokeResponse
	if err := c.requestHandler(c.baseURL+"/organizations/"+orgID+"/api_keys/"+strconv.FormatInt(keyID, 10), "DELETE", nil, &v); err != nil {
		return OrgApiKeyRevokeResponse{}, err
	}
	return v, nil
}

// RevokePermissionFromProject Revokes project access from the user associated with the specified permission `id`. You can retrieve a user's permission `id` by listing project access.
func (c Client) RevokePermissionFromProject(projectID string, permissionID string) (ProjectPermission, error) {
	var v ProjectPermission
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/permissions/"+permissionID, "DELETE", nil, &v); err != nil {
		return ProjectPermission{}, err
	}
	return v, nil
}

// SendNeonAuthTestEmail Sends a test email using the configured email server settings to verify SMTP connectivity and credentials.
// The request body must include the SMTP server settings
// (`host`, `port`, `username`, `password`, `sender_email`, `sender_name`) and the `recipient_email` address.
func (c Client) SendNeonAuthTestEmail(projectID string, branchID string, cfg SendNeonAuthTestEmailRequest) (SendNeonAuthTestEmailResponse, error) {
	var v SendNeonAuthTestEmailResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/auth/send_test_email", "POST", cfg, &v); err != nil {
		return SendNeonAuthTestEmailResponse{}, err
	}
	return v, nil
}

// SetDefaultProjectBranch Sets the specified branch as the project's default branch.
// The default designation is automatically removed from the previous default branch.
// For more information, see [Manage branches](https://neon.com/docs/manage/branches/).
func (c Client) SetDefaultProjectBranch(projectID string, branchID string) (BranchOperations, error) {
	var v BranchOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/set_as_default", "POST", nil, &v); err != nil {
		return BranchOperations{}, err
	}
	return v, nil
}

// SetOrganizationSpendingLimit Sets the monthly spending limit for the specified organization.
// To remove a previously configured limit, send a DELETE request to this endpoint.
// When a limit is configured, email notifications are sent at 80% and 100% of the limit.
// Computes are not suspended when the limit is reached.
// Available to organization admins on Launch and Scale plans only.
func (c Client) SetOrganizationSpendingLimit(orgID string, cfg SpendingLimitUpdateRequest) (SpendingLimitResponse, error) {
	var v SpendingLimitResponse
	if err := c.requestHandler(c.baseURL+"/organizations/"+orgID+"/billing/spending_limit", "PUT", cfg, &v); err != nil {
		return SpendingLimitResponse{}, err
	}
	return v, nil
}

// SetSnapshotSchedule Updates the backup schedule for the specified branch.
// The schedule defines how often automatic snapshots are created (e.g., `hourly`, `daily`).
// **Note**: This endpoint is currently in Beta.
func (c Client) SetSnapshotSchedule(projectID string, branchID string, cfg BackupSchedule) (EmptyResponse, error) {
	var v EmptyResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/backup_schedule", "PUT", cfg, &v); err != nil {
		return EmptyResponse{}, err
	}
	return v, nil
}

// StartAnonymization Starts the anonymization process for an anonymized branch that is in the initialized, error, or anonymized state.
// This will apply all defined masking rules to anonymize sensitive data in the branch databases.
// The branch must be an anonymized branch to start anonymization.
// **Note**: This endpoint is currently in Beta.
func (c Client) StartAnonymization(projectID string, branchID string) (AnonymizedBranchStatusResponse, error) {
	var v AnonymizedBranchStatusResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/anonymize", "POST", nil, &v); err != nil {
		return AnonymizedBranchStatusResponse{}, err
	}
	return v, nil
}

// StartProjectEndpoint Starts a compute endpoint.
// The compute endpoint is ready to use after the last operation in the chain finishes successfully.
// An `endpoint_id` has an `ep-` prefix.
// For information about compute endpoints, see [Manage computes](https://neon.com/docs/manage/endpoints/).
func (c Client) StartProjectEndpoint(projectID string, endpointID string) (EndpointOperations, error) {
	var v EndpointOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints/"+endpointID+"/start", "POST", nil, &v); err != nil {
		return EndpointOperations{}, err
	}
	return v, nil
}

// SuspendProjectEndpoint Suspends the specified compute endpoint.
// An `endpoint_id` has an `ep-` prefix.
// For information about compute endpoints, see [Manage computes](https://neon.com/docs/manage/endpoints/).
func (c Client) SuspendProjectEndpoint(projectID string, endpointID string) (EndpointOperations, error) {
	var v EndpointOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/endpoints/"+endpointID+"/suspend", "POST", nil, &v); err != nil {
		return EndpointOperations{}, err
	}
	return v, nil
}

// TransferNeonAuthProviderProject Transfers ownership of your Neon-managed auth project to your own auth provider account.
func (c Client) TransferNeonAuthProviderProject(cfg NeonAuthTransferAuthProviderProjectRequest) (NeonAuthTransferAuthProviderProjectResponse, error) {
	var v NeonAuthTransferAuthProviderProjectResponse
	if err := c.requestHandler(c.baseURL+"/projects/auth/transfer_ownership", "POST", cfg, &v); err != nil {
		return NeonAuthTransferAuthProviderProjectResponse{}, err
	}
	return v, nil
}

// TransferProjectsFromOrgToOrg Transfers selected projects, identified by their IDs, from your organization to another specified organization.
func (c Client) TransferProjectsFromOrgToOrg(sourceOrgID string, cfg TransferProjectsToOrganizationRequest) (EmptyResponse, error) {
	var v EmptyResponse
	if err := c.requestHandler(c.baseURL+"/organizations/"+sourceOrgID+"/projects/transfer", "POST", cfg, &v); err != nil {
		return EmptyResponse{}, err
	}
	return v, nil
}

// TransferProjectsFromUserToOrg DEPRECATED. Personal accounts have been migrated to organizations, making this operation no longer applicable.
func (c Client) TransferProjectsFromUserToOrg(cfg TransferProjectsToOrganizationRequest) (EmptyResponse, error) {
	var v EmptyResponse
	if err := c.requestHandler(c.baseURL+"/users/me/projects/transfer", "POST", cfg, &v); err != nil {
		return EmptyResponse{}, err
	}
	return v, nil
}

// UpdateBranchNeonAuthOauthProvider Updates a OAuth provider for the specified project.
func (c Client) UpdateBranchNeonAuthOauthProvider(projectID string, branchID string, oauthProviderID NeonAuthOauthProviderId, cfg NeonAuthUpdateOAuthProviderRequest) (NeonAuthOauthProvider, error) {
	var v NeonAuthOauthProvider
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/auth/oauth_providers/"+string(oauthProviderID), "PATCH", cfg, &v); err != nil {
		return NeonAuthOauthProvider{}, err
	}
	return v, nil
}

// UpdateMaskingRules Updates the masking rules for the specified anonymized branch.
// Masking rules define how sensitive data should be anonymized using PostgreSQL Anonymizer.
// **Note**: This endpoint is currently in Beta.
func (c Client) UpdateMaskingRules(projectID string, branchID string, cfg MaskingRulesUpdateRequest) (MaskingRulesResponse, error) {
	var v MaskingRulesResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/masking_rules", "PATCH", cfg, &v); err != nil {
		return MaskingRulesResponse{}, err
	}
	return v, nil
}

// UpdateNeonAuthAllowLocalhost Updates the localhost allow setting for the specified branch's Neon Auth integration.
// When enabled, authentication flows work from `localhost` without adding it to the redirect URI whitelist.
func (c Client) UpdateNeonAuthAllowLocalhost(projectID string, branchID string, cfg UpdateNeonAuthAllowLocalhostRequest) (NeonAuthAllowLocalhostResponse, error) {
	var v NeonAuthAllowLocalhostResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/auth/allow_localhost", "PATCH", cfg, &v); err != nil {
		return NeonAuthAllowLocalhostResponse{}, err
	}
	return v, nil
}

// UpdateNeonAuthConfig Updates the auth configuration for the branch.
// Currently supports updating the application name used in auth emails.
func (c Client) UpdateNeonAuthConfig(projectID string, branchID string, cfg NeonAuthConfigUpdate) (NeonAuthConfigResponse, error) {
	var v NeonAuthConfigResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/auth/config", "PATCH", cfg, &v); err != nil {
		return NeonAuthConfigResponse{}, err
	}
	return v, nil
}

// UpdateNeonAuthEmailAndPasswordConfig Updates the email and password authentication configuration for the specified branch's Neon Auth integration.
// Only the fields provided in the request body are updated.
func (c Client) UpdateNeonAuthEmailAndPasswordConfig(projectID string, branchID string, cfg NeonAuthEmailAndPasswordConfigUpdate) (NeonAuthEmailAndPasswordConfig, error) {
	var v NeonAuthEmailAndPasswordConfig
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/auth/email_and_password", "PATCH", cfg, &v); err != nil {
		return NeonAuthEmailAndPasswordConfig{}, err
	}
	return v, nil
}

// UpdateNeonAuthEmailProvider Updates the email provider configuration for the specified branch's Neon Auth integration.
// The email provider handles transactional messages such as verification emails and password reset links.
func (c Client) UpdateNeonAuthEmailProvider(projectID string, branchID string, cfg NeonAuthEmailServerConfig) (NeonAuthEmailServerConfig, error) {
	var v NeonAuthEmailServerConfig
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/auth/email_provider", "PATCH", cfg, &v); err != nil {
		return NeonAuthEmailServerConfig{}, err
	}
	return v, nil
}

// UpdateNeonAuthEmailServer DEPRECATED, use `/projects/{project_id}/branches/{branch_id}/auth/email_provider` instead. Updates the email server configuration for the specified project.
func (c Client) UpdateNeonAuthEmailServer(projectID string, cfg NeonAuthEmailServerConfig) (NeonAuthEmailServerConfig, error) {
	var v NeonAuthEmailServerConfig
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/auth/email_server", "PATCH", cfg, &v); err != nil {
		return NeonAuthEmailServerConfig{}, err
	}
	return v, nil
}

// UpdateNeonAuthMagicLinkPlugin Updates the magic link plugin configuration for Neon Auth.
// The magic link plugin enables passwordless authentication via email magic links.
func (c Client) UpdateNeonAuthMagicLinkPlugin(projectID string, branchID string, cfg NeonAuthMagicLinkConfigUpdate) (NeonAuthMagicLinkConfig, error) {
	var v NeonAuthMagicLinkConfig
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/auth/plugins/magic-link", "PATCH", cfg, &v); err != nil {
		return NeonAuthMagicLinkConfig{}, err
	}
	return v, nil
}

// UpdateNeonAuthOauthProvider DEPRECATED, use `/projects/{project_id}/branches/{branch_id}/auth/oauth_providers/{oauth_provider_id}` instead. Updates a OAuth provider for the specified project.
func (c Client) UpdateNeonAuthOauthProvider(projectID string, oauthProviderID NeonAuthOauthProviderId, cfg NeonAuthUpdateOAuthProviderRequest) (NeonAuthOauthProvider, error) {
	var v NeonAuthOauthProvider
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/auth/oauth_providers/"+string(oauthProviderID), "PATCH", cfg, &v); err != nil {
		return NeonAuthOauthProvider{}, err
	}
	return v, nil
}

// UpdateNeonAuthOrganizationPlugin Updates the organization plugin configuration for Neon Auth.
// The organization plugin enables multi-tenant organization support.
func (c Client) UpdateNeonAuthOrganizationPlugin(projectID string, branchID string, cfg NeonAuthOrganizationConfigUpdate) (NeonAuthOrganizationConfig, error) {
	var v NeonAuthOrganizationConfig
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/auth/plugins/organization", "PATCH", cfg, &v); err != nil {
		return NeonAuthOrganizationConfig{}, err
	}
	return v, nil
}

// UpdateNeonAuthPhoneNumberPlugin Updates the phone number plugin configuration for Neon Auth.
// Only the fields provided in the request body are updated; omitted fields retain their current values.
// The phone number plugin enables phone-based OTP authentication.
// OTP codes are delivered via the `send.otp` webhook event with `delivery_preference: "sms"`.
// A webhook must be configured with the `send.otp` event enabled for SMS delivery to work.
func (c Client) UpdateNeonAuthPhoneNumberPlugin(projectID string, branchID string, cfg NeonAuthPhoneNumberConfigUpdate) (NeonAuthPhoneNumberConfig, error) {
	var v NeonAuthPhoneNumberConfig
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/auth/plugins/phone-number", "PATCH", cfg, &v); err != nil {
		return NeonAuthPhoneNumberConfig{}, err
	}
	return v, nil
}

// UpdateNeonAuthUserRole Updates the role of a user in the Neon Auth user directory for the specified branch.
// The role controls the user's level of access within the Neon Auth integration.
func (c Client) UpdateNeonAuthUserRole(projectID string, branchID string, authUserID string, cfg UpdateNeonAuthUserRoleRequest) (UpdateNeonAuthUserRoleResponse, error) {
	var v UpdateNeonAuthUserRoleResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/auth/users/"+authUserID+"/role", "PUT", cfg, &v); err != nil {
		return UpdateNeonAuthUserRoleResponse{}, err
	}
	return v, nil
}

// UpdateNeonAuthWebhookConfig Updates the webhook configuration for the specified branch's Neon Auth integration.
// Webhooks notify an external endpoint when auth events occur, such as user creation or sign-in.
func (c Client) UpdateNeonAuthWebhookConfig(projectID string, branchID string, cfg NeonAuthWebhookConfig) (NeonAuthWebhookConfig, error) {
	var v NeonAuthWebhookConfig
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/auth/webhooks", "PUT", cfg, &v); err != nil {
		return NeonAuthWebhookConfig{}, err
	}
	return v, nil
}

// UpdateOrganizationMember Updates the role of an existing member in the specified organization.
// The requested role must be valid for the organization.
// Only organization admins can call this endpoint.
func (c Client) UpdateOrganizationMember(orgID string, memberID string, cfg OrganizationMemberUpdateRequest) (Member, error) {
	var v Member
	if err := c.requestHandler(c.baseURL+"/organizations/"+orgID+"/members/"+memberID, "PATCH", cfg, &v); err != nil {
		return Member{}, err
	}
	return v, nil
}

// UpdateProject Updates the specified project.
// Configurable properties include the project name, default compute settings, history retention period, and IP allowlist.
func (c Client) UpdateProject(projectID string, cfg ProjectUpdateRequest) (UpdateProjectRespObj, error) {
	var v UpdateProjectRespObj
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID, "PATCH", cfg, &v); err != nil {
		return UpdateProjectRespObj{}, err
	}
	return v, nil
}

// UpdateProjectBranch Updates the specified branch.
// For more information, see [Manage branches](https://neon.com/docs/manage/branches/).
func (c Client) UpdateProjectBranch(projectID string, branchID string, cfg BranchUpdateRequest) (BranchOperations, error) {
	var v BranchOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID, "PATCH", cfg, &v); err != nil {
		return BranchOperations{}, err
	}
	return v, nil
}

// UpdateProjectBranchDataAPI Updates the Neon Data API configuration for the specified branch.
// You can optionally provide settings to update the Data API configuration.
// The schema cache is always refreshed as part of this operation.
func (c Client) UpdateProjectBranchDataAPI(projectID string, branchID string, databaseName string, cfg *DataAPIUpdateRequest) (EmptyResponse, error) {
	var v EmptyResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/data-api/"+databaseName, "PATCH", cfg, &v); err != nil {
		return EmptyResponse{}, err
	}
	return v, nil
}

// UpdateProjectBranchDatabase Updates the specified database in the branch.
// For related information, see [Manage databases](https://neon.com/docs/manage/databases/).
func (c Client) UpdateProjectBranchDatabase(projectID string, branchID string, databaseName string, cfg DatabaseUpdateRequest) (DatabaseOperations, error) {
	var v DatabaseOperations
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/databases/"+databaseName, "PATCH", cfg, &v); err != nil {
		return DatabaseOperations{}, err
	}
	return v, nil
}

// UpdateProjectBranchFunction Updates the function's mutable metadata — currently only the display
// `name`. A string sets the display name; `null` clears it, after which
// the function's `name` falls back to its slug. Leading and trailing
// whitespace is trimmed; a whitespace-only name is rejected. Acts only
// on a function owned by the branch: a slug that is only inherited from
// an ancestor branch returns 404 — rename it on the branch that owns
// it. Like every other change on a branch, a rename is isolated per
// branch: a branch forked before the rename keeps the name it had at
// fork time.
// **Note**: This endpoint is currently in Private Beta.
func (c Client) UpdateProjectBranchFunction(projectID string, branchID string, slug string, cfg NeonFunctionUpdateRequest) (NeonFunctionResponse, error) {
	var v NeonFunctionResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/functions/"+slug, "PATCH", cfg, &v); err != nil {
		return NeonFunctionResponse{}, err
	}
	return v, nil
}

// UpdateProjectEndpoint Updates the specified compute endpoint.
// An `endpoint_id` has an `ep-` prefix. A `branch_id` has a `br-` prefix.
// For more information about compute endpoints, see [Manage computes](https://neon.com/docs/manage/endpoints/).
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

// UpdateSnapshot Updates the specified snapshot.
// **Note**: This endpoint is currently in Beta.
func (c Client) UpdateSnapshot(projectID string, snapshotID string, cfg SnapshotUpdateRequest) (UpdateSnapshotRespObj, error) {
	var v UpdateSnapshotRespObj
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/snapshots/"+snapshotID, "PATCH", cfg, &v); err != nil {
		return UpdateSnapshotRespObj{}, err
	}
	return v, nil
}



type  struct{}

type  struct{}

type  struct{}

type  struct{}

type  struct{}

type  struct{}

type  struct{}

type  struct{}

type  struct{}

type  struct{}

type  struct{}

type  struct{}

type  struct{}

type  struct{}

type  struct{}

type  struct{}

type  struct{}

type  struct{}

type  struct{}

type  struct{}

type  struct{}

type  struct{}

type  struct{}

type  struct{}

type AcceptProjectTransferRequestReqObj struct {
OrgID *string `json:"org_id,omitempty"`
}

type ActiveRegionsResponse struct {
// Regions The list of active regions
Regions []RegionResponse `json:"regions"`
}

// AddProjectJWKSRequest AddTypeDefinition a new JWKS to a specific endpoint of a project
type AddProjectJWKSRequest struct {
// BranchID Branch ID
BranchID *string `json:"branch_id,omitempty"`
// JwksURL The URL that lists the JWKS
JwksURL string `json:"jwks_url"`
// JwtAudience The name of the required JWT Audience to be used
JwtAudience *string `json:"jwt_audience,omitempty"`
// ProviderName The name of the authentication provider (e.g., Clerk, Stytch, Auth0)
ProviderName string `json:"provider_name"`
// RoleNames DEPRECATED. This field should only be used when using Neon RLS. The roles the JWKS should be mapped to. By default, the JWKS is mapped to the `authenticator`, `authenticated` and `anonymous` roles.
RoleNames *[]string `json:"role_names,omitempty"`
// SkipRoleCreation DEPRECATED. This field should only be used when using Neon RLS. If true, the role creation will be skipped.
SkipRoleCreation *bool `json:"skip_role_creation,omitempty"`
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
CreatedAt *time.Time `json:"created_at,omitempty"`
Object AnnotationObjectData `json:"object"`
UpdatedAt *time.Time `json:"updated_at,omitempty"`
Value AnnotationValueData `json:"value"`
}

type AnnotationObjectData struct {
ID string `json:"id"`
Type string `json:"type"`
}

type AnnotationResponse struct {
Annotation AnnotationData `json:"annotation"`
}

// AnnotationValueData Annotation properties.
type AnnotationValueData struct{}

type AnnotationsMapResponse struct {
Annotations AnnotationsMapResponseAnnotations `json:"annotations"`
}

type AnnotationsMapResponseAnnotations struct{}

// AnonymizationRunMetadata Metadata about the most recent anonymization attempt for the branch.
type AnonymizationRunMetadata struct {
// CompletedAt Timestamp indicating when the latest anonymization attempt completed.
// Populated even if the attempt failed.
CompletedAt *time.Time `json:"completed_at,omitempty"`
// MaskedColumns Number of columns that had masking rules applied during the attempt.
MaskedColumns *int `json:"masked_columns,omitempty"`
// StartedAt Timestamp indicating when the latest anonymization attempt started.
StartedAt *time.Time `json:"started_at,omitempty"`
// TriggeredBy UUID of the user who triggered the latest anonymization attempt.
TriggeredBy *string `json:"triggered_by,omitempty"`
// TriggeredByUsername Username of the user who triggered the latest anonymization attempt.
TriggeredByUsername *string `json:"triggered_by_username,omitempty"`
}

type AnonymizedBranchStatusResponse struct {
// BranchID The ID of the anonymized branch
BranchID string `json:"branch_id"`
// CreatedAt A timestamp indicating when the anonymized branch was created
CreatedAt time.Time `json:"created_at"`
// FailedAt A timestamp indicating when the anonymized branch operation failed (if applicable)
FailedAt *time.Time `json:"failed_at,omitempty"`
LastRun *AnonymizationRunMetadata `json:"last_run,omitempty"`
// ProjectID The ID of the project
ProjectID string `json:"project_id"`
// State The current state of the anonymized branch. Possible values: created, initialized, initialization_error, anonymizing, anonymized, error
State string `json:"state"`
// StatusMessage A descriptive message about the current status or any errors
StatusMessage *string `json:"status_message,omitempty"`
// UpdatedAt A timestamp indicating when the anonymized branch status was last updated
UpdatedAt time.Time `json:"updated_at"`
}

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
CreatedAt time.Time `json:"created_at"`
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

type AuthDetailsResponse struct {
AccountID string `json:"account_id"`
AuthData *string `json:"auth_data,omitempty"`
AuthMethod string `json:"auth_method"`
}

type AvailablePreloadLibrary struct {
Description string `json:"description"`
IsDefault bool `json:"is_default"`
IsExperimental bool `json:"is_experimental"`
LibraryName string `json:"library_name"`
Version string `json:"version"`
}

type BackupSchedule struct {
Schedule []BackupScheduleItem `json:"schedule"`
}

type BackupScheduleItem struct {
// Day The day of the week or month to take the snapshot (if applicable).
Day *int `json:"day,omitempty"`
// Frequency How often to take snapshots. Must be one of the following values:
//   - `hourly`
//   - `daily`
//   - `weekly`
//   - `monthly`
//   - `yearly`
Frequency string `json:"frequency"`
// Hour The hour of the day to take the snapshot (if applicable).
Hour *int `json:"hour,omitempty"`
// Month The month of the year to take the snapshot (if applicable).
Month *int `json:"month,omitempty"`
// RetentionSeconds How long to keep a snapshot (in seconds) before it's automatically deleted.
// If not set, the snapshot is kept indefinitely.
RetentionSeconds *int `json:"retention_seconds,omitempty"`
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
OrbPortalURL *string `json:"orb_portal_url,omitempty"`
PaymentMethod BillingPaymentMethod `json:"payment_method"`
PaymentSource PaymentSource `json:"payment_source"`
PlanDetails *PlanDetails `json:"plan_details,omitempty"`
// QuotaResetAtLast The last time the quota was reset. Defaults to the date-time the account is created.
QuotaResetAtLast time.Time `json:"quota_reset_at_last"`
// SpendingLimitCents Monthly spending cap in cents for V3 paid plans. When set,
// notifications are sent at 80% and 100% of this limit. `null`
// means no limit is configured.
SpendingLimitCents *int64 `json:"spending_limit_cents,omitempty"`
State BillingAccountState `json:"state"`
SubscriptionType BillingSubscriptionType `json:"subscription_type"`
// TaxID The tax identification number for the billing account, displayed on invoices.
TaxID *string `json:"tax_id,omitempty"`
// TaxIDType The type of the tax identification number based on the country.
TaxIDType *string `json:"tax_id_type,omitempty"`
}

// BillingAccountState State of the billing account.
type BillingAccountState string

const (
BillingAccountStateUNKNOWN BillingAccountState = "UNKNOWN"
BillingAccountStateActive BillingAccountState = "active"
BillingAccountStateDeactivated BillingAccountState = "deactivated"
BillingAccountStateDeleted BillingAccountState = "deleted"
BillingAccountStateSuspended BillingAccountState = "suspended"
)

// BillingPaymentMethod Indicates whether and how an account makes payments.
type BillingPaymentMethod string

const (
BillingPaymentMethodUNKNOWN BillingPaymentMethod = "UNKNOWN"
BillingPaymentMethodAwsMp BillingPaymentMethod = "aws_mp"
BillingPaymentMethodAzureMp BillingPaymentMethod = "azure_mp"
BillingPaymentMethodDirectPayment BillingPaymentMethod = "direct_payment"
BillingPaymentMethodNone BillingPaymentMethod = "none"
BillingPaymentMethodSharedPaymentToken BillingPaymentMethod = "shared_payment_token"
BillingPaymentMethodSponsorship BillingPaymentMethod = "sponsorship"
BillingPaymentMethodStaff BillingPaymentMethod = "staff"
BillingPaymentMethodStripe BillingPaymentMethod = "stripe"
BillingPaymentMethodTrial BillingPaymentMethod = "trial"
BillingPaymentMethodVercelMp BillingPaymentMethod = "vercel_mp"
)

// BillingSubscriptionType Type of subscription to Neon Cloud.
// Notice that for users without billing account this will be "UNKNOWN"
type BillingSubscriptionType string

const (
BillingSubscriptionTypeUNKNOWN BillingSubscriptionType = "UNKNOWN"
BillingSubscriptionTypeAwsMarketplace BillingSubscriptionType = "aws_marketplace"
BillingSubscriptionTypeBusiness BillingSubscriptionType = "business"
BillingSubscriptionTypeDirectSales BillingSubscriptionType = "direct_sales"
BillingSubscriptionTypeDirectSalesV3 BillingSubscriptionType = "direct_sales_v3"
BillingSubscriptionTypeFreeV2 BillingSubscriptionType = "free_v2"
BillingSubscriptionTypeFreeV3 BillingSubscriptionType = "free_v3"
BillingSubscriptionTypeLaunch BillingSubscriptionType = "launch"
BillingSubscriptionTypeLaunchV3 BillingSubscriptionType = "launch_v3"
BillingSubscriptionTypeScale BillingSubscriptionType = "scale"
BillingSubscriptionTypeScaleV3 BillingSubscriptionType = "scale_v3"
BillingSubscriptionTypeVercelPgLegacy BillingSubscriptionType = "vercel_pg_legacy"
)

type Branch struct {
ActiveTimeSeconds int64 `json:"active_time_seconds"`
ComputeTimeSeconds int64 `json:"compute_time_seconds"`
// CpuUsedSec CPU seconds used by all of the branch's compute endpoints, including deleted ones.
// This value is reset at the beginning of each billing period.
// Examples:
// 1. A branch that uses 1 CPU for 1 second is equal to `cpu_used_sec=1`.
// 2. A branch that uses 2 CPUs simultaneously for 1 second is equal to `cpu_used_sec=2`.
CpuUsedSec int64 `json:"cpu_used_sec"`
// CreatedAt A timestamp indicating when the branch was created
CreatedAt time.Time `json:"created_at"`
CreatedBy *BranchCreatedBy `json:"created_by,omitempty"`
// CreationSource The branch creation source
CreationSource string `json:"creation_source"`
CurrentState BranchState `json:"current_state"`
DataTransferBytes int64 `json:"data_transfer_bytes"`
// Default Whether the branch is the project's default branch
Default bool `json:"default"`
// ExpiresAt The timestamp when the branch is scheduled to expire and be automatically deleted. Must be set by the client following the [RFC 3339, section 5.6](https://tools.ietf.org/html/rfc3339#section-5.6) format with precision up to seconds (such as 2025-06-09T18:02:16Z). Deletion is performed by a background job and may not occur exactly at the specified time.
// 
// Access to this feature is currently limited to participants in the Early Access Program.
ExpiresAt *time.Time `json:"expires_at,omitempty"`
// ID The branch ID. This value is generated when a branch is created. A `branch_id` value has a `br` prefix. For example: `br-small-term-683261`.
ID string `json:"id"`
// InitSource The source of initialization for the branch. Valid values are `schema-only` and `parent-data` (default).
//   * `schema-only` - creates a new root branch containing only the schema. Use `parent_id` to specify the source branch. Optionally, you can provide `parent_lsn` or `parent_timestamp` to branch from a specific point in time or LSN. These fields define which branch to copy the schema from and at what point—they do not establish a parent-child relationship between the `parent_id` branch and the new schema-only branch.
//   * `parent-data` - creates the branch with both schema and data from the parent.
InitSource *string `json:"init_source,omitempty"`
// LastResetAt A timestamp indicating when the branch was last reset
LastResetAt *time.Time `json:"last_reset_at,omitempty"`
// LogicalSize The logical size of the branch, in bytes
LogicalSize *int64 `json:"logical_size,omitempty"`
// Name The branch name
Name string `json:"name"`
// ParentID The `branch_id` of the parent branch
ParentID *string `json:"parent_id,omitempty"`
// ParentLsn The Log Sequence Number (LSN) on the parent branch from which this branch was created.
// When restoring a branch using the `POST /projects/{project_id}/branches/{branch_id}/restore` endpoint,
// this value isn’t finalized until all operations related to the restore have completed successfully.
ParentLsn *string `json:"parent_lsn,omitempty"`
// ParentTimestamp The point in time on the parent branch from which this branch was created.
// When restoring a branch using the `POST /projects/{project_id}/branches/{branch_id}/restore` endpoint,
// this value isn’t finalized until all operations related to the restore have completed successfully.
// After all the operations completed, this value might stay empty.
ParentTimestamp *time.Time `json:"parent_timestamp,omitempty"`
PendingState *BranchState `json:"pending_state,omitempty"`
// Primary DEPRECATED. Use `default` field.
// Whether the branch is the project's primary branch
Primary *bool `json:"primary,omitempty"`
// ProjectID The ID of the project to which the branch belongs
ProjectID string `json:"project_id"`
// Protected Whether the branch is protected
Protected bool `json:"protected"`
Recovery *BranchRecoveryInfo `json:"recovery,omitempty"`
RestoreStatus *BranchRestoreStatus `json:"restore_status,omitempty"`
// RestoredAs ID of the target branch which was replaced when this branch was restored
RestoredAs *string `json:"restored_as,omitempty"`
// RestoredFrom ID of the snapshot that was the restore source for this branch
RestoredFrom *string `json:"restored_from,omitempty"`
// RestrictedActions A list of actions that are currently restricted for this branch and the reason why.
RestrictedActions *[]BranchRestrictedAction `json:"restricted_actions,omitempty"`
// StateChangedAt A UTC timestamp indicating when the `current_state` began
StateChangedAt time.Time `json:"state_changed_at"`
// TtlIntervalSeconds The time-to-live (TTL) duration originally configured for the branch, in seconds. This read-only value represents the interval between the time `expires_at` was set and the expiration timestamp itself. It is preserved to ensure the same TTL duration is reapplied when resetting the branch from its parent, and only updates when a new `expires_at` value is set.
// 
// Access to this feature is currently limited to participants in the Early Access Program.
TtlIntervalSeconds *int `json:"ttl_interval_seconds,omitempty"`
// UpdatedAt A timestamp indicating when the branch was last updated
UpdatedAt time.Time `json:"updated_at"`
WrittenDataBytes int64 `json:"written_data_bytes"`
}

type BranchAiGateway struct {
// BaseURL The AI-gateway endpoint root for this branch — an OpenAI-compatible
// base URL. No dialect path is included; clients append the route
// (e.g. `/ai-gateway/openai/v1/responses`) themselves.
BaseURL string `json:"base_url"`
// Enabled Always `true` in 200 responses. Present for forward compatibility,
// mirroring BranchStorage.enabled.
Enabled bool `json:"enabled"`
}

type BranchAnonymizedCreateRequest struct {

AnnotationCreateValueRequest
}

type BranchCreateRequest struct {
Branch *BranchCreateRequestBranch `json:"branch,omitempty"`
Endpoints *[]BranchCreateRequestEndpointOptions `json:"endpoints,omitempty"`
}

type BranchCreateRequestBranch struct {
// Archived Whether to create the branch as archived
Archived *bool `json:"archived,omitempty"`
// ExpiresAt The timestamp when the branch is scheduled to expire and be automatically deleted. Must be set by the client following the [RFC 3339, section 5.6](https://tools.ietf.org/html/rfc3339#section-5.6) format with precision up to seconds (such as 2025-06-09T18:02:16Z). Deletion is performed by a background job and may not occur exactly at the specified time.
// 
// Access to this feature is currently limited to participants in the Early Access Program.
ExpiresAt *time.Time `json:"expires_at,omitempty"`
// InitSource The source of initialization for the branch. Valid values are `schema-only` and `parent-data` (default).
//   * `schema-only` - creates a new root branch containing only the schema. Use `parent_id` to specify the source branch. Optionally, you can provide `parent_lsn` or `parent_timestamp` to branch from a specific point in time or LSN. These fields define which branch to copy the schema from and at what point—they do not establish a parent-child relationship between the `parent_id` branch and the new schema-only branch.
//   * `parent-data` - creates the branch with both schema and data from the parent.
InitSource *string `json:"init_source,omitempty"`
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
}

type BranchCreateRequestEndpointOptions struct {
AutoscalingLimitMaxCu *ComputeUnit `json:"autoscaling_limit_max_cu,omitempty"`
AutoscalingLimitMinCu *ComputeUnit `json:"autoscaling_limit_min_cu,omitempty"`
Provisioner *Provisioner `json:"provisioner,omitempty"`
Settings *EndpointSettingsData `json:"settings,omitempty"`
SuspendTimeoutSeconds *SuspendTimeoutSeconds `json:"suspend_timeout_seconds,omitempty"`
Type EndpointType `json:"type"`
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

type BranchRecoverResponse struct {
BranchResponse
EndpointsOptionalResponse
}

// BranchRecoveryInfo Recovery information for a deleted branch. Only present when listing deleted branches
// with `include_deleted=true`.
// 
// This is part of the Branch Recovery feature, which is in preview and not available to all users.
type BranchRecoveryInfo struct {
// DeletedAt Timestamp when the branch was deleted
DeletedAt time.Time `json:"deleted_at"`
// DeletionMethod How the branch was deleted: 'user' for manual deletion, 'ttl' for TTL expiration
DeletionMethod string `json:"deletion_method"`
// RecoverableUntil Timestamp when the recovery window expires and the branch will be permanently deleted
RecoverableUntil time.Time `json:"recoverable_until"`
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

// BranchRestoreStatus Could be `restored`, `finalized` or `detaching`.
// A `restored` branch becomes permanently `finalized` when you call `finalizeRestoreBranch`
// A `restored` or `finalized` branch may begin `detaching` as a one-time performance optimisation, after which it will continue in its original state
type BranchRestoreStatus string

// BranchRestrictedAction An action that is currently restricted for the branch and the reason why.
type BranchRestrictedAction struct {
// Name The name of a restricted action. Possible values include `restore`, `delete-rw-endpoint`.
Name string `json:"name"`
// Reason A human-readable explanation of why the action is restricted.
Reason string `json:"reason"`
}

type BranchSchemaCompareResponse struct {
Diff *string `json:"diff,omitempty"`
}

type BranchSchemaJSON struct {
Tables []object `json:"tables"`
}

type BranchSchemaResponse struct {
Json *BranchSchemaJSON `json:"json,omitempty"`
Sql *string `json:"sql,omitempty"`
}

// BranchState The branch’s state, indicating if it is initializing, ready for use, or archived.
//   * 'init' - the branch is being created but is not available for querying.
//   * 'resetting' - the branch is being reset to a specific point in time or LSN and is not yet available for querying.
//   * 'ready' - the branch is fully operational and ready for querying. Expect normal query response times.
//   * 'archived' - the branch is stored in cost-effective archival storage. Expect slow query response times.
type BranchState string

type BranchStorage struct {
// Enabled Always `true` in 200 responses. Present for forward compatibility: a
// future version may add intermediate states; callers should treat `true`
// as "storage is usable for this branch right now."
Enabled bool `json:"enabled"`
// ForcePathStyle Whether the S3 client must use path-style addressing
// (bucket-in-path rather than virtual-hosted subdomain).
// Always true: the wildcard TLS cert covers one level of subdomain
// (*.storage.<suffix>), so the branch ID occupies that label and the
// bucket name must travel in the request path, not as a further
// subdomain. Callers must set the S3 SDK's ForcePathStyle (or
// equivalent) to true.
ForcePathStyle bool `json:"force_path_style"`
// Region The AWS region for this branch's storage. The platform normalizes
// the us-east-1 convention server-side: a non-empty region string is
// always returned in 200 responses (e.g. `"us-east-1"` for the S3
// default region).
Region string `json:"region"`
// S3Endpoint The S3-compatible endpoint URL for this branch.
S3Endpoint string `json:"s3_endpoint"`
}

type BranchUpdateRequest struct {
Branch BranchUpdateRequestBranch `json:"branch"`
}

type BranchUpdateRequestBranch struct {
// ExpiresAt The timestamp when the branch is scheduled to expire and be automatically deleted. Must be set by the client following the [RFC 3339, section 5.6](https://tools.ietf.org/html/rfc3339#section-5.6) format with precision up to seconds (such as 2025-06-09T18:02:16Z). Deletion is performed by a background job and may not occur exactly at the specified time. If this field is set to null, the expiration timestamp is removed.
// 
// Access to this feature is currently limited to participants in the Early Access Program.
ExpiresAt *time.Time `json:"expires_at,omitempty"`
Name *string `json:"name,omitempty"`
Protected *bool `json:"protected,omitempty"`
}

type BranchesCountResponse struct {
Count int `json:"count"`
}

type BranchesResponse struct {
Branches []Branch `json:"branches"`
}

type Bucket struct {
AccessLevel BucketAccessLevel `json:"access_level"`
// CreatedAt When the bucket was created. For a bucket inherited from an
// ancestor branch this is the ancestor's creation time (the branch
// fork never re-creates the bucket).
CreatedAt time.Time `json:"created_at"`
// Name The bucket name (unique within a branch).
Name string `json:"name"`
}

// BucketAccessLevel Controls anonymous access to objects in the bucket.
// - `private`: all reads and writes require authenticated requests (default).
// - `public_read`: anonymous `GetObject`/`HeadObject` requests succeed; listing,
//   writes, and deletes still require authenticated requests.
type BucketAccessLevel string

const (
BucketAccessLevelPrivate BucketAccessLevel = "private"
BucketAccessLevelPublicRead BucketAccessLevel = "public_read"
)

type BucketObject struct {
// Etag The object's entity tag (content hash).
Etag string `json:"etag"`
// Key The full object key.
Key string `json:"key"`
// LastModified The time the object was last modified.
LastModified time.Time `json:"last_modified"`
// Size The object size in bytes.
Size int64 `json:"size"`
}

type BucketObjectsDeletePrefixResponse struct {
// Deleted The number of objects soft-deleted under the prefix. 0 when no live
// object matched the prefix on this branch.
Deleted int64 `json:"deleted"`
}

type BucketObjectsListResponse struct {
// Folders Common prefixes (folder names) collapsed under the requested
// `delimiter`. Empty when no `delimiter` was supplied.
Folders []string `json:"folders"`
// IsTruncated True when more results exist beyond this page.
IsTruncated bool `json:"is_truncated"`
// NextCursor Pagination cursor to pass as `cursor` on the next request. Empty
// when the listing is not truncated.
NextCursor *string `json:"next_cursor,omitempty"`
// Objects whose keys did not collapse into a folder.
Objects []BucketObject `json:"objects"`
// Prefix The prefix that was applied to this listing (echoed back).
Prefix string `json:"prefix"`
}

type BucketResponse struct {
Bucket Bucket `json:"bucket"`
}

type BucketsListResponse struct {
Buckets []Bucket `json:"buckets"`
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

type ConnectionURIsOptionalResponse struct {
ConnectionURIs *[]ConnectionDetails `json:"connection_uris,omitempty"`
}

type ConnectionURIsResponse struct {
ConnectionURIs []ConnectionDetails `json:"connection_uris"`
}

type ConsumptionHistoryGranularity string

const (
ConsumptionHistoryGranularityDaily ConsumptionHistoryGranularity = "daily"
ConsumptionHistoryGranularityHourly ConsumptionHistoryGranularity = "hourly"
ConsumptionHistoryGranularityMonthly ConsumptionHistoryGranularity = "monthly"
)

type ConsumptionHistoryPerBranchV2 struct {
// BranchID The branch ID
BranchID string `json:"branch_id"`
Periods []ConsumptionHistoryPerPeriodV2 `json:"periods"`
// ProjectID The project that owns the branch
ProjectID string `json:"project_id"`
}

type ConsumptionHistoryPerBranchV2Response struct {
Branches []ConsumptionHistoryPerBranchV2 `json:"branches"`
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

type ConsumptionHistoryPerPeriodV2 struct {
Consumption []ConsumptionHistoryPerTimeframeV2 `json:"consumption"`
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

type ConsumptionHistoryPerProjectV2 struct {
Periods []ConsumptionHistoryPerPeriodV2 `json:"periods"`
// ProjectID The project ID
ProjectID string `json:"project_id"`
}

type ConsumptionHistoryPerProjectV2Response struct {
Projects []ConsumptionHistoryPerProjectV2 `json:"projects"`
}

type ConsumptionHistoryPerTimeframe struct {
// ActiveTimeSeconds Seconds. The amount of time the compute endpoints have been active.
ActiveTimeSeconds int `json:"active_time_seconds"`
// ComputeTimeSeconds Seconds. The number of CPU seconds used by compute endpoints, including compute endpoints that have been deleted.
ComputeTimeSeconds int `json:"compute_time_seconds"`
// DataStorageBytesHour Bytes-Hour. The amount of storage consumed hourly.
DataStorageBytesHour *int `json:"data_storage_bytes_hour,omitempty"`
// LogicalSizeBytes Bytes. The amount of logical size consumed.
LogicalSizeBytes *int `json:"logical_size_bytes,omitempty"`
// LogicalSizeBytesHour Bytes-Hour. The amount of logical size consumed hourly.
LogicalSizeBytesHour *int `json:"logical_size_bytes_hour,omitempty"`
// SyntheticStorageSizeBytes Bytes. The space occupied in storage. Synthetic storage size combines the logical data size and Write-Ahead Log (WAL) size for all branches.
SyntheticStorageSizeBytes int `json:"synthetic_storage_size_bytes"`
// TimeframeEnd The specified end date-time for the reported consumption.
TimeframeEnd time.Time `json:"timeframe_end"`
// TimeframeStart The specified start date-time for the reported consumption.
TimeframeStart time.Time `json:"timeframe_start"`
// WrittenDataBytes Bytes. The amount of written data for all branches.
WrittenDataBytes int `json:"written_data_bytes"`
}

type ConsumptionMetricValue struct {
MetricName string `json:"metric_name"`
Value int `json:"value"`
}

type CreateBranchNeonAuthNewUserRequest struct {
Email string `json:"email"`
Name *string `json:"name,omitempty"`
}

type CreateCredentialRequest struct {
// Name Free-form customer label for the credential.
Name *string `json:"name,omitempty"`
// PrincipalType Principal type for the credential. Only `user` is customer-managed
// and accepted here. `function` and `system` credentials are
// platform-internal (e.g. function-serve auto-mint, presign signer)
// and are never issued through the customer-facing API.
PrincipalType string `json:"principal_type"`
Scopes []CredentialScope `json:"scopes"`
}

type CreateCredentialResponse struct {
// ApiToken Bearer token; returned exactly once.
ApiToken string `json:"api_token"`
BranchID string `json:"branch_id"`
CreatedAt time.Time `json:"created_at"`
// ExpiresAt When the credential expires; absent means never expires.
ExpiresAt *time.Time `json:"expires_at,omitempty"`
// Name Customer-supplied label, echoed back from the request. Absent when not provided.
Name *string `json:"name,omitempty"`
// S3SecretAccessKey nsk_live_<64 hex>; the AWS_SECRET_ACCESS_KEY, returned exactly once.
S3SecretAccessKey string `json:"s3_secret_access_key"`
Scopes []CredentialScope `json:"scopes"`
// TokenID Opaque credential id (e.g. nak_live_<32hex>).
TokenID string `json:"token_id"`
// TokenIDShort First 12 hex chars of token_id; safe to log.
TokenIDShort string `json:"token_id_short"`
}

type CreateProjectBranchReqObj struct {
AnnotationCreateValueRequest
BranchCreateRequest
}

type CreateSnapshotRespObj struct {
Operations array `json:"operations"`
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

type CredentialMeta struct {
BranchID *string `json:"branch_id,omitempty"`
CreatedAt time.Time `json:"created_at"`
// ExpiresAt When the credential expires; absent means never expires. The
// verifier refuses to authenticate after `expires_at <= now()`.
ExpiresAt *time.Time `json:"expires_at,omitempty"`
FunctionID *string `json:"function_id,omitempty"`
LastUsedAt *time.Time `json:"last_used_at,omitempty"`
// Name Customer-supplied label; absent when not provided at issuance.
Name *string `json:"name,omitempty"`
PrincipalType string `json:"principal_type"`
RevokedAt *time.Time `json:"revoked_at,omitempty"`
Scopes []CredentialScope `json:"scopes"`
// TokenID Opaque credential id (e.g. nak_live_<32hex>).
TokenID string `json:"token_id"`
TokenIDShort string `json:"token_id_short"`
}

// CredentialScope A single capability a credential may exercise. A credential is granted
// a set of these; it may only perform actions explicitly listed in its
// scopes.
type CredentialScope string

const (
CredentialScopeAiGateway:invoke CredentialScope = "ai_gateway:invoke"
CredentialScopeFunctions:invoke CredentialScope = "functions:invoke"
CredentialScopeStorage:read CredentialScope = "storage:read"
CredentialScopeStorage:write CredentialScope = "storage:write"
)

type CurrentUserAuthAccount struct {
Email string `json:"email"`
Image string `json:"image"`
// Login DEPRECATED. Use `email` field.
Login string `json:"login"`
Name string `json:"name"`
Provider IdentityProviderId `json:"provider"`
}

type CurrentUserInfoResponse struct {
// ActiveSecondsLimit Control plane observes active endpoints of a user this amount of wall-clock time.
ActiveSecondsLimit int64 `json:"active_seconds_limit"`
AuthAccounts []CurrentUserAuthAccount `json:"auth_accounts"`
BillingAccount *BillingAccount `json:"billing_account,omitempty"`
BranchesLimit int64 `json:"branches_limit"`
ComputeSecondsLimit *int64 `json:"compute_seconds_limit,omitempty"`
Email string `json:"email"`
ID string `json:"id"`
Image string `json:"image"`
LastName string `json:"last_name"`
// Login DEPRECATED. Use `email` field.
Login string `json:"login"`
// MaxAutoscalingLimit The maximum autoscaling limit in Compute Units.
// A value of 0 indicates no limit is configured.
MaxAutoscalingLimit float64 `json:"max_autoscaling_limit"`
Name string `json:"name"`
Plan string `json:"plan"`
ProjectsLimit int64 `json:"projects_limit"`
}

type CursorPaginationResponse struct {
Pagination *CursorPagination `json:"pagination,omitempty"`
}

// DataAPICreateRequest Create Neon Data API
type DataAPICreateRequest struct {
// AddDefaultGrants Grant all permissions to the tables in the public schema to authenticated users
AddDefaultGrants *bool `json:"add_default_grants,omitempty"`
// AuthProvider The authentication provider to use for the Neon Data API
AuthProvider *string `json:"auth_provider,omitempty"`
// JwksURL The URL that lists the JWKS
JwksURL *string `json:"jwks_url,omitempty"`
// JwtAudience WARNING - using this setting will only reject tokens with a
// different audience claim. Tokens without audience claim will still
// be accepted.
JwtAudience *string `json:"jwt_audience,omitempty"`
// ProviderName The name of the authentication provider (e.g., Clerk, Stytch, Auth0)
ProviderName *string `json:"provider_name,omitempty"`
Settings *DataAPISettings `json:"settings,omitempty"`
// SkipAuthSchema Skip creating the auth schema and RLS functions
SkipAuthSchema *bool `json:"skip_auth_schema,omitempty"`
}

// DataAPICreateResponse Neon Data API created successfully
type DataAPICreateResponse struct {
URL string `json:"url"`
}

// DataAPIReponse Neon Data API response
type DataAPIReponse struct {
// AvailableSchemas List of available database schemas (SubZero only)
AvailableSchemas *[]string `json:"available_schemas,omitempty"`
Settings *DataAPIReponseSettings `json:"settings,omitempty"`
// Status The status of the Neon Data API deployment
Status string `json:"status"`
// URL The URL of the Neon Data API
URL string `json:"url"`
}

// DataAPIReponseSettings Configuration settings for the Data API (SubZero only)
type DataAPIReponseSettings struct{}

// DataAPISettings Configuration settings for the Neon Data API
type DataAPISettings struct {
// DbAggregatesEnabled Enable aggregates feature
DbAggregatesEnabled *bool `json:"db_aggregates_enabled,omitempty"`
// DbAnonRole Database role to use for anonymous requests
DbAnonRole *string `json:"db_anon_role,omitempty"`
// DbExtraSearchPath Extra schemas to add to the search path
DbExtraSearchPath *string `json:"db_extra_search_path,omitempty"`
// DbMaxRows Maximum number of rows that can be returned in a single request
DbMaxRows *int `json:"db_max_rows,omitempty"`
// DbSchemas List of schemas to expose via the API. Default: ["public"]
DbSchemas *[]string `json:"db_schemas,omitempty"`
// JwtCacheMaxLifetime Maximum lifetime for JWT cache in seconds
JwtCacheMaxLifetime *int `json:"jwt_cache_max_lifetime,omitempty"`
// JwtRoleClaimKey JWT claim key to use for role extraction
JwtRoleClaimKey *string `json:"jwt_role_claim_key,omitempty"`
// OpenapiMode OpenAPI specification mode (ignore-privileges, disabled)
OpenapiMode *string `json:"openapi_mode,omitempty"`
// ServerCorsAllowedOrigins CORS allowed origins
ServerCorsAllowedOrigins *string `json:"server_cors_allowed_origins,omitempty"`
// ServerTimingEnabled Enable server timing headers
ServerTimingEnabled *bool `json:"server_timing_enabled,omitempty"`
}

// DataAPIUpdateRequest Update Neon Data API
type DataAPIUpdateRequest struct {
Settings *DataAPISettings `json:"settings,omitempty"`
}

type DatabaseCreateRequest struct {
Database DatabaseCreateRequestDatabase `json:"database"`
}

type DatabaseCreateRequestDatabase struct {
// Name The name of the database
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
AutoscalingLimitMaxCu *ComputeUnit `json:"autoscaling_limit_max_cu,omitempty"`
AutoscalingLimitMinCu *ComputeUnit `json:"autoscaling_limit_min_cu,omitempty"`
PgSettings *PgSettingsData `json:"pg_settings,omitempty"`
PgbouncerSettings *PgbouncerSettingsData `json:"pgbouncer_settings,omitempty"`
SuspendTimeoutSeconds *SuspendTimeoutSeconds `json:"suspend_timeout_seconds,omitempty"`
}

type DeleteNeonAuthIntegrationReqObj struct {
DeleteData *bool `json:"delete_data,omitempty"`
}

type DisableNeonAuthReqObj struct {
DeleteData *bool `json:"delete_data,omitempty"`
}

// EmptyResponse Empty response.
type EmptyResponse struct{}

type EnableNeonAuthIntegrationRequest struct {
AuthProvider NeonAuthSupportedAuthProvider `json:"auth_provider"`
DatabaseName *string `json:"database_name,omitempty"`
}

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
CreationSource string `json:"creation_source"`
CurrentState EndpointState `json:"current_state"`
// Disabled Whether to restrict connections to the compute endpoint.
// Enabling this option schedules a suspend compute operation.
// A disabled compute endpoint cannot be enabled by a connection or
// console action.
Disabled bool `json:"disabled"`
// Host The hostname of the compute endpoint. This is the hostname specified when connecting to a Neon database.
Host string `json:"host"`
// ID The compute endpoint ID. Compute endpoint IDs have an `ep-` prefix. For example: `ep-little-smoke-851426`
ID string `json:"id"`
// LastActive A timestamp indicating when the compute endpoint was last active
LastActive *time.Time `json:"last_active,omitempty"`
// Name Optional name of the compute endpoint
Name *string `json:"name,omitempty"`
// PasswordlessAccess Whether to permit passwordless access to the compute endpoint
PasswordlessAccess bool `json:"passwordless_access"`
PendingState *EndpointState `json:"pending_state,omitempty"`
// PoolerEnabled DEPRECATED. Whether to enable connection pooling for the compute endpoint.
// The recommended way to enable connection pooling is to append `-pooler` to the endpoint ID in the connection string.
// See [How to use connection pooling](https://neon.com/docs/connect/connection-pooling#how-to-use-connection-pooling)
PoolerEnabled bool `json:"pooler_enabled"`
PoolerMode EndpointPoolerMode `json:"pooler_mode"`
// ProjectID The ID of the project to which the compute endpoint belongs
ProjectID string `json:"project_id"`
Provisioner Provisioner `json:"provisioner"`
// ProxyHost DEPRECATED. Use the "host" property instead.
ProxyHost string `json:"proxy_host"`
// RegionID The region identifier
RegionID string `json:"region_id"`
Settings EndpointSettingsData `json:"settings"`
// StartedAt A timestamp indicating when the compute endpoint was last started
StartedAt *time.Time `json:"started_at,omitempty"`
SuspendTimeoutSeconds SuspendTimeoutSeconds `json:"suspend_timeout_seconds"`
// SuspendedAt A timestamp indicating when the compute endpoint was last suspended
SuspendedAt *time.Time `json:"suspended_at,omitempty"`
Type EndpointType `json:"type"`
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
// Name Optional name of the compute endpoint
Name *string `json:"name,omitempty"`
// PasswordlessAccess NOT YET IMPLEMENTED. Whether to permit passwordless access to the compute endpoint.
PasswordlessAccess *bool `json:"passwordless_access,omitempty"`
// PoolerEnabled DEPRECATED. Whether to enable connection pooling for the compute endpoint.
// The recommended way to enable connection pooling is to append `-pooler` to the endpoint ID in the connection string.
// See [How to use connection pooling](https://neon.com/docs/connect/connection-pooling#how-to-use-connection-pooling)
PoolerEnabled *bool `json:"pooler_enabled,omitempty"`
PoolerMode *EndpointPoolerMode `json:"pooler_mode,omitempty"`
Provisioner *Provisioner `json:"provisioner,omitempty"`
// RegionID The region where the compute endpoint will be created. Only the project's `region_id` is permitted.
RegionID *string `json:"region_id,omitempty"`
Settings *EndpointSettingsData `json:"settings,omitempty"`
SuspendTimeoutSeconds *SuspendTimeoutSeconds `json:"suspend_timeout_seconds,omitempty"`
Type EndpointType `json:"type"`
}

type EndpointOperations struct {
EndpointResponse
OperationsResponse
}

// EndpointSettingsData A collection of settings for a compute endpoint
type EndpointSettingsData struct {
PgSettings *PgSettingsData `json:"pg_settings,omitempty"`
PgbouncerSettings *PgbouncerSettingsData `json:"pgbouncer_settings,omitempty"`
PreloadLibraries *PreloadLibraries `json:"preload_libraries,omitempty"`
}

// EndpointState The state of the compute endpoint
type EndpointState string

const (
EndpointStateActive EndpointState = "active"
EndpointStateIdle EndpointState = "idle"
EndpointStateInit EndpointState = "init"
)

// EndpointType The compute endpoint type. Either `read_write` or `read_only`.
type EndpointType string

const (
EndpointTypeReadOnly EndpointType = "read_only"
EndpointTypeReadWrite EndpointType = "read_write"
)

type EndpointUpdateRequest struct {
Endpoint EndpointUpdateRequestEndpoint `json:"endpoint"`
}

type EndpointUpdateRequestEndpoint struct {
AutoscalingLimitMaxCu *ComputeUnit `json:"autoscaling_limit_max_cu,omitempty"`
AutoscalingLimitMinCu *ComputeUnit `json:"autoscaling_limit_min_cu,omitempty"`
// BranchID DEPRECATED: This field will be removed in a future release.
// The destination branch ID. The destination branch must not have an existing read-write endpoint.
BranchID *string `json:"branch_id,omitempty"`
// Disabled Whether to restrict connections to the compute endpoint.
// Enabling this option schedules a suspend compute operation.
// A disabled compute endpoint cannot be enabled by a connection or
// console action. However, the compute endpoint is periodically
// enabled by check_availability operations.
Disabled *bool `json:"disabled,omitempty"`
// Name Optional name of the compute endpoint
Name *string `json:"name,omitempty"`
// PasswordlessAccess NOT YET IMPLEMENTED. Whether to permit passwordless access to the compute endpoint.
PasswordlessAccess *bool `json:"passwordless_access,omitempty"`
// PoolerEnabled DEPRECATED. Whether to enable connection pooling for the compute endpoint.
// The recommended way to enable connection pooling is to append `-pooler` to the endpoint ID in the connection string.
// See [How to use connection pooling](https://neon.com/docs/connect/connection-pooling#how-to-use-connection-pooling)
PoolerEnabled *bool `json:"pooler_enabled,omitempty"`
PoolerMode *EndpointPoolerMode `json:"pooler_mode,omitempty"`
Provisioner *Provisioner `json:"provisioner,omitempty"`
Settings *EndpointSettingsData `json:"settings,omitempty"`
SuspendTimeoutSeconds *SuspendTimeoutSeconds `json:"suspend_timeout_seconds,omitempty"`
}

type EndpointsOptionalResponse struct {
Endpoints *[]Endpoint `json:"endpoints,omitempty"`
}

type EndpointsResponse struct {
Endpoints []Endpoint `json:"endpoints"`
}

type FinalizeRestoreBranchReqObj struct {
Name *string `json:"name,omitempty"`
}

type GetConsumptionHistoryPerBranchV2RespObj struct {
ConsumptionHistoryPerBranchV2Response
PaginationResponse
}

type GetConsumptionHistoryPerProjectRespObj struct {
ConsumptionHistoryPerProjectResponse
PaginationResponse
}

type GetOrganizationMembersRespObj struct {
CursorPaginationResponse
OrganizationMembersResponse
}

type GetProjectAdvisorSecurityIssuesRespObj struct {
Issues array `json:"issues"`
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
IdentityProviderIdGithub IdentityProviderId = "github"
IdentityProviderIdGoogle IdentityProviderId = "google"
IdentityProviderIdHasura IdentityProviderId = "hasura"
IdentityProviderIdKeycloak IdentityProviderId = "keycloak"
IdentityProviderIdMicrosoft IdentityProviderId = "microsoft"
IdentityProviderIdMicrosoftv2 IdentityProviderId = "microsoftv2"
IdentityProviderIdVercelmp IdentityProviderId = "vercelmp"
)

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
RoleNames *[]string `json:"role_names,omitempty"`
// UpdatedAt The date and time when the JWKS was last modified
UpdatedAt time.Time `json:"updated_at"`
}

type JWKSResponse struct {
Jwks JWKS `json:"jwks"`
}

type ListCredentialsResponse struct {
Credentials []CredentialMeta `json:"credentials"`
}

type ListNeonAuthIntegrationsResponse struct {
Data []NeonAuthIntegration `json:"data"`
}

type ListNeonAuthOauthProvidersResponse struct {
Providers []NeonAuthOauthProvider `json:"providers"`
}

type ListOperations struct {
OperationsResponse
PaginationResponse
}

type ListProjectBranchFunctionsRespObj struct {
CursorPaginationResponse
NeonFunctionsListResponse
}

type ListProjectBranchesRespObj struct {
AnnotationsMapResponse
BranchesResponse
CursorPaginationResponse
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

type ListSnapshotsRespObj struct {
Snapshots array `json:"snapshots"`
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

type MaskingRule struct {
// ColumnName The name of the column to be masked
ColumnName string `json:"column_name"`
// DatabaseName The name of the database containing the table to be masked
DatabaseName string `json:"database_name"`
// MaskingFunction The PostgreSQL Anonymizer masking function to apply.
// Can be a predefined function (e.g., 'anon.random_string(10)', 'anon.fake_email()')
// or a custom function definition (e.g., 'anon.hash(column_name)')
MaskingFunction *string `json:"masking_function,omitempty"`
// MaskingValue A literal value to set on the column when masking.
MaskingValue *string `json:"masking_value,omitempty"`
// SchemaName The name of the schema containing the table to be masked
SchemaName string `json:"schema_name"`
// TableName The name of the table containing the column to be masked
TableName string `json:"table_name"`
}

type MaskingRulesResponse struct {
// MaskingRules List of masking rules for the branch
MaskingRules []MaskingRule `json:"masking_rules"`
}

type MaskingRulesUpdateRequest struct {
// MaskingRules List of masking rules to apply to the branch.
// This will replace all existing masking rules for the branch.
MaskingRules []MaskingRule `json:"masking_rules"`
}

// MemberRole The role of the organization member. Some role values may not be
// available for all organizations.
type MemberRole string

const (
MemberRoleAdmin MemberRole = "admin"
MemberRoleCollaborator MemberRole = "collaborator"
MemberRoleEditor MemberRole = "editor"
MemberRoleMember MemberRole = "member"
MemberRoleViewer MemberRole = "viewer"
)

type MemberUserInfo struct {
// DeactivatedAt Timestamp of when the user account was deactivated.
// Absent for active users. When present, the UI should render a
// "Deactivated" badge inline next to the member.
DeactivatedAt *time.Time `json:"deactivated_at,omitempty"`
Email string `json:"email"`
// HasMfa Whether the member has MFA (TOTP) enabled
HasMfa *bool `json:"has_mfa,omitempty"`
}

type MemberWithUser struct {
Member Member `json:"member"`
User MemberUserInfo `json:"user"`
}

type NeonAuthAddDomainToRedirectURIWhitelistRequest struct {
AuthProvider NeonAuthSupportedAuthProvider `json:"auth_provider"`
Domain string `json:"domain"`
}

type NeonAuthAddOAuthProviderRequest struct {
ClientID *string `json:"client_id,omitempty"`
ClientSecret *string `json:"client_secret,omitempty"`
ID NeonAuthOauthProviderId `json:"id"`
MicrosoftTenantID *string `json:"microsoft_tenant_id,omitempty"`
}

type NeonAuthAllowLocalhostResponse struct {
// AllowLocalhost Whether to allow localhost connections
AllowLocalhost bool `json:"allow_localhost"`
}

type NeonAuthConfigResponse struct {
// Name The application name used in auth emails and communications.
Name string `json:"name"`
}

type NeonAuthConfigUpdate struct {
// Name The application name used in auth emails and communications.
Name string `json:"name"`
}

type NeonAuthCreateAuthProviderSDKKeysRequest struct {
AuthProvider NeonAuthSupportedAuthProvider `json:"auth_provider"`
ProjectID string `json:"project_id"`
}

type NeonAuthCreateIntegrationRequest struct {
AuthProvider NeonAuthSupportedAuthProvider `json:"auth_provider"`
BranchID string `json:"branch_id"`
DatabaseName *string `json:"database_name,omitempty"`
ProjectID string `json:"project_id"`
RoleName *string `json:"role_name,omitempty"`
}

type NeonAuthCreateIntegrationResponse struct {
AuthProvider NeonAuthSupportedAuthProvider `json:"auth_provider"`
AuthProviderProjectID string `json:"auth_provider_project_id"`
BaseURL *string `json:"base_url,omitempty"`
JwksURL string `json:"jwks_url"`
PubClientKey string `json:"pub_client_key"`
SchemaName string `json:"schema_name"`
SecretServerKey string `json:"secret_server_key"`
TableName string `json:"table_name"`
}

type NeonAuthCreateNewUserRequest struct {
AuthProvider NeonAuthSupportedAuthProvider `json:"auth_provider"`
Email string `json:"email"`
Name *string `json:"name,omitempty"`
ProjectID string `json:"project_id"`
}

type NeonAuthCreateNewUserResponse struct {
// ID of newly created user
ID string `json:"id"`
}

type NeonAuthDeleteDomainFromRedirectURIWhitelistItem struct {
Domain string `json:"domain"`
}

type NeonAuthDeleteDomainFromRedirectURIWhitelistRequest struct {
AuthProvider NeonAuthSupportedAuthProvider `json:"auth_provider"`
Domains []NeonAuthDeleteDomainFromRedirectURIWhitelistItem `json:"domains"`
}

type NeonAuthEmailAndPasswordConfig struct {
// AutoSignInAfterVerification Whether users are automatically signed in after verifying their email
AutoSignInAfterVerification bool `json:"auto_sign_in_after_verification"`
// DisableSignUp Whether to disable new user sign ups
DisableSignUp bool `json:"disable_sign_up"`
EmailVerificationMethod NeonAuthEmailVerificationMethod `json:"email_verification_method"`
// Enabled Whether email and password authentication is enabled
Enabled bool `json:"enabled"`
// RequireEmailVerification Whether email verification is required before users can sign in
RequireEmailVerification bool `json:"require_email_verification"`
// SendVerificationEmailOnSignIn Whether to send a verification email when users sign in
SendVerificationEmailOnSignIn bool `json:"send_verification_email_on_sign_in"`
// SendVerificationEmailOnSignUp Whether to send a verification email when users sign up
SendVerificationEmailOnSignUp bool `json:"send_verification_email_on_sign_up"`
}

type NeonAuthEmailAndPasswordConfigUpdate struct {
// AutoSignInAfterVerification Whether users are automatically signed in after verifying their email
AutoSignInAfterVerification *bool `json:"auto_sign_in_after_verification,omitempty"`
// DisableSignUp Whether to disable new user sign ups
DisableSignUp *bool `json:"disable_sign_up,omitempty"`
EmailVerificationMethod *NeonAuthEmailVerificationMethod `json:"email_verification_method,omitempty"`
// Enabled Whether email and password authentication is enabled
Enabled *bool `json:"enabled,omitempty"`
// RequireEmailVerification Whether email verification is required before users can sign in
RequireEmailVerification *bool `json:"require_email_verification,omitempty"`
// SendVerificationEmailOnSignIn Whether to send a verification email when users sign in
SendVerificationEmailOnSignIn *bool `json:"send_verification_email_on_sign_in,omitempty"`
// SendVerificationEmailOnSignUp Whether to send a verification email when users sign up
SendVerificationEmailOnSignUp *bool `json:"send_verification_email_on_sign_up,omitempty"`
}

// NeonAuthEmailVerificationMethod The email verification method to use.
// - `link`: Sends a verification link via email
// - `otp`: Sends a one-time password (OTP) via email
type NeonAuthEmailVerificationMethod string

const (
NeonAuthEmailVerificationMethodLink NeonAuthEmailVerificationMethod = "link"
NeonAuthEmailVerificationMethodOtp NeonAuthEmailVerificationMethod = "otp"
)

type NeonAuthIntegration struct {
AuthProvider NeonAuthSupportedAuthProvider `json:"auth_provider"`
AuthProviderProjectID string `json:"auth_provider_project_id"`
BaseURL *string `json:"base_url,omitempty"`
BranchID string `json:"branch_id"`
CreatedAt time.Time `json:"created_at"`
DbName string `json:"db_name"`
JwksURL string `json:"jwks_url"`
// Name The application name used in auth emails and communications. Defaults to the Neon project name.
Name *string `json:"name,omitempty"`
OwnedBy NeonAuthProviderProjectOwnedBy `json:"owned_by"`
TransferStatus *NeonAuthProviderProjectTransferStatus `json:"transfer_status,omitempty"`
}

type NeonAuthMagicLinkConfig struct {
// DisableSignUp Whether to disable sign-up via magic link
DisableSignUp bool `json:"disable_sign_up"`
// Enabled Whether the magic link plugin is enabled
Enabled bool `json:"enabled"`
// ExpiresIn Time in minutes before the magic link expires
ExpiresIn int32 `json:"expires_in"`
}

type NeonAuthMagicLinkConfigUpdate struct {
// DisableSignUp Whether to disable sign-up via magic link
DisableSignUp *bool `json:"disable_sign_up,omitempty"`
// Enabled Whether the magic link plugin is enabled
Enabled *bool `json:"enabled,omitempty"`
// ExpiresIn Time in minutes before the magic link expires
ExpiresIn *int32 `json:"expires_in,omitempty"`
}

type NeonAuthOauthProvider struct {
ClientID *string `json:"client_id,omitempty"`
ClientSecret *string `json:"client_secret,omitempty"`
ID NeonAuthOauthProviderId `json:"id"`
Type NeonAuthOauthProviderType `json:"type"`
}

type NeonAuthOauthProviderId string

const (
NeonAuthOauthProviderIdGithub NeonAuthOauthProviderId = "github"
NeonAuthOauthProviderIdGoogle NeonAuthOauthProviderId = "google"
NeonAuthOauthProviderIdMicrosoft NeonAuthOauthProviderId = "microsoft"
NeonAuthOauthProviderIdVercel NeonAuthOauthProviderId = "vercel"
)

type NeonAuthOauthProviderType string

const (
NeonAuthOauthProviderTypeShared NeonAuthOauthProviderType = "shared"
NeonAuthOauthProviderTypeStandard NeonAuthOauthProviderType = "standard"
)

type NeonAuthOrganizationConfig struct {
// CreatorRole The role assigned to the user who creates an organization
CreatorRole string `json:"creator_role"`
// Enabled Whether the organization plugin is enabled
Enabled bool `json:"enabled"`
// MembershipLimit Maximum number of members per organization
MembershipLimit int32 `json:"membership_limit"`
// OrganizationLimit Maximum number of organizations a user can create
OrganizationLimit int32 `json:"organization_limit"`
// SendInvitationEmail Whether to send invitation emails when inviting members to an organization
SendInvitationEmail bool `json:"send_invitation_email"`
}

type NeonAuthOrganizationConfigUpdate struct {
// CreatorRole The role assigned to the user who creates an organization
CreatorRole *string `json:"creator_role,omitempty"`
// Enabled Whether the organization plugin is enabled
Enabled *bool `json:"enabled,omitempty"`
// MembershipLimit Maximum number of members per organization
MembershipLimit *int32 `json:"membership_limit,omitempty"`
// OrganizationLimit Maximum number of organizations a user can create
OrganizationLimit *int32 `json:"organization_limit,omitempty"`
// SendInvitationEmail Whether to send invitation emails when inviting members to an organization
SendInvitationEmail *bool `json:"send_invitation_email,omitempty"`
}

type NeonAuthPhoneNumberConfig struct {
// Enabled Whether the phone number plugin is enabled
Enabled bool `json:"enabled"`
// OtpExpiresIn Time in seconds before the OTP expires
OtpExpiresIn *int `json:"otp_expires_in,omitempty"`
}

type NeonAuthPhoneNumberConfigUpdate struct {
// Enabled Whether the phone number plugin is enabled
Enabled *bool `json:"enabled,omitempty"`
// OtpExpiresIn Time in seconds before the OTP expires
OtpExpiresIn *int `json:"otp_expires_in,omitempty"`
}

// NeonAuthPluginConfigs Aggregated plugin configurations for Neon Auth
type NeonAuthPluginConfigs struct {
AllowLocalhost *bool `json:"allow_localhost,omitempty"`
EmailAndPassword *NeonAuthEmailAndPasswordConfig `json:"email_and_password,omitempty"`
EmailProvider *NeonAuthEmailServerConfig `json:"email_provider,omitempty"`
MagicLink *NeonAuthMagicLinkConfig `json:"magic_link,omitempty"`
OauthProviders *[]NeonAuthOauthProvider `json:"oauth_providers,omitempty"`
Organization *NeonAuthOrganizationConfig `json:"organization,omitempty"`
PhoneNumber *NeonAuthPhoneNumberConfig `json:"phone_number,omitempty"`
}

type NeonAuthProviderProjectOwnedBy string

const (
NeonAuthProviderProjectOwnedByNeon NeonAuthProviderProjectOwnedBy = "neon"
NeonAuthProviderProjectOwnedByUser NeonAuthProviderProjectOwnedBy = "user"
)

type NeonAuthProviderProjectTransferStatus string

const (
NeonAuthProviderProjectTransferStatusFinished NeonAuthProviderProjectTransferStatus = "finished"
NeonAuthProviderProjectTransferStatusInitiated NeonAuthProviderProjectTransferStatus = "initiated"
)

type NeonAuthRedirectURIWhitelistDomain struct {
AuthProvider NeonAuthSupportedAuthProvider `json:"auth_provider"`
Domain string `json:"domain"`
}

type NeonAuthRedirectURIWhitelistResponse struct {
Domains []NeonAuthRedirectURIWhitelistDomain `json:"domains"`
}

type NeonAuthSupportedAuthProvider string

const (
NeonAuthSupportedAuthProviderBetterAuth NeonAuthSupportedAuthProvider = "better_auth"
NeonAuthSupportedAuthProviderMock NeonAuthSupportedAuthProvider = "mock"
NeonAuthSupportedAuthProviderStack NeonAuthSupportedAuthProvider = "stack"
)

type NeonAuthTransferAuthProviderProjectResponse struct {
// URL for completing the process of ownership transfer
URL string `json:"url"`
}

type NeonAuthUpdateOAuthProviderRequest struct {
ClientID *string `json:"client_id,omitempty"`
ClientSecret *string `json:"client_secret,omitempty"`
MicrosoftTenantID *string `json:"microsoft_tenant_id,omitempty"`
}

type NeonAuthWebhookConfig struct {
Enabled bool `json:"enabled"`
EnabledEvents *[]string `json:"enabled_events,omitempty"`
TimeoutSeconds *int `json:"timeout_seconds,omitempty"`
WebhookURL *string `json:"webhook_url,omitempty"`
}

type NeonFunction struct {
// ActiveDeployment The most recent deployment whose build completed successfully.
// This is the deployment that serves invocations. Omitted until a
// deployment succeeds.
ActiveDeployment * `json:"active_deployment,omitempty"`
CreatedAt string `json:"created_at"`
// CurrentDeployment The most recent deployment, regardless of build status. It may
// still be building or it may have failed. Omitted until the first
// deployment is created.
CurrentDeployment * `json:"current_deployment,omitempty"`
// ID Opaque, stable function identifier.
ID string `json:"id"`
// InvocationURL URL at which the function is invoked. The host carries `<branch_id>-<slug>` as its first DNS label under a Neon-managed functions domain, and the URL ends with a trailing slash so paths concatenate onto it. Empty string when the function has no servable invoke host (e.g. a deployment without an invocation front-door).
InvocationURL string `json:"invocation_url"`
// Name Free-form display name.
Name string `json:"name"`
// Slug Branch-unique, lowercase DNS-label. Forms the invocation URL's host together with the branch id. Immutable.
Slug string `json:"slug"`
}

type NeonFunctionDeployment struct {
CreatedAt string `json:"created_at"`
// Environment The NAMES of the deployment's environment variables, sorted.
// Values are encrypted at rest and are never returned — they are
// write-only. To change a value, deploy the variable with the new
// value; to remove a variable, deploy it with an empty value.
Environment *[]string `json:"environment,omitempty"`
// Error Human-readable reason the deployment build failed. Present only
// when `status` is `failed`.
Error *string `json:"error,omitempty"`
// ID The deployment id, which is the platform version number (monotonic per function).
ID int32 `json:"id"`
MemoryMib int32 `json:"memory_mib"`
Runtime string `json:"runtime"`
// Status Build lifecycle status of the deployment.
Status string `json:"status"`
}

type NeonFunctionDeploymentResponse struct {
Deployment NeonFunctionDeployment `json:"deployment"`
}

type NeonFunctionResponse struct {
Function NeonFunction `json:"function"`
}

type NeonFunctionUpdateRequest struct {
// Name New display name for the function. `null` clears the display
// name; the function's `name` then falls back to its slug. Leading
// and trailing whitespace is trimmed; a whitespace-only name is
// rejected.
Name string `json:"name"`
}

type Operation struct {
Action OperationAction `json:"action"`
// BranchID The branch ID
BranchID *string `json:"branch_id,omitempty"`
// CreatedAt A timestamp indicating when the operation was created
CreatedAt time.Time `json:"created_at"`
// EndpointID The endpoint ID
EndpointID *string `json:"endpoint_id,omitempty"`
// Error The error that occurred
Error *string `json:"error,omitempty"`
// FailuresCount The number of times the operation failed
FailuresCount int32 `json:"failures_count"`
// ID The operation ID
ID string `json:"id"`
// ProjectID The Neon project ID
ProjectID string `json:"project_id"`
// RetryAt A timestamp indicating when the operation was last retried
RetryAt *time.Time `json:"retry_at,omitempty"`
Status OperationStatus `json:"status"`
// TotalDurationMs The total duration of the operation in milliseconds
TotalDurationMs int32 `json:"total_duration_ms"`
// UpdatedAt A timestamp indicating when the operation status was last updated
UpdatedAt time.Time `json:"updated_at"`
}

// OperationAction The action performed by the operation
type OperationAction string

const (
OperationActionApplyConfig OperationAction = "apply_config"
OperationActionApplySchemaFromBranch OperationAction = "apply_schema_from_branch"
OperationActionApplyStorageConfig OperationAction = "apply_storage_config"
OperationActionCheckAvailability OperationAction = "check_availability"
OperationActionCreateBranch OperationAction = "create_branch"
OperationActionCreateCompute OperationAction = "create_compute"
OperationActionCreateTimeline OperationAction = "create_timeline"
OperationActionDeleteTimeline OperationAction = "delete_timeline"
OperationActionDetachParentBranch OperationAction = "detach_parent_branch"
OperationActionDisableMaintenance OperationAction = "disable_maintenance"
OperationActionFinalizeMigration OperationAction = "finalize_migration"
OperationActionImportData OperationAction = "import_data"
OperationActionMarkMigrationPrepared OperationAction = "mark_migration_prepared"
OperationActionPrepareSecondaryPageserver OperationAction = "prepare_secondary_pageserver"
OperationActionPrewarmReplica OperationAction = "prewarm_replica"
OperationActionPromoteReplica OperationAction = "promote_replica"
OperationActionReplaceSafekeeper OperationAction = "replace_safekeeper"
OperationActionSetStorageNonDirty OperationAction = "set_storage_non_dirty"
OperationActionStartCompute OperationAction = "start_compute"
OperationActionStartReservedCompute OperationAction = "start_reserved_compute"
OperationActionSuspendCompute OperationAction = "suspend_compute"
OperationActionSwapBindingId OperationAction = "swap_binding_id"
OperationActionSwitchPageserver OperationAction = "switch_pageserver"
OperationActionSyncDbsAndRolesFromCompute OperationAction = "sync_dbs_and_roles_from_compute"
OperationActionTenantAttach OperationAction = "tenant_attach"
OperationActionTenantAttachSafekeepers OperationAction = "tenant_attach_safekeepers"
OperationActionTenantDetach OperationAction = "tenant_detach"
OperationActionTenantDetachSafekeepers OperationAction = "tenant_detach_safekeepers"
OperationActionTenantIgnore OperationAction = "tenant_ignore"
OperationActionTenantReattach OperationAction = "tenant_reattach"
OperationActionTimelineArchive OperationAction = "timeline_archive"
OperationActionTimelineMarkInvisible OperationAction = "timeline_mark_invisible"
OperationActionTimelineUnarchive OperationAction = "timeline_unarchive"
OperationActionTimelineUpdateProtectedConfig OperationAction = "timeline_update_protected_config"
OperationActionUpdateCatalog OperationAction = "update_catalog"
)

type OperationResponse struct {
Operation Operation `json:"operation"`
}

// OperationStatus The status of the operation
type OperationStatus string

const (
OperationStatusCancelled OperationStatus = "cancelled"
OperationStatusCancelling OperationStatus = "cancelling"
OperationStatusError OperationStatus = "error"
OperationStatusFailed OperationStatus = "failed"
OperationStatusFinished OperationStatus = "finished"
OperationStatusRunning OperationStatus = "running"
OperationStatusScheduling OperationStatus = "scheduling"
OperationStatusSkipped OperationStatus = "skipped"
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
// AllowHipaaProjects If true, allow account to mark projects as HIPAA
AllowHipaaProjects *bool `json:"allow_hipaa_projects,omitempty"`
// CreatedAt A timestamp indicting when the organization was created
CreatedAt time.Time `json:"created_at"`
Handle string `json:"handle"`
ID string `json:"id"`
// ManagedBy Organizations created via the Console or the API are managed by `console`.
// Organizations created by other methods can't be deleted via the Console or the API.
ManagedBy string `json:"managed_by"`
Name string `json:"name"`
Plan string `json:"plan"`
// RequireMfa If true, all members must have MFA enabled to access this organization
RequireMfa *bool `json:"require_mfa,omitempty"`
// UpdatedAt A timestamp indicating when the organization was updated
UpdatedAt time.Time `json:"updated_at"`
}

type OrganizationInvitationsResponse struct {
Invitations []Invitation `json:"invitations"`
}

type OrganizationInviteCreateRequest struct {
Email string `json:"email"`
Role MemberRole `json:"role"`
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
type PgSettingsData struct{}

// PgVersion The major Postgres version number. Generally available versions are `14`, `15`, `16`, `17`, and `18`. `19` is being rolled out and is only accepted in regions where it has been enabled; requesting it in a region where it is not yet available returns an error.
type PgVersion int

// PgbouncerSettingsData DEPRECATED. A raw representation of PgBouncer settings. This schema is deprecated and will be removed after 2026-06-20.
type PgbouncerSettingsData struct{}

type PlanDetails struct {
Name string `json:"name"`
Version *PlanVersion `json:"version,omitempty"`
}

type PlanVersion struct {
Major int `json:"major"`
Minor int `json:"minor"`
}

// PreloadLibraries The shared libraries to preload into the project's compute instances.
type PreloadLibraries struct {
EnabledLibraries *[]string `json:"enabled_libraries,omitempty"`
UseDefaults *bool `json:"use_defaults,omitempty"`
}

// PresignRequest Options for the presigned URL. The `operation` selects upload (`PUT`)
// or download (`GET`); the remaining fields are optional.
type PresignRequest struct {
// ContentType The `Content-Type` to bind into the signed request. Only meaningful
// for `upload`: when set, the caller MUST send the same `Content-Type`
// header on the `PUT`, and the value is echoed back in the response
// `headers`. Ignored for `download`.
ContentType *string `json:"content_type,omitempty"`
// ExpiresInSeconds How long the presigned URL stays valid, in seconds. Defaults to 900
// (15 minutes); capped at 604800 (7 days).
ExpiresInSeconds *int64 `json:"expires_in_seconds,omitempty"`
// Operation The transfer direction. `upload` returns a presigned `PUT` URL;
// `download` returns a presigned `GET` URL.
Operation string `json:"operation"`
}

// PresignResponseHeaders Headers the caller MUST send verbatim on the request (e.g.
// `Content-Type` when it was signed on an upload). May be empty.
type PresignResponseHeaders struct{}

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
// Omitted when observed no activity for endpoints of this project.
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
DataTransferBytes int64 `json:"data_transfer_bytes"`
DefaultEndpointSettings *DefaultEndpointSettings `json:"default_endpoint_settings,omitempty"`
EffectiveProjectPermission *string `json:"effective_project_permission,omitempty"`
// HipaaEnabledAt A timestamp indicating when HIPAA was enabled for this project
HipaaEnabledAt *time.Time `json:"hipaa_enabled_at,omitempty"`
// HistoryRetentionSeconds The number of seconds to retain the shared history for all branches in this project.
HistoryRetentionSeconds int32 `json:"history_retention_seconds"`
// ID The project ID
ID string `json:"id"`
// MaintenanceScheduledFor A timestamp indicating when project update begins. If set, computes might experience a brief restart around this time.
MaintenanceScheduledFor *time.Time `json:"maintenance_scheduled_for,omitempty"`
// MaintenanceStartsAt A timestamp indicating when project maintenance begins. If set, the project is placed into maintenance mode at this time.
MaintenanceStartsAt *time.Time `json:"maintenance_starts_at,omitempty"`
// Name The project name
Name string `json:"name"`
OrgID *string `json:"org_id,omitempty"`
Owner *ProjectOwnerData `json:"owner,omitempty"`
OwnerID string `json:"owner_id"`
PgVersion PgVersion `json:"pg_version"`
// PlatformID The cloud platform identifier. Currently, only AWS is supported, for which the identifier is `aws`.
PlatformID string `json:"platform_id"`
Provisioner Provisioner `json:"provisioner"`
// ProxyHost The proxy host for the project. This value combines the `region_id`, the `platform_id`, and the Neon domain (`neon.tech`).
ProxyHost string `json:"proxy_host"`
// QuotaResetAt DEPRECATED. Use `consumption_period_end` from the getProject endpoint instead.
// A timestamp indicating when the project quota resets.
QuotaResetAt *time.Time `json:"quota_reset_at,omitempty"`
// RegionID The region identifier
RegionID string `json:"region_id"`
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

type ProjectAuditLogLevel string

const (
ProjectAuditLogLevelBase ProjectAuditLogLevel = "base"
ProjectAuditLogLevelExtended ProjectAuditLogLevel = "extended"
ProjectAuditLogLevelFull ProjectAuditLogLevel = "full"
)

type ProjectCreateRequest struct {
Project ProjectCreateRequestProject `json:"project"`
}

type ProjectCreateRequestProject struct {
AutoscalingLimitMaxCu *ComputeUnit `json:"autoscaling_limit_max_cu,omitempty"`
AutoscalingLimitMinCu *ComputeUnit `json:"autoscaling_limit_min_cu,omitempty"`
Branch *ProjectCreateRequestProjectBranch `json:"branch,omitempty"`
DefaultEndpointSettings *DefaultEndpointSettings `json:"default_endpoint_settings,omitempty"`
// HistoryRetentionSeconds The number of seconds to retain the shared history for all branches in this project.
// The default is 1 day (86400 seconds).
HistoryRetentionSeconds *int32 `json:"history_retention_seconds,omitempty"`
// Name The project name. If not specified, the name will be identical to the generated project ID
Name *string `json:"name,omitempty"`
// OrgID Organization id in case the project created belongs to an organization.
// If not present, project is owned by a user and not by org.
OrgID *string `json:"org_id,omitempty"`
PgVersion *PgVersion `json:"pg_version,omitempty"`
Provisioner *Provisioner `json:"provisioner,omitempty"`
// RegionID The region identifier. Refer to our [Regions](https://neon.com/docs/introduction/regions) documentation for supported regions. Values are specified in this format: `aws-us-east-1`
RegionID *string `json:"region_id,omitempty"`
Settings *ProjectSettingsData `json:"settings,omitempty"`
// StorePasswords Whether or not passwords are stored for roles in the Neon project. Storing passwords facilitates access to Neon features that require authorization.
StorePasswords *bool `json:"store_passwords,omitempty"`
}

type ProjectCreateRequestProjectBranch struct {
Annotations *AnnotationValueData `json:"annotations,omitempty"`
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
// Omitted when observed no activity for endpoints of this project.
ComputeLastActiveAt *time.Time `json:"compute_last_active_at,omitempty"`
// CpuUsedSec DEPRECATED. Use data from the getProject endpoint instead.
CpuUsedSec int64 `json:"cpu_used_sec"`
// CreatedAt A timestamp indicating when the project was created
CreatedAt time.Time `json:"created_at"`
// CreationSource The project creation source
CreationSource string `json:"creation_source"`
DefaultEndpointSettings *DefaultEndpointSettings `json:"default_endpoint_settings,omitempty"`
// DeletedAt A timestamp indicating when the project was deleted
DeletedAt *time.Time `json:"deleted_at,omitempty"`
EffectiveProjectPermission *string `json:"effective_project_permission,omitempty"`
// HipaaEnabledAt A timestamp indicating when HIPAA was enabled for this project
HipaaEnabledAt *time.Time `json:"hipaa_enabled_at,omitempty"`
// HistoryRetentionSeconds The number of seconds to retain the shared history for all branches in this project.
HistoryRetentionSeconds *int32 `json:"history_retention_seconds,omitempty"`
// ID The project ID
ID string `json:"id"`
// MaintenanceStartsAt A timestamp indicating when project maintenance begins. If set, the project is placed into maintenance mode at this time.
MaintenanceStartsAt *time.Time `json:"maintenance_starts_at,omitempty"`
// Name The project name
Name string `json:"name"`
// OrgID Organization id if the project belongs to an organization.
// Permissions for the project will be given to organization members as defined by the organization admins.
// The permissions of the project do not depend on the user that created the project if a project belongs to an organization.
OrgID *string `json:"org_id,omitempty"`
// OrgName Organization name if the project belongs to an organization.
OrgName *string `json:"org_name,omitempty"`
OwnerID string `json:"owner_id"`
PgVersion PgVersion `json:"pg_version"`
// PlatformID The cloud platform identifier. Currently, only AWS is supported, for which the identifier is `aws`.
PlatformID string `json:"platform_id"`
Provisioner Provisioner `json:"provisioner"`
// ProxyHost The proxy host for the project. This value combines the `region_id`, the `platform_id`, and the Neon domain (`neon.tech`).
ProxyHost string `json:"proxy_host"`
// QuotaResetAt DEPRECATED. Use `consumption_period_end` from the getProject endpoint instead.
// A timestamp indicating when the project quota resets
QuotaResetAt *time.Time `json:"quota_reset_at,omitempty"`
// RecoverableUntil A timestamp indicating the project will be recoverable until this date and time.
RecoverableUntil *time.Time `json:"recoverable_until,omitempty"`
// RegionID The region identifier
RegionID string `json:"region_id"`
Settings *ProjectSettingsData `json:"settings,omitempty"`
// StorePasswords Whether or not passwords are stored for roles in the Neon project. Storing passwords facilitates access to Neon features that require authorization.
StorePasswords bool `json:"store_passwords"`
// SyntheticStorageSize The current space occupied by the project in storage, in bytes. Synthetic storage size combines the logical data size and Write-Ahead Log (WAL) size for all branches in a project.
SyntheticStorageSize *int64 `json:"synthetic_storage_size,omitempty"`
// UpdatedAt A timestamp indicating when the project was last updated
UpdatedAt time.Time `json:"updated_at"`
}

type ProjectOwnerData struct {
BranchesLimit int `json:"branches_limit"`
Email string `json:"email"`
Name string `json:"name"`
SubscriptionType BillingSubscriptionType `json:"subscription_type"`
}

type ProjectPermission struct {
GrantedAt time.Time `json:"granted_at"`
GrantedToEmail string `json:"granted_to_email"`
ID string `json:"id"`
RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

type ProjectPermissions struct {
ProjectPermissions []ProjectPermission `json:"project_permissions"`
}

// ProjectQuota Per-project consumption quotas. If a quota is exceeded, all active computes
// are automatically suspended and cannot be started via API calls or incoming connections.
// 
// The exception is `logical_size_bytes`, which is enforced per branch.
// If a branch exceeds its `logical_size_bytes` quota, computes can still be started,
// but write operations will fail—allowing data to be deleted to free up space.
// Computes on other branches are not affected.
// 
// Setting `logical_size_bytes` overrides any lower value set by the `neon.max_cluster_size` Postgres setting.
// 
// Quotas are enforced using per-project consumption metrics with the same names.
// These metrics reset at the start of each billing period. `logical_size_bytes`
// is also an exception—it reflects the total data stored in a branch and does not reset.
// 
// A zero or empty quota value means “unlimited.”
type ProjectQuota struct {
// ActiveTimeSeconds The total amount of wall-clock time allowed to be spent by the project's compute endpoints.
ActiveTimeSeconds *int64 `json:"active_time_seconds,omitempty"`
// ComputeTimeSeconds The total amount of CPU seconds allowed to be spent by the project's compute endpoints.
ComputeTimeSeconds *int64 `json:"compute_time_seconds,omitempty"`
// DataTransferBytes Total amount of data transferred from all of a project's branches using the proxy.
DataTransferBytes *int64 `json:"data_transfer_bytes,omitempty"`
// LogicalSizeBytes Limit on the logical size of every project's branch.
// 
// If a branch exceeds its `logical_size_bytes` quota, computes can still be started,
// but write operations will fail—allowing data to be deleted to free up space.
// Computes on other branches are not affected.
// 
// Setting `logical_size_bytes` overrides any lower value set by the `neon.max_cluster_size` Postgres setting.
LogicalSizeBytes *int64 `json:"logical_size_bytes,omitempty"`
// WrittenDataBytes Total amount of data written to all of a project's branches.
WrittenDataBytes *int64 `json:"written_data_bytes,omitempty"`
}

type ProjectRecoverResponse struct {
BranchesResponse
ProjectResponse
}

type ProjectResponse struct {
Project Project `json:"project"`
}

type ProjectSettingsData struct {
AllowedIps *AllowedIps `json:"allowed_ips,omitempty"`
AuditLogLevel *ProjectAuditLogLevel `json:"audit_log_level,omitempty"`
// BlockPublicConnections When set, connections from the public internet
// are disallowed. This supersedes the AllowedIPs list.
// This parameter is under active development and its semantics may change in the future.
BlockPublicConnections *bool `json:"block_public_connections,omitempty"`
// BlockVpcConnections When set, connections using VPC endpoints are disallowed.
// This parameter is under active development and its semantics may change in the future.
BlockVpcConnections *bool `json:"block_vpc_connections,omitempty"`
// EnableLogicalReplication Sets wal_level=logical for all compute endpoints in this project.
// All active endpoints will be suspended.
// Once enabled, logical replication cannot be disabled.
EnableLogicalReplication *bool `json:"enable_logical_replication,omitempty"`
Hipaa *bool `json:"hipaa,omitempty"`
MaintenanceWindow *MaintenanceWindow `json:"maintenance_window,omitempty"`
PreloadLibraries *PreloadLibraries `json:"preload_libraries,omitempty"`
Quota *ProjectQuota `json:"quota,omitempty"`
}

type ProjectTransferRequestResponse struct {
// CreatedAt The timestamp when the transfer request was created
CreatedAt time.Time `json:"created_at"`
// ExpiresAt The timestamp when the transfer request will expire
ExpiresAt time.Time `json:"expires_at"`
// ID The unique identifier for the transfer request
ID string `json:"id"`
// ProjectID The ID of the project that is being transferred
ProjectID string `json:"project_id"`
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
Name *string `json:"name,omitempty"`
Settings *ProjectSettingsData `json:"settings,omitempty"`
}

// ProjectsApplicationsMapResponse A map where key is a project ID and a value is a list of installed applications.
type ProjectsApplicationsMapResponse struct {
Applications ProjectsApplicationsMapResponseApplications `json:"applications"`
}

type ProjectsApplicationsMapResponseApplications struct{}

type ProjectsIntegrationsMapResponseIntegrations struct{}

type ProjectsResponse struct {
Projects []ProjectListItem `json:"projects"`
// UnavailableProjectIDs A list of project IDs indicating which projects are known to exist, but whose details could not
// be fetched within the requested (or implicit) time limit
UnavailableProjectIDs *[]string `json:"unavailable_project_ids,omitempty"`
}

// Provisioner The Neon compute provisioner.
// Specify the `k8s-neonvm` provisioner to create a compute endpoint that supports Autoscaling.
// 
// Provisioner can be one of the following values:
// * k8s-pod
// * k8s-neonvm
// * serverless-platform
// 
// Clients must expect, that any string value that is not documented in the description above should be treated as a error. UNKNOWN value if safe to treat as an error too.
type Provisioner string

type RestoreSnapshotReqObj struct {
FinalizeRestore *bool `json:"finalize_restore,omitempty"`
Name *string `json:"name,omitempty"`
TargetBranchID *string `json:"target_branch_id,omitempty"`
}

type RestoredSnapshot struct {
BranchResponse
EndpointsOptionalResponse
OperationsResponse
}

type Role struct {
// AuthenticationMethod Authentication method configured for this role. Valid options: `password`, `oauth`, `no_login`
AuthenticationMethod *string `json:"authentication_method,omitempty"`
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
// NoLogin Whether to create a role that cannot login.
NoLogin *bool `json:"no_login,omitempty"`
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

type SendNeonAuthTestEmailRequest struct{}

type SendNeonAuthTestEmailResponse struct {
// ErrorMessage The error message from the email server.
ErrorMessage *string `json:"error_message,omitempty"`
// Success Whether the test email was sent successfully.
Success bool `json:"success"`
}

type Snapshot struct {
CreatedAt string `json:"created_at"`
// DiffSize Incremental storage size in bytes since the previous scheduled snapshot, when the snapshot is billed on incremental (diff) usage.
// 
// When absent, either the incremental size has not been calculated yet and the snapshot is not being charged, or the snapshot is charged at full logical size (in that case `full_size` is set).
DiffSize *int64 `json:"diff_size,omitempty"`
ExpiresAt *string `json:"expires_at,omitempty"`
// FullSize Full logical size of the snapshot in bytes at the time it was taken.
// 
// When absent, the logical size has not been calculated yet and the snapshot is not being charged.
// 
// When present, a value of 0 means the snapshot is not being charged.
FullSize *int64 `json:"full_size,omitempty"`
ID string `json:"id"`
Lsn *string `json:"lsn,omitempty"`
Manual *bool `json:"manual,omitempty"`
Name string `json:"name"`
SourceBranchID *string `json:"source_branch_id,omitempty"`
Timestamp *string `json:"timestamp,omitempty"`
}

type SnapshotSchedule struct {
BackupSchedule
}

type SnapshotUpdateRequest struct {
Snapshot SnapshotUpdateRequestSnapshot `json:"snapshot"`
}

type SnapshotUpdateRequestSnapshot struct {
// ExpiresAt The date and time when the snapshot will expire.
// 
// Omit to leave the current expiration unchanged. Send `null` to
// clear the expiration so the snapshot never expires. A future
// timestamp sets the absolute expiration.
ExpiresAt *time.Time `json:"expires_at,omitempty"`
Name *string `json:"name,omitempty"`
}

type SpendingLimitResponse struct {
// SpendingLimitCents Monthly spending cap in cents. `null` indicates that no limit
// is currently configured.
SpendingLimitCents int64 `json:"spending_limit_cents"`
}

type SpendingLimitUpdateRequest struct {
// SpendingLimitCents Monthly spending cap in cents. Must be positive. To remove a
// previously configured limit, send a DELETE request to the
// spending_limit endpoint — `0` and `null` are rejected here.
// The cap is alert-only: notifications fire at 80% and 100%, but
// computes are not suspended. Setting a cap below the period's
// already-accrued spend is permitted and will trigger the
// over-limit notification on the next worker run.
SpendingLimitCents int64 `json:"spending_limit_cents"`
}

// SuspendTimeoutSeconds Duration of inactivity in seconds after which the compute endpoint is
// automatically suspended. The value `0` means use the default value.
// The value `-1` means never suspend. The default value is `300` seconds (5 minutes).
// The minimum value is `60` seconds (1 minute).
// The maximum value is `604800` seconds (1 week). For more information, see
// [Scale to zero configuration](https://neon.com/docs/manage/endpoints#scale-to-zero-configuration).
type SuspendTimeoutSeconds int64

type TransferProjectsToOrganizationRequest struct {
// DestinationOrgID The destination organization identifier
DestinationOrgID string `json:"destination_org_id"`
// ProjectIDs The list of projects ids to transfer. Maximum of 400 project ids
ProjectIDs []string `json:"project_ids"`
}

type UpdateNeonAuthAllowLocalhostRequest struct {
// AllowLocalhost Whether to allow localhost connections
AllowLocalhost bool `json:"allow_localhost"`
}

type UpdateNeonAuthUserRoleRequest struct {
// Roles Array of roles to assign to the user
Roles []string `json:"roles"`
}

type UpdateNeonAuthUserRoleResponse struct {
// ID of the updated user
ID string `json:"id"`
}

type UpdateProjectRespObj struct {
OperationsResponse
ProjectResponse
}

type UpdateSnapshotRespObj struct {
Snapshot
}

type VPCEndpoint struct {
// Label A descriptive label for the VPC endpoint
Label string `json:"label"`
// VpcEndpointID The VPC endpoint ID
VpcEndpointID string `json:"vpc_endpoint_id"`
}

type VPCEndpointAssignment struct {
Label string `json:"label"`
}

type VPCEndpointDetails struct {
// ExampleRestrictedProjects A list of example projects that are restricted to use this VPC endpoint.
// There are at most 3 projects in the list, even if more projects are restricted.
ExampleRestrictedProjects []string `json:"example_restricted_projects"`
// Label A descriptive label for the VPC endpoint
Label string `json:"label"`
// NumRestrictedProjects The number of projects that are restricted to use this VPC endpoint.
NumRestrictedProjects int `json:"num_restricted_projects"`
// State The current state of the VPC endpoint. Possible values are
// `new` (just configured, pending acceptance) or `accepted`
// (VPC connection was accepted by Neon).
State string `json:"state"`
// VpcEndpointID The VPC endpoint ID
VpcEndpointID string `json:"vpc_endpoint_id"`
}

type VPCEndpointWithRegion struct {

VPCEndpoint
}

type VPCEndpointsResponse struct {
Endpoints []VPCEndpoint `json:"endpoints"`
}
