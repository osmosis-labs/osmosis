create_pool_tx_hash=$(osmosisd tx gamm create-pool --pool-file=./tests/localosmosis/scripts/uosmoUionBalancerPool.json --from lo-test1 --chain-id=localosmosis --home $HOME/.osmosisd-local --keyring-backend test --fees=10000uosmo --gas=1000000 --output=json --yes | jq .txhash)

stripped_string=${create_pool_tx_hash//\"/}

echo $stripped_string

sleep 7

osmosisd q tx "$stripped_string"