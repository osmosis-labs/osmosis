touch stake-valtoken.json
nano stake-valtoken.json 
```
{
	"weights": "5stake,5valtoken",
	"initial-deposit": "1000000stake,1000000valtoken",
	"swap-fee": "0.01",
	"exit-fee": "0.01",
	"future-governor": "168h"
}
```

# setup correct genesis for epochs
# - should try delegation
# - should check gauge is created correctly
# - should check rewards are moving correctly for staking rewards to gauges
# - should check twap changes per epoch for swap operation(min)
# - should check delegation amount is changing per epoch
# - should add query for getting intermediary account address and further intermediary account info and check delegation
# - should try undelegation
# - check adding one more lock is not creating a new gauge
# - check how removing superfluid asset is affecting existing superfluid delegations and new delegation
# - Check the case when superfluid asset is enabled again

# - Should check the case - undelegating when superfluid asset is removed

# - Delegation exists for a while even though asset-twap is zero already - probably 1 epoch delay I think
# - Delegation not found for a while after enabling again instantly - probably 1 epoch delay I think - check the timing between TWAP and delegation
# - Coins on current epoch is released on next epoch - which should be instantly claimed at gauge
# - Error for adding more tokens + synthetic lockup combination

# - should try redelegation
# - should try producing slashing on multi-nodes (is there a way to do this on single node?)


osmosisd tx gamm create-pool --pool-file=./stake-valtoken.json --from=validator --keyring-backend=test --chain-id=testing --broadcast-mode=block --yes
osmosisd tx gamm swap-exact-amount-in 100000valtoken 50000 --swap-route-pool-ids=1 --swap-route-denoms=stake --from=validator --keyring-backend=test --chain-id=testing --broadcast-mode=block --yes
osmosisd query gamm pool 1

osmosisd query bank balances $(osmosisd keys show -a validator --keyring-backend=test)
osmosisd tx lockup lock-tokens 10000000000000000000gamm/pool/1 --duration=504h --from=validator --keyring-backend=test --chain-id=testing --broadcast-mode=block --yes
osmosisd query lockup  lock-by-id 1
osmosisd query staking validators

osmosisd tx gov submit-proposal set-superfluid-assets-proposal --title="set superfluid assets" --description="set superfluid assets description" --superfluid-assets="gamm/pool/1" --deposit=10000000stake --from=validator --chain-id=testing --keyring-backend=test --broadcast-mode=block --yes
osmosisd tx gov vote 1 yes --from=validator --keyring-backend=test --chain-id=testing --broadcast-mode=block --yes

osmosisd tx gov submit-proposal remove-superfluid-assets-proposal --title="remove superfluid assets" --description="remove superfluid assets description" --superfluid-assets="gamm/pool/1" --deposit=10000000stake --from=validator --chain-id=testing --keyring-backend=test --broadcast-mode=block --yes
osmosisd tx gov vote 2 yes --from=validator --keyring-backend=test --chain-id=testing --broadcast-mode=block --yes

osmosisd query gov proposals

osmosisd query superfluid asset-twap gamm/pool/1
osmosisd query superfluid all-intermediary-accounts

osmosisd tx superfluid delegate 1 osmovaloper1sv8m28x9kjdavt8uqsv0x49kzmfzvyghtprg6m --from=validator --keyring-backend=test --chain-id=testing --broadcast-mode=block --yes
osmosisd tx superfluid undelegate 1 --from=validator --keyring-backend=test --chain-id=testing  --broadcast-mode=block --yes
osmosisd tx superfluid redelegate 1 osmovaloper1sv8m28x9kjdavt8uqsv0x49kzmfzvyghtprg6m --from=validator --keyring-backend=test --chain-id=testing  --broadcast-mode=block --yes

osmosisd query staking delegation osmo1x6kg843vfzc4xmr2awg32fu4d8yq5hhy02n53r osmovaloper1sv8m28x9kjdavt8uqsv0x49kzmfzvyghtprg6m
136154.000000000000000000

osmosisd query distribution rewards $(osmosisd keys show -a validator --keyring-backend=test)
osmosisd query incentives active-gauges 

osmosisd keys add acc2 --keyring-backend=test
osmosisd tx bank send validator $(osmosisd keys show -a acc2 --keyring-backend=test) 1000000stake,1000000valtoken,10000000000000000000gamm/pool/1 --keyring-backend=test --chain-id=testing  --broadcast-mode=block --yes
osmosisd tx lockup lock-tokens 10000000000000000000gamm/pool/1 --duration=504h --from=acc2 --keyring-backend=test --chain-id=testing --broadcast-mode=block --yes
osmosisd query lockup lock-by-id 2
osmosisd tx superfluid delegate 2 osmovaloper1sv8m28x9kjdavt8uqsv0x49kzmfzvyghtprg6m --from=acc2 --keyring-backend=test --chain-id=testing --broadcast-mode=block --yes
