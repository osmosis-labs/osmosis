#!/bin/bash

osmosisd tx gov submit-proposal software-upgrade v8 --title "v8 upgrade" --description "v8 upgrade proposal" --upgrade-height 75 --upgrade-info "" --chain-id $1 --from val -b block --yes --keyring-backend test
osmosisd tx gov deposit 1 10000000stake --from val --chain-id $1 -b block --yes --keyring-backend test
osmosisd tx gov vote 1 yes --from val --chain-id $1 -b block --yes --keyring-backend test