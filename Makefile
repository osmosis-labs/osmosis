#!/usr/bin/make -f

VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
LEDGER_ENABLED ?= true
SDK_PACK := $(shell go list -m github.com/cosmos/cosmos-sdk | sed  's/ /\@/g')
GO_VERSION := $(shell cat go.mod | grep -E 'go [0-9].[0-9]+' | cut -d ' ' -f 2)
DOCKER := $(shell which docker)
BUILDDIR ?= $(CURDIR)/build
E2E_UPGRADE_VERSION := "v15"


GO_MAJOR_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f1)
GO_MINOR_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f2)

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
whitespace += $(whitespace)
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

###############################################################################
###                                  Build                                  ###
###############################################################################

check_version:
ifneq ($(GO_MINOR_VERSION),19)
	@echo "ERROR: Go version 1.19 is required for this version of Osmosis."
	exit 1
endif

all: install lint test

BUILD_TARGETS := build install

build: BUILD_ARGS=-o $(BUILDDIR)/

$(BUILD_TARGETS): check_version go.sum $(BUILDDIR)/
	GOWORK=off go $@ -mod=readonly $(BUILD_FLAGS) $(BUILD_ARGS) ./...

$(BUILDDIR)/:
	mkdir -p $(BUILDDIR)/

# Cross-building for arm64 from amd64 (or viceversa) takes
# a lot of time due to QEMU virtualization but it's the only way (afaik)
# to get a statically linked binary with CosmWasm

build-reproducible: build-reproducible-amd64 build-reproducible-arm64

build-reproducible-amd64: go.sum $(BUILDDIR)/
	$(DOCKER) buildx create --name osmobuilder || true
	$(DOCKER) buildx use osmobuilder
	$(DOCKER) buildx build \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg GIT_VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(COMMIT) \
		--build-arg RUNNER_IMAGE=alpine:3.16 \
		--platform linux/amd64 \
		-t osmosis:local-amd64 \
		--load \
		-f Dockerfile .
	$(DOCKER) rm -f osmobinary || true
	$(DOCKER) create -ti --name osmobinary osmosis:local-amd64
	$(DOCKER) cp osmobinary:/bin/osmosisd $(BUILDDIR)/osmosisd-linux-amd64
	$(DOCKER) rm -f osmobinary

build-reproducible-arm64: go.sum $(BUILDDIR)/
	$(DOCKER) buildx create --name osmobuilder || true
	$(DOCKER) buildx use osmobuilder
	$(DOCKER) buildx build \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg GIT_VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(COMMIT) \
		--build-arg RUNNER_IMAGE=alpine:3.16 \
		--platform linux/arm64 \
		-t osmosis:local-arm64 \
		--load \
		-f Dockerfile .
	$(DOCKER) rm -f osmobinary || true
	$(DOCKER) create -ti --name osmobinary osmosis:local-arm64
	$(DOCKER) cp osmobinary:/bin/osmosisd $(BUILDDIR)/osmosisd-linux-arm64
	$(DOCKER) rm -f osmobinary

build-linux: go.sum
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

build-contract-tests-hooks:
	mkdir -p $(BUILDDIR)
	go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/ ./cmd/contract_tests

go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@GOWORK=off go mod verify

draw-deps:
	@# requires brew install graphviz or apt-get install graphviz
	go get github.com/RobotsAndPencils/goviz
	@goviz -i ./cmd/osmosisd -d 2 | dot -Tpng -o dependency-graph.png

clean:
	rm -rf $(CURDIR)/artifacts/

distclean: clean
	rm -rf vendor/

###############################################################################
###                                  Proto                                  ###
###############################################################################

proto-all: proto-format proto-gen

proto:
	@echo
	@echo "=========== Generate Message ============"
	@echo
	./scripts/protocgen.sh
	@echo
	@echo "=========== Generate Complete ============"
	@echo

test:
	@go test -v ./x/...

docs:
	@echo
	@echo "=========== Generate Message ============"
	@echo
	./scripts/generate-docs.sh

	statik -src=client/docs/static -dest=client/docs -f -m
	@if [ -n "$(git status --porcelain)" ]; then \
        echo "\033[91mSwagger docs are out of sync!!!\033[0m";\
        exit 1;\
    else \
        echo "\033[92mSwagger docs are in sync\033[0m";\
    fi
	@echo
	@echo "=========== Generate Complete ============"
	@echo
.PHONY: docs

protoVer=v0.8
protoImageName=osmolabs/osmo-proto-gen:$(protoVer)
containerProtoGen=cosmos-sdk-proto-gen-$(protoVer)
containerProtoFmt=cosmos-sdk-proto-fmt-$(protoVer)

proto-gen:
	@echo "Generating Protobuf files"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoGen}$$"; then docker start -a $(containerProtoGen); else docker run --name $(containerProtoGen) -v $(CURDIR):/workspace --workdir /workspace $(protoImageName) \
		sh ./scripts/protocgen.sh; fi

