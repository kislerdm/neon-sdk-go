# End-to-end example

The example illustrates how to create a Neon project using the Neon Go SDK and to query the project's default database.

## Prerequisites

- [go](https://go.dev/dl/)
- [API Key](https://neon.tech/docs/manage/api-keys/)

## Steps

Follow the steps to run the example:

1. Download the module's dependencies:

```commandline
go mod download
```

2. Export the Neon API key to the process environment:

```commandline
export NEON_API_KEY=##YOU-API-KEY##
```

3. Compile and run the application:

```commandline
go run main.go
```

It's expected to see the current UTC timestamp printed to the stdout, for example:

```commandline
current UTC timestamp from database: 2024-12-08 11:08:46.67138 +0000 UTC
```
