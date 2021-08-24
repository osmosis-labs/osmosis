#!/bin/bash
# Based on the work of Joe (Chorus-One) for Microtick - https://github.com/microtick/bounties/tree/main/statesync
# Further based on the work of Bitcanna.
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

  NODE2_IP="http://5.9.106.185"
  RPC2="http://$NODE2_IP"
  RPC_PORT2=2000
  P2P_PORT2=2001

  #If you want to use a third StateSync Server... 
  #DOMAIN_3=seed1.bitcanna.io     # If you want to use domain names 
  #NODE3_IP=$(dig $DOMAIN_1 +short
  #RPC3="http://$NODE3_IP"
  #RPC_PORT3=26657
  #P2P_PORT3=26656

INTERVAL=1000

LATEST_HEIGHT=$(curl -s $RPC1:$RPC_PORT1/block | jq -r .result.block.header.height);
BLOCK_HEIGHT=$((($(($LATEST_HEIGHT / $INTERVAL)) -10) * $INTERVAL)); #Mark addition
  
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
#  NODE1_ID=$(curl -s "$RPC1:$RPC_PORT1/status" | jq -r .result.node_info.id)
#  NODE2_ID=$(curl -s "$RPC2:$RPC_PORT2/status" | jq -r .result.node_info.id)


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



# SED MAGIC TO PUT IN CONFIG FILE
#  sed -i.bak -E "s|^(enable[[:space:]]+=[[:space:]]+).*$|\1true| ; \
#  s|^(rpc_servers[[:space:]]+=[[:space:]]+).*$|\1\"http://$NODE1_IP:$RPC_PORT1,http://$NODE2_IP:$RPC_PORT2\"| ; \
#  s|^(trust_height[[:space:]]+=[[:space:]]+).*$|\1$BLOCK_HEIGHT| ; \
#  s|^(trust_hash[[:space:]]+=[[:space:]]+).*$|\1\"$TRUST_HASH\"| ; \
#  s|^(persistent_peers[[:space:]]+=[[:space:]]+).*$|\1\"${NODE1_ID}@${NODE1_IP}:${P2P_PORT1},${NODE2_ID}@${NODE2_IP}:${P2P_PORT2}\"| ; \
#  s|^(seeds[[:space:]]+=[[:space:]]+).*$|\1\"d6aa4c9f3ccecb0cc52109a95962b4618d69dd3f@seed1.bitcanna.io:26656,23671067d0fd40aec523290585c7d8e91034a771@seed2.bitcanna.io:16656\"|" $HOME/.bcna/config/config.toml

 
#  sed -E -i 's/minimum-gas-prices = \".*\"/minimum-gas-prices = \"0.01bcna\"/' $HOME/.bcna/config/app.toml

osmosisd unsafe-reset-all
osmosisd start
