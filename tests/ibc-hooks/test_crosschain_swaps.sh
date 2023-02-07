#!/bin/bash
set -o errexit -o nounset -o pipefail -o xtrace
shopt -s expand_aliases

alias osmosisd="~/devel/osmosis/build/osmosisd"
alias junod="~/devel/juno/bin/junod"

# setup the keys
# echo "bottom loan skill merry east cradle onion journey palm apology verb edit desert impose absurd oil bubble sweet glove shallow size build burst effort" | ~/devel/osmosis/build/osmosisd --keyring-backend test keys add validator --recover || echo "key exists"
# echo "increase bread alpha rigid glide amused approve oblige print asset idea enact lawn proof unfold jeans rabbit audit return chuckle valve rather cactus great" | ~/devel/osmosis/build/osmosisd --keyring-backend test keys add faucet --recover || echo "key exists"

# echo "bottom loan skill merry east cradle onion journey palm apology verb edit desert impose absurd oil bubble sweet glove shallow size build burst effort" | ~/devel/juno/bin/junod --keyring-backend test keys add validator --recover || echo "key exists"
# echo "increase bread alpha rigid glide amused approve oblige print asset idea enact lawn proof unfold jeans rabbit audit return chuckle valve rather cactus great" | ~/devel/juno/bin/junod --keyring-backend test keys add faucet --recover || echo "key exists"

export JUNO_CHANNEL="channel-0"
export OSMOSIS_CHANNEL="channel-0"
export VALIDATOR_OSMO=$(osmosisd keys show validator -a)
export VALIDATOR_JUNO=$(junod keys show validator -a)

# Give the validator some tokens
# send tokens to osmosisd
osmosisd tx bank send faucet "$VALIDATOR_OSMO" 10000000000uosmo -y
junod tx ibc-transfer transfer transfer $JUNO_CHANNEL "$VALIDATOR_OSMO" 60000000ujuno --from validator -y --gas auto --gas-prices 0.1ujuno --gas-adjustment 1.3
sleep 16 # wait for the roundtrip
export DENOM=$(osmosisd q bank balances "$VALIDATOR_OSMO" -o json | jq -r '.balances[] | select(.denom | contains("ibc")) | .denom')
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
osmosisd tx gamm create-pool --pool-file sample_pool.json --from validator --yes -b block
sleep 6
# get the pool id
export POOL_ID=$(osmosisd query gamm pools -o json | jq -r '.pools[-1].id')

# Store the swaprouter contract
osmosisd tx wasm store ./bytecode/swaprouter.wasm --from validator --gas auto --gas-prices 0.1uosmo --gas-adjustment 1.3 -y -b block
# get the code id
sleep 6
SWAPROUTER_CODE_ID=$(osmosisd query wasm list-code -o json | jq -r '.code_infos[-1].code_id')
# Instantiate the swaprouter contract
MSG=$(jenv -c '{"owner": $VALIDATOR_OSMO}')
osmosisd tx wasm instantiate "$SWAPROUTER_CODE_ID" "$MSG" --from validator --admin $VALIDATOR_OSMO --label swaprouter --yes -b block
export SWAPROUTER_ADDRESS=$(osmosisd query wasm list-contract-by-code "$SWAPROUTER_CODE_ID" -o json | jq -r '.contracts | [last][0]')

# Configure the swaprouter
MSG=$(jenv -c '{"set_route":{"input_denom":$DENOM,"output_denom":"uosmo","pool_route":[{"pool_id":$POOL_ID,"token_out_denom":"uosmo"}]}}')
osmosisd tx wasm execute "$SWAPROUTER_ADDRESS" "$MSG" --from validator -y

# Store the crosschainswap contract
osmosisd tx wasm store ./bytecode/crosschain_swaps.wasm --from validator --gas auto --gas-prices 0.1uosmo --gas-adjustment 1.3 -y -b block
CROSSCHAIN_SWAPS_CODE_ID=$(osmosisd query wasm list-code -o json | jq -r '.code_infos[-1].code_id')
# Instantiate the crosschainswap contract
MSG=$(jenv -c '{"swap_contract": $SWAPROUTER_ADDRESS, "channels": [["juno", $OSMOSIS_CHANNEL]]}')
osmosisd tx wasm instantiate "$CROSSCHAIN_SWAPS_CODE_ID" "$MSG" --from validator --admin $VALIDATOR_OSMO --label=crosschain_swaps --yes -b block
export CROSSCHAIN_SWAPS_ADDRESS=$(osmosisd query wasm list-contract-by-code "$CROSSCHAIN_SWAPS_CODE_ID" -o json | jq -r '.contracts | [last][0]')

echo "Crosschain Swaps contract deployed at $CROSSCHAIN_SWAPS_ADDRESS"

# Send a crosschain swap
# MEMO=$(jenv -c '{"wasm": {"contract": $CROSSCHAIN_SWAPS_ADDRESS, "msg": {"osmosis_swap":{"swap_amount":"100","output_denom":"uosmo","slippage":{"twap": {"slippage_percentage":"20", "window_seconds": 10}},"receiver":$VALIDATOR_JUNO, "on_failed_delivery": "do_nothing"}}}}')
# junod tx ibc-transfer transfer transfer $JUNO_CHANNEL $CROSSCHAIN_SWAPS_ADDRESS 100ujuno \
#     --from validator -y --gas auto --gas-prices 0.1ujuno --gas-adjustment 1.3 \
#     --memo "$MEMO"
