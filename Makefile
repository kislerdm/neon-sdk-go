.DEFAULT_GOAL := help

DIR := $(PWD)
APP := `basename $(DIR)`
OS := `uname | tr '[:upper:]' '[:lower:]'`
ARCH := `uname -m`

.PHONY: help
help: ## Prints help message.
	@ grep -h -E '^[a-zA-Z0-9_-].+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[1m%-30s\033[0m %s\n", $$1, $$2}'

PATH_SPEC := $(PWD)/openAPIDefinition.json
PATH_SDK := $(PWD)

.PHONY: generate-sdk
generate-sdk: ## Generates the SDK codebase using code generator.
	@ cd generator && \
		go mod tidy && \
		CGO_ENABLED=0 go run cmd/main.go --output $(PATH_SDK) --input $(PATH_SPEC)

.PHONY: tests
tests: go.mod.test ## Run tests.
	@ cd $(DIR) && \
 		cp go.mod go.mod.dump && cp go.mod.test go.mod && \
 		go mod tidy && \
  		go test -timeout 3m --tags=unittest -v -coverprofile=.coverage.out . -coverpkg=. && \
		go tool cover -func .coverage.out && rm .coverage.out &&\
		mv go.mod.dump go.mod

.PHONY: build
build: ## Compiles the binary.
	@ cd $(DIR) && \
 		test -d bin || mkdir -p bin && \
 		go mod tidy && \
  		CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build -o bin/$(APP)-$(OS)-$(ARCH) -ldflags="-s -w" ./cmd/main.go

.PHONY: testacc
testacc: ## Runs smoke tests.
	@ source .env && TF_ACC=1 go test acc_test.go

# official
SPEC_URL := https://neon.tech/api_spec/release/v2.json

# from James, VP of Eng. from Neon
SPEC_URL := https://d216pytvakpmhr.cloudfront.net/api_spec/main/v2.json

.PHONY: fetch-specs
fetch-specs: ## Downloads API specs.
	@ curl -SLo openAPIDefinition_new.json $(SPEC_URL)
