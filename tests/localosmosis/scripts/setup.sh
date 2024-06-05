#!/bin/sh

CHAIN_ID=localsymphony
OSMOSIS_HOME=$HOME/.symphonyd
CONFIG_FOLDER=$OSMOSIS_HOME/config
MONIKER=val
STATE='false'

MNEMONIC="bottom loan skill merry east cradle onion journey palm apology verb edit desert impose absurd oil bubble sweet glove shallow size build burst effort"
POOLSMNEMONIC="traffic cool olive pottery elegant innocent aisle dial genuine install shy uncle ride federal soon shift flight program cave famous provide cute pole struggle"

while getopts s flag
do
    case "${flag}" in
        s) STATE='true';;
    esac
done

install_prerequisites () {
    apk add dasel
}

edit_genesis () {

    GENESIS=$CONFIG_FOLDER/genesis.json

    # Update staking module
    dasel put string -f $GENESIS '.app_state.staking.params.bond_denom' 'note'
    dasel put string -f $GENESIS '.app_state.staking.params.unbonding_time' '240s'

    # Update bank module
    dasel put string -f $GENESIS '.app_state.bank.denom_metadata.[0].description' 'Registered denom uion for localsymphony testing'
    dasel put string -f $GENESIS '.app_state.bank.denom_metadata.[0].denom_units.[0].denom' 'uion'
    dasel put string -f $GENESIS '.app_state.bank.denom_metadata.[0].denom_units.[0].exponent' 0
    dasel put string -f $GENESIS '.app_state.bank.denom_metadata.[0].base' 'uion'
    dasel put string -f $GENESIS '.app_state.bank.denom_metadata.[0].display' 'uion'
    dasel put string -f $GENESIS '.app_state.bank.denom_metadata.[0].name' 'uion'
    dasel put string -f $GENESIS '.app_state.bank.denom_metadata.[0].symbol' 'uion'

    dasel put string -f $GENESIS '.app_state.bank.denom_metadata.[1].description' 'Registered denom note for localsymphony testing'
    dasel put string -f $GENESIS '.app_state.bank.denom_metadata.[1].denom_units.[0].denom' 'note'
    dasel put string -f $GENESIS '.app_state.bank.denom_metadata.[1].denom_units.[0].exponent' 0
    dasel put string -f $GENESIS '.app_state.bank.denom_metadata.[1].base' 'note'
    dasel put string -f $GENESIS '.app_state.bank.denom_metadata.[1].display' 'note'
    dasel put string -f $GENESIS '.app_state.bank.denom_metadata.[1].name' 'note'
    dasel put string -f $GENESIS '.app_state.bank.denom_metadata.[1].symbol' 'note'

    # Update crisis module
    dasel put string -f $GENESIS '.app_state.crisis.constant_fee.denom' 'note'

    # Update gov module
    dasel put string -f $GENESIS '.app_state.gov.voting_params.voting_period' '60s'
    dasel put string -f $GENESIS '.app_state.gov.deposit_params.min_deposit.[0].denom' 'note'

    # Update epochs module
    dasel put string -f $GENESIS '.app_state.epochs.epochs.[1].duration' "60s"

    # Update poolincentives module
    dasel put string -f $GENESIS '.app_state.poolincentives.lockable_durations.[0]' "120s"
    dasel put string -f $GENESIS '.app_state.poolincentives.lockable_durations.[1]' "180s"
    dasel put string -f $GENESIS '.app_state.poolincentives.lockable_durations.[2]' "240s"
    dasel put string -f $GENESIS '.app_state.poolincentives.params.minted_denom' "note"

    # Update incentives module
    dasel put string -f $GENESIS '.app_state.incentives.lockable_durations.[0]' "1s"
    dasel put string -f $GENESIS '.app_state.incentives.lockable_durations.[1]' "120s"
    dasel put string -f $GENESIS '.app_state.incentives.lockable_durations.[2]' "180s"
    dasel put string -f $GENESIS '.app_state.incentives.lockable_durations.[3]' "240s"
    dasel put string -f $GENESIS '.app_state.incentives.params.distr_epoch_identifier' "hour"

    # Update mint module
    dasel put string -f $GENESIS '.app_state.mint.params.mint_denom' "note"
    dasel put string -f $GENESIS '.app_state.mint.params.epoch_identifier' "hour"

    # Update poolmanager module
    dasel put string -f $GENESIS '.app_state.poolmanager.params.pool_creation_fee.[0].denom' "note"

    # Update txfee basedenom
    dasel put string -f $GENESIS '.app_state.txfees.basedenom' "note"

    # Update wasm permission (Nobody or Everybody)
    dasel put string -f $GENESIS '.app_state.wasm.params.code_upload_access.permission' "Everybody"

    # Update concentrated-liquidity (enable pool creation)
    dasel put bool -f $GENESIS '.app_state.concentratedliquidity.params.is_permissionless_pool_creation_enabled' true
}

