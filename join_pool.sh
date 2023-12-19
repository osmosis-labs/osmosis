estimate_query=$(osmosisd tx gamm join-pool --pool-id 1 --max-amounts-in 1000000uosmo,1000000uion --share-amount-out 10000 --from osmo1cyyzpxplxdzkeea7kwsydadg87357qnahakaks --chain-id localosmosis --keyring-backend test --fees=10000uosmo --gas=1000000 --yes --dry-run)

echo $estimate_query

swap_tx_hash=$(osmosisd tx gamm join-pool --pool-id 1 --max-amounts-in 1000000uosmo,1000000uion --share-amount-out 10000 --from lo-test1 --chain-id localosmosis --keyring-backend test --fees=10000uosmo --gas=1000000 --yes --output=json | jq .txhash)

stripped_string=${swap_tx_hash//\"/}

echo $stripped_string

sleep 7

swap_gas_used=$(osmosisd q tx "$stripped_string" --output=json | jq .gas_used)

echo "swap gas used" $swap_gas_used