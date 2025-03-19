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
LOCAL_PLUGIN_DIR=~/.terraform.d/plugins/github.com/agilize/jumpcloud/$(VERSION)/$(OS_ARCH)

.PHONY: all build clean test fmt lint vet mod-tidy install release

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

fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

lint:
	@echo "Linting code..."
	$(GOLINT) run

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
	@echo "  make fmt               Format Go code"
	@echo "  make lint              Run linters"
	@echo "  make vet               Run Go vet"
	@echo "  make mod-tidy          Tidy Go modules"
	@echo "  make mod-vendor        Download all dependencies"
	@echo "  make install           Install provider to local Terraform plugin directory"
	@echo "  make docs              Generate documentation"
	@echo "  make release           Create release artifacts for different platforms" 