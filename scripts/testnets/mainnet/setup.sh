#### Script for taking a modified mainnet export and starting up a node
# FIXME incomplete / WIP , not functional

#Parameters
export MAIN_GENESIS_URL="https://transfer.sh/o9B26h/genesis.json"
export CHAIN_ID="osmosis-testnet-main-4"
export VERSION="v4.2.0"


#TODO replace these with generate and write to file
export VALIDATOR_SEED="oven thank broccoli giant neither swamp betray moment birth lady wage student bicycle craft permit avoid burden tortoise oxygen file fix penalty two onion"
export FAUCET_SEED="catch spider raise grass flush audit result off auction stone best day soap stay organ canoe test spoon edit relief want warrior siren act"


sudo apt-get update
sudo apt-get upgrade -y
sudo apt-get install curl build-essential jq git wget liblz4-tool aria2 make -y

curl https://raw.githubusercontent.com/canha/golang-tools-install-script/master/goinstall.sh | bash

export GOROOT=/root/.go
export PATH=$GOROOT/bin:$PATH
export GOPATH=/root/go
export PATH=$GOPATH/bin:$PATH

cd /root/

rm /root/osmosis/ -rf 
git clone https://github.com/osmosis-labs/osmosis
cd osmosis
git checkout $VERSION
make install

rm /root/.osmosisd/ -rf

osmosisd config chain-id $CHAIN_ID
osmosisd init "testnet-validator" --chain-id=$CHAIN_ID

rm /root/.osmosisd/config/genesis.json

curl -o /root/.osmosisd/config/genesis.json $MAIN_GENESIS_URL

##TODO following section should be part of modification script, not setup

echo $VALIDATOR_SEED | osmosisd keys add validator --recover --keyring-backend=test
echo $FAUCET_SEED | osmosisd keys add faucet --recover --keyring-backend=test

osmosisd add-genesis-account validator 1000000000000uosmo,1000000000uion --keyring-backend=test
osmossid add-genesis-account faucet 1000000000000uosmo,1000000000uion --keyring-backend=test