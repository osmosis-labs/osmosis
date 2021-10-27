
#!/bin/bash
# microtick and bitcanna contributed significantly here.

# PRINT EVERY COMMAND
set -uxe

export GOPATH=~/go
export PATH=$PATH:~/go/bin
go install ./...


# MAKE HOME FOLDER AND GET GENESIS
osmosisd init test 
wget -O ~/.osmosisd/config/genesis.json https://cloudflare-ipfs.com/ipfs/QmXRvBT3hgoXwwPqbK6a2sXUuArGM8wPyo1ybskyyUwUxs  

INTERVAL=1500

# GET TRUST HASH AND TRUST HEIGHT

LATEST_HEIGHT=$(curl -s 162.55.132.230:2001/block | jq -r .result.block.header.height);
BLOCK_HEIGHT=$(($LATEST_HEIGHT-$INTERVAL)) 
TRUST_HASH=$(curl -s "162.55.132.230:2001/block?height=$BLOCK_HEIGHT" | jq -r .result.block_id.hash)


# TELL USER WHAT WE ARE DOING
echo "TRUST HEIGHT: $BLOCK_HEIGHT"
echo "TRUST HASH: $TRUST_HASH"


# export state sync vars
export OSMOSISD_STATESYNC_ENABLE=true
export OSMOSISD_STATESYNC_RPC_SERVERS="162.55.132.230:2001,162.55.132.230:2001"
export OSMOSISD_STATESYNC_TRUST_HEIGHT=$BLOCK_HEIGHT
export OSMOSISD_STATESYNC_TRUST_HASH=$TRUST_HASH
export OSMOSISD_P2P_PERSISTENT_PEERS="63aba59a7da5197c0fbcdc13e760d7561791dca8@162.55.132.230:2000"

osmosisd unsafe-reset-all
osmosisd start
