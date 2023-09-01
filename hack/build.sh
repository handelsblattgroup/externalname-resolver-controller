#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

if [ -z "${BIN}" ]; then
    echo "BIN must be set"
    exit 1
fi
if [ -z "${PKG}" ]; then
    echo "PKG must be set"
    exit 1
fi
if [ -z "${ARCH}" ]; then
    echo "ARCH must be set"
    exit 1
fi
if [ -z "${VERSION}" ]; then
    echo "VERSION must be set"
    exit 1
fi
if [ -z "${COMMIT}" ]; then
    echo "COMMIT must be set"
    exit 1
fi
if [ -z "${OS}" ]; then
    echo "OS must be set"
    exit 1
fi

export CGO_ENABLED=0
export GOARCH="${ARCH}"
export GOOS="${OS}"
export GOCACHE=/go/cache
export GO111MODULE=auto

#go build -v -x                                                                                      \
go build                                                                                     \
    -installsuffix "static"                                                                         \
    -ldflags "-X ${PKG}/pkg/version.GitCommit=${COMMIT} -X ${PKG}/pkg/version.Version=${VERSION}" 	\
    -mod=vendor                                                                                     \
    -o $GOPATH/bin/${BIN}                                                                           \
    ./cmd/${BIN}/main.go