.PHONY: all build buildstatic help  

MODULE        = $(shell env GO111MODULE=on $(GO) list -m)
VERSION       = $(shell git describe --tags)
DATE          = $(shell date +%FT%T%z)
OUT           = "release/kdt"
PLATFORMS     := linux/amd64 windows/amd64 darwin/amd64
TEMP 					= $(subst /, ,$@)
OS 						= $(word 1, $(TEMP))
ARCH 					= $(word 2, $(TEMP))
IMAGE_NAME    = "kondukto-cli"
IMAGE_VERSION = $(shell echo $(VERSION)|cut -d "-" -f1)

export GO111MODULE=on

define hash
	sha256sum $(1) > $(1).sha256
endef

default: help

build:
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build \
			-tags prod \
			-ldflags '-s -w -X main.Version=$(VERSION) -extldflags=-static' \
			-o $(OUT)-$(OS)


help:
	@echo $(IMAGE_VERSION)

all: $(PLATFORMS)

$(PLATFORMS):
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build \
			-tags prod \
			-ldflags '-s -w -X main.Version=$(VERSION) -extldflags=-static' \
			-o $(OUT)-$(OS)
	$(call hash, $(OUT)-$(OS))
