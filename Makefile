.PHONY: all build help docker clean

MODULE        = $(shell env GO111MODULE=on $(GO) list -m)
COMMIT        = $(shell git rev-parse --short HEAD)
TAG           = $(shell git describe --tags --abbrev=0)
DATE          = $(shell git log -1 --format=%cd --date=format:"%Y%m%d")
BUILD_DIR     = _release
OUT           = $(BUILD_DIR)/kdt
PLATFORMS     := linux/amd64 windows/amd64 darwin/amd64 darwin/arm64
TEMP 	      = $(subst /, ,$@)
OS 	      = $(word 1, $(TEMP))
ARCH 	      = $(word 2, $(TEMP))

VERSION_TAG = $(shell echo $(TAG)|cut -d "-" -f1)
VERSION      := $(VERSION_TAG)-$(COMMIT)

export GO111MODULE=on

define hash
	cd $(BUILD_DIR) && sha256sum $(1) > $(1).sha256
endef

default: help

clean:
	rm -r $(BUILD_DIR)
	go clean

help:
	@echo 
	@echo "Available commands: all | image | help"
	@echo " make all   -- to build kdt in all supported environments"
	@echo " make image -- to build docker image"

all: $(PLATFORMS)

$(PLATFORMS):
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build \
			-tags prod \
			-buildmode exe \
			-ldflags '-s -w -X github.com/kondukto-io/kdt/cmd.Version=$(VERSION) -extldflags=-static' \
			-o $(OUT)-$(OS)-$(ARCH)
	$(call hash,kdt-$(OS)-$(ARCH))
