#!/bin/bash
#
# This script can be used to manually test the crosschain swaps. It is meant as a guide and not to be run directly
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
echo "increase bread alpha rigid glide amused approve oblige print asset idea enact lawn proof unfold jeans rabbit audit return chuckle valve rather cactus great" | osmosisd --keyring-backend test keys add faucet --recover || echo "key exists"

export VALIDATOR=$(osmosisd keys show validator -a)
export CHANNEL_ID="channel-0"
args="--keyring-backend test --gas auto --gas-prices 0.1uosmo --gas-adjustment 1.3 --broadcast-mode block --yes"
TX_FLAGS=($args)

# send money to the validator on both chains
chainA tx bank send faucet "$VALIDATOR" 1000000000uosmo "${TX_FLAGS[@]}"
chainB tx bank send faucet "$VALIDATOR" 1000000000uosmo "${TX_FLAGS[@]}"

# Give the validator some tokens
# send tokens to chainB
chainA tx ibc-transfer transfer transfer $CHANNEL_ID "$VALIDATOR" 600000000uosmo --from validator "${TX_FLAGS[@]}"
sleep 22 # wait for the roundtrip
export DENOM=$(chainB q bank balances "$VALIDATOR" -o json | jq -r '.balances[] | select(.denom | contains("ibc")) | .denom')
echo chainB q bank balances "$VALIDATOR"
sleep 1

# create the sample_pool.json file
cat >sample_pool.json <<EOF
{
        "weights": "1${DENOM},1uosmo",
        "initial-deposit": "1000000${DENOM},1000000uosmo",
        "swap-fee": "0.01",
        "exit-fee": "0.01",
        "future-governor": "168h"
}
EOF

# create the pool
chainB tx gamm create-pool --pool-file sample_pool.json --from validator --yes -b block
sleep 6
# get the pool id
export POOL_ID=$(chainB query gamm pools -o json | jq -r '.pools[-1].id')

# Store the swaprouter contract
chainB tx wasm store ./bytecode/swaprouter.wasm --from validator "${TX_FLAGS[@]}"
# get the code id
sleep 6
SWAPROUTER_CODE_ID=$(chainB query wasm list-code -o json | jq -r '.code_infos[-1].code_id')
# Instantiate the swaprouter contract
MSG=$(jenv -c '{"owner": $VALIDATOR}')
chainB tx wasm instantiate "$SWAPROUTER_CODE_ID" "$MSG" --from validator --admin $VALIDATOR --label swaprouter --yes -b block
export SWAPROUTER_ADDRESS=$(chainB query wasm list-contract-by-code "$SWAPROUTER_CODE_ID" -o json | jq -r '.contracts | [last][0]')

# Configure the swaprouter
MSG=$(jenv -c '{"set_route":{"input_denom":$DENOM,"output_denom":"uosmo","pool_route":[{"pool_id":$POOL_ID,"token_out_denom":"uosmo"}]}}')
chainB tx wasm execute "$SWAPROUTER_ADDRESS" "$MSG" --from validator -y

# Store the crosschainswap contract
chainB tx wasm store ./bytecode/crosschain_swaps.wasm --from validator --gas auto --gas-prices 0.1uosmo --gas-adjustment 1.3 -y -b block
CROSSCHAIN_SWAPS_CODE_ID=$(chainB query wasm list-code -o json | jq -r '.code_infos[-1].code_id')
# Instantiate the crosschainswap contract
MSG=$(jenv -c '{"swap_contract": $SWAPROUTER_ADDRESS, "channels": [["osmo", $CHANNEL_ID]]}')
chainB tx wasm instantiate "$CROSSCHAIN_SWAPS_CODE_ID" "$MSG" --from validator --admin $VALIDATOR --label=crosschain_swaps --yes -b block
export CROSSCHAIN_SWAPS_ADDRESS=$(chainB query wasm list-contract-by-code "$CROSSCHAIN_SWAPS_CODE_ID" -o json | jq -r '.contracts | [last][0]')

balances=$(chainA query bank balances "$VALIDATOR" -o json | jq '.balances')

# Send a crosschain swap
MEMO=$(jenv -c '{"wasm": {"contract": $CROSSCHAIN_SWAPS_ADDRESS, "msg": {"osmosis_swap":{"swap_amount":"100","output_denom":"uosmo","slippage":{"twap": {"slippage_percentage":"20", "window_seconds": 10}},"receiver":$VALIDATOR, "on_failed_delivery": "do_nothing"}}}}')
chainA tx ibc-transfer transfer transfer $CHANNEL_ID $CROSSCHAIN_SWAPS_ADDRESS 100uosmo \
    --from validator -y "${TX_FLAGS[@]}" \
    --memo "$MEMO"

sleep 20 # wait for the roundtrip

new_balances=$(chainA query bank balances "$VALIDATOR" -o json | jq '.balances')
echo "old balances: $balances, new balances: $new_balances"
