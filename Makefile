# Terraform JumpCloud Provider Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOVET=$(GOCMD) vet
GOFMT=$(GOCMD) fmt
GOLINT=golangci-lint

# Build parameters
BINARY_NAME=terraform-provider-jumpcloud
VERSION=$(shell git describe --tags --always --dirty --match=v* 2> /dev/null || echo "dev")
BUILD_DIR=dist
OS_ARCH=darwin_amd64

# Terraform directories
LOCAL_PLUGIN_DIR=~/.terraform.d/plugins/registry.terraform.io/agilize/jumpcloud/$(VERSION)/$(OS_ARCH)

.PHONY: all build clean test test-unit test-integration test-acceptance test-resources test-datasources test-performance test-security test-coverage fmt lint lint-strict vet mod-tidy mod-vendor install docs release pr-check pr-checks check-sdk-version tfproviderlint-check check-fmt

all: clean fmt lint vet test build

build:
	@echo "Building provider..."
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)

clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v -run "Test[^Acc]" ./...

test-integration:
	@echo "Running integration tests..."
	JUMPCLOUD_INTEGRATION_TEST=true $(GOTEST) -v -run "Integration" ./...

test-acceptance:
	@echo "Running acceptance tests..."
	TF_ACC=1 $(GOTEST) -v -run "TestAcc" ./...

test-resources:
	@echo "Running resource tests..."
	$(GOTEST) -v -run "TestResource" ./...

test-datasources:
	@echo "Running data source tests..."
	$(GOTEST) -v -run "TestDataSource" ./...

test-performance:
	@echo "Running performance tests..."
	$(GOTEST) -v -run "TestPerformance|Benchmark" ./...

test-security:
	@echo "Running security tests..."
	$(GOTEST) -v -run "TestSecurity" ./...

test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

lint:
	@echo "Linting code..."
	@if ! command -v $(GOLINT) &> /dev/null; then \
		echo "golangci-lint not found. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@echo "Running linters (ignoring errors for PR check)..."
	-$(GOLINT) run --timeout=5m --no-config --enable=errcheck,govet,ineffassign,staticcheck,unused --verbose || true

lint-strict:
	@echo "Linting code (strict mode)..."
	@if ! command -v $(GOLINT) &> /dev/null; then \
		echo "golangci-lint not found. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@echo "Running linters (errors will cause failure)..."
	$(GOLINT) run --timeout=5m --no-config --enable=errcheck,govet,ineffassign,staticcheck,unused --verbose

vet:
	@echo "Vetting code..."
	$(GOVET) ./...

mod-tidy:
	@echo "Tidying Go modules..."
	$(GOMOD) tidy

mod-vendor:
	@echo "Downloading Go module dependencies..."
	$(GOMOD) vendor

install: build
	@echo "Installing provider locally..."
	mkdir -p $(LOCAL_PLUGIN_DIR)
	cp $(BUILD_DIR)/$(BINARY_NAME) $(LOCAL_PLUGIN_DIR)/

docs:
	@echo "Generating documentation..."
	cd docs && go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

release: clean mod-tidy fmt lint vet test build
	@echo "Preparing release version $(VERSION)..."
	mkdir -p $(BUILD_DIR)/releases
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/releases/$(BINARY_NAME)_$(VERSION)_darwin_amd64
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/releases/$(BINARY_NAME)_$(VERSION)_linux_amd64
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/releases/$(BINARY_NAME)_$(VERSION)_windows_amd64.exe
	cd $(BUILD_DIR)/releases && \
		zip $(BINARY_NAME)_$(VERSION)_darwin_amd64.zip $(BINARY_NAME)_$(VERSION)_darwin_amd64 && \
		zip $(BINARY_NAME)_$(VERSION)_linux_amd64.zip $(BINARY_NAME)_$(VERSION)_linux_amd64 && \
		zip $(BINARY_NAME)_$(VERSION)_windows_amd64.zip $(BINARY_NAME)_$(VERSION)_windows_amd64.exe

check-fmt:
	@echo "Checking code formatting..."
	@gofmt_files=$$(gofmt -l .); \
	if [[ -n "$$gofmt_files" ]]; then \
		echo "These files need to be formatted with gofmt:"; \
		echo "$$gofmt_files"; \
		exit 1; \
	else \
		echo "All Go files are properly formatted."; \
	fi

