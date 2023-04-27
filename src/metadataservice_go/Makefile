-include env
include env

UNAME := $(shell uname)
GO_PATH := $(shell which go)

ifeq ($(UNAME), Linux)
MDS_BUILD_PATH = ${MDS_BUILD_PATH_LINUX}
SOURCE = .
ifeq ($(GO_PATH),)
	GO_PATH = ${GO_PATH_LINUX}
endif
endif

ifeq ($(UNAME), Darwin)
MDS_BUILD_PATH = ${MDS_BUILD_PATH_DARWIN}
SOURCE = source
endif

ifeq ($(GO_PATH),)
	GO_PATH = ${GO_REMOTE_PATH}
endif

.PHONY: help
## help: prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: tidy
## tidy: tidy up go modules
tidy:
	@${GO_PATH} mod tidy

.PHONY: vendor
## vendor: vendoring the modules
vendor:
	@${GO_PATH} mod vendor -v

.PHONY: clean
## clean: clean build path for Darwin
clean:
	@echo "Cleaning"
	@${GO_PATH} clean
	@rm -rf ${MDS_BUILD_PATH}

.PHONY: clean-all
## clean-all: clean everything
clean-all:
	@echo "Cleaning"
	@${GO_PATH} clean
	@rm -rf ${MDS_BUILD_PATH_LINUX}
	@rm -rf ${MDS_BUILD_PATH_DARWIN}

.PHONY: build
## build: build the MDS
build: clean build-mds

.PHONY: build-mds
## build-mds: build the MDS component (OS-dependent)
build-mds:
	@echo "Building MDS ..."
	@${GO_PATH} build -o ${MDS_BUILD_PATH}${MDS_BINARY} ./cmd/mds/*.go

## run-mds: run the MDS component
run-mds:
	@echo "Running MDS ..."
	@${GO_PATH} run ./cmd/mds

## run-mock-client: run the MockClient
run-mock-client:
	@echo "Running Mock-Client ..."
	@${GO_PATH} run ./cmd/mockclient


## git-commit: commit all files
git-commit:
	@echo "Commit"
	@git add . ; git commit -m 'auto push';

## git-push: push to main
git-push:
	@echo "Pushing to git main"
	@git add . ; git commit -m 'auto push'; git push origin main

## git-pull: pull from main
git-pull:
	@echo "Pulling from git main"
	@git pull origin main


### MDS LV -GATEWAYS
## run-mds: run the MDS component server A
rmda:
	@echo "Running MDS A..."
	@${GO_PATH} run ./cmd/rmda

## run-mds: run the MDS component server B
rmdb:
	@echo "Running MDS B..."
	@${GO_PATH} run ./cmd/rmdb

## run-mds: run the MDS component server C
rmdc:
	@echo "Running MDS C..."
	@${GO_PATH} run ./cmd/rmdc

## run-mock-client: run the MockClient A
mcka:
	@echo "Running Mock-Client A..."
	${GO_PATH}  run ./cmd/mcka


## run-mock-client: run the MockClient B
mckb:
	@echo "Running Mock-Client B..."
	@${GO_PATH} run ./cmd/mckb

## run-mock-client: run the MockClient C
mckc:
	@echo "Running Mock-Client C..."
	@${GO_PATH} run ./cmd/mckc

## run-mock-client: run the MockClient that triggers pubsub events in loop
mckpub:
	@echo "Running Mock-Client PUB SUB..."
	@${GO_PATH} run ./cmd/mckpub


### MDS LV - END

.PHONY: protos
## protos: generate the protos
protos:
	@echo "Generating the Protos ..."
	@rm -rf ./protos;
	@mkdir ./protos
	@protoc --experimental_allow_proto3_optional -I ../protos ../protos/*.proto  --go_out=./protos
	@protoc --experimental_allow_proto3_optional -I ../protos ../protos/*.proto  --go-grpc_out=require_unimplemented_servers=false:./protos


.PHONY: build-docker-images
## build-docker-images: build docker images
build-docker-images:
	@echo "Building Docker images ..."
	@docker build --force-rm -t ${DOCKER_REGISTRY}/geds-mds-go -f ./deployment/docker/images/geds-mds-go.dockerfile .
	@docker tag ${DOCKER_REGISTRY}/geds-mds-go ${DOCKER_REGISTRY}/geds-mds-go:${DOCKER_TAG_ENV}

.PHONY: push-docker-images
## push-docker-images: push docker images
push-docker-images:
	@echo "Pushing Docker images ..."
	@docker push ${DOCKER_REGISTRY}/geds-mds-go:${DOCKER_TAG_ENV}