add_genesis_accounts () {

    symphonyd add-genesis-account symphony1p7mp7r9f9f6sf2c95ht42ncm6ga96ha8xghdeg 100000000000note,100000000000uion,100000000000stake,100000000000uusdc,100000000000uweth --home $OSMOSIS_HOME
    # note such large amounts are set for e2e tests on FE 
    symphonyd add-genesis-account symphony1c605nvcw94rvvehrcdfj85qe09ulseyt0efhk7 9999999999999999999999999999999999999999999999999note,9999999999999999999999999999999999999999999999999uion,100000000000stake,100000000000uusdc,100000000000uweth --home $OSMOSIS_HOME
    symphonyd add-genesis-account symphony1jpr5824frn5472qm73ckfe2c3rh6vrn4lvlgj7 100000000000note,100000000000uion,100000000000stake,100000000000uusdc,100000000000uweth --home $OSMOSIS_HOME
    symphonyd add-genesis-account symphony1amr6zrvs0hymf62qd5mwvshx94ul8cgfu9jtxn 100000000000note,100000000000uion,100000000000stake,100000000000uusdc,100000000000uweth --home $OSMOSIS_HOME
    symphonyd add-genesis-account symphony1egts9ayaqr6t54ahs62awmz5smuf764uu5f5xv 100000000000note,100000000000uion,100000000000stake,100000000000uusdc,100000000000uweth --home $OSMOSIS_HOME
    symphonyd add-genesis-account symphony1450weujlqvtd0d5z59v388jmzwyk3e6qhlj5r5 100000000000note,100000000000uion,100000000000stake,100000000000uusdc,100000000000uweth --home $OSMOSIS_HOME
    symphonyd add-genesis-account symphony12mdnm5yv5dfz37qsu0eu60x8qwxxl0x7sqqzn0 100000000000note,100000000000uion,100000000000stake,100000000000uusdc,100000000000uweth --home $OSMOSIS_HOME
    symphonyd add-genesis-account symphony1ar8mfrrtkwlm62wgu88d0cfleng5gl8y062gsn 100000000000note,100000000000uion,100000000000stake,100000000000uusdc,100000000000uweth --home $OSMOSIS_HOME
    symphonyd add-genesis-account symphony1kvgujs5yg9h6l6e265smwx99fmnnmc4af5v0ah 100000000000note,100000000000uion,100000000000stake,100000000000uusdc,100000000000uweth --home $OSMOSIS_HOME
    symphonyd add-genesis-account symphony1ww5e3y7ptw8h3lc0cumxe5lmcu3m53dn7qyn4k 100000000000note,100000000000uion,100000000000stake,100000000000uusdc,100000000000uweth --home $OSMOSIS_HOME
    symphonyd add-genesis-account symphony1tsehv6f0v7ce4gy7574thxnp6v8jx7jm4evkpe 100000000000note,100000000000uion,100000000000stake,100000000000uusdc,100000000000uweth --home $OSMOSIS_HOME
    symphonyd add-genesis-account symphony1fg5d24fgmxgux2p8e6xm8vjdjza8xy3ju6ta6m 1000000000000note,1000000000000uion,1000000000000stake,1000000000000uusdc,1000000000000uweth --home $OSMOSIS_HOME

    echo $MNEMONIC | symphonyd keys add $MONIKER --recover --keyring-backend=test --home $OSMOSIS_HOME
    echo $POOLSMNEMONIC | symphonyd keys add pools --recover --keyring-backend=test --home $OSMOSIS_HOME
    symphonyd gentx $MONIKER 500000000note --keyring-backend=test --chain-id=$CHAIN_ID --home $OSMOSIS_HOME

    symphonyd collect-gentxs --home $OSMOSIS_HOME
}

edit_config () {

    # Remove seeds
    dasel put string -f $CONFIG_FOLDER/config.toml '.p2p.seeds' ''

    # Expose the rpc
    dasel put string -f $CONFIG_FOLDER/config.toml '.rpc.laddr' "tcp://0.0.0.0:26657"
    
    # Expose pprof for debugging
    # To make the change enabled locally, make sure to add 'EXPOSE 6060' to the root Dockerfile
    # and rebuild the image.
    dasel put string -f $CONFIG_FOLDER/config.toml '.rpc.pprof_laddr' "0.0.0.0:6060"
}

