.PHONY: help build-examples lint fmt fmt-check clean deps deps-update security-scan docs

# Variables
APP_NAME := ipc_plugin_examples
APP_PLUGIN := plugin
APP_HOST := host
APP_MANAGER := manager

# Help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build
build-examples: ## Build the application
	@echo "Building $(APP_NAME)..."
	CGO_ENABLED=0 GOOS=linux go build -o bin/$(APP_PLUGIN) examples/plugin/main.go
	CGO_ENABLED=0 GOOS=linux go build -o bin/$(APP_HOST) examples/host/main.go
	CGO_ENABLED=0 GOOS=linux go build -o bin/$(APP_MANAGER) examples/manager/main.go

# Development
dev-setup: ## Set up development environment
	@echo "Setting up development environment..."
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest

# Code quality
lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

fmt-check: ## Check if code is formatted
	@echo "Checking code formatting..."
	test -z $$(gofmt -l .)

# Clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install dependencies
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Update dependencies
deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Security
security-scan: ## Run security scan
	@echo "Running security scan..."
	gosec ./...

# Generate documentation for the package
docs: ## Generate documentation
	@echo "Generating documentation..."
	gomarkdoc --output ./docs/DOCUMENTATION.md
	go doc -all > docs/documentation.txt

# Default target
.DEFAULT_GOAL := help