# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.13.0] - 2025-02-19

### Added

- Added the following methods to manage Neon integration with third-party authentication providers:
  - `CreateProjectIdentityAuthProviderSDKKeys`;
  - `CreateProjectIdentityIntegration`;
  - `DeleteProjectIdentityIntegration`;
  - `ListProjectIdentityIntegrations`;
  - `TransferProjectIdentityAuthProviderProject`.

## [v0.12.0] - 2025-02-17

### Added

- Added the method `GetProjectBranchSchemaComparison` to compare the schema of the database in one branch to the schema of that database from another branch.   
- Added the method `CountProjectBranches` to retrieve the total number of branches in the project.
- Added the methods to manage VPC endpoints:
  - `AssignOrganizationVPCEndpoint`: assigns a VPC endpoint to the project in the organization.
  - `GetOrganizationVPCEndpointDetails`: reads details about the VPC endpoint assigned to the project in the organization.
  - `DeleteOrganizationVPCEndpoint`: deletes the VPC endpoint for the project in the organization.
  - `ListOrganizationVPCEndpoints`: lists VPC endpoints for the organization.
  - `AssignProjectVPCEndpoint`: sets, or updates VPC endpoint restrictions for the project.
  - `ListProjectVPCEndpoints`: lists VPC endpoints restrictions for the project.
- Added support for sorting and pagination for the method `ListProjectBranches` to improve branches listing experience.
- Added explicit timeout to limit the delay when listing the projects using the methods `ListProjects` and `ListSharedProjects`.
- Added the method `TransferProjectsFromOrgToOrg` to migrate the project to the organization.
- Added the field `InitSource` to the type `BranchCreateRequestBranch`.
- Added the field `MaintenanceScheduledFor` to the type `Project` to indicate when the project's computes might be restarted.
- Added the field `NoLogin` to the type `RoleCreateRequestRole` to create roles without login permissions.
- Added the field `UnavailableProjectIDs` to the type `ProjectsResponse` to indicate existing projects which details could not be retrieved within the set time scope of a single request. 
- Added the following `OperationAction` categories:
  - `OperationActionApplySchemaFromBranch`: apply_schema_from_branch;
  - `OperationActionImportData`: "apply_schema_from_branch".
- Added the following `IdentityProviderId` category:
  - `IdentityProviderIdMicrosoftv2`: "microsoftv2"

### Changed

- **[BREAKING]** Changed the type `AddProjectJWKSRequest`: the field's type `RoleNames` is the slice instead of the pointer to a slice.

### Removed

- **[BREAKING]** Removed the field `SchemaInitializationType` from the type `BranchCreateRequestBranch`. 

## [v0.11.0] - 2024-12-08

The release incorporates the up-to-date [API contract](openAPIDefinition.json) as of 2024-12-08 10:35:00 GMT.

### Added

- Added the attribute `Archived` to the type `BranchCreateRequestBranch` to create archived branches.

### Removed

- **[BREAKING]** Removed the attribute `ProtectedBranchOnly` from the type `AllowedIps`. 

## [v0.10.0] - 2024-11-21

The release incorporates the up-to-date [API contract](openAPIDefinition.json) as of 2024-11-21 00:12:18 GMT.

### Added

- Added the `MaintenanceWindow` attribute of the `ProjectSettingsData` type to configure the Neon project's
  maintenance window.

## [v0.9.0] - 2024-11-09

The release incorporates the up-to-date [API contract](openAPIDefinition.json) as of 2024-11-09 11:12:00 GMT.

### Added

- Added the methods to manipulate JWKS for the project: 
  - `AddProjectJWKS`: adds the JWKS URL to the project.
  - `GetProjectJWKS`: retrieve all JWKS URLs configured for the project.
  - `DeleteProjectJWKS`: deletes the JWKS URL from the project.

- Added the method `GetOrganization` to read the organization's info

- Added the methods to manage the organization's membership invitations:
  - `GetOrganizationInvitations`: retrieves the memberships invitations for the organization. 
  - `GetOrganizationMember`: reads the info about the organization's members.
  - `GetOrganizationMembers`: retrieves the info about all organization's members.

- Added Vercel as the identity provider's enum value.

### Changed

- **[BREAKING]** Changed the type of the attributes `Primary` in the `Branch` struct to the pointer to bool from bool.

## [v0.8.0] - 2024-10-28

