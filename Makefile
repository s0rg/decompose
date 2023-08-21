SHELL=/usr/bin/env bash

BIN=bin/decompose
CMD=./cmd/decompose
COP=cover.out

GIT_TAG=`git describe --abbrev=0 2>/dev/null || echo -n "no-tag"`
GIT_HASH=`git rev-parse --short HEAD 2>/dev/null || echo -n "no-git"`
BUILD_AT=`date +%FT%T%z`

LDFLAGS=-w -s -X main.gitHash=${GIT_HASH} -X main.buildDate=${BUILD_AT} -X main.gitVersion=${GIT_TAG}

export CGO_ENABLED=0

.PHONY: build

build: vet
	go build -ldflags "${LDFLAGS}" -o "${BIN}" "${CMD}"

vet:
	go vet ./...

test: vet
	CGO_ENABLED=1 go test -race -count 1 -v -tags=test -coverprofile="${COP}" ./...

test-cover: test
	go tool cover -func="${COP}"

lint:
	golangci-lint run

markdown-fix:
	# https://github.com/executablebooks/mdformat
	mdformat .

clean:
	[ -f "${BIN}" ] && rm "${BIN}"
	[ -f "${COP}" ] && rm "${COP}"
