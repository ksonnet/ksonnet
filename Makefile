VERSION = dev-$(shell date +%FT%T%z)

GO = go
GO_FLAGS = -ldflags="-X main.version=$(VERSION)"

all: kubecfg

kubecfg:
	$(GO) build $(GO_FLAGS) .

test:
	$(GO) test ./cmd/...

clean:
	$(RM) ./kubecfg

.PHONY: all test clean kubecfg
