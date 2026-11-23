SHELL := /usr/bin/env bash

.PHONY: fmt test vet build check

fmt:
	gofmt -w cmd internal schemas

test:
	go test -race -cover ./...

vet:
	go vet ./...

build:
	go build -trimpath -o bin/obsctl ./cmd/obsctl

check:
	test -z "$$(gofmt -l .)"
	go vet ./...
	go test -race ./...
	go build -trimpath -o bin/obsctl ./cmd/obsctl