The release incorporates the up-to-date [API contract](openAPIDefinition.json) as of 2024-10-28 22:27:00 GMT.

### Added

- **[BREAKING]** Added the argument `search` to the method `ListProjectBranches` to list branches by name or id.
  A partial name or id values may be specified to filter results.
- Added the property `BillingAccount` of the type `CurrentUserInfoResponse`: added the `State` to define the 
  billing account's state.
- Added the attribute `CreatedBy` to the type `ApiKeysListResponseItem` to define the user who created the API key. 
- Added the endpoint `GetActiveRegions` to list active regions to provision databases.

## [v0.7.0] - 2024-10-20

The release incorporates the up-to-date [API contract](openAPIDefinition.json) as of 2024-10-20 11:15:00 GMT.

### Added

- Added the attributes to reflect when and who created, or revoked API keys:
  - Added `CreatedBy` to `ApiKeyCreateResponse` and `ApiKeysListResponseItem`;
  - Added `CreatedAt` and `CreatedBy` to `ApiKeyRevokeResponse`.

- Added the attribute `StateChangedAt` to the type `Branch` to reflect when the "current_state" began.

### Removed

- **[BREAKING]** Removed the method to list projects consumption `ListProjectsConsumption`.

## [v0.6.2] - 2024-10-08

The release incorporates the up-to-date [API contract](openAPIDefinition.json) as of 2024-10-08 00:11:00 GMT.

### Added

- Added the mapping of the enum value "archived" to the type `BranchState`.
- Added the mapping of the enum values "timeline_archive" and "timeline_unarchive" to the type `OperationAction`.

## [v0.6.1] - 2024-10-01

The release incorporates the up-to-date [API contract](openAPIDefinition.json) as of 2024-10-01 06:44:00 GMT.

### Added

- Added "azure_mp" as a `BillingPaymentMethod` enum value.

### Changed

- Changed the response type of the method `ListProjects`, `ListProjectsRespObj`. It includes the map 
  `ProjectsApplicationsMapResponse` which contains all applications available in the project.

### Deprecated

- Deprecated `Login` field in the types `CurrentUserAuthAccount` and `CurrentUserInfoResponse`.

## [v0.6.0] - 2024-09-28

The release incorporates the up-to-date [API contract](openAPIDefinition.json) as of 2024-09-28 21:53:00 GMT.

### Added

- Added the method `TransferProjectsFromUserToOrg` to migrate personal projects to organisation.
- Added the payment method attribute to the struct `BillingAccount`.
- Added `Business` subscription type's value.
- Added the attribute `CreatedBy` to the struct `Branch` to indicate who created the branch.

### Changed

- **[BREAKING]** Change the signatures of the methods:
  - `ListProjectBranches`;
  - `CreateProjectBranch`;
  - `GetProjectBranch`;
  - `GetProjectBranchSchema`.

## [v0.5.0] - 2024-06-24

The release incorporates the up-to-date [API contract](openAPIDefinition.json) as of 2024-06-24 22:03:00 GMT.

### Added

- Added support of the following ENUMs as type aliases:
  - The type `BillingSubscriptionType` as the attribute of the response of the method `GetCurrentUserInfo`.
  - The type `BranchState` as attributes of the response of the method `GetProjectBranch`.
  - The type `ConsumptionHistoryGranularity` as the argument of the method `GetConsumptionHistoryPerAccount` and `GetConsumptionHistoryPerProject`.
  - The types `EndpointPoolerMode`, `EndpointState` as the attributes of the struct `Endpoint`, which defines the 
    response of the method `GetProjectEndpoint`.
  - The type `EndpointType` as the attribute which defines the endpoint's type to create an endpoint, or define the 
    options of the branch's endpoints.
  - The type `IdentityProviderId` as the attribute of the struct `CurrentUserAuthAccount` which defines the response
    of the method `GetCurrentUserInfo`.
  - The types `OperationAction` and `OperationStatus` as the attributes of the struct `Operation` which defines the 
    response of several endpoints which include the operations.
  - The type `Provisioner` which defines the Neon compute provisioner's type.
