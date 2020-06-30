# set default shell
SHELL = bash -e -o pipefail

default: build

VERSION ?= latest
## Docker related
DOCKER_EXTRA_ARGS        ?=
DOCKER_REGISTRY          ?=
DOCKER_REPOSITORY        ?=
DOCKER_TAG               ?= ${VERSION}
IMAGE_NAME               := ${DOCKER_REGISTRY}${DOCKER_REPOSITORY}onos-stress:${DOCKER_TAG}
DOCKER_BUILD_ARGS        ?=${DOCKER_EXTRA_ARGS} --build-arg version="${VERSION}"

help:
	@echo "Usage: make [<target>]"
	@echo "where available targets are:"
	@echo
	@echo "build      		 : Build Onos Stress docker image"
	@echo "help              : Print this help"
	@echo

build:
	docker build $(DOCKER_BUILD_ARGS) -t ${IMAGE_NAME} -f Dockerfile .
