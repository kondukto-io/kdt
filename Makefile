BINARY := kdt
VERSION := 1.0.3
PLATFORMS := linux/amd64 windows/amd64

temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))

release: $(PLATFORMS)

$(PLATFORMS):
	mkdir -p release
	GOOS=$(os) GOARCH=$(arch) CGO_ENABLED=0 go build -o release/$(BINARY)-$(VERSION)-$(os)

.PHONY: release $(PLATFORMS)
