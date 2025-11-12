.PHONY: build test lint run-local run-remote docker-build clean help

BINARY_NAME=echo-mcp
DOCKER_IMAGE=echo-mcp:latest

help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the echo-mcp binary
	go build -o $(BINARY_NAME) ./cmd/echo-mcp

test: ## Run all tests
	go test -v -race -coverprofile=coverage.out ./...

lint: ## Run linters
	golangci-lint run
	gofmt -l -s .

run-local: build ## Run server in local mode (stdio)
	./$(BINARY_NAME) --mode=local

run-remote: build ## Run server in remote mode (HTTP on port 8080)
	./$(BINARY_NAME) --mode=remote --port=8080

docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE) .

clean: ## Remove build artifacts
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	rm -rf bin/ dist/

coverage: test ## Generate HTML coverage report
	go tool cover -html=coverage.out -o coverage.html

mod-download: ## Download Go module dependencies
	go mod download
	go mod tidy

all: lint test build ## Run lint, test, and build
