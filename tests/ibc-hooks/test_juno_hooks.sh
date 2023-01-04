#!/bin/bash
set -o errexit -o nounset -o pipefail -o xtrace
shopt -s expand_aliases

alias osmosisd="~/devel/osmosis/build/osmosisd"
alias junod="~/devel/juno/bin/junod"

# setup the keys
echo "bottom loan skill merry east cradle onion journey palm apology verb edit desert impose absurd oil bubble sweet glove shallow size build burst effort" | ~/devel/osmosis/build/osmosisd --keyring-backend test keys add validator --recover || echo "key exists"
echo "increase bread alpha rigid glide amused approve oblige print asset idea enact lawn proof unfold jeans rabbit audit return chuckle valve rather cactus great" | ~/devel/osmosis/build/osmosisd --keyring-backend test  keys add faucet --recover || echo "key exists"

echo "bottom loan skill merry east cradle onion journey palm apology verb edit desert impose absurd oil bubble sweet glove shallow size build burst effort" | ~/devel/juno/bin/junod --keyring-backend test keys add validator --recover || echo "key exists"
echo "increase bread alpha rigid glide amused approve oblige print asset idea enact lawn proof unfold jeans rabbit audit return chuckle valve rather cactus great" | ~/devel/juno/bin/junod --keyring-backend test keys add faucet --recover || echo "key exists"

# store and instantiate the contract
junod tx wasm store ./bytecode/counter.wasm --from validator  --gas auto --gas-prices 0.1ujuno --gas-adjustment 1.3 -y
CONTRACT_ID=$( junod query wasm list-code -o json | jq -r '.code_infos[-1].code_id')
junod tx wasm instantiate "$CONTRACT_ID" '{"count": 0}' --from validator --no-admin --label=counter --yes

# get the contract address
export CONTRACT_ADDRESS=$(junod query wasm list-contract-by-code 1 -o json | jq -r '.contracts | [last][0]')

# Fund the validator
osmosisd tx bank send faucet osmo12smx2wdlyttvyzvzg54y2vnqwq2qjateuf7thj 1000000000uosmo -y

# send ibc transaction to execite the contract
MEMO=$(jenv -c '{"wasm":{"contract":$CONTRACT_ADDRESS,"msg": {"increment": {}} }}' )
osmosisd tx ibc-transfer transfer transfer channel-0 $CONTRACT_ADDRESS 10uosmo \
         --from validator -y --gas auto --gas-prices 0.1uosmo --gas-adjustment 1.3 \
         --memo "$MEMO"
