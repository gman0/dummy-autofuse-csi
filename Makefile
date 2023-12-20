IMAGE_BUILD_TOOL ?= podman

SHELL      = /bin/bash
GIT_BRANCH = $(shell git rev-parse --abbrev-ref HEAD)
GIT_COMMIT = $(shell git rev-parse HEAD)
GIT_SHA    = $(shell git rev-parse --short HEAD)
GIT_TAG    = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
GIT_DIRTY  = $(shell test -n "`git status --porcelain`" && echo "dirty" || echo "clean")

IMAGE_TAG := ${GIT_BRANCH}
ifneq ($(GIT_TAG),)
    IMAGE_TAG = ${GIT_TAG}
endif

all: autofs csi fs workload

autofs:
	DEBUG=1 DONTSTRIP=1 $(MAKE) -C $@

csi:
	$(MAKE) -C $@

fs:
	$(MAKE) -C $@

workload:
	$(MAKE) -C $@

image: all
	$(IMAGE_BUILD_TOOL) build                                      \
		--build-arg RELEASE=$(IMAGE_TAG)                           \
		--build-arg GITREF=$(GIT_COMMIT)                           \
		--build-arg CREATED=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ") \
		-t registry.cern.ch/rvasek/dummy-autofuse-csi:$(IMAGE_TAG) \
		-f ./Dockerfile .

clean:
	$(MAKE) -C autofs clean
	$(MAKE) -C csi clean
	$(MAKE) -C fs clean
	$(MAKE) -C workload clean

.PHONY: all autofs csi fs workload image clean
