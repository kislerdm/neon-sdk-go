# Go SDK for Neon Postgres SaaS Platform

[![logo](fig/logo.png)](https://neon.tech)

[![Go Report Card](https://goreportcard.com/badge/github.com/kislerdm/neon-sdk-go)](https://goreportcard.com/report/github.com/kislerdm/neon-sdk-go)
[![codecov](https://codecov.io/gh/kislerdm/neon-sdk-go/branch/master/graph/badge.svg?token=F6SF7VX3G3)](https://codecov.io/gh/kislerdm/neon-sdk-go)

- [How to use](#how-to-use)
    + [Prerequisites](#prerequisites)
    + [Installation](#installation)
    + [Code Snippets](#code-snippets)
        - [Authentication with the Neon Platform](#authentication)
            * [Variadic Function](#variadic-function)
            * [Environment Variables Evaluation](#environment-variables-evaluation)
        - [Mock](#mock)
- [Development](#development)
    + [API V2](#api-v2)
- [Contribution](#contribution)

The SDK to manage [Neon Platform](https://neon.tech) programmatically.

> Neon is a fully managed serverless PostgreSQL with a generous free tier. Neon separates storage and compute and offers
> modern developer features such as serverless, branching, bottomless storage, and more. Neon is open source and written
> in Rust.

Find more about Neon [here](https://neon.tech/docs/introduction/about/).

## How to use

### Prerequisites

- [go ~> 1.17](https://go.dev/dl/)
- [API Key](https://neon.tech/docs/manage/api-keys/)

### Installation

Add the SDK as a module dependency:

```commandline
go get github.com/kislerdm/neon-sdk-go
```

Run to specify the release version:

```commandline
go get github.com/kislerdm/neon-sdk-go@{{.Ver}}
```

Where `{{.Ver}}` is the release version.

### Code Snippets

#### Authentication

Authentication with the Neon Platform is implemented
using [variadic functions](https://gobyexample.com/variadic-functions) and environment variables evaluation in the
following order:

1. Variadic function client's argument;
2. Environment variable `NEON_API_KEY`.

Note that if the API key is provided as the variadic function argument, key from the environment variable `NEON_API_KEY`
will be ignored.

##### Variadic Function

```go
package main

import (
	"log"

	neon "github.com/kislerdm/neon-sdk-go"
)

func main() {
	client, err := neon.NewClient(neon.WithAPIKey("{{.NeonApiKey}}"))
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

##### Environment Variables Evaluation

**_Requirement_**: a valid Neon [API key](https://neon.tech/docs/manage/api-keys/) must be exported as the environment variable `NEON_API_KEY`.

```go
package main

import (
	"log"

	neon "github.com/kislerdm/neon-sdk-go"
)

func main() {
	client, err := neon.NewClient()
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

#### Mock

The SDK provides the http client's mock for unit tests. An example snippet is shown below.

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

## Development

The SDK codebase is generated using the [OpenAPI](https://spec.openapis.org/) from
the [API reference page](https://neon.tech/api-reference/v2/). The generator application codebase can be
found [here](generator).

### API V2

The SDK supports V2 only because V1 has been deprecated.

## Contribution

The SDK is distributed under the [MIT license](LICENSE). Please feel free to open an issue ticket, or PR to contribute.
