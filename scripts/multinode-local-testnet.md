# Multi Node Local Testnet Script

This script creates a multi node local testnet with three validator nodes on a single machine. Note: The default weights of these validators is 5:5:4 respectively. That means in order to keep the chain running, at a minimum Validator1 and Validator2 must be running in order to keep greater than 66% power online.

## Instructions

Clone the osmosis repo

Checkout the branch you are looking to test

Make install / reload profile

Give the script permission with `chmod +x multinode-local-testnet.sh`

Run with `./multinode-local-testnet.sh` (allow ~45 seconds to run, required sleep commands due to multiple transactions)

## Logs

Validator1: `tmux a -t validator1`

Validator2: `tmux a -t validator2`

Validator3: `tmux a -t validator3`

## Directories

Validator1: `$HOME/.osmosisd/validator1`

Validator2: `$HOME/.osmosisd/validator2`

Validator3: `$HOME/.osmosisd/validator3`

## Ports

Validator1: `1317, 9090, 9091, 26658, 26657, 26656, 6060`

Validator2: `1316, 9088, 9089, 26655, 26654, 26653, 6061`

Validator3: `1315, 9086, 9087, 26652, 26651, 26650, 6062`

Ensure to include the `--home` flag or `--node` flag when using a particular node.

## Examples

Validator2: `osmosisd status --node "tcp://localhost:26654"`

Validator3: `osmosisd status --node "tcp://localhost:26651"`

or

Validator1: `osmosisd keys list --keyring-backend test --home $HOME/.osmosisd/validator1`

Validator2: `osmosisd keys list --keyring-backend test --home $HOME/.osmosisd/validator2`
