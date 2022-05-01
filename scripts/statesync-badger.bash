# microtick and bitcanna contributed significantly here.
# badgerdb works
# This defaults to badgerdb.  Read the script, some stuff is commented out based on your system. 


# PRINT EVERY COMMAND
set -ux

# uncomment the three lines below to build osmosis

export GOPATH=~/go
export PATH=$PATH:~/go/bin

go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=badgerdb' -tags badgerdb ./...

# MAKE HOME FOLDER AND GET GENESIS
osmosisd init osmosis-rocks
wget -O ~/.osmosisd/config/genesis.json https://cloudflare-ipfs.com/ipfs/QmXRvBT3hgoXwwPqbK6a2sXUuArGM8wPyo1ybskyyUwUxs

# this will let tendermint know that we want rocks
sed -i 's/goleveldb/badgerdb/' ~/.osmosisd/config/config.toml


# Uncomment if resyncing a server
# osmosisd unsafe-reset-all --home /osmosis/osmosis



INTERVAL=1500

# GET TRUST HASH AND TRUST HEIGHT

LATEST_HEIGHT=$(curl -s https://osmosis.validator.network/block | jq -r .result.block.header.height);
BLOCK_HEIGHT=$(($LATEST_HEIGHT-$INTERVAL))
TRUST_HASH=$(curl -s "https://osmosis.validator.network/block?height=$BLOCK_HEIGHT" | jq -r .result.block_id.hash)


# TELL USER WHAT WE ARE DOING
echo "TRUST HEIGHT: $BLOCK_HEIGHT"
echo "TRUST HASH: $TRUST_HASH"


# export state sync vars
export OSMOSISD_P2P_MAX_NUM_OUTBOUND_PEERS=200
export OSMOSISD_STATESYNC_ENABLE=true
export OSMOSISD_STATESYNC_RPC_SERVERS="https://osmosis.validator.network:443,https://rpc.osmosis.notional.ventures:443,https://rpc-osmosis.ecostake.com:443"
export OSMOSISD_STATESYNC_TRUST_HEIGHT=$BLOCK_HEIGHT
export OSMOSISD_STATESYNC_TRUST_HASH=$TRUST_HASH


# THIS WILL FAIL BECAUSE THE APP VERSION IS CORRECTLY SET IN OSMOSIS
osmosisd start --db_backend badgerdb



# THIS WILL FIX THE APP VERSION, contributed by callum and claimens
git clone https://github.com/faddat/tendermint
cd tendermint
git checkout update-tmdb
go install -tags badgerdb ./...
tendermint set-app-version 1 --home ~/.osmosisd

# THERE, NOW IT'S SYNCED AND YOU CAN PLAY
osmosisd start --db_backend badgerdb
