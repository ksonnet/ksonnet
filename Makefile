VERSION = dev-$(shell date +%FT%T%z)

GO = go
GO_FLAGS = -ldflags="-X main.version=$(VERSION) $(GO_LDFLAGS)"
GOFMT = gofmt

JSONNET_FILES = lib/kubecfg_test.jsonnet examples/guestbook.jsonnet
# TODO: Simplify this once ./... ignores ./vendor
GO_PACKAGES = ./cmd/... ./utils/...

all: kubecfg

kubecfg:
	$(GO) build $(GO_FLAGS) .

test: gotest jsonnettest

gotest:
	$(GO) test $(GO_FLAGS) $(GO_PACKAGES)

jsonnettest: kubecfg $(JSONNET_FILES)
#	TODO: use `kubecfg check` once implemented
	./kubecfg -J lib show $(JSONNET_FILES) >/dev/null

vet:
	$(GO) vet $(GO_FLAGS) $(GO_PACKAGES)

fmt:
	$(GOFMT) -s -w $(shell $(GO) list -f '{{.Dir}}' $(GO_PACKAGES))

clean:
	$(RM) ./kubecfg

.PHONY: all test clean vet fmt
.PHONY: kubecfg
