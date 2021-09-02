
#!/bin/bash
# microtick and bitcanna contributed significantly here.


set -e

# Change for your custom chain
export GOPATH=~/go
export PATH=$PATH:~/go/bin
go install ./...


# MAKE HOME FOLDER AND GET GENESIS
osmosisd init test 
wget -O $HOME/.osmosisd/config/genesis.json $GENESIS 
  

INTERVAL=1500

# GET TRUST HASH AND TRUST HEIGHT

LATEST_HEIGHT=$(curl -s $RPC1:$RPC_PORT1/block | jq -r .result.block.header.height);
BLOCK_HEIGHT=$(($LATEST_HEIGHT-$INTERVAL)) 
TRUST_HASH=$(curl -s "$RPC1:$RPC_PORT1/block?height=$BLOCK_HEIGHT" | jq -r .result.block_id.hash)


# TELL USER WHAT WE ARE DOING
echo "TRUST HEIGHT: $BLOCK_HEIGHT"
echo "TRUST HASH: $TRUST_HASH"


# export state sync vars
export OSMOSISD_STATESYNC_ENABLE=true
export OSMOSISD_STATESYNC_RPC_SERVERS="162.55.132.230:2001,162.55.132.230:2001"
export OSMOSISD_STATESYNC_TRUST_HEIGHT=$BLOCK_HEIGHT
export OSMOSISD_STATESYNC_TRUST_HASH=$TRUST_HASH

osmosisd unsafe-reset-all
osmosisd start