- Added the method `GetProjectBranchSchema` to retrieve the database schema, see details [here](https://api-docs.neon.tech/reference/getprojectbranchschema).
- Added the methods to retrieve the consumption metrics: 
  - `GetConsumptionHistoryPerAccount` allows to read the account's consumption history, see details [here](https://api-docs.neon.tech/reference/getconsumptionhistoryperaccount).
  - `GetConsumptionHistoryPerProject` allows to read the consumption history for a list of projects, see details [here](https://api-docs.neon.tech/reference/getconsumptionhistoryperproject).
- Added the method `GetCurrentUserOrganizations` to read all organization which a given user belongs to.
- Added support of the organization ID (`orgID` argument) when using the following methods:
  - `ListProjectsConsumption`, see details [here](https://api-docs.neon.tech/reference/listprojectsconsumption).
- Added the name, the address and the tax information to the billing details of the account: `BillingAccount` struct.

### Changed

- All arguments which end with the suffices Id/Ids, Url/Urls, Uri/Uris will follow the Go convention. For example,
  the query parameter `project_ids` will correspond to the method's argument `projectIDs`.

### Deprecated

- The method `SetPrimaryProjectBranch` is deprecated, please use the method `SetDefaultProjectBranch` instead.
- The label "primary" branch and the attributes `Primary` is deprecated for the label "default" and the respective
  attribute `Default`. See the struct `Branch` for example.
- The attribute `ProxyHost` of the struct `Endpoint` is deprecated, please use the attribute `Host` instead.
- The attribute `CpuUsedSec` of the structs `Project` and `ProjectListItem` is deprecated, 
  please use the attribute `ComputeTimeSeconds` instead.
- The attribute `QuotaResetAt` of the structs `Project` and `ProjectListItem` is deprecated, 
  please use the attribute `ConsumptionPeriodEnd` instead.

## [v0.4.9] - 2024-04-13

The release incorporates the up-to-date [API contract](openAPIDefinition.json) as of 2024-04-13 11:00:00 GMT.

### Added

- Added the filtering argument `orgID` to the method `ListProjects` to enhance the projects listing functionality. 
- Added the method `RestartProjectEndpoint` to restart the project's endpoint. Find details [here](https://api-docs.neon.tech/reference/restartprojectendpoint).
- Added the filter `ProtectedBranchesOnly` provision the list of allowed IP addresses only for the protected branches.
- Added the field `ComputeReleaseVersion` to the struct `Endpoint` to reflect the version of the compute resources.

### Changed

- Changed the type of the `HistoryRetentionSeconds` to `int32` attribute.

### Fixed

- Fixed the method `GetConnectionURI` by correcting the logic of building the request query.

## [v0.4.8] - 2024-03-22

The release incorporates the up-to-date [API contract](openAPIDefinition.json) as of 2024-03-21 23:20:00 GMT.

### Added

- Added the method `GetConnectionURI` to retrieve the connection URI for the given role and database, see [details](https://api-docs.neon.tech/reference/getconnectionuri).
- The branch can be set as protected at creation, or update using the attribute `Protected`.

## [v0.4.7] - 2024-03-07

The release incorporates the up-to-date [API contract](openAPIDefinition.json) as of 2024-03-05 00:08:00 GMT.

### Added

- Added the `RestoreProjectBranch` method to restore the branch to a specified state: [details](https://api-docs.neon.tech/reference/restoreprojectbranch).

## [v0.4.6] - 2024-02-26

The release incorporates the up-to-date [API contract](openAPIDefinition.json) as of 2024-02-25 00:08:00 GMT.

### Removed

- [**BREAKING**] Removed the method `VerifyUserPassword`.

## [v0.4.5] - 2024-02-22

The release incorporates the up-to-date [API contract](openAPIDefinition.json) as of 2024-02-22 00:08:00 GMT.

### Changed

- [**BREAKING**] The signature of the method `VerifyUserPassword` includes the password sent for verification.

## [v0.4.4] - 2024-02-18

The release incorporates the up-to-date [API contract](openAPIDefinition.json) as of 2024-02-18 11:14:00 GMT.

### Added

- [**BREAKING**] Added the `search` argument to the signature of the methods `ListProjects` and `ListSharedProjects`. 
It allows to filter the search results by the project name or ID (see [more](https://api-docs.neon.tech/reference/listprojects)).
- Added the methods [`GetCurrentUserAuthInfo`](https://api-docs.neon.tech/reference/getcurrentuserinfo) and 
[`VerifyUserPassword`](https://api-docs.neon.tech/reference/verifyuserpassword) to query the user identity.

### Fixed

- Trailing "?"-sign for empty request queries.

## [v0.4.3] - 2024-01-19

The release incorporates the up-to-date [API contract](openAPIDefinition.json) as of 2024-01-19 13:00:00 GMT.

### Added

- Added support to manage the project's permissions:
  - List permissions: `ListProjectPermissions`;
  - Grant project's permissions to the user: `GrantPermissionToProject`;
  - Revoke project's permissions from the user: `RevokePermissionFromProject`.

### Changed

- [**BREAKING**] Changed the type of the attribute `Ips` in the struct `AllowedIps`.

## [v0.4.2] - 2024-01-11

The release incorporates the up-to-date [API contract](openAPIDefinition.json) as of 2024-01-11 05:27:00 GMT.

### Added

- Added support to list the projects shared with the account. Find more [here](https://neon.tech/docs/manage/projects#share-a-project).

## [v0.4.1] - 2023-12-22

The release incorporates the up-to-date [API contract](openAPIDefinition.json) as of 2023-12-22 01:09:10 GMT.

### Added

- Activation of the data [logical replication](https://neon.tech/docs/introduction/logical-replication) for CDC.

## [v0.4.0] - 2023-12-21

### Changed

- [**BREAKING**] All optional fields present in the structs of the API request and response payloads are now defined as pointers.

## [v0.3.2] - 2023-12-20

The release incorporates the up-to-date [API contract](openAPIDefinition.json) as of 2023-12-20 10:37:20 GMT.

### Added

- Project level configuration of IP addresses allowed to connect to the endpoints.
- `LastResetAt` attribute indicating the last time when a branch was reset. 

## [v0.3.1] - 2023-11-02

The release incorporates the up-to-date [API contract](openAPIDefinition.json) as of 2023-11-02 13:54:20 GMT.

### Added

- [**BREAKING**] `ListProjectsConsumption` has the arguments "from" and "to" to filter by the billing period time now.
- [**BREAKING**] `ProjectConsumption` contains the billing period's information including the period's start and end timestamps now.
- `ProjectsConsumptionResponse` contains the number of billing periods `PeriodsInResponse` now.

### Deleted

- [**BREAKING**] `ProjectConsumption` does not contain the attributes `ComputeLastActiveAt` and `ID` now.

### Fixed

- [Linting] Private method `requestHandler` is assigned to the `Client` instead of a pointer to it.

## [v0.3.0] - 2023-10-26

### Changed

- The interface `Client` is removed and the output of the `NewClient` is the pointer to the struct now. It will facilitate stub of the SDK client.
- [**BREAKING**] `NewClient` requires `Config` to initialise a Neon client. It's meant to improve security by eliminating
  support of environment variables for authentication by default. It also simplifies the codebase.
  
  **Example**
  ```go
  package main

  import (
        "log"
  
        neon "github.com/kislerdm/neon-sdk-go"
  )
  
  func main() {
        client, err := neon.NewClient(neon.Config{Key: "{{.NeonApiKey}}"})
        if err != nil {
            panic(err)
        }
  
        v, err := client.ListProjects()
        if err != nil {
            panic(err)
        }
  
        log.Printf("%d projects found", len(v.Projects))
  }
  ```
- [**BREAKING**] Removed support of variadic functions used to configure SDK client.

## [v0.2.5] - 2023-10-22

The release incorporates the up-to-date [API contract](openAPIDefinition.json) as of 2023-10-11 00:08:16 GMT.

### Fixed

- [**BREAKING**] Method `UpdateProject` returns the object of the type `UpdateProjectRespObj` (combination of `OperationsResponse` and `ProjectResponse`) now.
  **Note** that it reverts corresponding change made in [v0.2.4](#v024---2023-09-29).

### Changed

- The struct `ApiKeyCreateResponse` (the response type of the method `CreateApiKey`) contains the attributes `CreatedAt` and `Name` now.

## [v0.2.4] - 2023-09-29

The release incorporates the up-to-date [API contract](openAPIDefinition.json) as of 2023-09-29 00:08:00 GMT.

### Changed

- [**BREAKING**] Method `UpdateProject` returns the object of the type `ProjectResponse` now
- The struct `CurrentUserInfoResponse` (the response type of the method `GetCurrentUserInfo`) contains the attribute `LastName` now

## [v0.2.3] - 2023-09-21

The release incorporates the up-to-date [API contract](openAPIDefinition.json) as of 2023-09-21 01:30:00 GMT.

### Added

- Method `GetCurrentUserInfo` to retrieve information about the current user.

## [v0.2.2] - 2023-09-19

The release incorporates the up-to-date [API contract](openAPIDefinition.json) as of 2023-09-14 00:08:27 GMT.

### Fixed

- ([#26](https://github.com/kislerdm/neon-sdk-go/issues/26)) Removed temporary workaround to map invalid
  "200 response" which used to be returned by the Neon API for non existed resources (find
  details [here](https://github.com/neondatabase/neon/issues/2159)).
  The rationale:
    - The workaround logic relies on the header `Content-Length`, but the server response _may not contain_ that header
      according to the [RFC#7230](https://datatracker.ietf.org/doc/html/rfc7230#section-3.3.2):
      > [...] Aside from the cases defined above, in the absence of Transfer-Encoding, an origin server SHOULD send a
      Content-Length header field when the payload body size is known prior to sending the complete header section.[...]
    - It was noticed that the workaround leads to unexpected behaviour for the method `ListProject`: it returns the
    error "object not found" for the API response which contains a non-empty list of projects. The issue appeared when 
    such list contains more than a couple of objects.
    - The reason caused implementation of the workaround was _presumably_ resolved on the Neon API side.

## [v0.2.1] - 2023-07-30

The release incorporates the up-to-date [API contract](openAPIDefinition.json) as of 2023-07-29 23:16:06 GMT.

### Changed

- The response type `CreatedBranch` of the method `CreateProjectBranch`.
- The type `Operation`: `TotalDurationMs` attribute added to reflect duration of the operation.

## [v0.2.0] - 2023-07-26

The release incorporates the up-to-date [API contract](openAPIDefinition.json) as of 2023-07-26 11:57:20 GMT.

### Added

- Pagination support for list methods
- [Preview] Method to query consumption metrics for every
  project ([details](https://api-docs.neon.tech/reference/listprojectsconsumption)):
    - `ListProjectsConsumption`

### Fixed

- ([#15](https://github.com/kislerdm/neon-sdk-go/issues/15)) Autoscaling limits are defined using floating point units
- ([#9](https://github.com/kislerdm/neon-sdk-go/issues/9)) The project's quota data struct is collocated with the
  project's settings data struct.

### Changed

- Module's documentation
- Order of the data structs' attributes, they will be sorted alphabetically

### Removed

- Support of go versions below 1.18

## [v0.1.4] - 2023-01-08

### Fixed

- Fixed the definition of the optional field `Settings` for the following types:
    - `EndpointUpdateRequestEndpoint`
    - `EndpointCreateRequestEndpoint`

_The reason_: to omit the JSON field `settings` when serialising the above types' objects when the field `Settings`
is `nil`.

## [v0.1.3] - 2023-01-06

### Fixed

- Fixed handling optional request body
- Fixed test for `CreateProjectBranch`
- Fixed definition of the _request_ data types with optional fields of the type `time.Time`

## [v0.1.2] - 2023-01-04

### Fixed

- Type `PgSettingsData` is defined as `map[string]interface{}` instead of empty `struct`.

### Changed

- Fixed the way environment variables are set in the tests: `t.Setenv` instead of `os.Setenv`.
- Reduced the [cyclomatic complexity](https://en.wikipedia.org/wiki/Cyclomatic_complexity) to be below 16.

## [v0.1.1] - 2023-01-04

### Fixed

- Aligned the SDK methods and types definitions with the up-to-date [API spec](https://neon.tech/api-reference/v2)

## [v0.1.0] - 2023-01-03

### Added

- Methods compliant with the [API V2](https://neon.tech/api-reference/v2/)
- Only required path parameters mapped onto the SDK methods
- SDK authentication using [variadic function](https://gobyexample.com/variadic-functions) and environment
  variable `NEON_API_KEY`. The evaluation order:
    1. Variadic function client's argument;
    2. Environment variables.
- Mock for the HTTP client:

```go
package main

import (
	"log"

	neon "github.com/kislerdm/neon-sdk-go"
)

func main() {
	client, err := neon.NewClient(neon.WithHTTPClient(neon.NewMockHTTPClient()))
	if err != nil {
		panic(err)
	}

	v, err := client.ListProjects()
	if err != nil {
		panic(err)
	}

	log.Printf("%d projects found", len(v.Projects))
}
```
