.PHONY: help build run clean test fmt deps proto proto-install branch pr sync

# Default target
.DEFAULT_GOAL := help

# Build configuration
BINARY_NAME := scheduler
BUILD_DIR := .
PROTO_DIR := proto
PROTO_GEN_DIR := $(PROTO_DIR)/gen
GO_FILES := $(shell find . -name "*.go" -not -path "./$(PROTO_GEN_DIR)/*")
PROTO_FILES := $(shell find $(PROTO_DIR) -name "*.proto")

# Tools
PROTOC := ./bin/protoc
PROTOC_GEN_GO := $(shell go env GOPATH)/bin/protoc-gen-go
PROTOC_GEN_GO_GRPC := $(shell go env GOPATH)/bin/protoc-gen-go-grpc

help: ## Show this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ { printf "  %-15s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

deps: ## Install/update Go dependencies
	@echo "Installing Go dependencies..."
	go mod tidy
	go mod download

proto-install: ## Install protobuf tools
	@echo "Installing protobuf tools..."
	@if [ ! -f $(PROTOC) ]; then \
		echo "Installing protoc..."; \
		curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v25.1/protoc-25.1-linux-x86_64.zip; \
		unzip -q protoc-25.1-linux-x86_64.zip -d protoc-temp; \
		mkdir -p bin include; \
		cp protoc-temp/bin/protoc bin/; \
		cp -r protoc-temp/include/* include/; \
		rm -rf protoc-temp protoc-25.1-linux-x86_64.zip; \
	fi
	@if [ ! -f $(PROTOC_GEN_GO) ]; then \
		echo "Installing protoc-gen-go..."; \
		go install google.golang.org/protobuf/cmd/protoc-gen-go@latest; \
	fi
	@if [ ! -f $(PROTOC_GEN_GO_GRPC) ]; then \
		echo "Installing protoc-gen-go-grpc..."; \
		go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest; \
	fi

proto: proto-install ## Generate protobuf code
	@echo "Generating protobuf code..."
	@mkdir -p $(PROTO_GEN_DIR)
	@export PATH="$(shell go env GOPATH)/bin:$$PATH"; \
	for proto_file in $(PROTO_FILES); do \
		echo "Processing $$proto_file..."; \
		$(PROTOC) \
			--go_out=$(PROTO_GEN_DIR) \
			--go_opt=paths=source_relative \
			--go-grpc_out=$(PROTO_GEN_DIR) \
			--go-grpc_opt=paths=source_relative \
			--proto_path=$(PROTO_DIR) \
			--proto_path=include \
			$$proto_file; \
	done
	@echo "Protobuf code generation complete"

build: proto ## Build the scheduler binary
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) .
	@echo "Build complete: $(BINARY_NAME)"

run: build ## Build and start the server
	@echo "Starting scheduler server..."
	./$(BINARY_NAME) run

test: proto ## Run tests
	@echo "Running tests..."
	go test -v ./...

fmt: ## Format Go code
	@echo "Formatting Go code..."
	go fmt ./...
	@echo "Code formatting complete"

clean: ## Remove build artifacts and generated code
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	rm -rf $(PROTO_GEN_DIR)
	rm -rf bin/ include/
	@echo "Clean complete"

lint: proto ## Run linter (if available)
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping lint check"; \
		echo "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

dev: ## Development mode - build, run with auto-reload (requires entr)
	@echo "Starting development mode..."
	@if command -v entr >/dev/null 2>&1; then \
		find . -name "*.go" -o -name "*.proto" | entr -r make run; \
	else \
		echo "entr not installed, running once..."; \
		make run; \
	fi

# Ensure proto generation happens before Go builds
$(BINARY_NAME): proto $(GO_FILES)
	go build -o $(BINARY_NAME) .

# Git workflow commands
branch: ## Create and switch to new feature branch (usage: make branch name=feature-name)
	@if [ -z "$(name)" ]; then \
		echo "Error: Please provide branch name (usage: make branch name=feature-name)"; \
		exit 1; \
	fi
	@current_branch=$$(git branch --show-current); \
	if [ "$$current_branch" != "main" ]; then \
		echo "Warning: Not on main branch (currently on $$current_branch)"; \
		echo "Switching to main first..."; \
		git checkout main; \
		git pull origin main; \
	fi
	@echo "Creating feature branch: feature/$(name)"
	@git checkout -b feature/$(name)

pr: ## Push current branch and create PR (requires gh CLI)
	@current_branch=$$(git branch --show-current); \
	if [ "$$current_branch" = "main" ]; then \
		echo "Error: Cannot create PR from main branch!"; \
		echo "Please create a feature branch first: make branch name=your-feature"; \
		exit 1; \
	fi
	@echo "Pushing branch $$current_branch to origin..."
	@git push -u origin $$current_branch
	@if command -v gh >/dev/null 2>&1; then \
		echo "Creating pull request..."; \
		gh pr create --title "$(shell git branch --show-current | sed 's/feature\///' | sed 's/-/ /g')" --body "## Summary\n\n<!-- Describe your changes -->\n\n## Test Plan\n\n- [ ] Tests pass\n- [ ] Code formatted\n- [ ] Manual testing completed"; \
	else \
		echo "gh CLI not installed. Please create PR manually at:"; \
		echo "https://github.com/$$(git remote get-url origin | sed 's/.*github.com[:/]//' | sed 's/.git$$//')/compare/$$current_branch"; \
	fi

sync: ## Pull latest changes from main into current branch
	@current_branch=$$(git branch --show-current); \
	if [ "$$current_branch" = "main" ]; then \
		echo "Pulling latest changes from origin/main..."; \
		git pull origin main; \
	else \
		echo "Syncing feature branch $$current_branch with main..."; \
		git fetch origin main; \
		git rebase origin/main; \
	fi

# Mark proto generation as dependency for key targets
test: proto
build: proto
run: proto