# `claim`

## Abstract

This specifies the `claim` module of Osmosis project, provide commands for claimable amount query and claim airdrop.
We apply real-time decay after `DurationUntilDecay` pass where monthly decay rate is `-10%` of inital airdrop amount.
When `DurationUntilDecay + DurationOfDecay` time passes, all unclaimed coins will be sent to the community pool.

## Genesis State

### Accounts

All genesis accounts have `1 Osmo` for claim fee.

### Claimables

Claimables are the maximum claimable amounts per address and it's configured by genesis.
It's determined by applying few rules to the snapshot of ATOM balance at cosmoshub-3.

### User actions
User need to accomplish `DelegateStake`, `Vote`, `AddLiquidity`, `Swap` actions to completely withdraw his claimables.
Each action is defined in enum in implementation
```
	ActionAddLiquidity  Action = 0
	ActionSwap          Action = 1
	ActionVote          Action = 2
	ActionDelegateStake Action = 3
```
All of these actions are monitored by registring claim **hooks** to governance, staking, gamm, lockup modules hooks.

### User withdrawn actions
A user can't withdraw twice for single action type, example `Vote`. To withdraw all, they need to do every action.
To do this, claim module manage withdrawn actions flag. It's using same strategy as user actions, but only count the actions that user has already withdrawn.
If Alice has done `DelegateStake` action, withdraw 25% of claimables, and then if she do `Vote` action, she is able to withdraw another 25% of claimables. Here, withdrawn actions flag is used to avoid double withdraw on `DelegateStake` action.

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
