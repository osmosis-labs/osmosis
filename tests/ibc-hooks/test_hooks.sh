#!/bin/bash
#
# This script can be used to manually test the ibc hooks. It is meant as a guide and not to be run directly 
# without taking into account the context in which it's being run.
# The script uses `jenv` (https://github.com/nicolaslara/jenv) to easily generate the json strings passed
# to some of the commands. If you don't want to use it you can generate the json manually or modify this script.
#
set -o errexit -o nounset -o pipefail -o xtrace
shopt -s expand_aliases

alias chainA="osmosisd --node http://localhost:26657 --chain-id localosmosis-a"
alias chainB="osmosisd --node http://localhost:36657 --chain-id localosmosis-b"

# setup the keys
echo "bottom loan skill merry east cradle onion journey palm apology verb edit desert impose absurd oil bubble sweet glove shallow size build burst effort" | osmosisd --keyring-backend test keys add validator --recover || echo "key exists"
echo "increase bread alpha rigid glide amused approve oblige print asset idea enact lawn proof unfold jeans rabbit audit return chuckle valve rather cactus great" | osmosisd --keyring-backend test  keys add faucet --recover || echo "key exists"

VALIDATOR=$(osmosisd keys show validator -a)

args="--keyring-backend test --gas auto --gas-prices 0.1uosmo --gas-adjustment 1.3 --broadcast-mode block --yes"
TX_FLAGS=($args)

# send money to the validator on both chains
chainA tx bank send faucet "$VALIDATOR" 1000000000uosmo "${TX_FLAGS[@]}"
chainB tx bank send faucet "$VALIDATOR" 1000000000uosmo "${TX_FLAGS[@]}"

# store and instantiate the contract
chainA tx wasm store ./bytecode/counter.wasm --from validator  "${TX_FLAGS[@]}"
CONTRACT_ID=$(chainA query wasm list-code -o json | jq -r '.code_infos[-1].code_id')
chainA tx wasm instantiate "$CONTRACT_ID" '{"count": 0}' --from validator --no-admin --label=counter "${TX_FLAGS[@]}"

# get the contract address
export CONTRACT_ADDRESS=$(chainA query wasm list-contract-by-code 1 -o json | jq -r '.contracts | [last][0]')

denom=$(chainA query bank balances "$CONTRACT_ADDRESS" -o json | jq -r '.balances[0].denom')
balance=$(chainA query bank balances "$CONTRACT_ADDRESS" -o json | jq -r '.balances[0].amount')

# send ibc transaction to execite the contract
MEMO=$(jenv -c '{"wasm":{"contract":$CONTRACT_ADDRESS,"msg": {"increment": {}} }}' )
chainB tx ibc-transfer transfer transfer channel-0 $CONTRACT_ADDRESS 10uosmo \
       --from validator -y  \
       --memo "$MEMO"

# wait for the ibc round trip
sleep 16

new_balance=$(chainA query bank balances "$CONTRACT_ADDRESS" -o json | jq -r '.balances[0].amount')
export ADDR_IN_CHAIN_A=$(chainA q ibchooks wasm-sender channel-0 "$VALIDATOR")
QUERY=$(jenv -c -r '{"get_total_funds": {"addr": $ADDR_IN_CHAIN_A}}')
funds=$(chainA query wasm contract-state smart "$CONTRACT_ADDRESS" "$QUERY" -o json | jq -c -r '.data.total_funds[]')
QUERY=$(jenv -c -r '{"get_count": {"addr": $ADDR_IN_CHAIN_A}}')
count=$(chainA query wasm contract-state smart "$CONTRACT_ADDRESS" "$QUERY" -o json |  jq -r '.data.count')

echo "funds: $funds, count: $count"
echo "denom: $denom, old balance: $balance, new balance: $new_balance"
