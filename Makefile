.DEFAULT_GOAL := help

help: ## Prints help message.
	@ grep -h -E '^[a-zA-Z0-9_-].+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[1m%-30s\033[0m %s\n", $$1, $$2}'

tests: ## Run tests.
	@ go test -timeout 3m --tags=unittest -v -coverprofile=.coverage.out ./... -coverpkg=./...
	@ go tool cover -func .coverage.out && rm .coverage.out
