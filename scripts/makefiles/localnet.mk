###############################################################################
###                                Localnet                                 ###
###############################################################################
#
# Please refer to https://github.com/osmosis-labs/osmosis/blob/main/tests/localosmosis/README.md for detailed 
# usage of localnet.

localnet-help:
	@echo "localnet subcommands"
	@echo ""
	@echo "Usage:"
	@echo "  make localnet-[command]"
	@echo ""
	@echo "Available Commands:"
	@echo "  keys                            Add keys for localnet"
	@echo "  init                            Initialize localnet"
	@echo "  build                           Build localnet"
	@echo "  start                           Start localnet"
	@echo "  start-with-state                Start localnet with state"
	@echo "  startd                          Start localnet in detached mode"
	@echo "  startd-with-state               Start localnet in detached mode with state"
	@echo "  stop                            Stop localnet"
	@echo "  clean                           Clean localnet"
	@echo "  state-export-init               Initialize localnet state export"
	@echo "  state-export-build              Build localnet state export"
	@echo "  state-export-start              Start localnet state export"
	@echo "  state-export-startd             Start localnet state export in detached mode"
	@echo "  state-export-stop               Stop localnet state export"
	@echo "  state-export-clean              Clean localnet state export"
	@echo "  cl-create-positions             Create concentrated liquidity positions"
	@echo "  cl-small-swap                   Perform small randomized swaps"
	@echo "  cl-large-swap                   Perform large swaps"
	@echo "  cl-external-incentive           Create external incentive"
	@echo "  cl-create-pool                  Create concentrated liquidity pool"
	@echo "  cl-claim-spread-rewards         Claim spread rewards"
	@echo "  cl-claim-incentives             Claim incentives"
	@echo "  cl-add-to-positions             Add to positions"
	@echo "  cl-withdraw-positions           Withdraw positions"
	@echo "  cl-positions-small-swaps        Create positions and perform small swaps"
	@echo "  cl-positions-large-swaps        Create positions and perform large swaps"
	@echo "  cl-refresh-subgraph-positions   Refresh subgraph positions"
	@echo "  cl-refresh-subgraph-genesis     Refresh subgraph genesis"
	@echo "  cl-create-bigbang-config        Create Big Bang configuration"
localnet: localnet-help

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
localnet-cl-refresh-subgraph-positions:
	go run ./tests/cl-genesis-positions --operation 0

# This script converts the positions data created by the
# cl-refresh-subgraph-positions makefile step into an Osmosis
# genesis. It writes the file under tests/cl-genesis-positions/genesis.json
localnet-cl-refresh-subgraph-genesis:
	go run ./tests/cl-genesis-positions --operation 1

# This script converts the positions data created by the
# cl-refresh-subgraph-positions makefile step into a Big Bang
# configuration file for spinning up testnets.
# It writes the file under tests/cl-genesis-positions/bigbang_positions.json
localnet-cl-create-bigbang-config:
	go run ./tests/cl-genesis-positions --operation 1 --big-bang
