##### Spins up and starts a single validator Osmosis testnet

#### Parameters
export CHAIN_ID="osmosis-clean-testnet-X"
export VERSION="v5.0.0"

#### Initial node setup
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


#### This section is for setting up local keys and genesis state
## TODO Should probably be separated out into another script


#TODO Should be replaced with dynamically generated where seed is written to a local file
echo "oven thank broccoli giant neither swamp betray moment birth lady wage student bicycle craft permit avoid burden tortoise oxygen file fix penalty two onion" | osmosisd keys add validator --recover --keyring-backend=test
echo "catch spider raise grass flush audit result off auction stone best day soap stay organ canoe test spoon edit relief want warrior siren act" | osmosisd keys add faucet --recover --keyring-backend=test
echo "name salt burden assume awkward copy morning any kangaroo crucial width brother organ casual brief scorpion actress lady hover figure idea another employ another" | osmosisd keys add clawback --recover --keyring-backend=test

echo "travel renew first fiction trick fly disease advance hunt famous absurd region" | osmosisd keys add keplr1 --recover --keyring-backend=test

osmosisd add-genesis-account validator 2000000000000uosmo --keyring-backend=test
osmosisd add-genesis-account faucet 2000000000000uosmo,2000000000uion,2000000000000ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2 --keyring-backend=test
osmosisd add-genesis-account clawback 2000000000uosmo,2000000000uion,2000000000000ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2 --keyring-backend=test
# osmosisd add-genesis-account keplr1 2000000000uosmo,2000000000uion,2000000000000ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2 --keyring-backend=test

osmosisd prepare-genesis mainnet $CHAIN_ID

osmosisd gentx validator 1000000000000uosmo --chain-id=$CHAIN_ID --commission-rate=0.05 --commission-max-change-rate=0.01 --commission-max-rate=1.0 --keyring-backend=test
osmosisd collect-gentxs

#TODO most of this should be done with jq instead of sed, for better readability
sed -i 's%osmo15qgrux35vf87hfx77efk6hw3lcc5slv0awc3qh%osmo1h5rcx73zj474nrkkcyf28tud47k8thy59pt529%g' /root/.osmosisd/config/genesis.json
sed -i 's%stake%uosmo%g' /root/.osmosisd/config/genesis.json
sed -i 's%"voting_period": "172800s"%"voting_period": "60s"%g' /root/.osmosisd/config/genesis.json
sed -i 's%"distr_epoch_identifier": "week"%"distr_epoch_identifier": "day"%g' /root/.osmosisd/config/genesis.json
sed -i 's%"epoch_identifier": "week"%"epoch_identifier": "day"%g' /root/.osmosisd/config/genesis.json
sed -i 's%"duration": "86400s"%"duration": "120s"%g' /root/.osmosisd/config/genesis.json
sed -i 's%"duration_until_decay": "3600s"%"duration_until_decay": "30s"%g' /root/.osmosisd/config/genesis.json
sed -i 's%"duration_of_decay": "18000s"%"duration_of_decay": "600s"%g' /root/.osmosisd/config/genesis.json

#state that was specific for v4 -> v5, should be removed for future setups
jq '.app_state.claim.module_account_balance.amount = "393880785"' /root/.osmosisd/config/genesis.json > /root/.osmosisd/config/genesis_2.json
jq '.app_state.claim.claim_records[0] = {"address": "osmo1h5rcx73zj474nrkkcyf28tud47k8thy59pt529","initial_claimable_amount": [{"denom": "uosmo","amount": "393880785"}],"action_completed": [false,false,false,false]}' /root/.osmosisd/config/genesis_2.json > /root/.osmosisd/config/genesis_3.json

jq '.app_state.poolincentives.distr_info = {"total_weight": "1000","records": [{"gauge_id": "0","weight": "1000"}]}' /root/.osmosisd/config/genesis_3.json > /root/.osmosisd/config/genesis.json

sed -i 's%minimum-gas-prices = ""%minimum-gas-prices = "0.01uosmo"%g' /root/.osmosisd/config/app.toml



osmosisd unsafe-reset-all
osmosisd start


