# This file is canonically located at fid-dev/go-hello-world
# Please make changes portable whenever possible.
# Please commit any portable changes back.

ARCH ?= amd64
OS ?= linux
BUILD_IMAGE ?= golang:1.20
GOPATH ?= $(shell go env GOPATH)
GOPATH_SRC := $(GOPATH)/src/
CURRENT_WORK_DIR := $(shell pwd)
ALL_FILES=$(shell find . -path ./vendor -prune -type f -o -name *.proto)
PKG=github.com/handelsblattgroup/externalname-resolver-controller

GIT_COMMIT := $(shell git rev-parse HEAD)
VERSION ?= $(GIT_COMMIT)

BIN ?= ""
TAG ?= $(VERSION)
IMAGE ?= ""
YES ?= ""

.PHONY: all test build clean container build-dirs clean-dirs push-container check-image

all: build

clean: clean-dirs

container: check-image dist/$(ARCH)/$(BIN)
	@docker build --quiet -t $(IMAGE):$(TAG) -f hack/release/Dockerfile .
	$(info container image built $(IMAGE):$(TAG))

ifeq ($(YES), 1)
push-container: check-image container
	docker push $(IMAGE):$(TAG)
else
push-container:
	$(warning push disabled. to enable set environment YES=1)
endif

ifndef IMAGE
check-image:
	  $(error env IMAGE is undefined)
else
check-image:
	  $(info target image is $(IMAGE))
endif

build: $(subst cmd, dist/$(ARCH), $(wildcard cmd/*))

dist/$(ARCH)/%: build-dirs
	$(info building binary $(notdir $@))
	@docker run \
		--rm \
		-u $$(id -u):$$(id -g) \
		-v "$$(pwd):/src" \
		-v "$$(pwd)/dist/$(OS)/$(ARCH):/go/bin" \
		-v "$$(pwd)/.gocache/:/go/cache" \
		-w /src \
		$(BUILD_IMAGE) \
		/bin/sh -c " \
			ARCH=$(ARCH) \
			OS=$(OS) \
			VERSION=$(VERSION) \
			COMMIT=$(GIT_COMMIT) \
			PKG=$(PKG) \
			BIN=$(notdir $@) \
			GO111MODULE=auto \
			./hack/build.sh \
		"

test: build-dirs
	$(info run test)
	@docker run \
		--rm \
		-u $$(id -u):$$(id -g) \
		-v "$$(pwd):/src" \
		-v "$$(pwd)/dist/$(OS)/$(ARCH):/go/bin" \
		-v "$$(pwd)/.gocache/:/go/cache" \
		-w /src \
		$(BUILD_IMAGE) \
		/bin/sh -c "CGO_ENABLED=1 GO111MODULE=auto GOCACHE=/go/cache go test -race -mod=vendor ./..."

build-dirs:
	@echo "build-dirs"
	@mkdir -p ./dist/$(OS)/$(ARCH)
	@mkdir -p ./.gocache

clean-dirs:
	$(info clean up cache and dist folders)
	@rm -rf ./dist
	@rm -rf ./.gocache