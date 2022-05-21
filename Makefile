#!/usr/bin/make -f

PACKAGES_SIMTEST=$(shell go list ./... | grep '/simulation')
VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
LEDGER_ENABLED ?= true
SDK_PACK := $(shell go list -m github.com/cosmos/cosmos-sdk | sed  's/ /\@/g')
DOCKER := $(shell which docker)
BUILDDIR ?= $(CURDIR)/build

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

# The below include contains the tools target.
include contrib/devtools/Makefile

###############################################################################
###                                  Build                                  ###
###############################################################################

all: install lint test

BUILD_TARGETS := build install

build: BUILD_ARGS=-o $(BUILDDIR)/

$(BUILD_TARGETS): go.sum $(BUILDDIR)/
	go $@ -mod=readonly $(BUILD_FLAGS) $(BUILD_ARGS) ./...

$(BUILDDIR)/:
	mkdir -p $(BUILDDIR)/

build-reproducible: go.sum
	$(DOCKER) rm latest-build || true
	$(DOCKER) run --volume=$(CURDIR):/sources:ro \
	--env TARGET_PLATFORMS='linux/amd64' \
	--env APP=osmosisd \
	--env VERSION=$(VERSION) \
	--env COMMIT=$(COMMIT) \
	--env LEDGER_ENABLED=$(LEDGER_ENABLED) \
	--name latest-build osmolabs/rbuilder:latest
	$(DOCKER) cp -a latest-build:/home/builder/artifacts/ $(CURDIR)/

build-linux: go.sum
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

build-contract-tests-hooks:
	mkdir -p $(BUILDDIR)
	go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/ ./cmd/contract_tests

build-e2e-chain-init:
	mkdir -p $(BUILDDIR)
	go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/ ./tests/e2e/chain_init

go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify

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

protoVer=v0.2
protoImageName=tendermintdev/sdk-proto-gen:$(protoVer)
containerProtoGen=osmosis-proto-gen-$(protoVer)
containerProtoFmt=osmosis-proto-fmt-$(protoVer)

proto-gen:
	@echo "Generating Protobuf files"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoGen}$$"; then docker start -a $(containerProtoGen); else docker run --name $(containerProtoGen) -v $(CURDIR):/workspace --workdir /workspace $(protoImageName) \
		sh ./scripts/protocgen.sh; fi

proto-format:
	@echo "Formatting Protobuf files"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoFmt}$$"; then docker start -a $(containerProtoFmt); else docker run --name $(containerProtoFmt) -v $(CURDIR):/workspace --workdir /workspace tendermintdev/docker-build-proto \
		find ./ -not -path "./third_party/*" -name "*.proto" -exec clang-format -i {} \; ; fi

###############################################################################
###                                 Devdoc                                  ###
###############################################################################

build-docs:
	@cd docs && \
	while read p; do \
		(git checkout $${p} && npm install && VUEPRESS_BASE="/$${p}/" npm run build) ; \
		mkdir -p ~/output/$${p} ; \
		cp -r .vuepress/dist/* ~/output/$${p}/ ; \
		cp ~/output/$${p}/index.html ~/output ; \
	done < versions ;

sync-docs:
	cd ~/output && \
	echo "role_arn = ${DEPLOYMENT_ROLE_ARN}" >> /root/.aws/config ; \
	echo "CI job = ${CIRCLE_BUILD_URL}" >> version.html ; \
	aws s3 sync . s3://${WEBSITE_BUCKET} --profile terraform --delete ; \
	aws cloudfront create-invalidation --distribution-id ${CF_DISTRIBUTION_ID} --profile terraform --path "/*" ;
.PHONY: sync-docs


###############################################################################
###                           Tests & Simulation                            ###
###############################################################################

PACKAGES_UNIT=$(shell go list ./... | grep -E -v 'simapp|e2e')
PACKAGES_E2E=$(shell go list ./... | grep '/e2e')
PACKAGES_SIM=$(shell go list ./... | grep '/simapp')
TEST_PACKAGES=./...

include sims.mk

test: test-unit test-build

test-all: check test-race test-cover

test-unit:
	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock norace' $(PACKAGES_UNIT)

test-race:
	@VERSION=$(VERSION) go test -mod=readonly -race -tags='ledger test_ledger_mock' $(PACKAGES_UNIT)

test-cover:
	@VERSION=$(VERSION) go test -mod=readonly -timeout 30m -coverprofile=coverage.txt -tags='norace' -covermode=atomic $(PACKAGES_UNIT)

test-sim:
	@VERSION=$(VERSION) go test -mod=readonly $(PACKAGES_SIM)

test-e2e:
	@VERSION=$(VERSION) go test -mod=readonly -timeout=25m -v $(PACKAGES_E2E)

test-e2e-skip-upgrade:
	@VERSION=$(VERSION) OSMOSIS_E2E_SKIP_UPGRADE=True go test -mod=readonly -timeout=25m -v $(PACKAGES_E2E)

lo-keys:
	. tests/localosmosis/keys.sh

lo-build:
	@docker build -t local:osmosis --build-arg BASE_IMG_TAG=debug -f tests/localosmosis/Dockerfile .

lo-start:
	@docker-compose -f tests/localosmosis/docker-compose.yml up

lo-remove:
	@docker-compose -f tests/localosmosis/docker-compose.yml down

benchmark:
	@go test -mod=readonly -bench=. $(PACKAGES_UNIT)

docker-build-debug:
	@docker build -t osmosis:debug --build-arg BASE_IMG_TAG=debug -f Dockerfile .

docker-build-e2e-chain-init:
	@docker build -t osmosis-e2e-chain-init:debug -f tests/e2e/chain_init/chain-init.Dockerfile .

###############################################################################
###                                Linting                                  ###
###############################################################################

lint:
	@echo "--> Running linter"
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run --timeout=10m

format:
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run ./... --fix
	@go run mvdan.cc/gofumpt -l -w .

###############################################################################
###                                Localnet                                 ###
###############################################################################


.PHONY: all build-linux install format lint \
	go-mod-cache draw-deps clean build build-contract-tests-hooks \
	test test-all test-build test-cover test-unit test-race benchmark
