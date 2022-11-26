
#!/bin/bash
# microtick and bitcanna contributed significantly here.




# PRINT EVERY COMMAND
set -ux

# uncomment the three lines below to build osmosis

go mod edit -replace github.com/tendermint/tm-db=github.com/baabeetaa/tm-db@pebble
go mod tidy
go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=pebbledb -X github.com/tendermint/tm-db.ForceSync=1' -tags pebbledb ./...


# MAKE HOME FOLDER AND GET GENESIS
osmosisd init test 
wget -O ~/.osmosisd/config/genesis.json https://github.com/osmosis-labs/osmosis/raw/main/networks/osmosis-1/genesis.json


INTERVAL=1500

# GET TRUST HASH AND TRUST HEIGHT

LATEST_HEIGHT=$(curl -s https://rpc.osmosis.zone/block | jq -r .result.block.header.height);
BLOCK_HEIGHT=$(($LATEST_HEIGHT-$INTERVAL))
TRUST_HASH=$(curl -s "https://rpc.osmosis.zone/block?height=$BLOCK_HEIGHT" | jq -r .result.block_id.hash)


# TELL USER WHAT WE ARE DOING
echo "TRUST HEIGHT: $BLOCK_HEIGHT"
echo "TRUST HASH: $TRUST_HASH"


# export state sync vars
export OSMOSISD_P2P_MAX_NUM_OUTBOUND_PEERS=0 #this is so that we only connect to Notional's peer
export OSMOSISD_P2P_PERSISTENT_PEERS=ad431b916e5c7f55b1d4a852db115510f31d2d3a@65.108.232.181:26656
export OSMOSISD_STATESYNC_ENABLE=true
export OSMOSISD_STATESYNC_RPC_SERVERS="https://rpc.osmosis.zone:443,http://65.108.232.181:26657"
export OSMOSISD_STATESYNC_TRUST_HEIGHT=$BLOCK_HEIGHT
export OSMOSISD_STATESYNC_TRUST_HASH=$TRUST_HASH




# THERE, NOW IT'S SYNCED AND YOU CAN PLAY
osmosisd start --db_backend pebbledb
