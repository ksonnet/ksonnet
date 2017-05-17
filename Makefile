VERSION = dev-$(shell date +%FT%T%z)

GO = go
GO_FLAGS = -ldflags="-X main.version=$(VERSION)"
GOFMT = gofmt

# TODO: Simplify this once ./... ignores ./vendor
GO_PACKAGES = ./cmd/... ./utils/...

all: kubecfg

kubecfg:
	$(GO) build $(GO_FLAGS) .

test:
	$(GO) test $(GO_PACKAGES)

vet:
	$(GO) vet $(GO_PACKAGES)

fmt:
	$(GOFMT) -s -w $(shell $(GO) list -f '{{.Dir}}' $(GO_PACKAGES))

clean:
	$(RM) ./kubecfg

.PHONY: all test clean vet fmt
.PHONY: kubecfg
