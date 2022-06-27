#!/bin/sh

# change staking denom to uosmo
osmosisd init --chain-id=localosmosis val
echo "satisfy adjust timber high purchase tuition stool faith fine install that you unaware feed domain license impose boss human eager hat rent enjoy dawn" | osmosisd keys add val --recover --keyring-backend=test
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="uosmo"' | sponge $HOME/.osmosisd/config/genesis.json
osmosisd add-genesis-account osmo1phaxpevm5wecex2jyaqty2a4v02qj7qmlmzk5a 100000000000uosmo,100000000000uion
osmosisd add-genesis-account osmo1cyyzpxplxdzkeea7kwsydadg87357qnahakaks 100000000000uosmo,100000000000uion
osmosisd add-genesis-account osmo18s5lynnmx37hq4wlrw9gdn68sg2uxp5rgk26vv 100000000000uosmo,100000000000uion
osmosisd add-genesis-account osmo1qwexv7c6sm95lwhzn9027vyu2ccneaqad4w8ka 100000000000uosmo,100000000000uion
osmosisd add-genesis-account osmo14hcxlnwlqtq75ttaxf674vk6mafspg8xwgnn53 100000000000uosmo,100000000000uion
osmosisd add-genesis-account osmo12rr534cer5c0vj53eq4y32lcwguyy7nndt0u2t 100000000000uosmo,100000000000uion
osmosisd add-genesis-account osmo1nt33cjd5auzh36syym6azgc8tve0jlvklnq7jq 100000000000uosmo,100000000000uion
osmosisd add-genesis-account osmo10qfrpash5g2vk3hppvu45x0g860czur8ff5yx0 100000000000uosmo,100000000000uion
osmosisd add-genesis-account osmo1f4tvsdukfwh6s9swrc24gkuz23tp8pd3e9r5fa 100000000000uosmo,100000000000uion
osmosisd add-genesis-account osmo1myv43sqgnj5sm4zl98ftl45af9cfzk7nhjxjqh 100000000000uosmo,100000000000uion
osmosisd add-genesis-account osmo14gs9zqh8m49yy9kscjqu9h72exyf295afg6kgk 100000000000uosmo,100000000000uion
osmosisd gentx val 500000000uosmo --keyring-backend=test --chain-id=localosmosis
osmosisd collect-gentxs
# update staking genesis
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["staking"]["params"]["unbonding_time"]="240s"' | sponge $HOME/.osmosisd/config/genesis.json
# update crisis variable to uosmo
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="uosmo"' | sponge $HOME/.osmosisd/config/genesis.json
# udpate gov genesis
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["gov"]["voting_params"]["voting_period"]="60s"' | sponge $HOME/.osmosisd/config/genesis.json
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="uosmo"' | sponge $HOME/.osmosisd/config/genesis.json
# update epochs genesis
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["epochs"]["epochs"][1]["duration"]="60s"' | sponge $HOME/.osmosisd/config/genesis.json
# update poolincentives genesis
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["poolincentives"]["lockable_durations"][0]="120s"' | sponge $HOME/.osmosisd/config/genesis.json
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["poolincentives"]["lockable_durations"][1]="180s"' | sponge $HOME/.osmosisd/config/genesis.json
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["poolincentives"]["lockable_durations"][2]="240s"' | sponge $HOME/.osmosisd/config/genesis.json
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["poolincentives"]["params"]["minted_denom"]="uosmo"' | sponge $HOME/.osmosisd/config/genesis.json
# update incentives genesis
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["incentives"]["lockable_durations"][0]="1s"' | sponge $HOME/.osmosisd/config/genesis.json
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["incentives"]["lockable_durations"][1]="120s"' | sponge $HOME/.osmosisd/config/genesis.json
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["incentives"]["lockable_durations"][2]="180s"' | sponge $HOME/.osmosisd/config/genesis.json
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["incentives"]["lockable_durations"][3]="240s"' | sponge $HOME/.osmosisd/config/genesis.json
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["incentives"]["params"]["distr_epoch_identifier"]="day"' | sponge $HOME/.osmosisd/config/genesis.json
# update mint genesis
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["mint"]["params"]["mint_denom"]="uosmo"' | sponge $HOME/.osmosisd/config/genesis.json
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["mint"]["params"]["epoch_identifier"]="day"' | sponge $HOME/.osmosisd/config/genesis.json
# update gamm genesis
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["gamm"]["params"]["pool_creation_fee"][0]["denom"]="uosmo"' | sponge $HOME/.osmosisd/config/genesis.json
# remove seeds
sed -i.bak -E 's#^(seeds[[:space:]]+=[[:space:]]+).*$#\1""#' ~/.osmosisd/config/config.toml
