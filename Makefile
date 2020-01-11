APP             = apmz
PACKAGE         = github.com/devigned/$(APP)
DATE            ?= $(shell date +%FT%T%z)
VERSION         ?= $(shell git rev-list -1 HEAD)
SHORT_VERSION   ?= $(shell git rev-parse --short HEAD)
GOBIN           ?= $(HOME)/go/bin
TOOLSBIN        = $(shell pwd)/tools/bin
GOFMT           = gofmt
GO              = go
PKGS            = $(or $(PKG),$(shell $(GO) list ./... | grep -vE "^$(PACKAGE)/templates/"))
GOLINT          = $(TOOLSBIN)/golint
BINDATA         = $(TOOLSBIN)/go-bindata
GOX             = $(TOOLSBIN)/gox

V = 0
Q = $(if $(filter 1,$V),,@)

.PHONY: all
all: install-tools generate fmt lint vet tidy build

.PHONY: install-tools
install-tools: ; $(info $(M) installing tools…)
	$(Q) make -C ./tools

.PHONY: build
build: lint tidy ; $(info $(M) buiding ./bin/$(APP))
	$Q $(GO) build -ldflags "-X $(PACKAGE)/cmd.GitCommit=$(VERSION)" -o ./bin/$(APP)

.PHONY: generate
generate: ; $(info $(M) running generate…)
	$Q $(BINDATA) -o ./pkg/data/bindata.go -pkg data -nocompress ./data/...

.PHONY: lint
lint: install-tools ; $(info $(M) running golint…) @ ## Run golint
	$(Q) $(GOLINT) -set_exit_status `go list ./... | grep -v /internal/ | grep -v /pkg/data `

.PHONY: fmt
fmt: ; $(info $(M) running gofmt…) @ ## Run gofmt on all source files
	@ret=0 && for d in $$($(GO) list -f '{{.Dir}}' ./... | grep -v /pkg/data | grep -v /internal/test/bash); do \
		$(GOFMT) -l -w $$d/*.go || ret=$$? ; \
	 done ; exit $$ret

.PHONY: vet
vet: ; $(info $(M) running vet…) @ ## Run vet
	$Q $(GO) vet ./...

.PHONY: tidy
tidy: ; $(info $(M) running tidy…) @ ## Run tidy
	$Q $(GO) mod tidy

.PHONY: build-debug
build-debug: ; $(info $(M) buiding debug...)
	$Q $(GO)  build -o ./bin/$(APP) -tags debug

.PHONY: test-generate ; $(info $(M) running test-generate…)
test-generate: ; $(info $(M) generating test data…)
	$Q $(BINDATA) -o ./internal/test/bash/bindata.go -pkg bash_test -nocompress ./cmd/bash/testdata/...

.PHONY: test
test: build test-generate ; $(info $(M) running go test…)
	$(Q) $(GO) test ./... -tags=noexit

.PHONY: test-cover
test-cover: install-tools ; $(info $(M) running go test…)
	$(Q) $(GO) test -tags=noexit -race -covermode atomic -coverprofile=profile.cov ./...

.PHONY: gox
gox: install-tools
	$(Q) $(GOX) -osarch="darwin/amd64 windows/amd64 linux/amd64" -ldflags "-X $(PACKAGE)/cmd.GitCommit=$(VERSION)" -output "./bin/$(SHORT_VERSION)/{{.Dir}}_{{.OS}}_{{.Arch}}"
	$(Q) tar -czvf ./bin/$(SHORT_VERSION)/$(APP)_darwin_amd64.tar.gz -C ./bin/$(SHORT_VERSION)/ $(APP)_darwin_amd64
	$(Q) tar -czvf ./bin/$(SHORT_VERSION)/$(APP)_linux_amd64.tar.gz -C ./bin/$(SHORT_VERSION)/ $(APP)_linux_amd64
	$(Q) tar -czvf ./bin/$(SHORT_VERSION)/$(APP)_windows_amd64.tar.gz -C ./bin/$(SHORT_VERSION)/ $(APP)_windows_amd64.exe

.PHONY: ci
ci: install-tools fmt lint vet tidy build test-cover
