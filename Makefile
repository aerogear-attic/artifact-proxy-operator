DOCKERHOST = docker.io
DOCKERORG = aerogear
IMAGENAME = artifact-proxy-operator
TAG = latest
USER=$(shell id -u)
PWD=$(shell pwd)
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