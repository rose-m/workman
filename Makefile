.PHONY: build fmt lint clean test install-tools install help

# Binary name
BINARY_NAME=workman

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOFMT=$(GOCMD) fmt
GOMOD=$(GOCMD) mod

help: ## Show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@$(GOBUILD) -o $(BINARY_NAME) .
	@echo "Build complete: ./$(BINARY_NAME)"

fmt: ## Format Go code
	@echo "Formatting code..."
	@$(GOFMT) ./...
	@echo "Format complete"

lint: ## Run linter (golangci-lint)
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found. Run 'make install-tools' to install it."; \
		exit 1; \
	fi

test: ## Run tests
	@echo "Running tests..."
	@$(GOTEST) -v ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -f $(BINARY_NAME)
	@echo "Clean complete"

tidy: ## Tidy and verify Go modules
	@echo "Tidying modules..."
	@$(GOMOD) tidy
	@$(GOMOD) verify

install-tools: ## Install development tools
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed"

run: build ## Build and run the application
	@./$(BINARY_NAME)

install: build ## Build and install to ~/.local/bin
	@echo "Installing $(BINARY_NAME) to ~/.local/bin..."
	@mkdir -p ~/.local/bin
	@cp $(BINARY_NAME) ~/.local/bin/$(BINARY_NAME)
	@echo "Install complete: ~/.local/bin/$(BINARY_NAME)"
	@echo "Make sure ~/.local/bin is in your PATH"

all: fmt lint build ## Format, lint, and build

.DEFAULT_GOAL := help
