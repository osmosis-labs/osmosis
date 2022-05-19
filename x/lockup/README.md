Commands
========

``` {.sh}
# 1 day 100stake lock-tokens command
osmosisd tx lockup lock-tokens 200stake --duration="86400s" --from=validator --chain-id=testing --keyring-backend=test --yes

# 5s 100stake lock-tokens command
osmosisd tx lockup lock-tokens 100stake --duration="5s" --from=validator --chain-id=testing --keyring-backend=test --yes

# begin unlock tokens, NOTE: add more gas when unlocking more than two locks in a same command
osmosisd tx lockup begin-unlock-tokens --from=validator --gas=500000 --chain-id=testing --keyring-backend=test --yes

# unlock tokens, NOTE: add more gas when unlocking more than two locks in a same command
osmosisd tx lockup unlock-tokens --from=validator --gas=500000 --chain-id=testing --keyring-backend=test --yes

# unlock specific period lock
osmosisd tx lockup unlock-by-id 1 --from=validator --chain-id=testing --keyring-backend=test --yes

# account balance
osmosisd query bank balances $(osmosisd keys show -a validator --keyring-backend=test)

# query module balance
osmosisd query lockup module-balance

# query locked amount
osmosisd query lockup module-locked-amount

# query lock by id
osmosisd query lockup lock-by-id 1

# query account unlockable coins
osmosisd query lockup account-unlockable-coins $(osmosisd keys show -a validator --keyring-backend=test)

# query account locks by denom past time
osmosisd query lockup account-locked-pasttime-denom $(osmosisd keys show -a validator --keyring-backend=test) 1611879610 stake

# query account locks past time
osmosisd query lockup account-locked-pasttime $(osmosisd keys show -a validator --keyring-backend=test) 1611879610

# query account locks by denom with longer duration
osmosisd query lockup account-locked-longer-duration-denom $(osmosisd keys show -a validator --keyring-backend=test) 5.1s stake

# query account locks with longer duration
osmosisd query lockup account-locked-longer-duration $(osmosisd keys show -a validator --keyring-backend=test) 5.1s

# query account locked coins
osmosisd query lockup account-locked-coins $(osmosisd keys show -a validator --keyring-backend=test)

# query account locks before time
osmosisd query lockup account-locked-beforetime $(osmosisd keys show -a validator --keyring-backend=test) 1611879610
```
