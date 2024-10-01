.PHONY: all build help docker clean

MODULE        = $(shell env GO111MODULE=on $(GO) list -m)
COMMIT        = $(shell git rev-parse --short HEAD)
TAG           = $(shell git describe --tags --abbrev=0)
VERSION_TAG   = $(shell echo $(TAG)| cut -d '-' -f 1)
DATE          = $(shell git log -1 --format=%cd --date=format:"%Y%m%d")
BUILD_DIR     = _release
OUT           = $(BUILD_DIR)/kdt
PLATFORMS     := linux/amd64 linux/arm64 windows/amd64 darwin/amd64 darwin/arm64
TEMP 	      = $(subst /, ,$@)
OS 	      	  = $(word 1, $(TEMP))
ARCH 	      = $(word 2, $(TEMP))

VERSION       := $(VERSION_TAG)

export GO111MODULE=on

define hash
	cd $(BUILD_DIR) && sha256sum $(1) > $(1).sha256
endef

clean:
	rm -r $(BUILD_DIR)
	rm coverage.out
	go clean

test:
	go test ./...

test_coverage:
	go test ./... -coverprofile=coverage.out

go.mod:
	go mod tidy
	go mod verify

vet:
	go vet

all: $(PLATFORMS)

$(PLATFORMS):
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build \
			-tags prod \
			-buildmode exe \
			-ldflags '-s -w -X github.com/kondukto-io/kdt/cmd.Version=$(VERSION) -extldflags=-static' \
			-o $(OUT)-$(OS)-$(ARCH)
	$(call hash,kdt-$(OS)-$(ARCH))
