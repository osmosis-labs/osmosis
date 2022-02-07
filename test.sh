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
# - Panic error for adding more tokens on top of synthetic lockup and removing it
# - Add tokens to lockup should be instantly affected on superfluid staking
# - - Add tokens is only called by owner? Is there any case other user add tokens to lock or any other reward distributor can add tokens to this lockup?
# - Remove tokens from lockup should be instantly affected on superfluid staking
# - - Remove tokens source is only one by slashing?
# (I think it's not mandatory since it should be updated on intermediary account side - but in case other slash exists, could add hook for refreshing that lockup)
# - Should clarify last TWAP and current twap thoughts
# - Fix unit tests
# - add unit test for superfluid delegate, add more tokens, undelegate and remaining amount is zero for intermediary account
# - add cli command for checking all superfluid assets
# - add query for intermediary account connected to lockup by id
# - BugCLI: Delegation exists after superfluid undelegate 
# - BugCLI: Delegation exists after superfluid delegate, add more tokens, undelegate 
# Thoughts: - probably issue with staking query - if it involves unbonding queries as well
# - Check delegation amount changes when add more tokens to existing lock
# - Undelegate take effect correctly for delegation amount



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

osmosisd tx superfluid delegate 1 osmovaloper1njvl27h9xmt2v9hu568ryqk8xc2wean8y4c2sc --from=validator --keyring-backend=test --chain-id=testing --broadcast-mode=block --yes
osmosisd tx superfluid undelegate 1 --from=validator --keyring-backend=test --chain-id=testing  --broadcast-mode=block --yes
osmosisd tx superfluid redelegate 1 osmovaloper1njvl27h9xmt2v9hu568ryqk8xc2wean8y4c2sc --from=validator --keyring-backend=test --chain-id=testing  --broadcast-mode=block --yes
osmosisd tx lockup begin-unlock-by-id 1 --from=validator --keyring-backend=test --chain-id=testing --broadcast-mode=block --yes

osmosisd query staking delegation osmo1p7za50gky60ufad6mr90vwu50zqe02rl2hkn8c osmovaloper1njvl27h9xmt2v9hu568ryqk8xc2wean8y4c2sc

osmosisd query distribution rewards $(osmosisd keys show -a validator --keyring-backend=test)
osmosisd query incentives active-gauges 

osmosisd keys add acc2 --keyring-backend=test
osmosisd tx bank send validator $(osmosisd keys show -a acc2 --keyring-backend=test) 1000000stake,1000000valtoken,10000000000000000000gamm/pool/1 --keyring-backend=test --chain-id=testing  --broadcast-mode=block --yes
osmosisd tx lockup lock-tokens 10000000000000000000gamm/pool/1 --duration=504h --from=acc2 --keyring-backend=test --chain-id=testing --broadcast-mode=block --yes
osmosisd query lockup lock-by-id 2
osmosisd tx superfluid delegate 2 osmovaloper1sv8m28x9kjdavt8uqsv0x49kzmfzvyghtprg6m --from=acc2 --keyring-backend=test --chain-id=testing --broadcast-mode=block --yes

osmosisd query superfluid all-superfluid-assets
osmosisd query superfluid connected-intermediary-account 1