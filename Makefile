SHA=$(shell git rev-parse --short HEAD)
VERSION=$(shell cat VERSION)
DIRTY=$(shell if `git diff-index --quiet HEAD --`; then echo false; else echo true;  fi)
# TODO add release flag
# LDFLAGS=-ldflags "-w -s -X github.com/chanzuckerberg/bff/util.GitSha=${SHA} -X github.com/chanzuckerberg/bff/util.Version=${VERSION} -X github.com/chanzuckerberg/bff/util.Dirty=${DIRTY}"

all: test install

setup:
	go get github.com/rakyll/gotest
	go install github.com/rakyll/gotest
	
lint: ## run the fast go linters
	gometalinter --vendor --fast ./...

lint-slow: ## run all linters, even the slow ones
	gometalinter --vendor --deadline 120s ./...

release: ## run a release
	./release
	git push
	goreleaser release --rm-dist

release-snapshot: ## run a release
	goreleaser release --rm-dist --snapshot

build: ## build the binary
	# go build ${LDFLAGS} .
	go build .

coverage: ## run the go coverage tool, reading file coverage.out
	go tool cover -html=coverage.out

test: ## run the tests
	gotest -coverprofile=coverage.txt -covermode=atomic ./...

install: ## install the bff binary in $GOPATH/bin
	# go install ${LDFLAGS} .
	go install .

help: ## display help for this makefile
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

clean: ## clean the repo
	rm bff 2>/dev/null || true
	go clean
	rm -rf dist

.PHONY: build clean coverage test install lint lint-slow release help
