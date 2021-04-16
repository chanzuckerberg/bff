SHA=$(shell git rev-parse --short HEAD)
VERSION=$(shell cat VERSION)
DIRTY=$(shell if `git diff-index --quiet HEAD --`; then echo false; else echo true;  fi)
LDFLAGS=-ldflags "-w -s -X github.com/chanzuckerberg/bff/util.GitSha=${SHA} -X github.com/chanzuckerberg/bff/util.Version=${VERSION} -X github.com/chanzuckerberg/bff/util.Dirty=${DIRTY}"
export GO111MODULE=on

all: test
.PHONY: all

setup: ## setup development dependencies
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.39.0
.PHONY: setup

lint: setup ## run the fast go linters
	./bin/golangci-lint run
.PHONY: lint

release: build ## run a release
	./bff bump
	git push
	goreleaser release --rm-dist
.PHONY: release

release-snapshot: ## run a release
	goreleaser release --rm-dist --snapshot
.PHONY: release-snapshot

build: deps ## build the binary
	go build ${LDFLAGS} .
.PHONY: build

coverage: ## run the go coverage tool, reading file coverage.txt
	go tool cover -html=coverage.txt
.PHONY: coverage

deps:
	go mod tidy
.PHONY: deps

test: deps ## run the tests
	go test -coverprofile=coverage.txt -covermode=atomic ./...
.PHONY: test

install: deps ## install the bff binary in $GOPATH/bin
	go install ${LDFLAGS} .
.PHONY: install

help: ## display help for this makefile
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
.PHONY: help

clean: ## clean the repo
	rm bff 2>/dev/null || true
	go clean
	rm -rf dist
.PHONY: clean

check-mod:
	go mod tidy
	git diff --exit-code -- go.mod go.sum
.PHONY: check-mod
