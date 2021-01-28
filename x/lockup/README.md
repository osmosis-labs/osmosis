# Commands

```sh
# 1 day 100stake lock-tokens command
osmosisd tx lockup lock-tokens 100stake --duration="86400000000000" --from=validator --chain-id=testing --keyring-backend=test --yes

# 5s 100stake lock-tokens command
osmosisd tx lockup lock-tokens 100stake --duration="5000000000" --from=validator --chain-id=testing --keyring-backend=test --yes

# unlock tokens
osmosisd tx lockup unlock-tokens --from=validator --chain-id=testing --keyring-backend=test --yes

# unlock specific period lock
osmosisd tx lockup unlock-by-id 1 --from=validator --chain-id=testing --keyring-backend=test --yes

# account balance
osmosisd query bank balances $(osmosisd keys show -a validator --keyring-backend=test)

# query module balance
osmosisd query lockup module-balance
```