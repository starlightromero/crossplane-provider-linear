# ====================================================================================
# provider-linear Makefile
# ====================================================================================

PROJECT_NAME := provider-linear
PROJECT_REPO := github.com/avodah-inc/provider-linear

# Terraform provider source
TERRAFORM_PROVIDER_SOURCE := terraform-community-providers/linear
TERRAFORM_PROVIDER_REPO := https://github.com/terraform-community-providers/terraform-provider-linear
TERRAFORM_PROVIDER_VERSION := 0.5.0

# Build variables
PLATFORMS ?= linux_amd64 linux_arm64
GO := go
GOFLAGS ?=
LDFLAGS := -s -w -X $(PROJECT_REPO)/internal/version.Version=$(VERSION)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Binary output
BIN_DIR := bin
PROVIDER_BIN := $(BIN_DIR)/$(PROJECT_NAME)

# Container image
REGISTRY ?= ghcr.io/avodah-inc
IMAGE := $(REGISTRY)/$(PROJECT_NAME)
IMAGE_TAG ?= $(VERSION)

# Tools
GOLANGCI_LINT_VERSION ?= v1.59.1
UPJET := $(GO) run github.com/crossplane/upjet/cmd/upjet

# ====================================================================================
# Targets
# ====================================================================================

.PHONY: all
all: generate build test lint

# ------------------------------------------------------------------------------------
# Generate
# ------------------------------------------------------------------------------------

.PHONY: generate
generate: ## Run Upjet code generation to produce CRDs and controllers
	@echo "==> Generating CRDs and controllers from Terraform provider..."
	$(GO) generate ./...
	@echo "==> Generation complete."

.PHONY: generate.init
generate.init: ## Initialize the Upjet generation pipeline
	@$(GO) run github.com/crossplane/upjet/cmd/scraper \
		-n $(TERRAFORM_PROVIDER_SOURCE) \
		-r $(TERRAFORM_PROVIDER_REPO) \
		-o config/schema.json

# ------------------------------------------------------------------------------------
# Build
# ------------------------------------------------------------------------------------

.PHONY: build
build: ## Build the provider binary
	@echo "==> Building $(PROJECT_NAME)..."
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(PROVIDER_BIN) ./cmd/provider/
	@echo "==> Build complete: $(PROVIDER_BIN)"

.PHONY: build.image
build.image: ## Build the provider container image
	docker build \
		--build-arg VERSION=$(VERSION) \
		-t $(IMAGE):$(IMAGE_TAG) \
		.

# ------------------------------------------------------------------------------------
# Test
# ------------------------------------------------------------------------------------

.PHONY: test
test: ## Run all tests (unit + property-based)
	@echo "==> Running tests..."
	$(GO) test -race -count=1 ./...
	@echo "==> Tests complete."

.PHONY: test.unit
test.unit: ## Run unit tests only
	$(GO) test -race -count=1 -short ./...

.PHONY: test.integration
test.integration: ## Run integration tests
	$(GO) test -race -count=1 -tags=integration ./...

.PHONY: test.coverage
test.coverage: ## Run tests with coverage report
	$(GO) test -race -count=1 -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "==> Coverage report: coverage.html"

# ------------------------------------------------------------------------------------
# Lint
# ------------------------------------------------------------------------------------

.PHONY: lint
lint: ## Run linters
	@echo "==> Running linters..."
	golangci-lint run ./...
	@echo "==> Lint complete."

.PHONY: lint.fix
lint.fix: ## Run linters with auto-fix
	golangci-lint run --fix ./...

# ------------------------------------------------------------------------------------
# Clean
# ------------------------------------------------------------------------------------

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BIN_DIR) coverage.out coverage.html

# ------------------------------------------------------------------------------------
# Security & Compliance
# ------------------------------------------------------------------------------------

.PHONY: scan.trivy
scan.trivy: ## Run Trivy filesystem scan
	trivy fs . --include-dev-deps

.PHONY: scan.sonar
scan.sonar: ## Run SonarQube scan
	sonar-scanner

.PHONY: validate.manifests
validate.manifests: ## Validate Kubernetes manifests with kubeconform
	kubeconform \
		-schema-location default \
		-schema-location "https://raw.githubusercontent.com/datreeio/CRDs-catalog/main/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json" \
		-strict \
		-summary \
		examples/ package/

# ------------------------------------------------------------------------------------
# Help
# ------------------------------------------------------------------------------------

.PHONY: help
help: ## Show this help message
	@grep -E '^[a-zA-Z._-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
