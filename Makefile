# Tooling settings
GOPATH        = $(shell go env GOPATH)
GO_BIN_DIR    = bin

# Release build metadata
MODULE        = $(shell env GO111MODULE=on go list -m)
COMMIT        = $(shell git rev-parse --short HEAD)
TAG           = $(shell git describe --tags --abbrev=0)
VERSION_TAG   = $(shell echo $(TAG)| cut -d '-' -f 1)
DATE          = $(shell git log -1 --format=%cd --date=format:"%Y%m%d")
BUILD_DIR     = _release
OUT           = $(BUILD_DIR)/kdt
IMAGE_NAME    = kondukto/kondukto-cli
PLATFORMS     := linux/amd64 linux/arm64 windows/amd64 darwin/amd64 darwin/arm64
TEMP          = $(subst /, ,$@)
OS            = $(word 1, $(TEMP))
ARCH          = $(word 2, $(TEMP))

VERSION       := $(VERSION_TAG)

export GO111MODULE=on

# Phony targets to prevent conflicts with files of the same name
.PHONY: all help docker clean test test_coverage \
        tidy test-local integration-test integration-test-local \
        lint nilaway gofumpt vulncheck operations operations-local

.DEFAULT_GOAL := help

help: ## Show this help
	@grep -E '^[a-zA-Z_.-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-24s\033[0m %s\n", $$1, $$2}'

# --- Dependencies ---

tidy: ## Tidy and verify Go module dependencies
	@echo "Tidying dependencies..."
	@go version
	@go mod tidy
	@go mod verify

# --- Build ---

all: $(PLATFORMS) ## Cross-compile release binaries for all platforms

$(PLATFORMS):
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build \
			-tags prod \
			-buildmode exe \
			-ldflags '-s -w -X github.com/kondukto-io/kdt/cmd.Version=$(VERSION) -extldflags=-static' \
			-o $(OUT)-$(OS)-$(ARCH)
	$(call hash,kdt-$(OS)-$(ARCH))

docker: linux/amd64 ## Build the Docker image from the linux/amd64 release binary
	@docker build -t $(IMAGE_NAME):latest -f Dockerfile .

# --- Tests ---

test: ## Run unit tests with race detector, coverage and JSON report (CI)
	@echo "Running unit tests..."
	@CGO_ENABLED=1 go test ./... -coverprofile=coverage.out -race -count 1 -json > test-report.json
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

test-local: ## Run unit tests with race detector and coverage (human-readable)
	@echo "Running unit tests..."
	@go test ./... -coverprofile=coverage.out -race -count 1
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

integration-test: ## Run integration tests (build tag: integration) with JSON report (CI)
	@echo "Running integration tests..."
	@CGO_ENABLED=1 go test -tags integration ./... -coverprofile=integration-coverage.out -race -count 1 -json > integration-test-report.json
	@go tool cover -html=integration-coverage.out -o integration-coverage.html
	@echo "Integration coverage report generated at integration-coverage.html"

integration-test-local: ## Run integration tests (build tag: integration) human-readable
	@echo "Running integration tests..."
	@go test -tags integration ./... -coverprofile=integration-coverage.out -race -count 1
	@go tool cover -html=integration-coverage.out -o integration-coverage.html
	@echo "Integration coverage report generated at integration-coverage.html"

test_coverage: ## Run unit tests writing coverage profile (kept for CI compatibility)
	@go test ./... -coverprofile=coverage.out

# --- Static analysis ---

lint: ## Run golangci-lint v2
	@echo "Running linters v2..."
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.7.2
	@"$(GOPATH)/$(GO_BIN_DIR)/golangci-lint" version
	@"$(GOPATH)/$(GO_BIN_DIR)/golangci-lint" run ./... --config=.golangci.yaml

nilaway: ## Run nilaway nil-safety analysis
	@echo "Running nilaway..."
	@go install go.uber.org/nilaway/cmd/nilaway@latest
	@"$(GOPATH)/$(GO_BIN_DIR)/nilaway" -include-pkgs="github.com/kondukto-io/kdt" ./...

gofumpt: ## Format code with gofumpt
	@echo "Running gofumpt..."
	@go install mvdan.cc/gofumpt@latest
	@"$(GOPATH)/$(GO_BIN_DIR)/gofumpt" -w .

vulncheck: ## Run govulncheck vulnerability scan
	@echo "Running vulncheck..."
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@"$(GOPATH)/$(GO_BIN_DIR)/govulncheck" -show verbose ./...

# --- Aggregates ---

operations: tidy all test integration-test lint nilaway vulncheck ## Run the full CI pipeline
	@echo ""
	@echo "All operations completed!"

operations-local: tidy all test-local integration-test-local lint nilaway vulncheck ## Run the full local pipeline
	@echo ""
	@echo "All local operations completed!"

# --- Housekeeping ---

clean: ## Remove build artifacts and generated reports
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html integration-coverage.out integration-coverage.html \
		test-report.json integration-test-report.json golangci-lint-report.xml
	@go clean

define hash
	cd $(BUILD_DIR) && sha256sum $(1) > $(1).sha256
endef
