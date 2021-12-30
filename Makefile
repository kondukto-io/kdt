.PHONY: all build help docker clean

MODULE        = $(shell env GO111MODULE=on $(GO) list -m)
COMMIT        = $(shell git rev-list --abbrev-commit --tags --max-count=1)
TAG           = $(shell git describe --tags --abbrev=0)
DATE          = $(shell git log -1 --format=%cd --date=format:"%Y%m%d")
BUILD_DIR     = _release
OUT           = $(BUILD_DIR)/kdt
PLATFORMS     := linux/amd64 windows/amd64 darwin/amd64 darwin/arm64
TEMP 	      = $(subst /, ,$@)
OS 	      = $(word 1, $(TEMP))
ARCH 	      = $(word 2, $(TEMP))
IMAGE_NAME    = "kondukto/kondukto-cli"
IMAGE_VERSION = $(shell echo $(VERSION)|cut -d "+" -f1)

VERSION      := $(TAG)+$(COMMIT)

export GO111MODULE=on

define hash
	sha256sum $(BUILD_DIR)/$(1) > $(BUILD_DIR)/$(1).sha256
endef

default: help

image:
	docker build -t $(IMAGE_NAME):$(IMAGE_VERSION) .

push-image:
	docker push $(IMAGE_NAME):$(IMAGE_VERSION)

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
