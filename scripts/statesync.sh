#!/bin/bash
# Based on the work of Joe (Chorus-One) for Microtick - https://github.com/microtick/bounties/tree/main/statesync
# Further based on the work of Bitcanna.
# Adapted for osmosis by Jacob Gadikian of Notional Validation.
# For now this is a test script that ensures that state sync is working. 
# You need config in two peers (avoid seed servers) this values in app.toml:
#     [state-sync]
#     snapshot-interval = 1000
#     snapshot-keep-recent = 10
# Simplifications from: https://binary-star.plus/cosmos-sdk-state-sync-guide/
# Pruning should be fine tuned also, for this testings is set to nothing
#     pruning = "nothing"

set -e

# Change for your custom chain


  
NODE1_IP="144.76.183.180"
RPC1="http://$NODE1_IP"
P2P_PORT1=2000
RPC_PORT1=2001

NODE2_IP="144.76.183.180"
RPC2="http://$NODE2_IP"
P2P_PORT2=2000
RPC_PORT2=2001

#If you want to use a third StateSync Server... 
#DOMAIN_3=seed1.bitcanna.io     # If you want to use domain names 
#NODE3_IP=$(dig $DOMAIN_1 +short
#RPC3="http://$NODE3_IP"
#RPC_PORT3=26657
#P2P_PORT3=26656

INTERVAL=1000

LATEST_HEIGHT=$(curl -s $RPC1:$RPC_PORT1/block | jq -r .result.block.header.height)
BLOCK_HEIGHT=$(($LATEST_HEIGHT-$INTERVAL))

echo "State syncing from $BLOCK_HEIGHT"
  
if [ $BLOCK_HEIGHT -eq 0 ]; then
  echo "Error: Cannot state sync to block 0; Latest block is $LATEST_HEIGHT and must be at least $INTERVAL; wait a few blocks!"
  exit 1
fi

TRUST_HASH=$(curl -s "$RPC1:$RPC_PORT1/block?height=$BLOCK_HEIGHT" | jq -r .result.block_id.hash)
if [ "$TRUST_HASH" == "null" ]; then
  echo "Error: Cannot find block hash. This shouldn't happen :/"
  exit 1
fi


# Not needed because of embedded seeds

echo "$RPC1:$RPC_PORT1/status"
echo "$RPC2:$RPC_PORT2/status"
NODE1_ID=$(curl -s "$RPC1:$RPC_PORT1/status" | jq -r .result.node_info.id)
NODE2_ID=$(curl -s "$RPC2:$RPC_PORT2/status" | jq -r .result.node_info.id)

echo "Node 1 id is: $NODE1_ID"
echo "Node 2 id is: $NODE2_ID"
echo "Trust hash is: $TRUST_HASH"


  #NODE3_ID=$(curl -s "$RPC3:$RPC_PORT3/status" | jq -r .result.node_info.id)

# [statesync]
# enable = true
# rpc_servers = "foo.net:26657,bar.com:26657"
# trust_height = 1964
# trust_hash = "6FD28DAAAC79B77F589AE692B6CD403412CE27D0D2629E81951607B297696E5B"
# trust_period = "336h"  # 2/3 of unbonding time  

export OSMOSISD_STATESYNC_ENABLE=true
export OSMOSISD_STATESYNC_RPC_SERVERS="$RPC1:$RPC_PORT1/status,$RPC2:$RPC_PORT2/status"
export OSMOSISD_STATESYNC_TRUST_HEIGHT=$BLOCK_HEIGHT
export OSMOSISD_STATESYNC_TRUST_HASH=$TRUST_HASH
export OSMOSISD_STATESYNC_TRUST_PERIOD="224h"
export OSMOSISD_P2P_PERSISTENT_PEERS="$NODE1_ID@$NODE1_IP:$P2P_PORT1,$NODE2_ID@$NODE2_IP:$P2P_PORT2"




osmosisd unsafe-reset-all
osmosisd start
