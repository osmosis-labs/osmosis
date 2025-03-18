#!/usr/bin/make -f

# the subcommands are located in the specific makefiles
include scripts/makefiles/build.mk
include scripts/makefiles/deps.mk
include scripts/makefiles/docker.mk
include scripts/makefiles/e2e.mk
include scripts/makefiles/lint.mk
include scripts/makefiles/localnet.mk
include scripts/makefiles/proto.mk
include scripts/makefiles/release.mk
include scripts/makefiles/sqs.mk
include scripts/makefiles/tests.mk

.DEFAULT_GOAL := help
help:
	@echo "Available top-level commands:"
	@echo ""
	@echo "Usage:"
	@echo "    make [command]"
	@echo ""
	@echo "  make build                 Build osmosisd binary"
	@echo "  make build-help            Show available build commands"
	@echo "  make deps                  Show available deps commands"
	@echo "  make docker                Show available docker commands"
	@echo "  make e2e                   Show available e2e commands"
	@echo "  make go-mock-update        Generate mock files"
	@echo "  make install               Install osmosisd binary"
	@echo "  make lint                  Show available lint commands"
	@echo "  make localnet              Show available localnet commands"
	@echo "  make proto                 Show available proto commands"
	@echo "  make release               Show available release commands"
	@echo "  make release-help          Show available release commands"
	@echo "  make run-querygen          Generating GRPC queries, and queryproto logic"
	@echo "  make sqs                   Show available sqs commands"
	@echo "  make test                  Show available test commands"
	@echo ""
	@echo "Run 'make [subcommand]' to see the available commands for each subcommand."

VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')

LEDGER_ENABLED ?= true
SDK_PACK := $(shell go list -m github.com/cosmos/cosmos-sdk | sed  's/ /\@/g')
BUILDDIR ?= $(CURDIR)/build
DOCKER := $(shell which docker)
E2E_UPGRADE_VERSION := "v28"
#SHELL := /bin/bash

# Go version to be used in docker images
GO_VERSION := $(shell cat go.mod | grep -E 'go [0-9].[0-9]+' | cut -d ' ' -f 2)
GO_MAJOR_MINOR := $(shell cat go.mod | grep -E 'go [0-9].[0-9]+' | cut -d ' ' -f 2 | cut -d '.' -f 1-2)
# currently installed Go version
GO_MODULE := $(shell cat go.mod | grep "module " | cut -d ' ' -f 2)
GO_MAJOR_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f1)
GO_MINOR_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f2)
# minimum supported Go version
GO_MINIMUM_MAJOR_VERSION = $(shell cat go.mod | grep -E 'go [0-9].[0-9]+' | cut -d ' ' -f2 | cut -d'.' -f1)
GO_MINIMUM_MINOR_VERSION = $(shell cat go.mod | grep -E 'go [0-9].[0-9]+' | cut -d ' ' -f2 | cut -d'.' -f2)
# message to be printed if Go does not meet the minimum required version
GO_VERSION_ERR_MSG = "ERROR: Go version $(GO_MINIMUM_MAJOR_VERSION).$(GO_MINIMUM_MINOR_VERSION)+ is required"

export GO111MODULE = on

# process build tags

build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

ifeq (cleveldb,$(findstring cleveldb,$(OSMOSIS_BUILD_OPTIONS)))
  build_tags += gcc
else ifeq (rocksdb,$(findstring rocksdb,$(OSMOSIS_BUILD_OPTIONS)))
  build_tags += gcc
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
whitespace := $(whitespace) $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# process linker flags

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=osmosis \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=osmosisd \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)"

ifeq (cleveldb,$(findstring cleveldb,$(OSMOSIS_BUILD_OPTIONS)))
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
else ifeq (rocksdb,$(findstring rocksdb,$(OSMOSIS_BUILD_OPTIONS)))
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=rocksdb
endif
ifeq (,$(findstring nostrip,$(OSMOSIS_BUILD_OPTIONS)))
  ldflags += -w -s
endif
ifeq ($(LINK_STATICALLY),true)
	ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'
