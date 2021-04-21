SHELL=/usr/bin/env bash

all: clean venus-auth

venus-auth:
	go build -o venus-auth ./cmd/server/*.go

lint:
	gofmt -s -w ./
	golangci-lint run

linux: clean
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o venus-auth ./cmd/server/*.go

clean:
	rm -rf venus-auth

.PHONY: clean