enable_cors () {

    # Enable cors on RPC
    dasel put string -f $CONFIG_FOLDER/config.toml -v "*" '.rpc.cors_allowed_origins.[]'
    dasel put string -f $CONFIG_FOLDER/config.toml -v "Accept-Encoding" '.rpc.cors_allowed_headers.[]'
    dasel put string -f $CONFIG_FOLDER/config.toml -v "DELETE" '.rpc.cors_allowed_methods.[]'
    dasel put string -f $CONFIG_FOLDER/config.toml -v "OPTIONS" '.rpc.cors_allowed_methods.[]'
    dasel put string -f $CONFIG_FOLDER/config.toml -v "PATCH" '.rpc.cors_allowed_methods.[]'
    dasel put string -f $CONFIG_FOLDER/config.toml -v "PUT" '.rpc.cors_allowed_methods.[]'

    # Enable unsafe cors and swagger on the api
    dasel put bool -f $CONFIG_FOLDER/app.toml -v "true" '.api.swagger'
    dasel put bool -f $CONFIG_FOLDER/app.toml -v "true" '.api.enabled-unsafe-cors'

    # Enable cors on gRPC Web
    dasel put bool -f $CONFIG_FOLDER/app.toml -v "true" '.grpc-web.enable-unsafe-cors'

    # Enable SQS & route caching
    dasel put string -f $CONFIG_FOLDER/app.toml -v "true" '.symphony-sqs.is-enabled'
    dasel put string -f $CONFIG_FOLDER/app.toml -v "true" '.symphony-sqs.route-cache-enabled'

    dasel put string -f $CONFIG_FOLDER/app.toml -v "redis" '.symphony-sqs.db-host'
}

run_with_retries() {
  cmd=$1
  success_msg=$2

  substring='code: 0'
  COUNTER=0

  while [ $COUNTER -lt 15 ]; do
    string=$(eval $cmd 2>&1)
    echo $string

    if [ "$string" != "${string%"$substring"*}" ]; then
      echo "$success_msg"
      break
    else
      COUNTER=$((COUNTER+1))
      sleep 0.5
    fi
  done
}

# Define the functions using the new function
create_two_asset_pool() {
  run_with_retries "symphonyd tx gamm create-pool --pool-file=$1 --from pools --chain-id=$CHAIN_ID --home $OSMOSIS_HOME --keyring-backend=test -b block --fees 5000note --yes" "create two asset pool: successful"
}

create_stable_pool() {
  run_with_retries "symphonyd tx gamm create-pool --pool-file=uwethUusdcStablePool.json --pool-type=stableswap --from pools --chain-id=$CHAIN_ID --home $OSMOSIS_HOME --keyring-backend=test -b block --fees 5000note --yes" "create two asset pool: successful"
}

create_three_asset_pool() {
  run_with_retries "symphonyd tx gamm create-pool --pool-file=nativeDenomThreeAssetPool.json --from pools --chain-id=$CHAIN_ID --home $OSMOSIS_HOME --keyring-backend=test -b block --fees 5000note --gas 900000 --yes" "create three asset pool: successful"
}

create_concentrated_pool() {
  run_with_retries "symphonyd tx concentratedliquidity create-pool uion note 1 \"0.0005\" --from pools --chain-id=$CHAIN_ID --home $OSMOSIS_HOME --keyring-backend=test -b block --fees 5000note --gas 900000 --yes" "create concentrated pool: successful"
}

create_concentrated_pool_positions () {
    # Define an array to hold the parameters that change for each command
    set "[-1620000] 3420000" "305450 315000" "315000 322500" "300000 309990" "[-108000000] 342000000" "[-108000000] 342000000"

    substring='code: 0'
    COUNTER=0
    # Loop through each set of parameters in the array
    for param in "$@"; do
        run_with_retries "symphonyd tx concentratedliquidity create-position 6 $param 5000000000note,1000000uion 0 0 --from pools --chain-id=$CHAIN_ID --home $OSMOSIS_HOME --keyring-backend=test -b block --fees 5000note --gas 900000 --yes"
    done
}

if [[ ! -d $CONFIG_FOLDER ]]
then
    echo $MNEMONIC | symphonyd init -o --chain-id=$CHAIN_ID --home $OSMOSIS_HOME --recover $MONIKER
    install_prerequisites
    edit_genesis
    add_genesis_accounts
    edit_config
    enable_cors
fi

symphonyd start --home $OSMOSIS_HOME &

if [[ $STATE == 'true' ]]
then
    echo "Creating pools"
    
    echo "note / uusdc balancer"
    create_two_asset_pool "noteUusdcBalancerPool.json"
    
    echo "note / uion balancer"
    create_two_asset_pool "noteUionBalancerPool.json"
    
    echo "uweth / uusdc stableswap"
    create_stable_pool
    
    echo "uusdc / uion balancer"
    create_two_asset_pool "uusdcUionBalancerPool.json"

    echo "stake / uion / note balancer"
    create_three_asset_pool

    echo "uion / note concentrated"
    create_concentrated_pool
    create_concentrated_pool_positions
fi
wait