proto-format:
	@echo "Formatting Protobuf files"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoFmt}$$"; then docker start -a $(containerProtoFmt); else docker run --name $(containerProtoFmt) -v $(CURDIR):/workspace --workdir /workspace tendermintdev/docker-build-proto \
		find ./ -not -path "./third_party/*" -name "*.proto" -exec clang-format -i {} \; ; fi

proto-image-build:
	@DOCKER_BUILDKIT=1 docker build -t $(protoImageName) -f ./proto/Dockerfile ./proto

proto-image-push:
	docker push $(protoImageName)

###############################################################################
###                                Querygen                                 ###
###############################################################################

run-querygen:
	@go run cmd/querygen/main.go

###############################################################################
###                           Tests & Simulation                            ###
###############################################################################

PACKAGES_UNIT=$(shell go list ./... ./osmomath/... ./osmoutils/... ./x/ibc-hooks/... | grep -E -v 'tests/simulator|e2e')
PACKAGES_E2E=$(shell go list ./... | grep '/e2e')
PACKAGES_SIM=$(shell go list ./... | grep '/tests/simulator')
TEST_PACKAGES=./...

test: test-unit test-build

test-all: test-race test-cover

test-unit:
	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock norace' $(PACKAGES_UNIT)

test-race:
	@VERSION=$(VERSION) go test -mod=readonly -race -tags='ledger test_ledger_mock' $(PACKAGES_UNIT)

test-cover:
	@VERSION=$(VERSION) go test -mod=readonly -timeout 30m -coverprofile=coverage.txt -tags='norace' -covermode=atomic $(PACKAGES_UNIT)

test-sim-suite:
	@VERSION=$(VERSION) go test -mod=readonly $(PACKAGES_SIM)

test-sim-app:
	@VERSION=$(VERSION) go test -mod=readonly -run ^TestFullAppSimulation -v $(PACKAGES_SIM)

test-sim-determinism:
	@VERSION=$(VERSION) go test -mod=readonly -run ^TestAppStateDeterminism -v $(PACKAGES_SIM)

test-sim-bench:
	@VERSION=$(VERSION) go test -benchmem -run ^BenchmarkFullAppSimulation -bench ^BenchmarkFullAppSimulation -cpuprofile cpu.out $(PACKAGES_SIM)

# test-e2e runs a full e2e test suite
# deletes any pre-existing Osmosis containers before running.
#
# Deletes Docker resources at the end.
# Utilizes Go cache.
test-e2e: e2e-setup test-e2e-ci e2e-remove-resources

# test-e2e-ci runs a full e2e test suite
# does not do any validation about the state of the Docker environment
# As a result, avoid using this locally.
test-e2e-ci:
	@VERSION=$(VERSION) OSMOSIS_E2E=True OSMOSIS_E2E_DEBUG_LOG=False OSMOSIS_E2E_UPGRADE_VERSION=$(E2E_UPGRADE_VERSION)  go test -mod=readonly -timeout=25m -v $(PACKAGES_E2E)

# test-e2e-debug runs a full e2e test suite but does
# not attempt to delete Docker resources at the end.
test-e2e-debug: e2e-setup
	@VERSION=$(VERSION) OSMOSIS_E2E=True OSMOSIS_E2E_DEBUG_LOG=True OSMOSIS_E2E_UPGRADE_VERSION=$(E2E_UPGRADE_VERSION) OSMOSIS_E2E_SKIP_CLEANUP=True go test -mod=readonly -timeout=25m -v $(PACKAGES_E2E) -count=1

# test-e2e-short runs the e2e test with only short tests.
# Does not delete any of the containers after running.
# Deletes any existing containers before running.
# Does not use Go cache.
test-e2e-short: e2e-setup
	@VERSION=$(VERSION) OSMOSIS_E2E=True OSMOSIS_E2E_DEBUG_LOG=True OSMOSIS_E2E_SKIP_UPGRADE=True OSMOSIS_E2E_SKIP_IBC=True OSMOSIS_E2E_SKIP_STATE_SYNC=True OSMOSIS_E2E_SKIP_CLEANUP=True go test -mod=readonly -timeout=25m -v $(PACKAGES_E2E) -count=1

test-mutation:
	@bash scripts/mutation-test.sh $(MODULES)

benchmark:
	@go test -mod=readonly -bench=. $(PACKAGES_UNIT)

build-e2e-script:
	mkdir -p $(BUILDDIR)
	go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/ ./tests/e2e/initialization/$(E2E_SCRIPT_NAME)

docker-build-debug:
	@DOCKER_BUILDKIT=1 docker build -t osmosis:${COMMIT} --build-arg BASE_IMG_TAG=debug --build-arg RUNNER_IMAGE=$(RUNNER_BASE_IMAGE_ALPINE) -f Dockerfile .
	@DOCKER_BUILDKIT=1 docker tag osmosis:${COMMIT} osmosis:debug

docker-build-e2e-init-chain:
	@DOCKER_BUILDKIT=1 docker build -t osmosis-e2e-init-chain:debug --build-arg E2E_SCRIPT_NAME=chain --platform=linux/x86_64 -f tests/e2e/initialization/init.Dockerfile .

docker-build-e2e-init-node:
	@DOCKER_BUILDKIT=1 docker build -t osmosis-e2e-init-node:debug --build-arg E2E_SCRIPT_NAME=node --platform=linux/x86_64 -f tests/e2e/initialization/init.Dockerfile .

e2e-setup: e2e-check-image-sha e2e-remove-resources
	@echo Finished e2e environment setup, ready to start the test

e2e-check-image-sha:
	tests/e2e/scripts/run/check_image_sha.sh

e2e-remove-resources:
	tests/e2e/scripts/run/remove_stale_resources.sh

.PHONY: test-mutation

###############################################################################
###                                Docker                                  ###
###############################################################################

RUNNER_BASE_IMAGE_DISTROLESS := gcr.io/distroless/static-debian11
RUNNER_BASE_IMAGE_ALPINE := alpine:3.16
RUNNER_BASE_IMAGE_NONROOT := gcr.io/distroless/static-debian11:nonroot

docker-build:
	@DOCKER_BUILDKIT=1 docker build \
		-t osmosis:local \
		-t osmosis:local-distroless \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg RUNNER_IMAGE=$(RUNNER_BASE_IMAGE_DISTROLESS) \
		--build-arg GIT_VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(COMMIT) \
		-f Dockerfile .

docker-build-distroless: docker-build

docker-build-alpine:
	@DOCKER_BUILDKIT=1 docker build \
		-t osmosis:local-alpine \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg RUNNER_IMAGE=$(RUNNER_BASE_IMAGE_ALPINE) \
		--build-arg GIT_VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(COMMIT) \
		-f Dockerfile .

docker-build-nonroot:
	@DOCKER_BUILDKIT=1 docker build \
		-t osmosis:local-nonroot \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg RUNNER_IMAGE=$(RUNNER_BASE_IMAGE_NONROOT) \
		--build-arg GIT_VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(COMMIT) \
		-f Dockerfile .

###############################################################################
###                                Linting                                  ###
###############################################################################

lint:
	@echo "--> Running linter"
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run --timeout=10m
	@docker run -v $(PWD):/workdir ghcr.io/igorshubovych/markdownlint-cli:latest "**/*.md"

