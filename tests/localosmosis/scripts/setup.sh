#!/bin/sh

CHAIN_ID=localosmosis
OSMOSIS_HOME=$HOME/.osmosisd
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
    dasel put -t string -f $GENESIS '.app_state.staking.params.bond_denom' -v 'uosmo'
    dasel put -t string -f $GENESIS '.app_state.staking.params.unbonding_time' -v '240s'

    # Update bank module
    dasel put -t string -f $GENESIS '.app_state.bank.denom_metadata.[].description' -v 'Registered denom uion for localosmosis testing'
    dasel put -t string -f $GENESIS '.app_state.bank.denom_metadata.[0].denom_units.[].denom' -v 'uion'
    dasel put -t string -f $GENESIS '.app_state.bank.denom_metadata.[0].denom_units.[0].exponent' -v 0
    dasel put -t string -f $GENESIS '.app_state.bank.denom_metadata.[0].base' -v 'uion'
    dasel put -t string -f $GENESIS '.app_state.bank.denom_metadata.[0].display' -v 'uion'
    dasel put -t string -f $GENESIS '.app_state.bank.denom_metadata.[0].name' -v 'uion'
    dasel put -t string -f $GENESIS '.app_state.bank.denom_metadata.[0].symbol' -v 'uion'

    dasel put -t string -f $GENESIS '.app_state.bank.denom_metadata.[].description' -v 'Registered denom uosmo for localosmosis testing'
    dasel put -t string -f $GENESIS '.app_state.bank.denom_metadata.[1].denom_units.[].denom' -v 'uosmo'
    dasel put -t string -f $GENESIS '.app_state.bank.denom_metadata.[1].denom_units.[0].exponent' -v 0
    dasel put -t string -f $GENESIS '.app_state.bank.denom_metadata.[1].base' -v 'uosmo'
    dasel put -t string -f $GENESIS '.app_state.bank.denom_metadata.[1].display' -v 'uosmo'
    dasel put -t string -f $GENESIS '.app_state.bank.denom_metadata.[1].name' -v 'uosmo'
    dasel put -t string -f $GENESIS '.app_state.bank.denom_metadata.[1].symbol' -v 'uosmo'

    # Update crisis module
    dasel put -t string -f $GENESIS '.app_state.crisis.constant_fee.denom' -v 'uosmo'

    # Update gov module
    dasel put -t string -f $GENESIS '.app_state.gov.voting_params.voting_period' -v '60s'
    dasel put -t string -f $GENESIS '.app_state.gov.params.voting_period' -v '60s'
    dasel put -t string -f $GENESIS '.app_state.gov.params.expedited_voting_period' -v '30s'
    dasel put -t string -f $GENESIS '.app_state.gov.params.min_deposit.[0].denom' -v 'uosmo'

    # Update epochs module
    dasel put -t string -f $GENESIS '.app_state.epochs.epochs.[1].duration' -v "60s"

    # Update poolincentives module
    dasel put -t string -f $GENESIS '.app_state.poolincentives.lockable_durations.[0]' -v "120s"
    dasel put -t string -f $GENESIS '.app_state.poolincentives.lockable_durations.[1]' -v "180s"
    dasel put -t string -f $GENESIS '.app_state.poolincentives.lockable_durations.[2]' -v "240s"
    dasel put -t string -f $GENESIS '.app_state.poolincentives.params.minted_denom' -v "uosmo"

    # Update incentives module
    dasel put -t string -f $GENESIS '.app_state.incentives.lockable_durations.[0]' -v "1s"
    dasel put -t string -f $GENESIS '.app_state.incentives.lockable_durations.[1]' -v "120s"
    dasel put -t string -f $GENESIS '.app_state.incentives.lockable_durations.[2]' -v "180s"
    dasel put -t string -f $GENESIS '.app_state.incentives.lockable_durations.[3]' -v "240s"
    dasel put -t string -f $GENESIS '.app_state.incentives.params.distr_epoch_identifier' -v "hour"

    # Update mint module
    dasel put -t string -f $GENESIS '.app_state.mint.params.mint_denom' -v "uosmo"
    dasel put -t string -f $GENESIS '.app_state.mint.params.epoch_identifier' -v "hour"

    # Update poolmanager module
    dasel put -t string -f $GENESIS '.app_state.poolmanager.params.pool_creation_fee.[0].denom' -v "uosmo"

    # Update txfee basedenom
    dasel put -t string -f $GENESIS '.app_state.txfees.basedenom' -v "uosmo"

    # Update wasm permission (Nobody or Everybody)
    dasel put -t string -f $GENESIS '.app_state.wasm.params.code_upload_access.permission' -v "Everybody"

    # Update concentrated-liquidity (enable pool creation)
    dasel put -t bool -f $GENESIS '.app_state.concentratedliquidity.params.is_permissionless_pool_creation_enabled' -v true
}

