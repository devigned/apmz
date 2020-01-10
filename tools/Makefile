GO 			?= go
LOCALBIN 	= $(shell pwd)/bin

default: install

install: $(LOCALBIN)/golint $(LOCALBIN)/gox $(LOCALBIN)/go-bindata
	@echo > /dev/null

$(LOCALBIN)/go-bindata:
	GOBIN=$(LOCALBIN) $(GO) get github.com/go-bindata/go-bindata/...@v3.1.2

$(LOCALBIN)/golint:
	GOBIN=$(LOCALBIN) $(GO) get golang.org/x/lint/golint

$(LOCALBIN)/gox:
	GOBIN=$(LOCALBIN) $(GO) get github.com/mitchellh/gox

.PHONY: reload
reload: clean install

.PHONY: clean
clean:
	rm -rf $(LOCALBIN)
