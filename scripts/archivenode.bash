#!/bin/bash



# Archive node script
# NB:  you can also download archives at quicksync:
# https://quicksync.io/networks/osmosis.html
# 2nd NB:  may explode but works for me. - Jacob


git checkout v1.0.1
make install
osmosisd init archive
wget -O ~/.osmosisd/config/genesis.json https://github.com/osmosis-labs/networks/raw/main/osmosis-1/genesis.json
osmosisd start --pruning nothing --p2p.seeds 085f62d67bbf9c501e8ac84d4533440a1eef6c45@95.217.196.54:26656,f515a8599b40f0e84dfad935ba414674ab11a668@osmosis.blockpane.com:26656
git checkout v3.1.0
make install
osmosisd start --pruning nothing --p2p.seeds 085f62d67bbf9c501e8ac84d4533440a1eef6c45@95.217.196.54:26656,f515a8599b40f0e84dfad935ba414674ab11a668@osmosis.blockpane.com:26656
git checkout v4.0.0
osmosisd start --pruning nothing --p2p.seeds 085f62d67bbf9c501e8ac84d4533440a1eef6c45@95.217.196.54:26656,f515a8599b40f0e84dfad935ba414674ab11a668@osmosis.blockpane.com:26656