add_genesis_accounts () {

    osmosisd add-genesis-account osmo12smx2wdlyttvyzvzg54y2vnqwq2qjateuf7thj 100000000000uosmo,100000000000uion,100000000000stake,100000000000uusdc,100000000000uweth --home $OSMOSIS_HOME
    # note such large amounts are set for e2e tests on FE 
    osmosisd add-genesis-account osmo1cyyzpxplxdzkeea7kwsydadg87357qnahakaks 9999999999999999999999999999999999999999999999999uosmo,9999999999999999999999999999999999999999999999999uion,100000000000stake,100000000000uusdc,100000000000uweth --home $OSMOSIS_HOME
    osmosisd add-genesis-account osmo18s5lynnmx37hq4wlrw9gdn68sg2uxp5rgk26vv 100000000000uosmo,100000000000uion,100000000000stake,100000000000uusdc,100000000000uweth --home $OSMOSIS_HOME
    osmosisd add-genesis-account osmo1qwexv7c6sm95lwhzn9027vyu2ccneaqad4w8ka 100000000000uosmo,100000000000uion,100000000000stake,100000000000uusdc,100000000000uweth --home $OSMOSIS_HOME
    osmosisd add-genesis-account osmo14hcxlnwlqtq75ttaxf674vk6mafspg8xwgnn53 100000000000uosmo,100000000000uion,100000000000stake,100000000000uusdc,100000000000uweth --home $OSMOSIS_HOME
    osmosisd add-genesis-account osmo12rr534cer5c0vj53eq4y32lcwguyy7nndt0u2t 100000000000uosmo,100000000000uion,100000000000stake,100000000000uusdc,100000000000uweth --home $OSMOSIS_HOME
    osmosisd add-genesis-account osmo1nt33cjd5auzh36syym6azgc8tve0jlvklnq7jq 100000000000uosmo,100000000000uion,100000000000stake,100000000000uusdc,100000000000uweth --home $OSMOSIS_HOME
    osmosisd add-genesis-account osmo10qfrpash5g2vk3hppvu45x0g860czur8ff5yx0 100000000000uosmo,100000000000uion,100000000000stake,100000000000uusdc,100000000000uweth --home $OSMOSIS_HOME
    osmosisd add-genesis-account osmo1f4tvsdukfwh6s9swrc24gkuz23tp8pd3e9r5fa 100000000000uosmo,100000000000uion,100000000000stake,100000000000uusdc,100000000000uweth --home $OSMOSIS_HOME
    osmosisd add-genesis-account osmo1myv43sqgnj5sm4zl98ftl45af9cfzk7nhjxjqh 100000000000uosmo,100000000000uion,100000000000stake,100000000000uusdc,100000000000uweth --home $OSMOSIS_HOME
    osmosisd add-genesis-account osmo14gs9zqh8m49yy9kscjqu9h72exyf295afg6kgk 100000000000uosmo,100000000000uion,100000000000stake,100000000000uusdc,100000000000uweth --home $OSMOSIS_HOME
    osmosisd add-genesis-account osmo1jllfytsz4dryxhz5tl7u73v29exsf80vz52ucc 1000000000000uosmo,1000000000000uion,1000000000000stake,1000000000000uusdc,1000000000000uweth --home $OSMOSIS_HOME

    echo $MNEMONIC | osmosisd keys add $MONIKER --recover --keyring-backend=test --home $OSMOSIS_HOME
    echo $POOLSMNEMONIC | osmosisd keys add pools --recover --keyring-backend=test --home $OSMOSIS_HOME
    osmosisd gentx $MONIKER 500000000uosmo --keyring-backend=test --chain-id=$CHAIN_ID --home $OSMOSIS_HOME

    osmosisd collect-gentxs --home $OSMOSIS_HOME
}

edit_config () {

    # Remove seeds
    dasel put -t string -f $CONFIG_FOLDER/config.toml '.p2p.seeds' -v ''

    # Expose the rpc
    dasel put -t string -f $CONFIG_FOLDER/config.toml '.rpc.laddr' -v "tcp://0.0.0.0:26657"
    
    # Expose pprof for debugging
    # To make the change enabled locally, make sure to add 'EXPOSE 6060' to the root Dockerfile
    # and rebuild the image.
    dasel put -t string -f $CONFIG_FOLDER/config.toml '.rpc.pprof_laddr' -v "0.0.0.0:6060"
}

