SHELL=/usr/bin/env bash

all: clean oauth-server

oauth-server:
	go build -o oauth-server ./cmd/server/*.go

lint:
	gofmt -s -w ./
	golangci-lint run


clean:
	rm -rf oauth-server
.PHONY: clean
