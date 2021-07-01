# run old binary on terminal1
go clean --modcache
git stash
git checkout v1.0.1
go install ./cmd/osmosisd/
Run the below commands
```
    #!/bin/bash
    rm -rf $HOME/.osmosisd/
    cd $HOME
    osmosisd init --chain-id=testing testing --home=$HOME/.osmosisd
    osmosisd keys add validator --keyring-backend=test --home=$HOME/.osmosisd
    osmosisd add-genesis-account $(osmosisd keys show validator -a --keyring-backend=test --home=$HOME/.osmosisd) 1000000000stake,1000000000valtoken --home=$HOME/.osmosisd
    osmosisd gentx validator 500000000stake --keyring-backend=test --home=$HOME/.osmosisd --chain-id=testing
    osmosisd gentx validator 500000000stake --commission-rate="0.0" --keyring-backend=test --home=$HOME/.osmosisd --chain-id=testing
    osmosisd collect-gentxs --home=$HOME/.osmosisd
    
    cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["gov"]["voting_params"]["voting_period"]="120s"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
    cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["staking"]["params"]["min_commission_rate"]="0.050000000000000000"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json

```

Create pool.json 
```
{
  "weights": "1stake,1valtoken",
  "initial-deposit": "100stake,20valtoken",
  "swap-fee": "0.01",
  "exit-fee": "0.01",
  "future-governor": "168h"
}
```

rm $HOME/.osmosisd/cosmovisor/current -rf
cosmovisor start

# operations on terminal2
osmosisd tx lockup lock-tokens 100stake --duration="5s" --from=validator --chain-id=testing --keyring-backend=test --yes
sleep 7
osmosisd tx gov submit-proposal software-upgrade upgrade-lockup-module-store-management --title="lockup module upgrade" --description="lockup module upgrade for gas efficiency"  --from=validator --upgrade-height=10 --deposit=10000000stake --chain-id=testing --keyring-backend=test -y
sleep 7
osmosisd tx gov vote 1 yes --from=validator --keyring-backend=test --chain-id=testing --yes
sleep 7
osmosisd tx gamm create-pool --pool-file="./pool.json"  --gas=3000000 --from=validator --chain-id=testing --keyring-backend=test --yes --broadcast-mode=block
sleep 7
osmosisd tx lockup lock-tokens 1000stake --duration="100s" --from=validator --chain-id=testing --keyring-backend=test --yes --broadcast-mode=block
sleep 7
osmosisd tx lockup lock-tokens 2000stake --duration="200s" --from=validator --chain-id=testing --keyring-backend=test --yes --broadcast-mode=block
sleep 7
osmosisd tx lockup lock-tokens 3000stake --duration="1s" --from=validator --chain-id=testing --keyring-backend=test --yes --broadcast-mode=block
sleep 7
osmosisd tx lockup begin-unlock-by-id 1 --from=validator --chain-id=testing --keyring-backend=test --yes --broadcast-mode=block
sleep 7
osmosisd tx lockup begin-unlock-by-id 3 --from=validator --chain-id=testing --keyring-backend=test --yes --broadcast-mode=block
sleep 7
osmosisd tx gov submit-proposal software-upgrade "v2" --title="lockup module upgrade" --description="lockup module upgrade for gas efficiency"  --from=validator --upgrade-height=20 --deposit=10000000stake --chain-id=testing --keyring-backend=test --yes  --broadcast-mode=block
sleep 7
osmosisd tx gov vote 1 yes --from=validator --keyring-backend=test --chain-id=testing --yes --broadcast-mode=block
osmosisd query gov proposal 1
osmosisd query upgrade plan
osmosisd query lockup account-locked-longer-duration $(osmosisd keys show -a --keyring-backend=test validator) 1s
osmosisd query gamm pools
osmosisd query staking validators
osmosisd query staking params

# on terminal1
Wait until consensus failure happen and stop binary using Ctrl + C
git checkout lockup_module_genesis_export
git checkout main

Update go mod file to use latest SDK changes: /Users/admin/go/pkg/mod/github.com/osmosis-labs/cosmos-sdk@v0.42.5-0.20210622202917-f4f6a08ac64b
go get github.com/osmosis-labs/cosmos-sdk@ea1ec79c739ba39639b9a24f824127ecc6650887
go: downloading github.com/osmosis-labs/cosmos-sdk v0.42.5-0.20210630100106-ea1ec79c739b
Upgrade Osmosis Cosmos SDK version to `v0.42.5-0.20210630100106-ea1ec79c739b`
go mod download github.com/cosmos/cosmos-sdk
git stash
git checkout min_commission_change_validation_change_ignore
go install ./cmd/osmosisd/
osmosisd start --home=$HOME/.osmosisd

# check on terminal2
osmosisd query lockup account-locked-longer-duration $(osmosisd keys show -a --keyring-backend=test validator) 1s
osmosisd query lockup account-locked-longer-duration $(osmosisd keys show -a --keyring-backend=test validator) 1s
osmosisd query lockup module-locked-amount
osmosisd query gamm pools
osmosisd query staking validators
osmosisd query staking params
osmosisd query bank balances $(osmosisd keys show -a --keyring-backend=test validator)
osmosisd tx staking edit-validator --commission-rate="0.1"  --from=validator --chain-id=testing --keyring-backend=test --yes --broadcast-mode=block
osmosisd tx staking edit-validator --commission-rate="0.08"  --from=validator --chain-id=testing --keyring-backend=test --yes --broadcast-mode=block

Result:
- pool exists
- lockup processed all correctly
- validator commission rate worked
- chain did not panic 