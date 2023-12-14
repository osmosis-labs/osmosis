#!/bin/bash

url=$1

# This script compares single hop quotes by running them against SQS and chain directly.

chain_amount_out=$(osmosisd q poolmanager estimate-swap-exact-amount-in 1248 10000000000ibc/D79E7D83AB399BFFF93433E54FAA480C191248FC556924A2A8351AE2638B3877 --swap-route-pool-ids 1248 --swap-route-denoms uosmo --node $url:26657)

sqs_custom_res=$(curl "$url/router/custom-quote?tokenIn=10000000000ibc/D79E7D83AB399BFFF93433E54FAA480C191248FC556924A2A8351AE2638B3877&tokenOutDenom=uosmo&poolIDs=1248")
sqs_custom_amount_out=$(echo $sqs_custom_res | jq .amount_out)

sqs_optimal_res=$(curl "$url/router/quote?tokenIn=10000000000ibc/D79E7D83AB399BFFF93433E54FAA480C191248FC556924A2A8351AE2638B3877&tokenOutDenom=uosmo")
sqs_optimal_amount_out=$(echo $sqs_optimal_res | jq .amount_out)

echo "chain_amount_out: $chain_amount_out"
echo "sqs_custom_amount_out: $sqs_custom_amount_out"
echo "sqs_optimal_amount_out: $sqs_optimal_amount_out"
