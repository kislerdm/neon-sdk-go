# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.1]

### Fixed

- Aligned the SDK methods and types definitions with the up-to-date [API spec](https://neon.tech/api-reference/v2)

## [v0.1.0] - 2022-01-03

### Added

- Methods compliant with the [API V2](https://neon.tech/api-reference/v2/)
- Only required path parameters mapped onto the SDK methods
- SDK authentication using [variadic function](https://gobyexample.com/variadic-functions) and environment variable `NEON_API_KEY`. The evaluation order:
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
