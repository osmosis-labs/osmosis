#!/bin/bash

url=$1

# This script compares single hop quotes by running them against SQS and chain directly.

chain_amount_out=$(osmosisd q poolmanager estimate-swap-exact-amount-in 1136 1000000ibc/C140AFD542AE77BD7DCC83F13FDD8C5E5BB8C4929785E6EC2F4C636F98F17901 --swap-route-pool-ids 1136 --swap-route-denoms ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2)

sqs_res=$(curl "$url/router/custom-quote?tokenIn=1000000ibc/C140AFD542AE77BD7DCC83F13FDD8C5E5BB8C4929785E6EC2F4C636F98F17901&tokenOutDenom=ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2&poolIDs=1136")

sqs_amount_out=$(echo $sqs_res | jq .amount_out)

echo "chain_amount_out: $chain_amount_out"
echo "sqs_amount_out: $sqs_amount_out"
