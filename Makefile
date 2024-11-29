SHELL := /bin/bash

.PHONY: lint-go
lint-go:
	golangci-lint run -v --config .golangci.yml

.PHONY: lint-proto
lint-proto:
	cd proto && buf lint

.PHONY: lint
lint: lint-go lint-proto

.PHONY: goimports
goimports:
	gofancyimports fix --local github.com/c4t-but-s4d/ctfcup-2024-igra -w $(shell find . -type f -name '*.go' -not -path "./pkg/proto/*")

.PHONY: test
test:
	go test -race -timeout 1m ./...

.PHONY: validate
validate: lint test

.PHONY: proto
proto:
	cd proto && buf generate
