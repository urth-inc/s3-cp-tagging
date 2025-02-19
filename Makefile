.PHONY: build test clean

VERSION ?= $(shell git describe --tags --always --dirty)
REVISION ?= $(shell git rev-parse --short HEAD)
BUILD_DATE ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')

LDFLAGS := -ldflags "-w -s -X main.Version=${VERSION} -X main.Revision=${REVISION} -X main.BuildDate=${BUILD_DATE}"
GO_BUILD_FLAGS := -trimpath

build:
	CGO_ENABLED=0 go build ${GO_BUILD_FLAGS} ${LDFLAGS} -o bin/s3-cp-tagging ./cmd/s3-cp-tagging

test:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

coverage-html: test
	go tool cover -html=coverage.out

clean:
	rm -rf bin/ 