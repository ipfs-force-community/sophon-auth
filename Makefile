SHELL=/usr/bin/env bash

git=$(subst -,.,$(shell git describe --always --match=NeVeRmAtCh --dirty 2>/dev/null || git rev-parse --short HEAD 2>/dev/null))

ldflags=-X=github.com/ipfs-force-community/sophon-auth/core.CurrentCommit=+git.$(git)
ifneq ($(strip $(LDFLAGS)),)
	ldflags+=-extldflags=$(LDFLAGS)
endif
GOFLAGS+=-ldflags="$(ldflags)"

all: clean sophon-auth

show-env:
	@echo '_________________build_environment_______________'
	@echo '| git commit=$(git)'
	@echo '-------------------------------------------------'

sophon-auth:show-env
	go build $(GOFLAGS) -o sophon-auth ./cmd/server/*.go

test:
	go test -race ./...

lint:
	golangci-lint run

linux: clean
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -o sophon-auth ./cmd/server/*.go

clean:
	rm -rf sophon-auth
.PHONY: clean 

static: clean
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -o sophon-auth ./cmd/server/*.go

gen:
	go generate ./...

.PHONY: docker
TAG:=test
docker:

ifdef DOCKERFILE
	cp $(DOCKERFILE) ./dockerfile
else
	curl -o dockerfile https://raw.githubusercontent.com/filecoin-project/venus-docs/master/script/docker/dockerfile
endif

	docker build --build-arg HTTPS_PROXY=$(BUILD_DOCKER_PROXY) --build-arg BUILD_TARGET=sophon-auth  -t sophon-auth .
	docker tag sophon-auth filvenus/sophon-auth:$(TAG)

ifdef PRIVATE_REGISTRY
	docker tag sophon-auth $(PRIVATE_REGISTRY)/filvenus/sophon-auth:$(TAG)
endif


docker-push: docker
	docker push $(PRIVATE_REGISTRY)/filvenus/sophon-auth:$(TAG)
