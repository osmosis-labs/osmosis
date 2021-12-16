##### Spins up and starts a single validator Osmosis testnet

#### Parameters
export CHAIN_ID="osmosis-clean-testnet-X"
export VERSION="v5.0.0"

#Constants
export GENESIS="/root/.osmosisd/config/genesis.json"
export TMP_GENESIS="/root/.osmosisd/config/genesis.json.tmp"

#### Initial node setup
sudo apt-get update
sudo apt-get upgrade -y
sudo apt-get install curl build-essential jq git wget make -y

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


#### This section is for setting up local keys and genesis state
## TODO Should probably be separated out into another script


#TODO Should be replaced with dynamically generated where seed is written to a local file
echo "oven thank broccoli giant neither swamp betray moment birth lady wage student bicycle craft permit avoid burden tortoise oxygen file fix penalty two onion" | osmosisd keys add validator --recover --keyring-backend=test
echo "catch spider raise grass flush audit result off auction stone best day soap stay organ canoe test spoon edit relief want warrior siren act" | osmosisd keys add faucet --recover --keyring-backend=test

osmosisd add-genesis-account validator 2000000000000uosmo --keyring-backend=test
osmosisd add-genesis-account faucet 2000000000000uosmo,2000000000uion --keyring-backend=test

jq '.app_state.poolincentives.distr_info = {"total_weight": "1000","records": [{"gauge_id": "0","weight": "1000"}]}' $GENESIS > $TMP_GENESIS && mv $TMP_GENESIS $GENESIS

osmosisd gentx validator 1000000000000uosmo --chain-id=$CHAIN_ID --commission-rate=0.05 --commission-max-change-rate=0.01 --commission-max-rate=1.0 --keyring-backend=test
osmosisd collect-gentxs

sed -i 's%stake%uosmo%g' $GENESIS

osmosisd unsafe-reset-all
osmosisd start


