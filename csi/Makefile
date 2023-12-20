GOLDFLAGS := "-w -s"
BUILD_DIR ?= build

GIT_BRANCH = $(shell git rev-parse --abbrev-ref HEAD)
GIT_COMMIT = $(shell git rev-parse HEAD)
GIT_SHA    = $(shell git rev-parse --short HEAD)
GIT_TAG    = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
GIT_DIRTY  = $(shell test -n "`git status --porcelain`" && echo "dirty" || echo "clean")

ifdef VERSION
	BINARY_VERSION = $(VERSION)
endif
BINARY_VERSION ?= ${GIT_TAG}

BASE_PKG = github.com/gman0/dummy-autofuse-csi
# Only set Version if building a tag or VERSION is set
ifneq ($(BINARY_VERSION),)
	LDFLAGS += -X ${BASE_PKG}/internal/version.version=${BINARY_VERSION}
endif

# Clear the "unreleased" string in BuildMetadata
ifneq ($(GIT_TAG),)
	LDFLAGS += -X ${BASE_PKG}/internal/version.metadata=
endif
LDFLAGS += -X ${BASE_PKG}/internal/version.commit=${GIT_COMMIT}
LDFLAGS += -X ${BASE_PKG}/internal/version.treestate=${GIT_DIRTY}

$(shell mkdir -p $(BUILD_DIR))

all: automount-reconciler automount-runner csi-driver

automount-reconciler:
	CGO_ENABLED=0 go build -ldflags $(GOLDFLAGS) -o $(BUILD_DIR)/$@ cmd/$@/main.go

automount-runner:
	CGO_ENABLED=0 go build -ldflags $(GOLDFLAGS) -o $(BUILD_DIR)/$@ cmd/$@/main.go

csi-driver:
	CGO_ENABLED=0 go build -ldflags $(GOLDFLAGS) -o $(BUILD_DIR)/$@ cmd/$@/main.go

clean:
	rm -rf $(BUILD_DIR)

.PHONY: clean
