# `claim`

## Abstract

This specifies the `claim` module of Osmosis project, provide commands for claimable amount query and claim airdrop.
We apply real-time decay after `DurationUntilDecay` pass where monthly decay rate is `-10%` of inital airdrop amount.
When `DurationUntilDecay + DurationOfDecay` time passes, all unclaimed coins will be sent to the community pool.

## Genesis State

### Accounts

All genesis accounts have `1 Osmo` for claim fee.

### Claimables

Claimables are configured by genesis.

### Airdrop Tools
There are tools to generate genesis from cosmos-hub snapshot.

#### Genesis generation

Generate genesis from cosmos-hub snapshot genesis and output snapshot of atom, osmo balance and percentage by address.
```sh
osmosisd export-airdrop-genesis uatom ../genesis.json --total-amount=100000000000000 --snapshot-output="../snapshot.json"
osmosisd export-airdrop-genesis uatom ../genesis.json --snapshot-output="../snapshot.json"
```

## Queries

Query claimable amount for a given address at the current time.
```sh
osmosisd query claim claimable $(osmosisd keys show -a validator --keyring-backend=test)
```

## Msgs

### (WIP) Actual claim commands will change

Claim full airdrop amount from `claim` module.
```sh
osmosisd tx claim claimable --from validator --keyring-backend=test --chain-id=testing --yes
```
