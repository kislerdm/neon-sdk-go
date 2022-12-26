.DEFAULT_GOAL := help

DIR := $(PWD)
APP := `basename $(DIR)`
OS := `uname | tr '[:upper:]' '[:lower:]'`
ARCH := `uname -m`


help: ## Prints help message.
	@ grep -h -E '^[a-zA-Z0-9_-].+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[1m%-30s\033[0m %s\n", $$1, $$2}'

tests: ## Run tests.
	@ cd $(DIR) && \
 		go mod tidy && \
  		go test -timeout 3m --tags=unittest -v -coverprofile=.coverage.out ./... -coverpkg=./... && \
		go tool cover -func .coverage.out && rm .coverage.out

build: ## Compiles the binary.
	@ cd $(DIR) && \
 		test -d bin || mkdir -p bin && \
 		go mod tidy && \
  		CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build -o bin/$(APP)-$(OS)-$(ARCH) -ldflags="-s -w" ./cmd/main.go
