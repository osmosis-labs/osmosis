# Superfluid Staking

## Abstract

Superfluid Staking provides the consensus layer more security with a
sort of "Proof of Useful Stake". Each person gets an amount of Osmo
representative of the value of their share of liquidity pool tokens
staked and delegated to validators, resulting in the security guarantee
of the consensus layer to also be based on GAMM LP shares. The OSMO
token is minted and burned in the context of Superfluid Staking.
Throughout all of this, OSMO's supply is preserved in queries to the
bank module.

### The process

All of the below methods are found under the [Superfluid
modules](https://github.com/osmosis-labs/osmosis/tree/main/x/superfluid).

- The `SuperfluidDelegate` method stores your share of bonded
  liquidity pool tokens, with `validateLock` as a verifier for lockup
  time.
- `GetSuperfluidOsmo` mints OSMO tokens each day for delegation as a
  representative of the value of your pool share. This amount is
  minted because the staking module at the moment requires staked
  tokens to be in OSMO. This amount is burned each day and re-minted
  to keep the representative amount of the value of your pool share
  accurate. The lockup duration is guaranteed from the underlying
  lockup module.
- `GetExpectedDelegationAmount` iterates over each (denom, delegate)
  pair and checks for how much OSMO we have delegated. The difference
  from the current balance to what is expected is burned / minted to
  match with the expected.
- A `messageServer` method executes the Superfluid delegate message.
- `syntheticLockup` is used to index bond holders and tracking their
  addresses for reward distribution or potentially slashing purposes.
  These track whether if your Superfluid stake is currently bonding or
  unbonding.
- An `IntermediaryAccount` is mostly used for the actual reward
  distribution or slashing events, and are responsible for
  establishing the connection between each superfluid staked lock and
  their delegation to the validator. These work by transferring the
  superfluid OSMO to their respective delegators. Rewards are linearly
  scaled based on how much you have locked for a given (validator,
  denom) pair. Rewards are first moved to the incentive gauges, then
  distributed from the gauges. In this way, we're using the existing
  gauge reward system for paying out superfluid staking rewards and
  tracking the amount you have superfluidly staked using the lockup
  module.
- Rewards are distributed per epoch, which is currently a day.
  `abci.go` checks whether or not the current block is at the
  beginning of the epoch using `BeginBlock`.
- Superfluid staking will continue to expand to other Osmosis pools
  based on governance proposals and vote turnouts.

### Example

If Alice has 500 GAMM tokens bonded to the ATOM \<\> OSMO, she will have
the equivalent value of OSMO minted, delegated to her chosen staker, and
burned for her each day with Superfluid staking. On the user side, all
she has to know is who she wants to delegate her tokens to. In order to
switch delegation, she has to unbond her tokens from the pool first and
then redeposit. Bob, who has a share of the same liquidity pool before
Superfluid Staking went live, also has to re-deposit into the pool for
the above process to kickstart.

### Why mint Osmo? How is this method safe and accurate?

Superfluid staking requires the minting of OSMO because in order to
stake on the Osmosis chain, OSMO tokens are required as the chosen
collateral. Synthetic Osmo is minted here as a representative of the
value of each superfluid staker's liquidity pool tokens.

The pool tokens are acquired by the user from normally staking in a
liquidity pool. They get minted an amount of OSMO equivalent to the
value of their GAMM pool tokens. This method is accurate because
querying the value OSMO every day allows for burning and minting
according to the difference in value of OSMO relative to the expected
delegation amount (as seen with
[GetExpectedDelegationAmount](https://github.com/osmosis-labs/osmosis/blob/main/x/superfluid/keeper/stake.go)).
It's like having a price oracle for fairly calculating the amount the
user has superfluidly staked.

On epoch (start of every day), we read from the lockup module how much
GAMM tokens we have locked which acts as an oracle for the
representative price of the GAMM token shares. The superfluid module has
"hooks" messages to refresh delegation amounts
(`RefreshIntermediaryDelegationAmounts`) and to increase delegation on
lockup (`IncreaseSuperfluidDelegation`). Then, we see whether or not the
superfluid OSMO currently delegated is worth more or less than this
expected delegation amount amount. If the OSMO is worth more, we do
instant undelegations and immediately burn the OSMO. If less, we mint
OSMO and update the amount delegated. A simplified diagram of this whole
process is found below:

<br/>

<p style="text-align:center;">

<img src="https://raw.githubusercontent.com/osmosis-labs/osmosis/main/x/superfluid/spec/superfluiddiagram.png" height="300"/>

</p>

</br>

This minting is safe because we strict constrain the permissions of Bank
(the module that burns and mints OSMO) to do what it's designed to do.
The authority is mediated through `mintOsmoTokensAndDelegate` and
`forceUndelegateAndBurnOsmoTokens` keeper methods called by the
`SuperfluidDelegate` and `SuperfluidUndelegate` message handlers for the
tokens. The hooks above that increase delegation and refresh delegation
amounts also call this keeper method.

The delegation is then verified to not already be associated with an
intermediary account (to prevent double-staking), and is always
delegated or withdrawn taking into account various multipliers for
synthetic OSMO value (its worth with respect to the liquidity pool, and
a risk modifier) to prevent mint inaccuracies. Before minting, we also
check that the message sender is the owner of the locked funds; that the
lock is not unlocking; is locked for at least the unbonding period, and
is bonded to a single asset. We also check to see if the lock isn't
already in superfluid and that the same lock isn't currently being
unbonded.

On the end of each epoch, we iterate through all intermediary accounts
to withdraw delegation rewards they may have received and put it all
into the perpetual gauges corresponding to each account for reward
delegation.

### Bonding, unbonding, slashing

Here, we describe how token bonding and unbonding works, and what
happens to your superfluid tokens in the case of a slashing event.

### Bonding

When bonding, your input tokens are locked up and you are given GAMM
pool tokens in exchange. These GAMM pool tokens represent a share of the
total liquidity pool, and allows you to get transaction fees or
participate in external incentive gauge token distributions. When
bonding, on top of the regular bonding transaction there will also be a
selection of validators. As stated above, OSMO is also minted and burned
each day and superfluidly staked to whoever you have chosen to be your
validator. You gain additional APR as a reward for bolstering the
Osmosis chain's consensus integrity by delegating.

### Unbonding

When unbonding, superfluid tokens get un-delegated. After making sure
that the unbond message sender is the owner of their corresponding
locked funds, the existing synthetic lockup is deleted and replaced with
a new synthetic lockup for unbonding purposes. The undelegated OSMO is
then instantly withdrawn from the intermediate account and validator
using the InstantUndelegate function. The OSMO that was originally used
for representing your LP shares are burnt. Moves the tracker for
unbonding, allows the underlying lock to start unlocking if desired

## Concepts

### SyntheticLockups

SyntheticLockups are synthetica of PeriodLocks, but different in the
sense that they store suffix, which is a combination of
bonding/unbonding status + validator address. This is mainly used to
track whether an individual lock that has been superfluid staked has an
bonding status or a unbonding status from the staking delegations.

### Intermediary Account

Intermediary Accounts establishes the connections between the superfluid
staked locks and delegations to the validator. Intermediary accounts
exists for every denom + validator combination, so that it would group
locks with the same denom + validator selection. Superfluid staking a
lock would mint equivalent amount of OSMO of the lock and send it to the
intermediary account and the intermediarry accounts would be delegating
to the specified validator.

### Intermediary Account Connection

Intermediary Accounts Connection serves the role of tracking the locks
that an Intermediary Account is dedicated to.

## State

### Superfluid Asset

A superfluid asset is a alternative asset (non-OSMO) that is allowed by
governance to be used for staking.

It can only be updated by governance proposals. We validate at proposal
creation time that the denom + pool exists. (Are we going to ignore edge
cases around a reference pool getting deleted it)

### Intermediary Accounts

Lots of questions to be answered here

### Dedicated Gauges

Each intermediary account has has dedicated gauge where it sends the
delegation rewards to. Gauges are distributing the rewards to end users
at the end of the epoch.

### Synthetic Lockups created

At the moment, one lock can only be fully bonded to one validator.

### Osmo Equivalent Multipliers

The Osmo Equivalent Multiplier for an asset is the multiplier it has for
its value relative to OSMO.

Different types of assets can have different functions for calculating
their multiplier. We currently support two asset types.

1. Native Token

The multiplier for OSMO is alway 1.

2. Gamm LP Shares

Currently we use the spot price for an asset based on a designated
osmo-basepair pool of an asset. The multiplier is set once per epoch, at
the beginning of the epoch. In the future, we will switch this out to
use a TWAP instead.

### State changes

The state of superfluid module state modifiers are classified into below
categories.

- [Proposals](07_proposals.md)
- [Messages](03_messages.md)
- [Epoch](04_epoch.md)
- [Hooks](06_hooks.md)

### Messages

### Superfluid Delegate

Owners of superfluid asset locks can submit `MsgSuperfluidDelegate`
transactions to delegate the Osmo in their locks to a selected
validator.

```{.go}
type MsgSuperfluidDelegate struct {
 Sender  string
 LockId  uint64
 ValAddr string
}
```

**State Modifications:**

- Safety Checks that are being done before running superfluid logic:
  - Check that `Sender` is the owner of `lock`
  - Check that `lock` corresponds to a single locked asset
  - Check that `lock` is not unlocking
  - Check that `lock` is locked for at least the unbonding period
  - Check that this `LockID` is not already superfluided
  - Check that the same lock isn't being unbonded
- Get the `IntermediaryAccount` for this lock's `Denom` and `ValAddr`
  pair.
  - Create it + a new gauge for the synthetic denom, if it does not
    yet exist.
- Create a SyntheticLockup.
- Calculate `Osmo` to delegate on behalf of this `lock`, as
  `Osmo Equivalent Multiplier` \* `# LP Shares` \*
  `Risk Adjustment Factor`
  - If this amount is less than 0.000001 `Osmo` (`1 uosmo`) reject
    the transaction, as it would be delegating `0 uosmo`
- Mint `Osmo` to match this amount and send to `IntermediaryAccount`
- Create a delegation from `IntermediaryAccount` to `Validator`
- Create a new perpetual `Gauge` for distributing staking payouts to
  locks of a synethic asset based on this `Validator` / `Denom` pair.
- Create a connection between this `lockID` and this
  `IntermediaryAccount`

### Superfluid Undelegate

```{.go}
type MsgSuperfluidUndelegate struct {
 Sender string
 LockId uint64
}
```

**State Modifications:**

- Lookup `lock` by `LockID`
- Check that `Sender` is the owner of `lock`
- Get the `IntermediaryAccount` for this `lockID`
- Delete the `SyntheticLockup` associated to this `lockID` + `ValAddr`
  pair
- Create a new `SyntheticLockup` which is unbonding
- Calculate the amount of `Osmo` delegated on behalf of this `lock` as
  `Osmo Equivalent Multipler` \* `# LP Shares` \*
  `Risk Adjustment Factor`
  - If this amount is less than 0.000001 `Osmo`, there is no
    delegated `Osmo` to undelegate and burn
- Use `InstantUndelegate` to instantly remove delegation from
  `IntermediaryAccount` to `Validator`
- Immediately burn undelegated `Osmo`
- Delete the connection between `lockID` and `IntermediaryAccount`

### Lock and Superfluid Delegate

```{.go}
type MsgLockAndSuperfluidDelegate struct {
 Sender string
 Coins sdk.Coins
 ValAddr string
}
```

This is effectively a multimsg tx of lockup's `MsgLockTokens` and
superfluid's `MsgSuperfluidDelegate`, but it is implemented as a single
msg, because currently we don't have a way of passing the lockid
outputted by `MsgLockTokens` as an input into the
`MsgSuperfluidDelegate` prior to execution.

**State Modifications:**

- Ensures that Coins has a length of only 1 (we use sdk.Coins instead
  of sdk.Coin in order to allow more flexibility in the future)
- Creates a lockup with Coins of a lock duration equivalent to the
  unstaking period from the staking module
  - Uses the lockup module's MsgServer
- Gets the lock id of the created lock, and uses it generate and
  execute a MsgSuperfluidDelegate message
  - Uses the SuperfluidDelegate function on this msg server

### Superfluid Unbond Lock

```{.go}
type MsgSuperfluidUnbondLock struct {
 Sender string
 LockId uint64
}
```

This message does all the functionality of `MsgSuperfluidUndelegate` but
also starts unbonding the underlying lock as well, allowing both the
unstaking and unlocking to complete at the same time. Without using this
function, a user will not be able to start unbonding their underlying
lock until after the the unstaking has finished.

**State Modifications:**

- This runs the functionality of `MsgSuperfluidUndelegate`
- It then triggers a force unbond of the underlying lock id

## Epochs

Overall Epoch sequence

- Epoch N ends, during AfterEpochEnd:
  - Distribute gauge rewards for all non-superfluid gauges
  - Mint new tokens
    - Issue new Osmo, and send to various modules (distribution,
      incentives, etc.)
    - 25% currently goes to `x/distribution` which funds `Staking`
      and `Superfluid` rewards
    - Rewards for `Superfluid` are based on the just updated
      delegation amounts, and queued for payout in the next epoch
- BeginBlock for Distribution
  - Distribute staking rewards to all of the 'lazy accounting'
    accumulators. (F1)
- Epoch N ends, during BeginBlock for superfluid **After**
  AfterEpochEnd:
  - Claim staking rewards for every `Intermediary Account`, put them
    into gauges.
  - Distribute Superfluid staking rewards from gauges to bonded
    Synthetic Lock owners
  - Update `Osmo Equivalent Multiplier` value for each LP token
    - (Currently spot price at epoch)
  - Refresh delegation amounts for all `Intermediary Accounts`
    - Calculate the expected delegation for this account as
      `Osmo Equivalent Multipler` _`# LP Shares`_
      `Risk adjustment`
      - If this is less than 0.000001 `Osmo` it will be rounded
        to 0
    - Lookup current delegation amount for `Intermediary Account`
      - If there is no delegation, treat the current delegation
        as 0
    - If expected amount \> current delegation:
      - Mint new `Osmo` and `Delegate` to `Validator`
    - If expected amount \< current delegation:
      - Use `InstantUndelegate` and burn the received `Osmo`

## Staking power updates

We need to be concerned with how/when validators enter and leave the
active set.

We expect the guarantee that there is an Intermediary account for every
(active validator, superfluid denom) pair, and every (unbonding
validator, superfluid denom) pair. (TODO: Where/why)

We also want to avoid resource exhaustion attacks. We relegate concerns
around upper-bounding the number of active + unbonding validators to the
staking module. This module is liable to potentially cause a 100-1000x
amplification factor on this workload.

### How we handle it now

- Intermediary accounts are not created on SetSuperfluidAsset
- They are created at-time-of-need on MsgSuperfluidDelegate
- Concerns: What happens if you delegate to an unbonding or jailed
  validator. Note: Isn't it same as normal delegation for unbonding
  validator?

## Other Module Hooks

-----;

In this section we describe the "hooks" that `superfluid` module
receives from other modules.

### AfterEpochEnd

On AfterEpochEnd, we iterate through all existing intermediary accounts
and withdraw delegation rewards they have received. Then we send the
collective rewards to the perpetual gauge corresponding to the
intermediary account. Then we update OSMO backing per share for the
specific pool. After the update, iteration through all intermediate
accounts happen, undelegating and bonding existing delegations for all
superfluid staking and use the updated spot price at epoch time to mint
and delegate.

### AfterAddTokensToLock

When a token is locked, we first check if the corresponding lock is
currently in the state of superfluid delegation. If it is, we run the
logic to add delegation via intermediary account.

### BeforeValidatorSlashed

Slashes the synthetic lockups and native lockups that is connected to
the to be slashed validator.

## Proposal Hooks

-----;

In this section we describe the proposals that is associated to
superfluid module.

### SetSuperfluidAssetsProposal

Enable multiple superfluid assets to be used for superfluid staking.

### RemoveSuperfluidAssetsProposal

Disable multiple assets from being used for superfluid staking.

## Events

-----;

### Messages

### MsgSuperfluidDelegate

| Type                | Attribute Key | Attribute Value |
| ------------------- | ------------- | --------------- |
| superfluid_delegate | lock_id       | {lock_id}       |
| superfluid_delegate | validator     | {validator}     |

### MsgSuperfluidUndelegate

| Type                  | Attribute Key | Attribute Value |
| --------------------- | ------------- | --------------- |
| superfluid_undelegate | lock_id       | {lock_id}       |

### MsgSuperfluidUnbondLock

| Type                   | Attribute Key | Attribute Value |
| ---------------------- | ------------- | --------------- |
| superfluid_unbond_lock | lock_id       | {lock_id}       |

### MsgLockAndSuperfluidDelegate

| Type                | Attribute Key  | Attribute Value |
| ------------------- | -------------- | --------------- |
| lock_tokens         | period_lock_id | {periodLockID}  |
| lock_tokens         | owner          | {owner}         |
| lock_tokens         | amount         | {amount}        |
| lock_tokens         | duration       | {duration}      |
| lock_tokens         | unlock_time    | {unlockTime}    |
| message             | action         | lock_tokens     |
| message             | sender         | {owner}         |
| transfer            | recipient      | {moduleAccount} |
| transfer            | sender         | {owner}         |
| transfer            | amount         | {amount}        |
| superfluid_delegate | lock_id        | {lock_id}       |
| superfluid_delegate | validator      | {validator}     |

## Proposals

### SetSuperfluidAssetsProposal

| Type                 | Attribute Key         | Attribute Value |
| -------------------- | --------------------- | --------------- |
| set_superfluid_asset | denom                 | {denom}         |
| set_superfluid_asset | superfluid_asset_type | {asset_type}    |

### RemoveSuperfluidAssetsProposal

| Type                    | Attribute Key | Attribute Value |
| ----------------------- | ------------- | --------------- |
| remove_superfluid_asset | denom         | {denom}         |

## Queries

### Params

```protobuf
message ParamsRequest {};

message ParamsResponse {
  // params defines the parameters of the module.
  Params params = 1 [ (gogoproto.nullable) = false ];
}

message Params {
  sdk.Dec minimum_risk_factor = 1; // serialized as string
}
```

The params query returns the params for the superfluid module. This
currently contains:

- `MinimumRiskFactor` which is an sdk.Dec that represents the discount
  to apply to all superfluid staked modules when calcultating their
  staking power. For example, if a specific denom has an OSMO
  equivalent value of 100 OSMO, but the the `MinimumRiskFactor` param
  is 0.05, then the denom will only get 95 OSMO worth of staking power
  when staked.

### AssetType

```protobuf
message AssetTypeRequest {
    string denom = 1;
};

message AssetTypeResponse {
    SuperfluidAssetType asset_type = 1;
};

enum SuperfluidAssetType {
  SuperfluidAssetTypeNative = 0;
  SuperfluidAssetTypeLPShare = 1;
}
```

The AssetType query returns what type of superfluid asset a denom is.
AssetTypes are meant for when we support more types of assets for
superfluid staking than just LP shares. Each AssetType has a different
algorithm used to get its "Osmo equivalent value".

We represent different types of superfluid assets as different enums.
Currently, only enum `1` is actually used. Enum value `0` is reserved
for the Native staking token for if we deprecate the legacy staking
workflow to have native staking also go through the superfluid module.
In the future, more enums will be added.

If this query errors, that means that a denom is not allowed to be used
for superfluid staking.

### AllAssets

```protobuf
message AllAssetsRequest {};

message AllAssetsResponse {
  repeated SuperfluidAsset assets = 1 [ (gogoproto.nullable) = false ];
};

message SuperfluidAsset {
  string denom = 1;
  SuperfluidAssetType asset_type = 2;
}
```

This parameterless query returns a list of all the superfluid staking
compatible assets. The return value includes a list of SuperfluidAssets,
which are pairs of `denom` with `SuperfluidAssetType` which was
described in the previous section.

This query does not currently support pagination, but may in the future.

### AssetMultiplier

```protobuf
message AssetMultiplierRequest {
    string denom = 1;
};

message AssetMultiplierResponse {
  OsmoEquivalentMultiplierRecord osmo_equivalent_multiplier = 1;
};

message OsmoEquivalentMultiplierRecord {
  int64 epoch_number = 1;
  string denom = 2;
  string multiplier = 3;
}
```

This query allows you to find the multiplier factor on a specific denom.
The Osmo-Equivalent-Multiplier Record for epoch N refers to the osmo
worth we treat a denom as having, for all of epoch N. For now, this is
the spot price at the last epoch boundary, and this is reset every
epoch. We currently don't store historical multipliers, so the epoch
parameter is kind of meaningless for now.

To calculate the staking power of the denom, one needs to multiply the
amount of the denom with `OsmoEquivalentMultipler` from this query with
the `MinimumRiskFactor` from the Params query endpoint.

`staking_power = amount * OsmoEquivalentMultipler * MinimumRiskFactor`

### ConnectedIntermediaryAccount

```protobuf
message ConnectedIntermediaryAccountRequest {
  uint64 lock_id = 1;
}

message ConnectedIntermediaryAccountResponse {
  SuperfluidIntermediaryAccountInfo account = 1;
}

message SuperfluidIntermediaryAccount {
  string denom = 1;
  string val_addr = 2;
  uint64 gauge_id = 3; // perpetual gauge for rewards distribution
}
```

Every superfluid denom and validator pair has an associated
"intermediary account", which does the actual delegation. This query
helps find the superfluid intermediary account for any superfluid
position.

That `lock_id` parameter passed in is the underlying lock id for the
superfluid, NOT the synthetic lock id.

This query can be used to find the validator a superfluid lock is
delegated to. The `gauge_id` also refers to the perpetual gauge that is
used to pay out the superfluid positions associated with this
intermediary account.

### AllIntermediaryAccounts

```{.protobuf}
message AllIntermediaryAccountsRequest {
  cosmos.base.query.v1beta1.PageRequest pagination = 1;
};

message AllIntermediaryAccountsResponse {
  repeated SuperfluidIntermediaryAccountInfo accounts = 1;
  cosmos.base.query.v1beta1.PageResponse pagination = 2;
};
```

This query returns a list of all superfluid intermediary accounts. It
supports pagination.

### SuperfluidDelegationAmount

```{.protobuf}
message SuperfluidDelegationAmountRequest {
  string delegator_address = 1;
  string validator_address = 2;
  string denom = 3;
}

message SuperfluidDelegationAmountResponse {
  repeated cosmos.base.v1beta1.Coin amount = 1 [];
}
```

This query returns the amount of underlying denom (i.e. lp share) for a
triplet of delegator, validator, and denom.

### SuperfluidDelegationsByDelegator

```{.protobuf}
message SuperfluidDelegationsByDelegatorRequest {
  string delegator_address = 1;
}

message SuperfluidDelegationsByDelegatorResponse {
  repeated SuperfluidDelegationRecord superfluid_delegation_records = 1;
  repeated cosmos.base.v1beta1.Coin total_delegated_coins = 2;
}

message SuperfluidDelegationRecord {
  string delegator_address = 1;
  string validator_address = 2;
  cosmos.base.v1beta1.Coin delegation_amount = 3;
}
```

This query returns a list of all the superfluid delegations of a
specific delegator. The return value includes, the validator delgated to
and the delegated coins (both denom and amount).

The return value of the query also includes the `total_delegated_coins`
which is the sum of all the delegations of that validator.

This query does require iteration that is linear with the number of
delegations a delegator has made, but for now until we support many
superfluid denoms, should be relatively bounded. Once that increases, we
will need to support pagination.

### SuperfluidDelegationsByValidatorDenom

```{.protobuf}
message SuperfluidDelegationsByValidatorDenomRequest {
  string validator_address = 1;
  string denom = 2;
}

message SuperfluidDelegationsByValidatorDenomResponse {
  repeated SuperfluidDelegationRecord superfluid_delegation_records = 1;
}
```

This query returns a list of all superfluid delegations that are with a
validator / superfluid denom pair. This query requires a lot of
iteration and should be used sparingly. We will need to add pagination
to make this usable.

### EstimateSuperfluidDelegatedAmountByValidatorDenom

```{.protobuf}
message EstimateSuperfluidDelegatedAmountByValidatorDenomRequest {
  string validator_address = 1;
  string denom = 2;
}

message EstimateSuperfluidDelegatedAmountByValidatorDenomResponse {
  repeated cosmos.base.v1beta1.Coin total_delegated_coins = 1;
}
```

This query returns the total amount of delegated coins for a validator /
superfluid denom pair. This query does NOT involve iteration, so should
be used instead of the above `SuperfluidDelegationsByValidatorDenom`
whenever possible. It is called an "Estimate" because it can have some
slight rounding errors, due to conversions between sdk.Dec and
sdk.Int\", but for the most part it should be very close to the sum of
the results of the previous query.

## Parameters

The superfluid module contains the following parameters:

| Key                 | Type    | Example |
| ------------------- | ------- | ------- |
| minimum_risk_factor | decimal | 0.01    |

## Slashing

Slashing works by gathering all accounts who were superfluidly staking
and delegated to the violating validator and slashing their underlying
lock collateral. The amount of tokens to slash are first calculated then
removed from the underlying and synthetic lock. Therefore, it is
important to select a reputable or reliable validator as to minimize
slashing risks on your tokens. At the moment we are slashing at latest
price rather than block height price. All slashed tokens go to the
community pool.

We first get a hook from the staking module, marking that a validator is
about to be slashed at a slashFactor of `f`, for an infraction at height
`h`.

The staking module handles slashing every delegation to that validator,
which will handle slashing the delegation from every intermediary
account. However, it is up to the superfluid module to then:

- Slash every constituent superfluid staking position for this
  validator.
- Slash every unbonding superfluid staking position to this validator.

We do this by:

- Collect all intermediate accounts to this validator
- For each IA, iterate over every lock to the underlying native denom.
- If the lock has a synthetic lockup, it gets slashed.
- The slash works by calculating the amount of tokens to slash.
- It removes these from the underlying lock and the synthetic lock.
- These coins are moved to the community pool.

### Nuances

- Slashed tokens go to the community pool, rather than being burned as
  in staking.
- We slash every unbonding, rather than just unbondings that started
  after the infraction height.
- We can "overslash" relative to the staking module. (For a slash
  factor of 5%, the staking module can often burn \<5% of active
  delegation, but superfluid will always slash 5%)

We slash every unbonding, purely because lockup module tracks things by
unbonding start time, whereas staking/slashing tracks things by height
we begin unbonding at. Thus we get a problem that we cannot convert
between these cleanly. Really there should be a storage of all
historical block height \<\> block times for everything in the unbonding
period, but this is not considered a near-term problem.

### Correcting overslashing

The overslashing possibility stems from a problem in the SDKs slashing
module, that really is a bug there, and superfluid is doing the correct
thing. <https://github.com/cosmos/cosmos-sdk/issues/1440>

Basically, slashes to unbondings and redelegations can lower the amount
that gets slashed from live delegations in the staking module today.

It turns out this edge case, where superfluid's intermediate account can
have more delegation than expected from its underlying collateral, is
already safely handled by the Superfluid refreshing logic.

The refreshing logic checks the total amount of tokens in locks to this
denom (Reading from the lockup accumulation store), calculates how many
osmo thats worth at the epochs new osmo worth for that asset, and then
uses that. Thus this safely handles this edge case, as it uses the new
'live' lockup amount.

## Minting

Superfluid module has the ability to arbitrarily mint and burn Osmo
through the `bank` module. This is potentially dangerous so we strictly
constrain it's ability to do so. This authority is mediated through the
`mintOsmoTokensAndDelegate` and `forceUndelegateAndBurnOsmoTokens`
keeper methods, which are in turn called by message handlers
(`SuperfluidDelegate` and `SuperfluidUndelegate`) as well as by hooks on
Epoch (`RefreshIntermediaryDelegationAmounts`) and Lockup
(`IncreaseSuperfluidDelegation`)

### Invariant

Each of these mechanisms maintains a local invariant between the amount
of Osmo minted and delegated by the `IntermediaryAccount`, and the
quantity of the underlying asset held by locks associated to the
account, modified by `OsmoEquivalentMultiplier` and `RiskAdjustment` for
the underlying asset. Namely that total minted/delegated =
`GetTotalSyntheticAssetsLocked` \* `GetOsmoEquivalentMultiplier` \*
`GetRiskAdjustment`

This can be equivalently expressed as `GetExpectedDelegationAmount`
being equal to the actual delegation amount.

## Message Handlers

### SuperfluidDelegate

In a `SuperfluidDelegate` transaction, we first verify that this lock is
not already associated to an `IntermediaryAccount`, and then use
`mintOsmoTokenAndDelegate` to properly balance the resulting change in
`GetExpectedDelegationAmount` from the increase in
`GetTotalSyntheticAssetsLocked`. i.e. we mint and delegate:
`GetOsmoEquivalentMultiplier` \* `GetRiskAdjustment` \*
`lock.Coins.Amount` new Osmo tokens.

### SuperfluidUndelegate

When a user submits a transaction to unlock their asset the invariant is
maintained by using `forceUndelegateAndBurnOsmoTokens` to remove an
amount of Osmo equal to `lockedCoin.Amount` \*
`GetOsmoEquivalentMultiplier` \* `GetRiskAdjustment`.

## Superfluid Hooks

### RefreshIntermediaryDelegationAmounts (AfterEpochEnd Hook)

In the `RefreshIntermediaryDelegationAmounts` method, calls are made to
`mintOsmoTokensAndDelegate` or `forceUndelegateAndBurnOsmoTokens` to
adjust the real delegation up or down to match
`GetExpectedDelegationAmount`.

### IncreaseSuperfluidDelegation (AfterAddTokensToLock Hook)

This is called as a result of a user adding more assets to a lock that
has already been associated to an `IntermediaryAccount`. The invariant
is maintained by using `mintOsmoTokenAndDelegate` to match the amount of
new asset locked \* `GetOsmoEquivalentMultiplier` \* `GetRiskAdjustment`
for the underlying asset.

### SlashLockupsForValidatorSlash (BeforeValidatorSlashed Hook)

During slashing the invariant is likely to be temporraily broken if the
referenced validator has any unbonding delegations. These unbonding
delegations are slashed first, which means that the amount delegated by
the `IntermediaryAccount` will be slashed by less than the
`SyntheticLock`s held by the account.

## See Also

### GetTotalSyntheticAssetsLocked

TODO - expand on this Uses `lockup` accumulator to find total amount of
synthetic locks for a given `IntermediaryAccount` (Superfluid Asset +
Validator pair)
