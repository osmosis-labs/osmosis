# microtick and bitcanna contributed significantly here.
# rocksdb works
# This defaults to rocksdb.  Read the script, some stuff is commented out based on your system. 


# PRINT EVERY COMMAND
set -ux

# uncomment the three lines below to build osmosis

export GOPATH=~/go
export PATH=$PATH:~/go/bin

# Use if building on a mac
# export CGO_CFLAGS="-I/opt/homebrew/Cellar/rocksdb/6.27.3/include"
# export CGO_LDFLAGS="-L/opt/homebrew/Cellar/rocksdb/6.27.3/lib -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4 -lzstd -L/opt/homebrew/Cellar/snappy/1.1.9/lib -L/opt/homebrew/Cellar/lz4/1.9.3/lib/ -L /opt/homebrew/Cellar/zstd/1.5.1/lib/"
go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=rocksdb' -tags rocksdb ./...

# MAKE HOME FOLDER AND GET GENESIS
osmosisd init osmosis-rocks
wget -O ~/.osmosisd/config/genesis.json https://cloudflare-ipfs.com/ipfs/QmXRvBT3hgoXwwPqbK6a2sXUuArGM8wPyo1ybskyyUwUxs

# this will let tendermint know that we want rocks
sed -i 's/goleveldb/rocksdb/' ~/.osmosisd/config/config.toml


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

# Rockdb won't make this folder, so we make it 
mkdir -p ~/.osmosisd/data/snapshots/metadata.db

# THIS WILL FAIL BECAUSE THE APP VERSION IS CORRECTLY SET IN OSMOSIS
osmosisd start --db_backend rocksdb 


# --db_path /tmp/osmosisd/db --db_rocksdb_options "compression=kNoCompression" --db_rocksdb_options "compaction_style=kCompactionStyleLevel" --db_rocksdb_options "level_compaction_dynamic_level_bytes=true" --db_rocksdb_options "num_levels=4" --db_rocksdb_options "max_bytes_for_level_base=104857600" --db_rocksdb_options "max_bytes_for_level_multiplier=1" --db_rocksdb_options "max_background_compactions=4" --db_rocksdb_options "max_background_flushes=4" --db_rocksdb_options "write_buffer_size=104857600" --db_rocksdb_options "target_file_size_base=104857600" --db_rocksdb_options "target_file_size_multiplier=1" --db_rocksdb_options "max_write_buffer_number=4" --db_rocksdb_options "min_write_buffer_number_to_merge=2" --db_rocksdb_options "max_grandparent_overlap_factor=10" --db_rocksdb_options "max_bytes_for_level_multiplier_additional=[4,4,4,4]" --db_rocksdb_options "compaction_pri=kOldestSmallestSeqFirst" --db_rocksdb_options "compaction_options_fifo={max_table_files_size=104857600}" --db_rocksdb_options "compaction_options_universal={size_ratio=1, min_merge_width=2, max_merge_width=2, max_size_amplification_percent=20}" --db_rocksdb_options "compaction_options_level_base={compression=kNoCompression, filter_policy=kNoFilter}" --db_rocksdb_options "compaction_options_level_base={compression=kSnappyCompression, filter_policy=

# THIS WILL FIX THE APP VERSION, contributed by callum and claimens
git clone https://github.com/notional-labs/tendermint
cd tendermint
git checkout remotes/origin/callum/app-version
go install -tags rocksdb ./...
tendermint set-app-version 1 --home ~/.osmosisd

# THERE, NOW IT'S SYNCED AND YOU CAN PLAY
osmosisd start --db_backend rocksdb
