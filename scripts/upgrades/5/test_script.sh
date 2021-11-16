# Create Pool to test JoinPool gas fee
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
osmosisd tx gamm create-pool --pool-file="./pool.json"  --gas=3000000 --from=validator --chain-id=testing --keyring-backend=test --yes --broadcast-mode=block

# Check that JoinPool gas fee and compare with JoinPool post-upgrade
osmosisd tx gamm join-pool --pool-id=1 --share-amount-out=5 --max-amounts-in=100stake


# Upgrade
osmosisd tx gov submit-proposal software-upgrade "v5" --title="v5 upgrade" --description="lv5 upgrade"  --from=validator --upgrade-height=20 --deposit=10000000stake --chain-id=testing --keyring-backend=test --yes  --broadcast-mode=block

# Test using asset that is not on whitelist as fees should fail
osmosisd tx lockup  lock-tokens 1uosmo --duration 1h --from=validator --keyring-backend=test --chain-id=testing --fees=1randtoken --yes


# Check that JoinPool gas fee has decreased post-upgrade
osmosisd tx gamm join-pool --pool-id=1 --share-amount-out=5 --max-amounts-in=100stake

