SHELL=/usr/bin/env bash

BIN=bin/decompose
COP=cover.out

CMD=./cmd/decompose
ALL=./...

GIT_TAG=`git describe --abbrev=0 2>/dev/null || echo -n "no-tag"`
GIT_HASH=`git rev-parse --short HEAD 2>/dev/null || echo -n "no-git"`
BUILD_AT=`date +%FT%T%z`

LDFLAGS=-w -s \
		-X main.buildDate=${BUILD_AT} \
		-X main.gitVersion=${GIT_TAG} \
		-X main.gitHash=${GIT_HASH}

export CGO_ENABLED=0

.PHONY: build
build:
	@go build -trimpath -ldflags "${LDFLAGS}" -o "${BIN}" "${CMD}"

.PHONY: vet
vet:
	@go vet "${ALL}"

.PHONY: test
test: vet
	@CGO_ENABLED=1 go test -v -race -count 1 -tags=test \
				-cover -coverpkg="${ALL}" -coverprofile="${COP}" \
				"${ALL}"

.PHONY: test-update
test-update:
	@echo "clean-up..."
	@find . -name "*.golden" -delete
	@echo "update"
	@go test -tags=test "./internal/builder" -update

.PHONY: test-cover
test-cover: test
	@go tool cover -func="${COP}"

.PHONY: lint
lint: vet
	@golangci-lint run

.PHONY: markdown-fix
markdown-fix:
	# https://github.com/executablebooks/mdformat
	mdformat .

.PHONY: clean
clean:
	[ -f "${BIN}" ] && rm "${BIN}"
	[ -f "${COP}" ] && rm "${COP}"
