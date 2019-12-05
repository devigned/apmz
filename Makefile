APP             = apmz
PACKAGE  		= github.com/devigned/$(APP)
DATE    		?= $(shell date +%FT%T%z)
VERSION 		?= $(shell git rev-list -1 HEAD)
SHORT_VERSION 	?= $(shell git rev-parse --short HEAD)
GOBIN      		?= $(HOME)/go/bin
GOFMT   		= gofmt
GO      		= go
PKGS     		= $(or $(PKG),$(shell $(GO) list ./... | grep -vE "^$(PACKAGE)/templates/"))
TOOLSBIN		= $(shell pwd)/tools/bin
GOLINT			= golint

V = 0
Q = $(if $(filter 1,$V),,@)

export PATH := $(TOOLSBIN):$(PATH)

.PHONY: all
all: fmt lint vet tidy build

install-tools: ; $(info $(M) installing tools…)
	$(Q) make -C ./tools

build: lint tidy ; $(info $(M) buiding ./bin/$(APP))
	$Q $(GO)  build -ldflags "-X $(PACKAGE)/cmd.GitCommit=$(VERSION)" -o ./bin/$(APP)

.PHONY: lint
lint: install-tools ; $(info $(M) running golint…) @ ## Run golint
	$(Q) $(GOLINT) -set_exit_status `go list ./... | grep -v /internal/`

.PHONY: fmt
fmt: ; $(info $(M) running gofmt…) @ ## Run gofmt on all source files
	@ret=0 && for d in $$($(GO) list -f '{{.Dir}}' ./...); do \
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

.PHONY: test
test: ; $(info $(M) running go test…)
	$(Q) $(GO) test ./... -tags=noexit

.PHONY: test-cover
test-cover: install-tools ; $(info $(M) running go test…)
	$(Q) $(GO) test -tags=noexit -race -covermode atomic -coverprofile=profile.cov ./...

.PHONY: gox
gox: install-tools
	$(Q) gox -osarch="darwin/amd64 windows/amd64 linux/amd64" -ldflags "-X $(PACKAGE)/cmd.GitCommit=$(VERSION)" -output "./bin/$(SHORT_VERSION)/{{.Dir}}_{{.OS}}_{{.Arch}}"
	$(Q) tar -czvf ./bin/$(SHORT_VERSION)/pub_darwin_amd64.tar.gz -C ./bin/$(SHORT_VERSION)/ pub_darwin_amd64
	$(Q) tar -czvf ./bin/$(SHORT_VERSION)/pub_linux_amd64.tar.gz -C ./bin/$(SHORT_VERSION)/ pub_linux_amd64
	$(Q) tar -czvf ./bin/$(SHORT_VERSION)/pub_windows_amd64.tar.gz -C ./bin/$(SHORT_VERSION)/ pub_windows_amd64.exe

.PHONY: ci
ci: install-tools fmt lint vet tidy build test-cover
