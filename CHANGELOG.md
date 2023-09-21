# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
