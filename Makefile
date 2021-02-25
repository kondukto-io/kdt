.PHONY: all build help docker clean

MODULE        = $(shell env GO111MODULE=on $(GO) list -m)
VERSION       = $(shell git describe --tags)
DATE          = $(shell date +%FT%T%z)
BUILD_DIR     = "_release"
OUT           = "$(BUILD_DIR)/kdt"
PLATFORMS     := linux/amd64 windows/amd64 darwin/amd64
TEMP 	      = $(subst /, ,$@)
OS 	      = $(word 1, $(TEMP))
ARCH 	      = $(word 2, $(TEMP))
IMAGE_NAME    = "kondukto/kondukto-cli"
IMAGE_VERSION = $(shell echo $(VERSION)|cut -d "-" -f1)

export GO111MODULE=on

define hash
	cd $(BUILD_DIR) && sha256sum $(1) > $(1).sha256
endef

default: help

image:
	docker build . -t $(IMAGE_NAME):$(IMAGE_VERSION)

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
			-ldflags '-s -w -X main.Version=$(VERSION) -extldflags=-static' \
			-o $(OUT)-$(OS)
	$(call hash, kdt-$(OS))
