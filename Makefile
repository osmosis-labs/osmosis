#!/usr/bin/make -f

VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')

LEDGER_ENABLED ?= true
SDK_PACK := $(shell go list -m github.com/cosmos/cosmos-sdk | sed  's/ /\@/g')
BUILDDIR ?= $(CURDIR)/build
DOCKER := $(shell which docker)
E2E_UPGRADE_VERSION := "v19"
#SHELL := /bin/bash

GO_VERSION := $(shell cat go.mod | grep -E 'go [0-9].[0-9]+' | cut -d ' ' -f 2)
GO_MODULE := $(shell cat go.mod | grep "module " | cut -d ' ' -f 2)
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
###                                  Build                                  ###
###############################################################################

check_version:
ifneq ($(GO_MINOR_VERSION),20)
	@echo "ERROR: Go version 1.20 is required for this version of Osmosis."
	exit 1
endif

all: install lint test

build: check_version go.sum
	mkdir -p $(BUILDDIR)/
	GOWORK=off go build -mod=readonly  $(BUILD_FLAGS) -o $(BUILDDIR)/ $(GO_MODULE)/cmd/osmosisd

build-all: check_version go.sum
	mkdir -p $(BUILDDIR)/
	GOWORK=off go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/ ./...

install: check_version go.sum
	GOWORK=off go install -mod=readonly $(BUILD_FLAGS) $(GO_MODULE)/cmd/osmosisd

install-with-autocomplete: check_version go.sum
	GOWORK=off go install -mod=readonly $(BUILD_FLAGS) $(GO_MODULE)/cmd/osmosisd
	@PARENT_SHELL=$$(ps -o ppid= -p $$PPID | xargs ps -o comm= -p); \
	if echo "$$PARENT_SHELL" | grep -q "zsh"; then \
		if ! grep -q ". <(osmosisd enable-cli-autocomplete zsh)" ~/.zshrc; then \
			echo ". <(osmosisd enable-cli-autocomplete zsh)" >> ~/.zshrc; \
			echo; \
			echo "Autocomplete enabled. Run 'source ~/.zshrc' to complete installation."; \
		else \
			echo; \
			echo "Autocomplete already enabled in ~/.zshrc"; \
		fi \
	elif echo "$$PARENT_SHELL" | grep -q "bash" && [ "$$(uname)" = "Darwin" ]; then \
		if ! grep -q -e "\. <(osmosisd enable-cli-autocomplete bash)" -e '\[\[ -r "/opt/homebrew/etc/profile.d/bash_completion.sh" \]\] && \. "/opt/homebrew/etc/profile.d/bash_completion.sh"' ~/.bash_profile; then \
			brew install bash-completion; \
			echo '[ -r "/opt/homebrew/etc/profile.d/bash_completion.sh" ] && . "/opt/homebrew/etc/profile.d/bash_completion.sh"' >> ~/.bash_profile; \
			echo ". <(osmosisd enable-cli-autocomplete bash)" >> ~/.bash_profile; \
			echo; \
			echo; \
			echo "Autocomplete enabled. Run 'source ~/.bash_profile' to complete installation."; \
		else \
			echo "Autocomplete already enabled in ~/.bash_profile"; \
		fi \
	elif echo "$$PARENT_SHELL" | grep -q "bash" && [ "$$(uname)" = "Linux" ]; then \
		if ! grep -q ". <(osmosisd enable-cli-autocomplete bash)" ~/.bash_profile; then \
			sudo apt-get install -y bash-completion; \
			echo '[ -r "/etc/bash_completion" ] && . "/etc/bash_completion"' >> ~/.bash_profile; \
			echo ". <(osmosisd enable-cli-autocomplete bash)" >> ~/.bash_profile; \
			echo; \
			echo "Autocomplete enabled. Run 'source ~/.bash_profile' to complete installation."; \
		else \
			echo; \
			echo "Autocomplete already enabled in ~/.bash_profile"; \
		fi \
	else \
		echo "Shell or OS not recognized. Skipping autocomplete setup."; \
	fi


