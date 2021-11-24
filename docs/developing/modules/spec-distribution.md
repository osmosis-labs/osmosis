# Distribution

::: warning Note:
Osmosis's Distribution module inherits from Cosmos SDK's [`distribution`](https://docs.cosmos.network/master/modules/distribution/) module. This document is a stub, and covers mainly important Osmosis-specific notes about how it is used.
:::

The `Distribution` module describes a mechanism that keeps track of collected fees and _passively_ distributes them to validators and delegators. In addition, the Distribution module also defines the [Community Pool](#community-pool), which are funds under the control of on-chain Governance.

## Concepts

### Validator & Delegator Rewards

::: warning IMPORTANT
Passive distribution means that validators and delegators will have to manually collect their fee rewards by submitting withdrawal transactions. Read up on how to do so with `osmosisd` [here](../osmosisd/distribution.md).
:::

Collected rewards are pooled globally and distrubuted to validators and delegators. Each validator has the opportunity to charge delegators commission on the rewards collected on behalf of the delegators. Fees are collected directly into a global reward pool and a validator proposer-reward pool. Due to the nature of passive accounting, whenever changes to parameters which affect the rate of reward distribution occur, withdrawal of rewards must also occur.

### Community Pool

The Community Pool is a reserve of tokens that is designated for funding projects that promote further adoption and stimulate growth for the Osmosis economy. The portion of seigniorage that is designated for ballot winners of the Exchange Rate Oracle is called the [Reward Weight](spec-treasury.md#reward-weight), a value governed by the Treasury. The rest of that seigniorage is all dedicated to the Community Pool.

::: warning Note:
As of Columbus-5, all seigniorage is burned, and the Community Pool no longer receives funding.
:::

## State

> This section was taken from the official Cosmos SDK docs, and placed here for your convenience to understand the Distribution module's parameters and genesis variables.

### FeePool

All globally tracked parameters for distribution are stored within
`FeePool`. Rewards are collected and added to the reward pool and
distributed to validators/delegators from here.

Note that the reward pool holds decimal coins (`DecCoins`) to allow
for fractions of coins to be received from operations like inflation.
When coins are distributed from the pool they are truncated back to
`sdk.Coins` which are non-decimal.

### Validator Distribution

Validator distribution information for the relevant validator is updated each time:

1.  delegation amount to a validator is updated,
2.  a validator successfully proposes a block and receives a reward,
3.  any delegator withdraws from a validator, or
4.  the validator withdraws it's commission.

### Delegation Distribution

Each delegation distribution only needs to record the height at which it last
withdrew fees. Because a delegation must withdraw fees each time it's
properties change (aka bonded tokens etc.) its properties will remain constant
and the delegator's _accumulation_ factor can be calculated passively knowing
only the height of the last withdrawal and its current properties.

## Message Types

### MsgSetWithdrawAddress

```go
type MsgSetWithdrawAddress struct {
	DelegatorAddress sdk.AccAddress `json:"delegator_address" yaml:"delegator_address"`
	WithdrawAddress  sdk.AccAddress `json:"withdraw_address" yaml:"withdraw_address"`
}
```


### MsgWithdrawDelegatorReward

```go
// msg struct for delegation withdraw from a single validator
type MsgWithdrawDelegatorReward struct {
	DelegatorAddress sdk.AccAddress `json:"delegator_address" yaml:"delegator_address"`
	ValidatorAddress sdk.ValAddress `json:"validator_address" yaml:"validator_address"`
}
```


### MsgWithdrawValidatorCommission

```go
type MsgWithdrawValidatorCommission struct {
	ValidatorAddress sdk.ValAddress `json:"validator_address" yaml:"validator_address"`
}
```


### MsgFundCommunityPool

```go
type MsgFundCommunityPool struct {
	Amount    sdk.Coins      `json:"amount" yaml:"amount"`
	Depositor sdk.AccAddress `json:"depositor" yaml:"depositor"`
}
```


## Proposals

### CommunityPoolSpendProposal

The Distribution module defines a special proposal that upon being passed, will disburse the coins specified in `Amount` to the `Recipient` account using funds from the Community Pool.

```go
type CommunityPoolSpendProposal struct {
	Title       string         `json:"title" yaml:"title"`
	Description string         `json:"description" yaml:"description"`
	Recipient   sdk.AccAddress `json:"recipient" yaml:"recipient"`
	Amount      sdk.Coins      `json:"amount" yaml:"amount"`
}
```

## Transitions

### Begin-Block

> This section derives from the official Cosmos SDK docs, and placed here for your convenience to understand the Distribution module's parameters.

At the beginning of the block, the Distribution module will set the proposer for determining distribution during endblock and distribute rewards for the previous block.

The fees received are transferred to the Distribution `ModuleAccount`, which tracks the flow of coins in and out of the module. Fees are also allocated to the proposer, community fund, and global pool:

- Proposer: When a validator is the proposer of a round, that validator and its delegators receive 1-5% of the fee rewards.
- Community fund: The reserve community tax is charged and distributed to the community pool. As of Columbus-5, this tax is no longer charged and the community pool no longer receives funding.
- Global pool: The remainder of the funds is allocated to the global pool, where they are distributed proportionally by voting power to all bonded validators independent of whether they voted. This allocation is called social distribution. Social distribution is applied to the proposer validator in addition to the proposer reward.

The proposer reward is calculated from pre-commits Tendermint messages in order to incentivize validators to wait and include additional pre-commits in the block. All provision rewards are added to a provision reward pool, which each validator holds individually (`ValidatorDistribution.ProvisionsRewardPool`).

```go
func AllocateTokens(feesCollected sdk.Coins, feePool FeePool, proposer ValidatorDistribution,
              sumPowerPrecommitValidators, totalBondedTokens, communityTax,
              proposerCommissionRate sdk.Dec)

     SendCoins(FeeCollectorAddr, DistributionModuleAccAddr, feesCollected)
     feesCollectedDec = MakeDecCoins(feesCollected)
     proposerReward = feesCollectedDec * (0.01 + 0.04
                       * sumPowerPrecommitValidators / totalBondedTokens)

     commission = proposerReward * proposerCommissionRate
     proposer.PoolCommission += commission
     proposer.Pool += proposerReward - commission

     communityFunding = feesCollectedDec * communityTax
     feePool.CommunityFund += communityFunding

     poolReceived = feesCollectedDec - proposerReward - communityFunding
     feePool.Pool += poolReceived

     SetValidatorDistribution(proposer)
     SetFeePool(feePool)
```

## Parameters

The subspace for the Distribution module is `distribution`.

```go
type GenesisState struct {
	...
	CommunityTax        sdk.Dec `json:"community_tax" yaml:"community_tax"`
	BaseProposerReward	sdk.Dec `json:"base_proposer_reward" yaml:"base_proposer_reward"`
	BonusProposerReward	sdk.Dec	`json:"bonus_proposer_reward" yaml:"bonus_proposer_reward"`
	WithdrawAddrEnabled bool 	`json:"withdraw_addr_enabled"`
	...
}
```

### CommunityTax

- type: `Dec`

### BaseProposerReward

- type: `Dec`

### BonusProposerReward

- type: `Dec`

### WithdrawAddrEnabled

- type: `bool`
