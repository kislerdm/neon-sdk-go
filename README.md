# Go SDK for Neon Postgres SaaS Platform

[![logo](fig/logo.png)](https://neon.tech)

[![Go Report Card](https://goreportcard.com/badge/github.com/kislerdm/neon-sdk-go)](https://goreportcard.com/report/github.com/kislerdm/neon-sdk-go)
[![codecov](https://codecov.io/gh/kislerdm/neon-sdk-go/branch/master/graph/badge.svg?token=F6SF7VX3G3)](https://codecov.io/gh/kislerdm/neon-sdk-go)
[![Licenses Check](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fkislerdm%2Fneon-sdk-go.svg?type=small)](https://app.fossa.com/reports/fcbd29f3-1d63-4437-9946-cb320a567c42)

- [How to use](#how-to-use)
    + [Prerequisites](#prerequisites)
    + [Installation](#installation)
    + [Code Snippets](#code-snippets)
        - [Default HTTP Client](#default-http-client)
        - [Custom HTTP Client](#custom-http-client)
        - [Mock](#mock)
- [Development](#development)
  + [Commands](#commands)
- [Contribution](#contribution)

The SDK to manage [Neon Platform](https://neon.tech) programmatically.

> Neon is a fully managed serverless PostgreSQL with a generous free tier. Neon separates storage and compute and offers
> modern developer features such as serverless, branching, bottomless storage, and more. Neon is open source and written
> in Rust.

Find more about Neon [here](https://neon.tech/docs/introduction/about/).

## How to use

### Prerequisites

- [go ~> 1.18](https://go.dev/dl/)
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

### Default HTTP Client

The following snippet demonstrates how to initialize SDK which will use default HTTP client. 

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

### Custom HTTP client

The SDK can initialized with a custom HTTP client.

```go
package main

import (
	"net/http"
	"log"
	"time"

	neon "github.com/kislerdm/neon-sdk-go"
)

func main() {
	myHTTPClient := &http.Client{Timeout: 30 * time.Second}

	client, err := neon.NewClient(neon.Config{Key: "{{.NeonApiKey}}", HTTPClient: myHTTPClient})
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
	client, err := neon.NewClient(neon.Config{HTTPClient: neon.NewMockHTTPClient()})
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

### Commands

**Prerequisites**:

- go ~> 1.18
- gnuMake / cmake

Run to see all available commands:

```commandline
make help
```

Run to generate the SDK codebase and store it to PWD, given the OpenAPI spec is available in the
file [`openAPIDefinition.json`](openAPIDefinition.json):

```commandline
make generate-sdk
```

Run to customise the locations:

```commandline
make generate-sdk PATH_SDK=##/PATH/TO/OUTPUT/SDK/CODE## PATH_SPEC=##/PATH/TO/SPEC.json##
```

Run to test generated SDK:

```commandline
make tests
```

Run to test generated SDK stored to `/PATH/TO/OUTPUT/SDK/CODE`:

```commandline
make tests DIR=/PATH/TO/OUTPUT/SDK/CODE
```

Run to test the [code generator](generator):

```commandline
make tests DIR=generator
```

Run to build the [code generator](generator):

```commandline
make build DIR=generator
```

## Contribution

The SDK is distributed under the [MIT license](LICENSE), find full list of dependencies'
licenses [here](https://app.fossa.com/reports/fcbd29f3-1d63-4437-9946-cb320a567c42).

Please feel free to open an issue ticket, or PR to contribute.
