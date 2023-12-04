# ------------------------------------------------------------------------------
# Configuration - Make
# ------------------------------------------------------------------------------

# Some sensible Make defaults: https://tech.davis-hansson.com/p/make/
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c

# ------------------------------------------------------------------------------
# Configuration - Repository
# ------------------------------------------------------------------------------

MAKEFLAGS += --no-print-directory
REPO_URL ?= github.com/kong/chart-migrate
REPO_INFO ?= $(shell git config --get remote.origin.url)
GO_MOD_MAJOR_VERSION ?= $(subst $(REPO_URL)/,,$(shell go list -m))
TAG ?= $(shell git describe --tags)

ifndef COMMIT
  COMMIT := $(shell git rev-parse --short HEAD)
endif

# ------------------------------------------------------------------------------
# Configuration - Golang
# ------------------------------------------------------------------------------

export GO111MODULE=on

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

.PHONY: test.integration
test.integration:
		GOFLAGS="-tags=integration_tests" \
		go test $(GOTESTFLAGS) \
		-v \
		-race \
		./test/integration

.PHONY: test.golden.update
test.golden.update:
	@./hack/update-golden.sh