# Cross-building for arm64 from amd64 (or viceversa) takes
# a lot of time due to QEMU virtualization but it's the only way (afaik)
# to get a statically linked binary with CosmWasm

build-reproducible: build-reproducible-amd64 build-reproducible-arm64

build-reproducible-amd64: go.sum
	mkdir -p $(BUILDDIR)
	$(DOCKER) buildx create --name osmobuilder || true
	$(DOCKER) buildx use osmobuilder
	$(DOCKER) buildx build \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg GIT_VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(COMMIT) \
		--build-arg RUNNER_IMAGE=alpine:3.17 \
		--platform linux/amd64 \
		-t osmosis:local-amd64 \
		--load \
		-f Dockerfile .
	$(DOCKER) rm -f osmobinary || true
	$(DOCKER) create -ti --name osmobinary osmosis:local-amd64
	$(DOCKER) cp osmobinary:/bin/osmosisd $(BUILDDIR)/osmosisd-linux-amd64
	$(DOCKER) rm -f osmobinary

build-reproducible-arm64: go.sum
	mkdir -p $(BUILDDIR)
	$(DOCKER) buildx create --name osmobuilder || true
	$(DOCKER) buildx use osmobuilder
	$(DOCKER) buildx build \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg GIT_VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(COMMIT) \
		--build-arg RUNNER_IMAGE=alpine:3.17 \
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
###                           Dependency Updates                            ###
###############################################################################

VERSION := 
MODFILES := ./go.mod ./osmoutils/go.mod ./osmomath/go.mod ./x/epochs/go.mod ./x/ibc-hooks/go.mod ./tests/cl-genesis-positions/go.mod ./tests/cl-go-client/go.mod
# run with VERSION argument specified
# e.g) make update-sdk-version VERSION=v0.45.1-0.20230523200430-193959b898ec
# This will change sdk dependencyu version for go.mod in root directory + all sub-modules in this repo.
update-sdk-version:
	@if [ -z "$(VERSION)" ]; then \
		echo "VERSION not set"; \
		exit 1; \
	fi
	@echo "Updating version to $(VERSION)"
	@for modfile in $(MODFILES); do \
		if [ -e "$$modfile" ]; then \
			sed -i '' 's|github.com/osmosis-labs/cosmos-sdk v[0-9a-z.\-]*|github.com/osmosis-labs/cosmos-sdk $(VERSION)|g' $$modfile; \
			cd `dirname $$modfile`; \
			go mod tidy; \
			cd - > /dev/null; \
		else \
			echo "File $$modfile does not exist"; \
		fi; \
	done

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

protoVer=v0.9
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

PACKAGES_UNIT=$(shell go list ./... ./osmomath/... ./osmoutils/... ./x/ibc-hooks/... ./x/epochs | grep -E -v 'tests/simulator|e2e')
PACKAGES_E2E := $(shell go list ./... | grep '/e2e' | awk -F'/e2e' '{print $$1 "/e2e"}' | uniq)
PACKAGES_SIM=$(shell go list ./... | grep '/tests/simulator')
TEST_PACKAGES=./...

test: test-unit test-build

test-all: test-race test-cover

test-unit:
	@VERSION=$(VERSION) SKIP_WASM_WSL_TESTS=$(SKIP_WASM_WSL_TESTS) go test -mod=readonly -tags='ledger test_ledger_mock norace' $(PACKAGES_UNIT)

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
	@VERSION=$(VERSION) OSMOSIS_E2E=True OSMOSIS_E2E_DEBUG_LOG=False OSMOSIS_E2E_UPGRADE_VERSION=$(E2E_UPGRADE_VERSION) go test -mod=readonly -timeout=25m -v $(PACKAGES_E2E) -p 4

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
	@DOCKER_BUILDKIT=1 docker build -t osmolabs/osmosis-e2e-init-chain:debug --build-arg E2E_SCRIPT_NAME=chain --platform=linux/x86_64 -f tests/e2e/initialization/init.Dockerfile .

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
RUNNER_BASE_IMAGE_ALPINE := alpine:3.17
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

