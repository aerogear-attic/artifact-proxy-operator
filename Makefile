DOCKERHOST = docker.io
DOCKERORG = aerogear
IMAGENAME = artifact-proxy-operator
TAG = latest
USER=$(shell id -u)
PWD=$(shell pwd)
PKG     = github.com/aerogear/artifact-proxy-operator
TOP_SRC_DIRS   = pkg
TEST_DIRS     ?= $(shell sh -c "find $(TOP_SRC_DIRS) -name \\*_test.go \
                   -exec dirname {} \\; | sort | uniq")
BIN_DIR := $(GOPATH)/bin
SHELL = /bin/bash

LDFLAGS=-ldflags "-w -s -X main.Version=${TAG}"

build_and_push: build_binary docker_build docker_push

.PHONY: build_binary
build_binary:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ./cmd/artifact-proxy-operator
	
.PHONY: docker_build
docker_build:
	docker build -t $(DOCKERHOST)/$(DOCKERORG)/$(IMAGENAME):$(TAG) .

.PHONY: docker_push
docker_push:
	docker push $(DOCKERHOST)/$(DOCKERORG)/$(IMAGENAME):$(TAG)

.PHONY: test-coveralls
test-coveralls:
	echo "mode: count" > coverage-all.out
	$(foreach test_dir,$(addprefix $(PKG)/,$(TEST_DIRS)),\
		go test -coverprofile=coverage.out -covermode=count $(test_dir);\
		tail -n +2 coverage.out >> coverage-all.out;)
	goveralls -coverprofile=coverage-all.out -service=travis-pro -repotoken $(COVERALLS_TOKEN)

.PHONY: test-html
test-html:
	echo "mode: count" > coverage-all.out
	$(foreach test_dir,$(addprefix $(PKG)/,$(TEST_DIRS)),\
		go test -coverprofile=coverage.out -covermode=count $(test_dir);\
		tail -n +2 coverage.out >> coverage-all.out;)
	go tool cover -html=coverage-all.out
