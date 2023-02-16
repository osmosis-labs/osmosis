#!/bin/bash

set -ex

# initialize Hermes relayer configuration
mkdir -p /root/.hermes/
touch /root/.hermes/config.toml
touch /root/.hermes/mnemonic.txt 
# setup Hermes relayer configuration
tee /root/.hermes/config.toml <<EOF
[global]
log_level = 'info'
[mode]
[mode.clients]
enabled = true
refresh = true
misbehaviour = true
[mode.connections]
enabled = false
[mode.channels]
enabled = false
[mode.packets]
enabled = true
clear_interval = 100
clear_on_start = true
tx_confirmation = true
[rest]
enabled = true
host = '0.0.0.0'
port = 3031
[telemetry]
enabled = true
host = '127.0.0.1'
port = 3001
[[chains]]
id = '$OSMO_A_E2E_CHAIN_ID'
rpc_addr = 'http://$OSMO_A_E2E_VAL_HOST:26657'
grpc_addr = 'http://$OSMO_A_E2E_VAL_HOST:9090'
websocket_addr = 'ws://$OSMO_A_E2E_VAL_HOST:26657/websocket'
rpc_timeout = '10s'
account_prefix = 'osmo'
key_name = 'val01-osmosis-a'
store_prefix = 'ibc'
max_gas = 6000000
gas_price = { price = 0.000, denom = 'uosmo' }
gas_adjustment = 1.0
clock_drift = '1m' # to accomdate docker containers
trusting_period = '239seconds'
trust_threshold = { numerator = '1', denominator = '3' }
[[chains]]
id = '$OSMO_B_E2E_CHAIN_ID'
rpc_addr = 'http://$OSMO_B_E2E_VAL_HOST:26657'
grpc_addr = 'http://$OSMO_B_E2E_VAL_HOST:9090'
websocket_addr = 'ws://$OSMO_B_E2E_VAL_HOST:26657/websocket'
rpc_timeout = '10s'
account_prefix = 'osmo'
key_name = 'val01-osmosis-b'
store_prefix = 'ibc'
max_gas = 6000000
gas_price = { price = 0.000, denom = 'uosmo' }
gas_adjustment = 1.0
clock_drift = '1m' # to accomdate docker containers
trusting_period = '239seconds'
trust_threshold = { numerator = '1', denominator = '3' }
EOF
tee /root/.hermes/mnemonic.txt<<EOF
${OSMO_B_E2E_VAL_MNEMONIC}
EOF
# import keys
hermes keys add --chain ${OSMO_B_E2E_CHAIN_ID} --key-name "val01-osmosis-b" --mnemonic-file /root/.hermes/mnemonic.txt
hermes keys add --chain ${OSMO_A_E2E_CHAIN_ID} --key-name "val01-osmosis-a" --mnemonic-file /root/.hermes/mnemonic.txt

# start Hermes relayer
hermes start
