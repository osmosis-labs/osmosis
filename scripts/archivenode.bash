#!/bin/bash



# Archive node script
# NB:  you can also download archives at quicksync:
# https://quicksync.io/networks/osmosis.html
# 2nd NB: you can change OSMOSISD_PRUNING=nothing to OSMOSISD_PRUNING=default OR you could also set the pruning settings manually with OSMOSISD_PRUNING=custom
# 3rd NB: you might want to use this to test different databases, and to do that my recommended technique is like:
# go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=rocksdb' -tags rocksdb ./...
# if you do not use the ldflags thing you won't use the chosen db for everything, so best use it.


export OSMOSISD_PRUNING=nothing
export OSMOSISD_DB_BACKEND=goleveldb

# VERSION THREE
echo "v3 took" > howlong
git checkout v3.x
osmosisd init speedrun
wget -O ~/.osmosisd/config/addrbook.json https://quicksync.io/addrbook.osmosis.json
wget -O ~/.osmosisd/config/genesis.json https://github.com/osmosis-labs/networks/raw/main/osmosis-1/genesis.json
time osmosisd start 
make install

# VERSION FOUR
echo "v4 took" >> howlong
git checkout v4.x
make install
osmosisd start 

# VERSION SIX
echo "v6 took" >> howlong
git checkout v6.x
make install
osmosisd start

# VERSION SEVEN
echo "v7 took" >> howlong
git checkout v7.x
make install