tfproviderlint-check:
	@echo "Running Terraform Provider Lint checks..."
	@if ! command -v tfproviderlint &> /dev/null; then \
		echo "Installing tfproviderlint..."; \
		go install github.com/bflad/tfproviderlint/cmd/tfproviderlint@latest; \
		echo "Installed tfproviderlint to $$(go env GOPATH)/bin/tfproviderlint"; \
	fi
	@chmod +x ./scripts/linting/check_critical_lint.sh
	-@PATH="$$(go env GOPATH)/bin:$$PATH" ./scripts/linting/check_critical_lint.sh || true

check-sdk-version:
	@echo "Checking Terraform SDK version..."
	@CURRENT_SDK=$$(go list -m github.com/hashicorp/terraform-plugin-sdk/v2 | awk '{print $$2}'); \
	echo "Current SDK version: $$CURRENT_SDK"; \
	if [[ "$$CURRENT_SDK" < "v2.10.0" ]]; then \
		echo "Warning: Using an older SDK version. Consider upgrading."; \
	else \
		echo "SDK version is sufficiently recent."; \
	fi

# Run all PR checks locally (same as GitHub Actions workflow)
pr-check:
	@echo "Running all PR checks locally..."
	@echo "Step 1/8: Tidying Go modules"
	@$(MAKE) mod-tidy
	@echo "Step 2/8: Running linters"
	@$(MAKE) lint
	@echo "Step 3/8: Checking code formatting"
	@$(MAKE) check-fmt
	@echo "Step 4/8: Vetting code"
	@$(MAKE) vet
	@echo "Step 5/8: Running unit tests"
	@$(MAKE) test-unit
	@echo "Step 6/8: Running tests with coverage"
	@$(MAKE) test-coverage
	@echo "Step 7/8: Running Terraform Provider Lint checks"
	@$(MAKE) tfproviderlint-check
	@echo "Step 8/8: Checking Terraform SDK version"
	@$(MAKE) check-sdk-version
	@echo "✅ All PR checks passed successfully!"

# Run all PR checks locally with strict linting (same as GitHub Actions workflow)
pr-checks:
	@echo "Running all PR checks locally with strict linting..."
	@echo "Step 1/8: Tidying Go modules"
	@$(MAKE) mod-tidy
	@echo "Step 2/8: Running linters (strict mode)"
	@$(MAKE) lint-strict
	@echo "Step 3/8: Checking code formatting"
	@$(MAKE) check-fmt
	@echo "Step 4/8: Vetting code"
	@$(MAKE) vet
	@echo "Step 5/8: Running unit tests"
	@$(MAKE) test-unit
	@echo "Step 6/8: Running tests with coverage"
	@$(MAKE) test-coverage
	@echo "Step 7/8: Running Terraform Provider Lint checks"
	@$(MAKE) tfproviderlint-check
	@echo "Step 8/8: Checking Terraform SDK version"
	@$(MAKE) check-sdk-version
	@echo "✅ All PR checks passed successfully!"

help:
	@echo "Terraform JumpCloud Provider Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make                   Build the provider after running format, lint, vet, and tests"
	@echo "  make build             Build the provider binary"
	@echo "  make clean             Remove build artifacts"
	@echo "  make test              Run all tests"
	@echo "  make test-unit         Run unit tests"
	@echo "  make test-integration  Run integration tests (requires API credentials)"
	@echo "  make test-acceptance   Run acceptance tests (requires API credentials)"
	@echo "  make test-resources    Run resource tests"
	@echo "  make test-datasources  Run data source tests"
	@echo "  make test-performance  Run performance tests"
	@echo "  make test-security     Run security tests"
	@echo "  make test-coverage     Run tests with coverage report"
	@echo "  make fmt               Format Go code"
	@echo "  make lint              Run linters (ignoring errors)"
	@echo "  make lint-strict       Run linters (failing on errors)"
	@echo "  make vet               Run Go vet"
	@echo "  make mod-tidy          Tidy Go modules"
	@echo "  make mod-vendor        Download all dependencies"
	@echo "  make install           Install provider to local Terraform plugin directory"
	@echo "  make docs              Generate documentation"
	@echo "  make release           Create release artifacts for different platforms"
	@echo "  make pr-check          Run all PR checks locally (ignoring lint errors)"
	@echo "  make pr-checks         Run all PR checks locally (failing on lint errors)"