enable_cors () {

    # Enable cors on RPC
    dasel put -t string -f $CONFIG_FOLDER/config.toml -v "*" '.rpc.cors_allowed_origins.[]'
    dasel put -t string -f $CONFIG_FOLDER/config.toml -v "Accept-Encoding" '.rpc.cors_allowed_headers.[]'
    dasel put -t string -f $CONFIG_FOLDER/config.toml -v "DELETE" '.rpc.cors_allowed_methods.[]'
    dasel put -t string -f $CONFIG_FOLDER/config.toml -v "OPTIONS" '.rpc.cors_allowed_methods.[]'
    dasel put -t string -f $CONFIG_FOLDER/config.toml -v "PATCH" '.rpc.cors_allowed_methods.[]'
    dasel put -t string -f $CONFIG_FOLDER/config.toml -v "PUT" '.rpc.cors_allowed_methods.[]'

    # Enable unsafe cors and swagger on the api
    dasel put -t bool -f $CONFIG_FOLDER/app.toml -v "true" '.api.swagger'
    dasel put -t bool -f $CONFIG_FOLDER/app.toml -v "true" '.api.enabled-unsafe-cors'

    # Enable cors on gRPC Web
    dasel put -t bool -f $CONFIG_FOLDER/app.toml -v "true" '.grpc-web.enable-unsafe-cors'

    # Enable SQS & route caching
    dasel put -t string -f $CONFIG_FOLDER/app.toml -v "true" '.osmosis-sqs.is-enabled'
    dasel put -t string -f $CONFIG_FOLDER/app.toml -v "true" '.osmosis-sqs.route-cache-enabled'

    dasel put -t string -f $CONFIG_FOLDER/app.toml -v "redis" '.osmosis-sqs.db-host'
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
  run_with_retries "osmosisd tx gamm create-pool --pool-file=$1 --from pools --chain-id=$CHAIN_ID --home $OSMOSIS_HOME --keyring-backend=test -b block --fees 5000uosmo --yes" "create two asset pool: successful"
}

create_stable_pool() {
  run_with_retries "osmosisd tx gamm create-pool --pool-file=uwethUusdcStablePool.json --pool-type=stableswap --from pools --chain-id=$CHAIN_ID --home $OSMOSIS_HOME --keyring-backend=test -b block --fees 5000uosmo --yes" "create two asset pool: successful"
}

create_three_asset_pool() {
  run_with_retries "osmosisd tx gamm create-pool --pool-file=nativeDenomThreeAssetPool.json --from pools --chain-id=$CHAIN_ID --home $OSMOSIS_HOME --keyring-backend=test -b block --fees 5000uosmo --gas 900000 --yes" "create three asset pool: successful"
}

create_concentrated_pool() {
  run_with_retries "osmosisd tx concentratedliquidity create-pool uion uosmo 1 \"0.0005\" --from pools --chain-id=$CHAIN_ID --home $OSMOSIS_HOME --keyring-backend=test -b block --fees 5000uosmo --gas 900000 --yes" "create concentrated pool: successful"
}

create_concentrated_pool_positions () {
    # Define an array to hold the parameters that change for each command
    set "[-1620000] 3420000" "305450 315000" "315000 322500" "300000 309990" "[-108000000] 342000000" "[-108000000] 342000000"

    substring='code: 0'
    COUNTER=0
    # Loop through each set of parameters in the array
    for param in "$@"; do
        run_with_retries "osmosisd tx concentratedliquidity create-position 6 $param 5000000000uosmo,1000000uion 0 0 --from pools --chain-id=$CHAIN_ID --home $OSMOSIS_HOME --keyring-backend=test -b block --fees 5000uosmo --gas 900000 --yes"
    done
}

if [[ ! -d $CONFIG_FOLDER ]]
then
    echo $MNEMONIC | osmosisd init -o --chain-id=$CHAIN_ID --home $OSMOSIS_HOME --recover $MONIKER
    install_prerequisites
    edit_genesis
    add_genesis_accounts
    edit_config
    enable_cors
fi

osmosisd start --home $OSMOSIS_HOME &

if [[ $STATE == 'true' ]]
then
    echo "Creating pools"

    echo "uosmo / uusdc balancer"
    create_two_asset_pool "uosmoUusdcBalancerPool.json"

    echo "uosmo / uion balancer"
    create_two_asset_pool "uosmoUionBalancerPool.json"

    echo "uweth / uusdc stableswap"
    create_stable_pool

    echo "uusdc / uion balancer"
    create_two_asset_pool "uusdcUionBalancerPool.json"

    echo "stake / uion / uosmo balancer"
    create_three_asset_pool

    echo "uion / uosmo concentrated"
    create_concentrated_pool
    create_concentrated_pool_positions
fi
wait
