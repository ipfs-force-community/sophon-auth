SHELL=/usr/bin/env bash

all: clean auth-server

auth-server:
	go build -o auth-server ./cmd/server/*.go

lint:
	gofmt -s -w ./
	golangci-lint run

linux: clean
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o auth-server ./cmd/server/*.go

clean:
	rm -rf auth-server

.PHONY: clean
