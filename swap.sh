estimate_query=$(osmosisd tx poolmanager swap-exact-amount-in 20000000uosmo 1 --swap-route-pool-ids 1 --swap-route-denoms uion --from osmo1cyyzpxplxdzkeea7kwsydadg87357qnahakaks --keyring-backend test --chain-id=localosmosis --fees 10000uosmo --output=json --yes --dry-run)

echo $estimate_query

swap_tx_hash=$(osmosisd tx poolmanager swap-exact-amount-in 20000000uosmo 1 --swap-route-pool-ids 1 --swap-route-denoms uion --from lo-test1 --keyring-backend test --chain-id=localosmosis --fees 10000uosmo --output=json --yes | jq .txhash)

stripped_string=${swap_tx_hash//\"/}

echo $stripped_string

sleep 7

swap_gas_used=$(osmosisd q tx "$stripped_string" --output=json | jq .gas_used)

echo "swap gas used" $swap_gas_used