format:
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run ./... --fix
	@go run mvdan.cc/gofumpt -l -w x/ app/ ante/ tests/
	@docker run -v $(PWD):/workdir ghcr.io/igorshubovych/markdownlint-cli:latest "**/*.md" --fix

mdlint:
	@echo "--> Running markdown linter"
	@docker run -v $(PWD):/workdir ghcr.io/igorshubovych/markdownlint-cli:latest "**/*.md"

markdown:
	@docker run -v $(PWD):/workdir ghcr.io/igorshubovych/markdownlint-cli:latest "**/*.md" --fix

###############################################################################
###                                Localnet                                 ###
###############################################################################

localnet-keys:
	. tests/localosmosis/scripts/add_keys.sh

localnet-init: localnet-clean localnet-build

localnet-build:
	@DOCKER_BUILDKIT=1 COMPOSE_DOCKER_CLI_BUILD=1 docker-compose -f tests/localosmosis/docker-compose.yml build

localnet-start:
	@STATE="" docker-compose -f tests/localosmosis/docker-compose.yml up

localnet-start-with-state:
	@STATE=-s docker-compose -f tests/localosmosis/docker-compose.yml up

localnet-startd:
	@STATE="" docker-compose -f tests/localosmosis/docker-compose.yml up -d

localnet-startd-with-state:
	@STATE=-s docker-compose -f tests/localosmosis/docker-compose.yml up -d

localnet-stop:
	@STATE="" docker-compose -f tests/localosmosis/docker-compose.yml down

localnet-clean:
	@rm -rfI $(HOME)/.osmosisd-local/

localnet-state-export-init: localnet-state-export-clean localnet-state-export-build 

localnet-state-export-build:
	@DOCKER_BUILDKIT=1 COMPOSE_DOCKER_CLI_BUILD=1 docker-compose -f tests/localosmosis/state_export/docker-compose.yml build

localnet-state-export-start:
	@docker-compose -f tests/localosmosis/state_export/docker-compose.yml up

localnet-state-export-startd:
	@docker-compose -f tests/localosmosis/state_export/docker-compose.yml up -d

localnet-state-export-stop:
	@docker-compose -f tests/localosmosis/docker-compose.yml down

localnet-state-export-clean: localnet-clean

.PHONY: all build-linux install format lint \
	go-mod-cache draw-deps clean build build-contract-tests-hooks \
	test test-all test-build test-cover test-unit test-race benchmark