# create 100 concentrated-liquidity positions in localosmosis at pool id 1
localnet-cl-create-positions:
	go run tests/cl-go-client/main.go --operation 0

# does 100 small randomized swaps in localosmosis at pool id 1
localnet-cl-small-swap:
	go run tests/cl-go-client/main.go --operation 1

# does 100 large swaps where the output of the previous swap is swapped back at the
# next swap. localosmosis at pool id 1
localnet-cl-large-swap:
	go run tests/cl-go-client/main.go --operation 2

# creates a gauge and waits for one epoch so that the gauge
# is converted into an incentive record for pool id 1.
localnet-cl-external-incentive:
	go run tests/cl-go-client/main.go --operation 3

# attempts to create a CL pool at id 1.
# if pool already exists, this is a no-op.
# if pool with different id is desired, tweak expectedPoolId
# in the script.
localnet-cl-create-pool:
	go run tests/cl-go-client/main.go --operation 4

# claims spread rewards for a random account for a random
# subset of positions.
localnet-cl-claim-spread-rewards:
	go run tests/cl-go-client/main.go --operation 5

# claims incentives for a random account for a random
# subset of positions.
localnet-cl-claim-incentives:
	go run tests/cl-go-client/main.go --operation 6

localnet-cl-add-to-positions:
	go run tests/cl-go-client/main.go --operation 7

localnet-cl-withdraw-positions:
	go run tests/cl-go-client/main.go --operation 8


# does both of localnet-cl-create-positions and localnet-cl-small-swap
localnet-cl-positions-small-swaps: localnet-cl-create-positions localnet-cl-small-swap

# does both of localnet-cl-create-positions and localnet-cl-large-swap
localnet-cl-positions-large-swaps: localnet-cl-create-positions localnet-cl-large-swap

# This script retrieves Uniswap v3 Ethereum position data
# from subgraph. It uses WETH / USDC pool. This is helpful
# for setting up somewhat realistic positions for testing
# in localosmosis. It writes the file under
# tests/cl-genesis-positions/subgraph_positions.json
cl-refresh-subgraph-positions:
	go run ./tests/cl-genesis-positions --operation 0

# This script converts the positions data created by the
# cl-refresh-subgraph-positions makefile step into an Osmosis
# genesis. It writes the file under tests/cl-genesis-positions/genesis.json
cl-refresh-subgraph-genesis:
	go run ./tests/cl-genesis-positions --operation 1

# This script converts the positions data created by the
# cl-refresh-subgraph-positions makefile step into a Big Bang
# configuration file for spinning up testnets.
# It writes the file under tests/cl-genesis-positions/bigbang_positions.json
cl-create-bigbang-config:
	go run ./tests/cl-genesis-positions --operation 1 --big-bang

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

GORELEASER_IMAGE := ghcr.io/goreleaser/goreleaser-cross:v$(GO_VERSION)
COSMWASM_VERSION := $(shell go list -m github.com/CosmWasm/wasmvm | sed 's/.* //')

ifdef GITHUB_TOKEN
release:
	docker run \
		--rm \
		-e GITHUB_TOKEN=$(GITHUB_TOKEN) \
		-e COSMWASM_VERSION=$(COSMWASM_VERSION) \
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

release-dry-run:
	docker run \
		--rm \
		-e COSMWASM_VERSION=$(COSMWASM_VERSION) \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/osmosisd \
		-w /go/src/osmosisd \
		$(GORELEASER_IMAGE) \
		release \
		--clean \
		--skip-publish

release-snapshot:
	docker run \
		--rm \
		-e COSMWASM_VERSION=$(COSMWASM_VERSION) \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/osmosisd \
		-w /go/src/osmosisd \
		$(GORELEASER_IMAGE) \
		release \
		--clean \
		--snapshot \
		--skip-validate \
		--skip-publish

.PHONY: all build-linux install format lint \
	go-mod-cache draw-deps clean build build-contract-tests-hooks \
	test test-all test-build test-cover test-unit test-race benchmark \
	release release-dry-run release-snapshot
