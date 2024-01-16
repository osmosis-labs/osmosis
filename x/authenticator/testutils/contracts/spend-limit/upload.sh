#!/bin/bash

#set -o xtrace

NODE="--node http://localhost:26657"
CHAIN="--chain-id testing"
KEYRING="--keyring-backend test"
FEES="--fees 62500uosmo"
GAS="--gas 25000000"
BROADCAST="-b sync"
OPTS="$KEYRING $CHAIN $NODE $FEES $GAS $BROADCAST"

SENDER="--from validator"
SENDEROTHER="--from user1"

user1=$(osmosisd keys show user1 -a --keyring-backend=test)
pubkey=$(osmosisd keys show user1 -p --keyring-backend test | jq -r .key)

RESP=$(osmosisd tx wasm store "./artifacts/spend_limit.wasm" $OPTS $SENDER -y -o json)
sleep 6

echo $RESP
RESP=$(osmosisd q tx $(echo "$RESP"| jq -r '.txhash') -o json)
CODE_ID=$(echo "$RESP" | jq -r '.logs[0].events[]| select(.type=="store_code").attributes[]| select(.key=="code_id").value')
CODE_HASH=$(echo "$RESP" | jq -r '.logs[0].events[]| select(.type=="store_code").attributes[]| select(.key=="code_checksum").value')

echo "-----------------------"
echo "* Code id: $CODE_ID"
echo "* Code checksum: $CODE_HASH"
echo "-----------------------"

sleep 6

#echo "-----------------------"
#echo "## List code"
#echo "-----------------------"
#osmosisd query wasm list-code --node=http://0.0.0.0:26657 -o json | jq


echo "-----------------------"
echo "## Create new contract instance"
echo "-----------------------"
INIT="{}"
RESP=$(osmosisd tx wasm instantiate --label "spend_limit" --no-admin "$CODE_ID" "$INIT" $SENDER $OPTS -y)
 
sleep 6
CONTRACT=$(osmosisd query wasm list-contract-by-code  "$CODE_ID" -o json | jq -r '.contracts[-1]')

echo "* Contract address: $CONTRACT"


