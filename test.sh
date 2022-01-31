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
# - Remove tokens from lockup should be instantly affected on superfluid staking
# (I think it's not mandatory since it should be updated on intermediary account side - but in case other slash exists, could add hook for refreshing that lockup)
# - Should clarify last TWAP and current twap thoughts
# - Bug: Delegation exists after superfluid delegate, add more tokens, undelegate

# - Check delegation amount changes when add more tokens to existing lock
# - Check the case adding more tokens to locks after starting redelegation or undelegation
# - Undelegate take effect correctly for delegation amount?
# - Add backwards testing for new changes
# - Fix unit tests
# - Write unit test scenario, user test scenario
# - Revert changes for go.mod, go.sum, .json, .sh files
# - Invariant: Total superfluid delegation amount by intermediary account should be same as superfluid delegated lockups' sum
# - simulation test should increase coverage for all of these manual testings to avoid unexpected issues

# removeTokensFromLock could be called for synthetic lockup?
# addTokensToLock could be called for synthetic lockup?

# Think of synthetic lockup accumulation store and synthetic lockup itself lifecycle
# - Could there be negative or more than the balance made during this cycle?
# 1) Add more tokens to lock
# 2) Remove tokens from lock
# 3) Redelegation
# 4) Undelegation
# 5) Create new synthetic lockup
# 6) Delete synthetic lockup

# - Delegation exists for a while even though asset-twap is zero already - probably 1 epoch delay I think
# - Delegation not found for a while after enabling again instantly - probably 1 epoch delay I think - check the timing between TWAP and delegation
# - Coins on current epoch is released on next epoch - which should be instantly claimed at gauge

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

osmosisd tx superfluid delegate 1 osmovaloper1l7tnl5fgcw4aad2lsl3hcgteg5fyl29mx63wcn --from=validator --keyring-backend=test --chain-id=testing --broadcast-mode=block --yes
osmosisd tx superfluid undelegate 1 --from=validator --keyring-backend=test --chain-id=testing  --broadcast-mode=block --yes
osmosisd tx superfluid redelegate 1 osmovaloper1l7tnl5fgcw4aad2lsl3hcgteg5fyl29mx63wcn --from=validator --keyring-backend=test --chain-id=testing  --broadcast-mode=block --yes

osmosisd query staking delegation osmo1098s7vzm9x0uwz8vpssp9zrz80kzd67kqhtpfq osmovaloper1l7tnl5fgcw4aad2lsl3hcgteg5fyl29mx63wcn

osmosisd query distribution rewards $(osmosisd keys show -a validator --keyring-backend=test)
osmosisd query incentives active-gauges 

osmosisd keys add acc2 --keyring-backend=test
osmosisd tx bank send validator $(osmosisd keys show -a acc2 --keyring-backend=test) 1000000stake,1000000valtoken,10000000000000000000gamm/pool/1 --keyring-backend=test --chain-id=testing  --broadcast-mode=block --yes
osmosisd tx lockup lock-tokens 10000000000000000000gamm/pool/1 --duration=504h --from=acc2 --keyring-backend=test --chain-id=testing --broadcast-mode=block --yes
osmosisd query lockup lock-by-id 2
osmosisd tx superfluid delegate 2 osmovaloper1sv8m28x9kjdavt8uqsv0x49kzmfzvyghtprg6m --from=acc2 --keyring-backend=test --chain-id=testing --broadcast-mode=block --yes
