###############################################################################
###                                 Tests                                   ###
###############################################################################

PACKAGES_UNIT=$(shell go list ./... ./osmomath/... ./osmoutils/... ./x/ibc-hooks/... ./x/epochs | grep -E -v 'tests/simulator|e2e')
PACKAGES_E2E := $(shell go list ./... | grep '/e2e' | awk -F'/e2e' '{print $$1 "/e2e"}' | uniq)
PACKAGES_SIM=$(shell go list ./... | grep '/tests/simulator')
TEST_PACKAGES=./...

test-help:
	@echo "test subcommands"
	@echo ""
	@echo "Usage:"
	@echo "  make test-[command]"
	@echo ""
	@echo "Available Commands:"
	@echo "  all                Run all tests"
	@echo "  unit               Run unit tests"
	@echo "  race               Run race tests"
	@echo "  cover              Run coverage tests"
	@echo "  sim-suite          Run sim suite tests"
	@echo "  sim-app            Run sim app tests"
	@echo "  sim-determinism    Run sim determinism tests"
	@echo "  sim-bench          Run sim benchmark tests"
	@echo "  e2e                Run e2e tests"
	@echo "  e2e-ci             Run e2e CI tests"
	@echo "  e2e-ci-scheduled   Run e2e CI scheduled tests"
	@echo "  e2e-debug          Run e2e debug tests"
	@echo "  e2e-short          Run e2e short tests"
	@echo "  mutation           Run mutation tests"
	@echo "  benchmark          Run benchmark tests"

test: test-help

test-all: test-race test-covertest-unit test-build

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

# test-e2e-ci runs a majority of e2e tests, only skipping the ones that are marked as scheduled tests
# does not do any validation about the state of the Docker environment
# As a result, avoid using this locally.
test-e2e-ci:
	@VERSION=$(VERSION) OSMOSIS_E2E=True OSMOSIS_E2E_DEBUG_LOG=False OSMOSIS_E2E_UPGRADE_VERSION=$(E2E_UPGRADE_VERSION) go test -mod=readonly -timeout=25m -v $(PACKAGES_E2E) -p 4

# test-e2e-ci-scheduled runs every e2e test available, and is only run on a scheduled basis
test-e2e-ci-scheduled:
	@VERSION=$(VERSION) OSMOSIS_E2E_SCHEDULED=True OSMOSIS_E2E=True OSMOSIS_E2E_DEBUG_LOG=False OSMOSIS_E2E_UPGRADE_VERSION=$(E2E_UPGRADE_VERSION) go test -mod=readonly -timeout=25m -v $(PACKAGES_E2E) -p 4

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

test-benchmark:
	@go test -mod=readonly -bench=. $(PACKAGES_UNIT)
