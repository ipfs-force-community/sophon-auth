SHELL=/usr/bin/env bash

git=$(subst -,.,$(shell git describe --always --match=NeVeRmAtCh --dirty 2>/dev/null || git rev-parse --short HEAD 2>/dev/null))


ldflags=-X=github.com/filecoin-project/venus-auth/core.CurrentCommit=+git.$(git)
ifneq ($(strip $(LDFLAGS)),)
	ldflags+=-extldflags=$(LDFLAGS)
endif
GOFLAGS+=-ldflags="$(ldflags)"

all: clean venus-auth


show-env:
	@echo '_________________build_environment_______________'
	@echo '| git commit=$(git)'
	@echo '-------------------------------------------------'

venus-auth:show-env $(BUILD_DEPS)
	go build $(GOFLAGS) -o venus-auth ./cmd/server/*.go

lint:
	golangci-lint run

linux: clean
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -o venus-auth ./cmd/server/*.go

clean:
	rm -rf venus-auth

.PHONY: clean 


.PHONY: docker


static: clean
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -o venus-auth ./cmd/server/*.go

docker:
	docker build --build-arg https_proxy=$(BUILD_DOCKER_PROXY) -t venus-auth .
