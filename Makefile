# Copyright 2017 The kubecfg authors
#
#
#    Licensed under the Apache License, Version 2.0 (the "License");
#    you may not use this file except in compliance with the License.
#    You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#    Unless required by applicable law or agreed to in writing, software
#    distributed under the License is distributed on an "AS IS" BASIS,
#    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#    See the License for the specific language governing permissions and
#    limitations under the License.

VERSION = dev-$(shell date +%FT%T%z)

GO = go
EXTRA_GO_FLAGS =
GO_FLAGS = -ldflags="-X main.version=$(VERSION) $(GO_LDFLAGS)" $(EXTRA_GO_FLAGS)
GOFMT = gofmt

KCFG_TEST_FILE = lib/kubecfg_test.jsonnet
GUESTBOOK_FILE = examples/guestbook.jsonnet
JSONNET_FILES = $(KCFG_TEST_FILE) $(GUESTBOOK_FILE)
# TODO: Simplify this once ./... ignores ./vendor
GO_PACKAGES = ./cmd/... ./utils/... ./pkg/... ./metadata/...

all: kubecfg

kubecfg:
	$(GO) build $(GO_FLAGS) .

test: gotest jsonnettest

gotest:
	$(GO) test $(GO_FLAGS) $(GO_PACKAGES)

jsonnettest: kubecfg $(JSONNET_FILES)
#	TODO: use `kubecfg check` once implemented
	./kubecfg -J lib show -f $(KCFG_TEST_FILE) -f $(GUESTBOOK_FILE) >/dev/null

vet:
	$(GO) vet $(GO_FLAGS) $(GO_PACKAGES)

fmt:
	$(GOFMT) -s -w $(shell $(GO) list -f '{{.Dir}}' $(GO_PACKAGES))

clean:
	$(RM) ./kubecfg

.PHONY: all test clean vet fmt
.PHONY: kubecfg
