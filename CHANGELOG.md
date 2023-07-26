# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