# check for nostrip option
ifeq (,$(findstring nostrip,$(OSMOSIS_BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
endif

# Note that this skips certain tests that are not supported on WSL
# This is a workaround to enable quickly running full unit test suite locally
# on WSL without failures. The failures are stemming from trying to upload
# wasm code. An OS permissioning issue.
is_wsl := $(shell uname -a | grep -i Microsoft)
ifeq ($(is_wsl),)
    # Not in WSL
    SKIP_WASM_WSL_TESTS := "false"
else
    # In WSL
    SKIP_WASM_WSL_TESTS := "true"
endif
###############################################################################
###                            Build & Install                              ###
###############################################################################

update-deps:
	@if [ -n "$(SDK_HASH)" ]; then \
		echo "Updating cosmos-sdk to hash $(SDK_HASH)"; \
		SDK_VERSION=$$(go get github.com/osmosis-labs/cosmos-sdk@$(SDK_HASH) 2>&1 | sed -n 's/.*github.com\/osmosis-labs\/cosmos-sdk@\([^ :]*\).*/\1/p'); \
		echo "Extracted SDK version: $${SDK_VERSION}"; \
		sed -i.bak "s|github.com/cosmos/cosmos-sdk => github.com/osmosis-labs/cosmos-sdk .*|github.com/cosmos/cosmos-sdk => github.com/osmosis-labs/cosmos-sdk $${SDK_VERSION}|" go.mod; \
	fi
	@if [ -n "$(COMET_HASH)" ]; then \
		echo "Updating cometbft to hash $(COMET_HASH)"; \
		COMET_VERSION=$$(go get github.com/osmosis-labs/cometbft@$(COMET_HASH) 2>&1 | sed -n 's/.*github.com\/osmosis-labs\/cometbft@\([^ :]*\).*/\1/p'); \
		echo "Extracted Comet version: $${COMET_VERSION}"; \
		sed -i.bak "s|github.com/cometbft/cometbft => github.com/osmosis-labs/cometbft .*|github.com/cometbft/cometbft => github.com/osmosis-labs/cometbft $${COMET_VERSION}|" go.mod; \
	fi
	@if [ -n "$(SDK_HASH)" ] || [ -n "$(COMET_HASH)" ]; then \
		go mod tidy; \
	fi

build: build-check-version go.sum
	@if [ -n "$(SDK_HASH)" ] || [ -n "$(COMET_HASH)" ]; then \
		cp go.mod go.mod.backup; \
		cp go.sum go.sum.backup; \
		$(MAKE) update-deps; \
	fi
	mkdir -p $(BUILDDIR)/
	GOWORK=off go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/ $(GO_MODULE)/cmd/osmosisd
	@if [ -n "$(SDK_HASH)" ] || [ -n "$(COMET_HASH)" ]; then \
		mv go.mod.backup go.mod; \
		mv go.sum.backup go.sum; \
		rm -f go.mod.bak; \
		go mod tidy; \
	fi

install: build-check-version go.sum
	@if [ -n "$(SDK_HASH)" ] || [ -n "$(COMET_HASH)" ]; then \
		cp go.mod go.mod.backup; \
		cp go.sum go.sum.backup; \
		$(MAKE) update-deps; \
	fi
	GOWORK=off go install -mod=readonly $(BUILD_FLAGS) $(GO_MODULE)/cmd/osmosisd
	@if [ -n "$(SDK_HASH)" ] || [ -n "$(COMET_HASH)" ]; then \
		mv go.mod.backup go.mod; \
		mv go.sum.backup go.sum; \
		rm -f go.mod.bak; \
		go mod tidy; \
	fi

###############################################################################
###                                Gen                                      ###
###############################################################################

run-querygen:
	@go run cmd/querygen/main.go


###############################################################################
###                                Go Mock                                  ###
###############################################################################

go-mock-update:
	mockgen -source=x/poolmanager/types/expected_keepers.go -destination=tests/mocks/pool_module.go -package=mocks
	mockgen -source=x/poolmanager/types/pool.go -destination=tests/mocks/pool.go -package=mocks
	mockgen -source=x/gamm/types/pool.go -destination=tests/mocks/cfmm_pool.go -package=mocks
	mockgen -source=x/concentrated-liquidity/types/cl_pool_extensionI.go -destination=tests/mocks/cl_pool.go -package=mocks

###############################################################################
###                                Release                                  ###
###############################################################################
GORELEASER_IMAGE := ghcr.io/goreleaser/goreleaser-cross:v$(GO_MAJOR_MINOR)
COSMWASM_VERSION := $(shell go list -m github.com/CosmWasm/wasmvm/v2 | sed 's/.* //')

ifdef GITHUB_TOKEN
ifdef S3_ENDPOINT
ifdef S3_REGION
ifdef AWS_ACCESS_KEY_ID
ifdef AWS_SECRET_ACCESS_KEY

release:
	docker run \
		--rm \
		-e GITHUB_TOKEN=$(GITHUB_TOKEN) \
		-e COSMWASM_VERSION=$(COSMWASM_VERSION) \
		-e S3_ENDPOINT=$(S3_ENDPOINT) \
		-e S3_REGION=$(S3_REGION) \
		-e AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) \
		-e AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY) \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/osmosisd \
		-w /go/src/osmosisd \
		$(GORELEASER_IMAGE) \
		release \
		--clean

else
release:
	@echo "Error: GITHUB_TOKEN is not defined. Please define it before running 'make release'."
endif
else
release:
	@echo "Error: S3_ENDPOINT is not defined. Please define it before running 'make release'."
endif
else
release:
	@echo "Error: S3_REGION is not defined. Please define it before running 'make release'."
endif
else
release:
	@echo "Error: AWS_ACCESS_KEY_ID is not defined. Please define it before running 'make release'."
endif
else
release:
	@echo "Error: AWS_SECRET_ACCESS_KEY is not defined. Please define it before running 'make release'."
endif

release-test:
	docker run \
		--rm \
		-e COSMWASM_VERSION=$(COSMWASM_VERSION) \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/osmosisd \
		-w /go/src/osmosisd \
		$(GORELEASER_IMAGE) \
		release \
		--snapshot --clean

.PHONY: all build-linux install format lint \
	go-mod-cache draw-deps clean build build-contract-tests-hooks \
	test test-all test-build test-cover test-unit test-race benchmark \
	release release-dry-run release-snapshot update-deps
