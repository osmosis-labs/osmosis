#!/bin/bash
git checkout v1.0.1
make install
osmosisd init archive
cp contrib/osmoarchive.service /etc/systemd/system
osmosisd start --pruning nothing --p2p.seed_nodes 085f62d67bbf9c501e8ac84d4533440a1eef6c45@95.217.196.54:26656,f515a8599b40f0e84dfad935ba414674ab11a668@osmosis.blockpane.com:26656
git checkout v3.1.0
make install
osmosisd start --pruning nothing --p2p.seed_nodes 085f62d67bbf9c501e8ac84d4533440a1eef6c45@95.217.196.54:26656,f515a8599b40f0e84dfad935ba414674ab11a668@osmosis.blockpane.com:26656
git checkout v4.0.0
osmosisd start --pruning nothing --p2p.seed_nodes 085f62d67bbf9c501e8ac84d4533440a1eef6c45@95.217.196.54:26656,f515a8599b40f0e84dfad935ba414674ab11a668@osmosis.blockpane.com:26656
