#!/bin/bash

# through this script, we are testing setting reward receiver for locks specifically through the following operations:
# 1. superfluid stake 
# 1-a) check that changing reward receiver changes the reward receiving address properly(staking rewards get distributed properly)
# 1-b) check that changing reward receiver back to original address makes the reward be distributed to original address properly
# 2. incentives
# 1-a) check that changing reward receiver changes the reward receiving address properly(incentives rewards get distributed properly)
# 2-b) check that changing reward receiver back to original address makes the reward be distributed to original address properly


rm -rf $HOME/.osmosisd/

osmosisd init --chain-id=testing testing --home=$HOME/.osmosisd
osmosisd keys add validator --keyring-backend=test --home=$HOME/.osmosisd
osmosisd add-genesis-account $(osmosisd keys show validator -a --keyring-backend=test) 100000000000stake,100000000000uosmo,100000000000uion --home=$HOME/.osmosisd
osmosisd gentx validator 500000000stake --keyring-backend=test --home=$HOME/.osmosisd --chain-id=testing
osmosisd collect-gentxs --home=$HOME/.osmosisd

update_genesis () {    
    cat $HOME/.osmosisd/config/genesis.json | jq "$1" > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
}
# update genesis app state first before starting chain
# update staking genesis
update_genesis '.app_state["staking"]["params"]["unbonding_time"]="120s"'

# update governance genesis
update_genesis '.app_state["gov"]["voting_params"]["voting_period"]="30s"'

# update epochs genesis
update_genesis '.app_state["epochs"]["epochs"][0]["identifier"]="min"'
update_genesis '.app_state["epochs"]["epochs"][0]["duration"]="60s"'

# update poolincentives genesis
update_genesis '.app_state["poolincentives"]["lockable_durations"][0]="120s"'
update_genesis '.app_state["poolincentives"]["lockable_durations"][1]="180s"'
update_genesis '.app_state["poolincentives"]["lockable_durations"][2]="240s"'

# update incentives genesis
update_genesis '.app_state["incentives"]["params"]["distr_epoch_identifier"]="min"'
update_genesis '.app_state["incentives"]["lockable_durations"][0]="1s"'
update_genesis '.app_state["incentives"]["lockable_durations"][1]="120s"'
update_genesis '.app_state["incentives"]["lockable_durations"][2]="180s"'
update_genesis '.app_state["incentives"]["lockable_durations"][3]="240s"'

# update mint genesis
update_genesis '.app_state["mint"]["params"]["epoch_identifier"]="min"'

# update gamm genesis
update_genesis '.app_state["gamm"]["params"]["pool_creation_fee"][0]["denom"]="stake"'

# update superfluid genesis
update_genesis '.app_state["superfluid"]["params"]["minimum_risk_factor"]="0.500000000000000000"'

tmux new -s validator -d osmosisd start --home=$HOME/.osmosisd

sleep 7
# create pool
osmosisd tx gamm create-pool --pool-file=./stake-uosmo.json --from=validator --keyring-backend=test --chain-id=testing --yes --fees=1000000stake
sleep 7

# do a swap in the pool created
osmosisd tx gamm swap-exact-amount-in 100000uosmo 50000 --swap-route-pool-ids=1 --swap-route-denoms=stake --from=validator --keyring-backend=test --chain-id=testing --yes --fees=10000stake
sleep 7

# create a lock up with lock duration 360h
osmosisd tx lockup lock-tokens 10000000000000000000gamm/pool/1 --duration=360h --from=validator --keyring-backend=test --chain-id=testing --broadcast-mode=block --yes --fees=10000stake
sleep 7

# register superfluid asset through governance
osmosisd tx gov submit-proposal set-superfluid-assets-proposal --title="set superfluid assets" --description="set superfluid assets description" --superfluid-assets="gamm/pool/1" --deposit=10000000stake --from=validator --chain-id=testing --keyring-backend=test --broadcast-mode=block --yes --fees=10000stake
sleep 7
osmosisd tx gov deposit 1 10000000stake --from=validator --keyring-backend=test --chain-id=testing --broadcast-mode=block --yes --fees=10000stake 
sleep 7
osmosisd tx gov vote 1 yes --from=validator --keyring-backend=test --chain-id=testing --yes --fees=100000stake
sleep 7

# Command to extract operator_address
operator_address=$(./osmosisd q staking validators | yq e '.validators[].operator_address' -)

# Use the operator_address in the osmosisd command
osmosisd tx superfluid delegate 1 ${operator_address} --from=validator --keyring-backend=test --chain-id=testing --broadcast-mode=block --yes --fees=10000stake
sleep 7

# add different account that would be set as lock reward receiver
osmosisd keys add test1 --keyring-backend=test 
sleep 7

# set lock reward receiver 
osmosisd tx lockup set-reward-receiver-address 1 $(./osmosisd keys show test1 -a --keyring-backend=test) --from=validator --keyring-backend=test --chain-id=testing --broadcast-mode=block --yes --fees=10000stake
sleep 7

# query lock reward receiver 
# osmosisd q lockup lock-reward-receiver 1

# the new address would be receiving the lock rewards, while the original account does not receive the rewards anymore
# osmosisd q bank balances $(osmosisd keys show test1 -a --keyring-backend=test)
# osmosisd q bank balances $(osmosisd keys show validator -a --keyring-backend=test)

# set the lock reward receiver back to the original address
osmosisd tx lockup set-reward-receiver-address 1 $(./osmosisd keys show validator -a --keyring-backend=test) --from=validator --keyring-backend=test --chain-id=testing --broadcast-mode=block --yes --fees=10000stake
sleep 7

# the original address is back to receiving the lock rewards, while the new account does not receive the rewards anymore
# osmosisd q bank balances $(osmosisd keys show test1 -a --keyring-backend=test)
# osmosisd q bank balances $(osmosisd keys show validator -a --keyring-backend=test)

# set it back to the new account for further testing
osmosisd tx lockup set-reward-receiver-address 1 $(./osmosisd keys show test1 -a --keyring-backend=test) --from=validator --keyring-backend=test --chain-id=testing --broadcast-mode=block --yes --fees=10000stake
sleep 7

# create a second lock for incentive testing: this one with 1s lock duration
osmosisd tx lockup lock-tokens 10000000000000000000gamm/pool/1 --duration=1s --from=validator --keyring-backend=test --chain-id=testing --broadcast-mode=block --yes --fees=10000stake
sleep 7

# now let's also try adding lock incentives as well
osmosisd tx incentives create-gauge gamm/pool/1 1000000uion 0 --duration 120s  --start-time 1688115600 --epochs 100 --from=validator --keyring-backend=test --chain-id=testing --broadcast-mode=block --yes --fees=10000stake
sleep 7

# querying lock reward receiving address shows that we are receiving rewards as expected each epoch.
# querying original reward receiving address shows that we are not receiving rewards as expected each epoch.
# osmosisd q bank balances $(osmosisd keys show test1 -a --keyring-backend=test)
# osmosisd q bank balances $(osmosisd keys show validator -a --keyring-backend=test)
