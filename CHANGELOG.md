<!--
Guiding Principles:

Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.

Usage:

Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry should ideally include a tag and
the Github issue reference in the following format:

* (<tag>) \#<issue-number> message

The issue numbers will later be link-ified during the release process so you do
not have to worry about including a link manually, but you can if you wish.

Types of changes (Stanzas):

"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"Client Breaking" for breaking CLI commands and REST routes used by end-users.
"API Breaking" for breaking exported APIs used by developers building on SDK.
"State Machine Breaking" for any changes that result in a different AppState
given same genesisState and txList.
Ref: https://keepachangelog.com/en/1.0.0/
-->

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## v19.0.0

### Features

* [#6034](https://github.com/osmosis-labs/osmosis/pull/6034) Taker fee

### Bug Fixes
* [#6190](https://github.com/osmosis-labs/osmosis/pull/6190) v19 upgrade handler superfluid fix
* [#6195](https://github.com/osmosis-labs/osmosis/pull/6195) (x/tokenfactory) Fix events for `mintTo` and `burnFrom`
* [#6195](https://github.com/osmosis-labs/osmosis/pull/6195) Fix panic edge case in superfluid AfterEpochEnd hook by surrounding CL multipler update with ApplyFuncIfNoError

### Misc Improvements

### Minor improvements & Bug Fixes

### Security

## v18.0.0

<<<<<<< HEAD
=======
### Misc Improvements

* [#6161](https://github.com/osmosis-labs/osmosis/pull/6161) Reduce CPU time of epochs

### Bug Fixes

* [#6162](https://github.com/osmosis-labs/osmosis/pull/6162) allow zero qualifying balancer shares in CL incentives

### Features

* [#6034](https://github.com/osmosis-labs/osmosis/pull/6034) feat(spike): taker fee
## v18.0.0

>>>>>>> 5c8fd80f (feat(spike): taker fee (#6034))
Fixes mainnet bugs w/ incorrect accumulation sumtrees, and CL handling for a balancer pool with 0 bonded shares.

### Improvements

* [#6144](https://github.com/osmosis-labs/osmosis/pull/6144) perf: Speedup compute time of Epoch
* [#6144](https://github.com/osmosis-labs/osmosis/pull/6144) misc: Move many Superfluid info logs to debug logs
* [#6161](https://github.com/osmosis-labs/osmosis/pull/6161) Reduce CPU time of epochs

### API breaks

* [#6071](https://github.com/osmosis-labs/osmosis/pull/6071) reduce number of returns for UpdatePosition and TicksToSqrtPrice functions

### Bug Fixes

* [#6162](https://github.com/osmosis-labs/osmosis/pull/6162) allow zero qualifying balancer shares in CL incentives
* [#6053](https://github.com/osmosis-labs/osmosis/pull/6053) monotonic sqrt with 36 decimals

## v17.0.0

### API breaks

* [#6014](https://github.com/osmosis-labs/osmosis/pull/6014) refactor: reduce the number of returns in superfluid migration
* [#5983](https://github.com/osmosis-labs/osmosis/pull/5983) refactor(CL): 6 return values in CL CreatePosition with a struct
* [#6004](https://github.com/osmosis-labs/osmosis/pull/6004) reduce number of returns for creating full range position
* [#6018](https://github.com/osmosis-labs/osmosis/pull/6018) golangci: add unused parameters linter

### Features

* [#5072](https://github.com/osmosis-labs/osmosis/pull/5072) IBC-hooks: Add support for async acks when processing onRecvPacket

### State Breaking

* [#5532](https://github.com/osmosis-labs/osmosis/pull/5532) fix: Fix x/tokenfactory genesis import denoms reset x/bank existing denom metadata
* [#5863](https://github.com/osmosis-labs/osmosis/pull/5863) fix: swap base/quote asset for CL spot price query
* [#5869](https://github.com/osmosis-labs/osmosis/pull/5869) fix negative interval accumulation with spread rewards
* [#5872](https://github.com/osmosis-labs/osmosis/pull/5872) fix negative interval accumulation with incentive rewards
* [#5883](https://github.com/osmosis-labs/osmosis/pull/5883) feat: Uninitialize empty ticks
* [#5874](https://github.com/osmosis-labs/osmosis/pull/5874) Remove Partial Migration from superfluid migration to CL
* [#5901](https://github.com/osmosis-labs/osmosis/pull/5901) Adding support for CW pools in ProtoRev
* [#5937](https://github.com/osmosis-labs/osmosis/pull/5937) feat: add SetScalingFactorController gov prop
* [#5949](https://github.com/osmosis-labs/osmosis/pull/5949) Add message to convert from superfluid / locks to native staking directly.
* [#5939](https://github.com/osmosis-labs/osmosis/pull/5939) Fix: Flip existing twapRecords base/quote price denoms 
* [#5938](https://github.com/osmosis-labs/osmosis/pull/5938) Chore: Fix valset amino codec

### BugFix

* [#5831](https://github.com/osmosis-labs/osmosis/pull/5831) Fix superfluid_delegations query
* [#5835](https://github.com/osmosis-labs/osmosis/pull/5835) Fix println's for "amountZeroInRemainingBigDec before fee" making it into production
* [#5841](https://github.com/osmosis-labs/osmosis/pull/5841) Fix protorev's out of gas erroring of the user's transcation.
* [#5930](https://github.com/osmosis-labs/osmosis/pull/5930) Updating Protorev Binary Search Range Logic with CL Pools
* [#5950](https://github.com/osmosis-labs/osmosis/pull/5950) fix: spot price for cosmwasm pool types

### Misc Improvements

* [#5534](https://github.com/osmosis-labs/osmosis/pull/5534) fix: fix the account number of x/tokenfactory module account
* [#5750](https://github.com/osmosis-labs/osmosis/pull/5750) feat: add cli commmand for converting proto structs to proto marshalled bytes
* [#5889](https://github.com/osmosis-labs/osmosis/pull/5889) provides an API for protorev to determine max amountIn that can be swapped based on max ticks willing to be traversed
* [#5849](https://github.com/osmosis-labs/osmosis/pull/5849) CL: Lower gas for leaving a position and withdrawing rewards
* [#5855](https://github.com/osmosis-labs/osmosis/pull/5855) feat(x/cosmwasmpool): Sending token_in_max_amount to the contract before running contract msg
* [#5893](https://github.com/osmosis-labs/osmosis/pull/5893) Export createPosition method in CL so other modules can use it in testing
* [#5870](https://github.com/osmosis-labs/osmosis/pull/5870) Remove v14/ separator in protorev rest endpoints
* [#5923](https://github.com/osmosis-labs/osmosis/pull/5923) CL: Lower gas for initializing ticks
* [#5927](https://github.com/osmosis-labs/osmosis/pull/5927) Add gas metering to x/tokenfactory trackBeforeSend hook
* [#5890](https://github.com/osmosis-labs/osmosis/pull/5890) feat: CreateCLPool & LinkCFMMtoCL pool into one gov-prop
* [#5959](https://github.com/osmosis-labs/osmosis/pull/5959) allow testing with different chain-id's in E2E testing
* [#5964](https://github.com/osmosis-labs/osmosis/pull/5964) fix e2e test concurrency bugs
* [#5948](https://github.com/osmosis-labs/osmosis/pull/5948) Parameterizing Pool Type Information in Protorev
* [#6001](https://github.com/osmosis-labs/osmosis/pull/6001) feat: improve set-env CLI cmd
* [#6012](https://github.com/osmosis-labs/osmosis/pull/6012) chore: add autocomplete to makefile

### Minor improvements & Bug Fixes

* [#5806](https://github.com/osmosis-labs/osmosis/issues/5806) ci: automatically close issues generated by the Broken Links Check action when a new run occurs.

## v16.1.1

### Security

* [#5824](https://github.com/osmosis-labs/osmosis/pull/5824) chore: cosmovisor hashes and v16.1.0 tag updates

### Features

* [#5796](https://github.com/osmosis-labs/osmosis/pull/5796) chore: add missing cli queries CL 

### Misc Improvements & Bug Fixes

* [#5831](https://github.com/osmosis-labs/osmosis/pull/5831) Fix the superfluid query
* [#5784](https://github.com/osmosis-labs/osmosis/pull/5784) Chore: Add amino name for tx msgs

## v16.1.0

### Security

* [#5822](https://github.com/osmosis-labs/osmosis/pull/5822) Revert "feat: lock existing position and sfs"

## v16.0.0
Osmosis Labs is excited to announce the release of v16.0.0, a major upgrade that includes a number of new features and improvements like introduction of new modules, updates existing APIs, and dependency updates. This upgrade aims to enhance capital efficiency by introducing SuperCharged Liquidity, introduce custom liquidity pools backed by CosmWasm smart contracts, and improve overall functionality.

New Modules and Features:

SuperCharged Liquidity Module (x/concentrated-liquidity):
- Introduces a game-changing pool model that enhances captical efficiency in Osmosis.

CosmWasm Pool Module (x/cosmwasmpool):
- Enables the creation and management of liquidity pools backed by CosmWasm smart contracts.

ProtoRev Changes (x/protorev):
- Modifies the payment schedule for the dev account from weekly to after every trade.
- Triggers backruns, joinPool, and exitPool using hooks.

TokenFactory before send hooks (x/tokenfactory):
- This enhancement allows for executing custom logic before sending tokens, providing more flexibility
and control over token transfers.


### Security

* Upgraded wasmvm to 1.2.3 in response to [CWA-2023-002](https://github.com/CosmWasm/advisories/blob/main/CWAs/CWA-2023-002.md)

### Features
  * [#3014](https://github.com/osmosis-labs/osmosis/issues/3014) implement x/concentrated-liquidity module.
  * [#5354](https://github.com/osmosis-labs/osmosis/pull/5354) implement x/cosmwasmpool module.
  * [#4659](https://github.com/osmosis-labs/osmosis/pull/4659) implement AllPools query in x/poolmanager.
  * [#4886](https://github.com/osmosis-labs/osmosis/pull/4886) Implement MsgSplitRouteSwapExactAmountIn and MsgSplitRouteSwapExactAmountOut that supports route splitting.
  * [#5045] (<https://github.com/osmosis-labs/osmosis/pull/5045>) Implement hook-based backrunning logic for ProtoRev
  * [#5281](https://github.com/osmosis-labs/osmosis/pull/5281) Add option to designate Reward Recipient to Lock and Incentives.
  * [#4827](https://github.com/osmosis-labs/osmosis/pull/4827) Protorev: Change highest liquidity pool updating from weekly to daily and change dev fee payout from weekly to after every trade.
  * [#5409](https://github.com/osmosis-labs/osmosis/pull/5409) x/gov: added expedited quorum param (Note: we set the expedited quorum to 2/3 in the upgrade handler)
  * [#4382](https://github.com/osmosis-labs/osmosis/pull/4382) Tokenfactory: Add Before send hooks

### API breaks
  * [#5375](https://github.com/osmosis-labs/osmosis/pull/5373) Add query and cli for lock reward receiver
  * [#4757](https://github.com/osmosis-labs/osmosis/pull/4752) Pagination for all intermediary accounts
  * [#5066](https://github.com/osmosis-labs/osmosis/pull/5066) Fixed bad stargate query declaration
  * [#4868](https://github.com/osmosis-labs/osmosis/pull/4868) Remove wasmEnabledProposals []wasm.ProposalType from NewOsmosisApp
  * [#4791](https://github.com/osmosis-labs/osmosis/pull/4791) feat(osmoutils): cosmwasm query and message wrappers
  * [#4549](https://github.com/osmosis-labs/osmosis/pull/4549) added single pool query
  * [#4659](https://github.com/osmosis-labs/osmosis/pull/4659) feat: implement AllPools query in x/poolmanager
  * [#4489](https://github.com/osmosis-labs/osmosis/pull/4489) Add unlocking lock id to BeginUnlocking response
  * [#4658](https://github.com/osmosis-labs/osmosis/pull/4658) refactor: unify pool query in pool manager, deprecate x/gamm, remove from CL module
  * [#4682](https://github.com/osmosis-labs/osmosis/pull/4682) feat(CL): x/poolmanager spot price query for concentrated liquidity
  * [#5138](https://github.com/osmosis-labs/osmosis/pull/5138) refactor: rename swap fee to spread factor
  * [#5020](https://github.com/osmosis-labs/osmosis/pull/5020) Add gas config to the client.toml
  * [#5459](https://github.com/osmosis-labs/osmosis/pull/5459) Create locktypes.LockQueryType.NoLock gauge. MsgCreateGauge takes pool id for new gauge type.
  * [#5503](https://github.com/osmosis-labs/osmosis/pull/5503) Deprecate gamm spot price query and add pool manager spot price query to stargate query whitelist.

## State Breaking
  * [#5380](https://github.com/osmosis-labs/osmosis/pull/5380) feat: add ica authorized messages in upgrade handler
  * [#5363](https://github.com/osmosis-labs/osmosis/pull/5363) fix: twap record upgrade handler
  * [#5265](https://github.com/osmosis-labs/osmosis/pull/5265) fix: expect single synthetic lock per native lock ID
  * [#4983](https://github.com/osmosis-labs/osmosis/pull/4983) implement gas consume on denom creation
  * [#4830](https://github.com/osmosis-labs/osmosis/pull/4830) Scale gas costs by denoms in gauge (AddToGaugeReward)
  * [#5511](https://github.com/osmosis-labs/osmosis/pull/5511) Scale gas costs by denoms in gauge (CreateGauge)
  * [#4336](https://github.com/osmosis-labs/osmosis/pull/4336) feat: make epochs standalone
  * [#4801](https://github.com/osmosis-labs/osmosis/pull/4801) refactor: remove GetTotalShares, GetTotalLiquidity and GetExitFee from PoolI
  * [#4951](https://github.com/osmosis-labs/osmosis/pull/4951) feat: implement pool liquidity query in pool manager, deprecate the one in gamm
  * [#5000](https://github.com/osmosis-labs/osmosis/pull/5000) osmomath.Power panics for base < 1 to temporarily restrict broken logic for such base.
  * [#5468](https://github.com/osmosis-labs/osmosis/pull/5468) fix: Reduce tokenfactory denom creation gas fee to 1_000_000

## Dependencies
  * [#4783](https://github.com/osmosis-labs/osmosis/pull/4783) Update wasmd to 0.31
  * [#5404](https://github.com/osmosis-labs/osmosis/pull/5404) Cosmwasm Cherry security patch
  * [#5320](https://github.com/osmosis-labs/osmosis/pull/5320) minor: huckleberry ibc patch

### Misc Improvements
  * [#5356](https://github.com/osmosis-labs/osmosis/pull/5356) Fix wrong restHandler for ReplaceMigrationRecordsProposal
  * [#5020](https://github.com/osmosis-labs/osmosis/pull/5020) Add gas config to the client.toml
  * [#5105](https://github.com/osmosis-labs/osmosis/pull/5105) Lint stableswap in the same manner as all of Osmosis
  * [#5065](https://github.com/osmosis-labs/osmosis/pull/5065) Use cosmossdk.io/errors
  * [#4549](https://github.com/osmosis-labs/osmosis/pull/4549) Add single pool price estimate queries
  * [#4767](https://github.com/osmosis-labs/osmosis/pull/4767) Disable create pool with non-zero exit fee
  * [#4847](https://github.com/osmosis-labs/osmosis/pull/4847) Update `make build` command to build only `osmosisd` binary
  * [#4891](https://github.com/osmosis-labs/osmosis/pull/4891) Enable CORS by default on localosmosis
  * [#4892](https://github.com/osmosis-labs/osmosis/pull/4847) Update Golang to 1.20
  * [#4893](https://github.com/osmosis-labs/osmosis/pull/4893) Update alpine docker base image to `alpine:3.17`
  * [#4907](https://github.com/osmosis-labs/osmosis/pull/4847) Add migrate-position cli
  * [#4912](https://github.com/osmosis-labs/osmosis/pull/4912) Export Position_lock_id mappings to GenesisState
  * [#4974](https://github.com/osmosis-labs/osmosis/pull/4974) Add lock id to `MsgSuperfluidUndelegateAndUnbondLockResponse`
  * [#2741](https://github.com/osmosis-labs/osmosis/pull/2741) Prevent updating the twap record if `ctx.BlockTime <= record.Time` or `ctx.BlockHeight <= record.Height`. Exception, can update the record created when creating the pool in the same block.
  * [#5165](https://github.com/osmosis-labs/osmosis/pull/5165) Improve error message when fully exiting from a pool.
  * [#5187](https://github.com/osmosis-labs/osmosis/pull/5187) Expand `IncentivizedPools` query to include internally incentivized CL pools.
  * [#5239](https://github.com/osmosis-labs/osmosis/pull/5239) Implement `GetTotalPoolShares` public keeper function for GAMM.
  * [#5261](https://github.com/osmosis-labs/osmosis/pull/5261) Allows `UpdateFeeTokenProposal` to take in multiple fee tokens instead of just one.
  * [#5265](https://github.com/osmosis-labs/osmosis/pull/5265) Ensure a lock cannot point to multiple synthetic locks. Deprecates `SyntheticLockupsByLockupID` in favor of `SyntheticLockupByLockupID`.
  * [#4950](https://github.com/osmosis-labs/osmosis/pull/4950) Add in/out tokens to Concentrated Liquidity's AfterConcentratedPoolSwap hook
  * [#4629](https://github.com/osmosis-labs/osmosis/pull/4629) Add amino proto annotations
  * [#4830](https://github.com/osmosis-labs/osmosis/pull/4830) Add gas cost when we AddToGaugeRewards, linearly increase with coins to add
  * [#5000](https://github.com/osmosis-labs/osmosis/pull/5000) osmomath.Power panics for base < 1 to temporarily restrict broken logic for such base.
  * [#4336](https://github.com/osmosis-labs/osmosis/pull/4336) Move epochs module into its own go.mod
  * [#5589](https://github.com/osmosis-labs/osmosis/pull/5589) Include linked balancer pool in incentivized pools query



## v15.1.2

### Security

* Upgraded ibc-go to 4.3.1 in response to [IBC huckleberry security advisory](https://forum.cosmos.network/t/ibc-security-advisory-huckleberry/10731)

### Misc Improvements

  * [#5129](https://github.com/osmosis-labs/osmosis/pull/5129) Relax twap record validation in init genesis to allow one of the spot prices to be non-zero when twap error is observed.

  * [#5134](https://github.com/osmosis-labs/osmosis/pull/5134) Update sdk fork with the change for correct block time in historical queries (#5134)

## v15.1.1

Same changes included in `v15.1.2` but redacted as tagged commit was not part of `v15.x` branch.

## v15.1.0

### Security

* Upgraded wasmvm to 1.1.2 in response to [CWA-2023-002](https://github.com/CosmWasm/advisories/blob/main/CWAs/CWA-2023-002.md)

### Features

* [#4829](https://github.com/osmosis-labs/osmosis/pull/4829) Add highest liquidity pool query in x/protorev
* [#4878](https://github.com/osmosis-labs/osmosis/pull/4878) Emit backrun event upon successful protorev backrun

### Misc Improvements

* [#4582](https://github.com/osmosis-labs/osmosis/pull/4582) Consistently generate build tags metadata, to return a comma-separated list without stray quotes. This affects the output from `version` CLI subcommand and server info API calls.


## v15.0.0

This release containts the following new modules:
- ProtoRev module (x/protorev). This module captures MEV via in-protocol cyclic arbitrage and distributes the revenue back to the protocol based on governance. Developed by the Skip team.
- Validator Set Preference module (x/valset-pref). This module gives users the ability to delegate to multiple validators according to their preference list.
- Pool Manager module (x/poolmanager). This module manages the infrastructure around pool creation and swaps. It serves as a unified entrypoint for any swap related message or query. This module is extracted from the pre-existing `x/gamm`. It is the first milestone on the path towards delivering concentrated liquidity.

### Features

  * [#4107](https://github.com/osmosis-labs/osmosis/pull/4107) Add superfluid unbond partial amount
  * [#4207](https://github.com/osmosis-labs/osmosis/pull/4207) Add support for Async Interchain Queries
  * [#4248](https://github.com/osmosis-labs/osmosis/pull/4248) Add panic recovery to `MultihopEstimateInGivenExactAmountOut`, `MultihopEstimateOutGivenExactAmountIn` and `RouteExactAmountOut`
  * [#3911](https://github.com/osmosis-labs/osmosis/pull/3911) Add Packet Forward Middleware
  * [#4244](https://github.com/osmosis-labs/osmosis/pull/4244) Consensus min gas fee of .0025 uosmo
  * [#4340](https://github.com/osmosis-labs/osmosis/pull/4340) Added rate limits according to: <https://www.mintscan.io/osmosis/proposals/427>
  * [#4207](https://github.com/osmosis-labs/osmosis/pull/4207) Integrate Async ICQ.

### Misc Improvements

  * [#4131](https://github.com/osmosis-labs/osmosis/pull/4141) Add GatherValuesFromStorePrefixWithKeyParser function to osmoutils.
  * [#4388](https://github.com/osmosis-labs/osmosis/pull/4388) Increase the max allowed contract size for non-proposal contracts to 3MB
  * [#4384](https://github.com/osmosis-labs/osmosis/pull/4384) migrate stXXX/XXX constant product pools 833, 817, 810 to stable swap
  * [#4461](https://github.com/osmosis-labs/osmosis/pull/4461) added rate limit quotas for a set of high value tokens
  * [#4819](https://github.com/osmosis-labs/osmosis/pull/4819) remove duplicate denom-authority-metadata query command
  * [#5028](https://github.com/osmosis-labs/osmosis/pull/5028) Change stakingTypes.Bankkeeper to simtypes.Bankkeeper

### API breaks

* [#3766](https://github.com/osmosis-labs/osmosis/pull/3766) Remove Osmosis gamm and twap `bindings` that were previously supported as custom wasm plugins.
* [#3905](https://github.com/osmosis-labs/osmosis/pull/3905) Deprecate gamm queries `NumPools`, `EstimateSwapExactAmountIn` and `EstimateSwapExactAmountOut`.
* [#3907](https://github.com/osmosis-labs/osmosis/pull/3907) Add `NumPools`, `EstimateSwapExactAmountIn` and `EstimateSwapExactAmountOut` query in poolmanager module to stargate whitelist.
* [#3880](https://github.com/osmosis-labs/osmosis/pull/3880) Switch usage of proto-generated SwapAmountInRoute and SwapAmountOutRoute in x/gamm to import the structs from x/poolmanager module.
* [#4489](https://github.com/osmosis-labs/osmosis/pull/4489) Add unlockingLockId to BeginUnlocking response.

### Bug Fix

* [#3715](https://github.com/osmosis-labs/osmosis/pull/3715) Fix x/gamm (golang API) CalculateSpotPrice, balancer.SpotPrice and Stableswap.SpotPrice base and quote asset.
* [#3746](https://github.com/osmosis-labs/osmosis/pull/3746) Make ApplyFuncIfNoErr logic preserve panics for OutOfGas behavior.
* [#4306](https://github.com/osmosis-labs/osmosis/pull/4306) Prevent adding more tokens to an already finished gauge
* [#4359](https://github.com/osmosis-labs/osmosis/pull/4359) Fix incorrect time delta due to nanoseconds in time causing twap jitter.
* [#4250](https://github.com/osmosis-labs/osmosis/pull/4250) Add denom metadata for uosmo, uion


## v14.0.1

### Bug fixes

* [#4132](https://github.com/osmosis-labs/osmosis/pull/4132) Fix CLI for EstimateSwapExactAmountIn and EstimateSwapExactAmountOut in x/gamm.
* [#4262](https://github.com/osmosis-labs/osmosis/pull/4262) Fix geometric twap genesis validation.

## v14.0.0

This release's main features are utility helpers for smart contract developers. This release contains:

- IBC composability work
  - IBC -> wasm hooks now gives sender information
  - IBC contracts can register a callback that forwards into a smart contract
  - This work is importable by external repositories, intended as an ecosystem standards
- Downtime detection tooling
  - There is now an on-chain query, allowing you to test if the chain is recovering from a downtime of a given duration.
    - The querier defines what recovering means, e.g. for a 1 hour downtime, do you consider the chain as recovering until at least 10 minutes since last 1 hr downtime? 
- Geometric TWAP
  - Every AMM pool now exposes a geometric TWAP, in addition to the existing arithmetic TWAP
* IBC features
  * Upgrade to IBC v4.2.0
* Cosmwasm
  * Upgrade to wasmd v0.30.x
* Update go build version to go 1.19

### Features

* [#2387](https://github.com/osmosis-labs/osmosis/pull/3838) Upgrade to IBC v4.2.0, and as a requirement for it wasmd to 0.30.0
* [#3609](https://github.com/osmosis-labs/osmosis/pull/3609) Add Downtime-detection module.
* [#2788](https://github.com/osmosis-labs/osmosis/pull/2788) Add logarithm base 2 implementation.
* [#3677](https://github.com/osmosis-labs/osmosis/pull/3677) Add methods for cloning and mutative multiplication on osmomath.BigDec.
* [#3676](https://github.com/osmosis-labs/osmosis/pull/3676) implement `PowerInteger` function on `osmomath.BigDec` 
* [#3678](https://github.com/osmosis-labs/osmosis/pull/3678) implement mutative `PowerIntegerMut` function on `osmomath.BigDec`.
* [#3708](https://github.com/osmosis-labs/osmosis/pull/3708) `Exp2` function to compute 2^decimal.
* [#3693](https://github.com/osmosis-labs/osmosis/pull/3693) Add `EstimateSwapExactAmountOut` query to stargate whitelist
* [#3731](https://github.com/osmosis-labs/osmosis/pull/3731) BigDec Power functions with decimal exponent.
* [#3847](https://github.com/osmosis-labs/osmosis/pull/3847) GeometricTwap and GeometricTwapToNow queries added to Stargate whitelist.
* [#3899](https://github.com/osmosis-labs/osmosis/pull/3899) Fixed osmoutils so its importable by chains that don't use the osmosis CosmosSDK fork 
  
### API breaks

* [#3763](https://github.com/osmosis-labs/osmosis/pull/3763) Move binary search and error tolerance code from `osmoutils` into `osmomath`
* [#3817](https://github.com/osmosis-labs/osmosis/pull/3817) Move osmoassert from `app/apptesting/osmoassert` to `osmoutils/osmoassert`.
* [#3771](https://github.com/osmosis-labs/osmosis/pull/3771) Move osmomath into its own go.mod
* [#3827](https://github.com/osmosis-labs/osmosis/pull/3827) Move osmoutils into its own go.mod

### Bug fixes

* [#3608](https://github.com/osmosis-labs/osmosis/pull/3608) Make it possible to state export from any directory.

## v13.1.2

Osmosis v13.1.2 is a minor patch release that includes several bug fixes and updates.

The main bug fix in this release is for the state export feature, which was not working properly in previous versions. This issue has now been resolved, and state export should work as expected in v13.1.2.

Additionally, the swagger files for v13 have been updated to improve compatibility and ensure that all API endpoints are properly documented.

### Misc Improvements

* [#3611](https://github.com/osmosis-labs/osmosis/pull/3611),[#3647](https://github.com/osmosis-labs/osmosis/pull/3647) Introduce osmocli, to automate thousands of lines of CLI boilerplate
* [#3634](https://github.com/osmosis-labs/osmosis/pull/3634) (Makefile) Ensure correct golang version in make build and make install. (Thank you @jhernandezb )
* [#3712](https://github.com/osmosis-labs/osmosis/pull/3712) replace `osmomath.BigDec` `Power` with `PowerInteger` 
* [#3711](https://github.com/osmosis-labs/osmosis/pull/3711) Use Dec instead of Int for additive `ErrTolerace` in `osmoutils`.
* [3647](https://github.com/osmosis-labs/osmosis/pull/3647), [3942](https://github.com/osmosis-labs/osmosis/pull/3942) (CLI) re-order the command line arguments for `osmosisd tx gamm join-swap-share-amount-out`

## v13.0.0

This release includes stableswap, and expands the IBC safety & composability functionality of Osmosis. The primary features are:

* Gamm:
  * Introduction of the stableswap pool type
  * Multi-hop spread factor reduction
  * Filtered queries to help front-ends
  * Adding a spot price v2 query
    * spotprice v1beta1 had baseassetdenom and quoteassetdenom backwards.
    * All contracts and integrators should switch to the v2 query from now on.
  * Adding more queries for contract developers
  * Force unpooling is now enableable by governance
* IBC features
  * Upgrade to IBC v3.4.0
  * Added IBC rate limiting, to increase safety of bridged assets
  * Allow ICS-20 to call into cosmwasm contracts
* Cosmwasm
  * Upgrade to cosmwasm v0.29.x
  * Inclusion of requested queries for contract developers

### Features

* [#2739](https://github.com/osmosis-labs/osmosis/pull/2739),[#3356](https://github.com/osmosis-labs/osmosis/pull/3356) Add pool type query, and add it to stargate whitelist
* [#2956](https://github.com/osmosis-labs/osmosis/issues/2956) Add queries for calculating amount of shares/tokens you get by providing X tokens/shares when entering/exiting a pool
* [#3217](https://github.com/osmosis-labs/osmosis/pull/3217) Add `CalcJoinPoolShares`, `CalcExitPoolCoinsFromShares`, `CalcJoinPoolNoSwapShares` to the registered Stargate queries list.
* [#3313](https://github.com/osmosis-labs/osmosis/pull/3313) Upgrade to IBC v3.4.0, allowing for IBC transfers with metadata.
* [#3335](https://github.com/osmosis-labs/osmosis/pull/3335) Add v2 spot price queries
  - The v1beta1 queries actually have base asset and quote asset reversed, so you were always getting 1/correct spot price. People fixed this by reordering the arguments.
  - This PR adds v2 queries for doing the correct thing, and giving people time to migrate from v1beta1 queries to v2.
  - It also changes cosmwasm to only allow the v2 queries, as no contracts on Osmosis mainnet uses the v1beta1 queries.

### Bug fixes

* [#2803](https://github.com/osmosis-labs/osmosis/pull/2803) Fix total pool liquidity CLI query.
* [#2914](https://github.com/osmosis-labs/osmosis/pull/2914) Remove out of gas panics from node logs
* [#2937](https://github.com/osmosis-labs/osmosis/pull/2937) End block ordering - staking after gov and module sorting.
* [#2923](https://github.com/osmosis-labs/osmosis/pull/2923) TWAP calculation now errors if it uses records that have errored previously.
* [#3312](https://github.com/osmosis-labs/osmosis/pull/3312) Add better panic catches within GAMM txs

### Misc Improvements

* [#2804](https://github.com/osmosis-labs/osmosis/pull/2804) Improve error handling and messages when parsing pool assets.
* [#3035](https://github.com/osmosis-labs/osmosis/pull/3035) Remove `PokePool` from `PoolI` interface. Define on a new WeightedPoolExtension` instead.
* [#3214](https://github.com/osmosis-labs/osmosis/pull/3214) Add basic CLI query support for TWAP.


## v12.0.0

This release includes several cosmwasm-developer and appchain-ecosystem affecting upgrades:

* TWAP - Time weighted average prices for all AMM pools
* Cosmwasm contract developer facing features
  * Enabling select queries for cosmwasm contracts
  * Add message responses to gamm messages, to remove the neccessity of bindings
  * Allow specifying denom metadata from tokenfactory
* Enabling Interchain accounts (for real this time)
* Upgrading IBC to v3.3.0
* Consistently makes authz work with ledger for all messages

The release also contains the following changes affecting Osmosis users and node operators

* Fixing State Sync
* Enabling expedited proposals

This upgrade also adds a number of safety and API boundary improving changes to the codebase.
While not state machine breaking, this release also includes the revamped Osmosis simulator,
which acts as a fuzz testing tool tailored for the SDK state machine.

### Breaking Changes

* [#2477](https://github.com/osmosis-labs/osmosis/pull/2477) Tokenfactory burn msg clash with sdk
  * TypeMsgBurn: from "burn" to "tf_burn"
  * TypeMsgMint: from "mint" to "tf_mint"
* [#2222](https://github.com/osmosis-labs/osmosis/pull/2222) Add scaling factors to MsgCreateStableswapPool
* [#1889](https://github.com/osmosis-labs/osmosis/pull/1825) Add proto responses to gamm LP messages:
  * MsgJoinPoolResponse: share_out_amount and token_in fields 
  * MsgExitPoolResponse: token_out field 
* [#1825](https://github.com/osmosis-labs/osmosis/pull/1825) Fixes Interchain Accounts (host side) by adding it to AppModuleBasics
* [#1994](https://github.com/osmosis-labs/osmosis/pull/1994) Removed bech32ibc module
* [#2016](https://github.com/osmosis-labs/osmosis/pull/2016) Add fixed 10000 gas cost for each Balancer swap
* [#2193](https://github.com/osmosis-labs/osmosis/pull/2193) Add TwapKeeper to the Osmosis app
* [#2227](https://github.com/osmosis-labs/osmosis/pull/2227) Enable charging fee in base denom for `CreateGauge` and `AddToGauge`.
* [#2283](https://github.com/osmosis-labs/osmosis/pull/2283) x/incentives: refactor `CreateGauge` and `AddToGauge` fees to use txfees denom
* [#2206](https://github.com/osmosis-labs/osmosis/pull/2283) Register all Amino interfaces and concrete types on the authz Amino codec. This will allow the authz module to properly serialize and de-serializes instances using Amino.
* [#2405](https://github.com/osmosis-labs/osmosis/pull/2405) Make SpotPrice have a max value of 2^160, and no longer be able to panic
* [#2473](https://github.com/osmosis-labs/osmosis/pull/2473) x/superfluid `AddNewSuperfluidAsset` now returns error, if any occurs instead of ignoring it.
* [#2714](https://github.com/osmosis-labs/osmosis/pull/2714) Upgrade wasmd to v0.28.0.
* Remove x/Bech32IBC
* [#3737](https://github.com/osmosis-labs/osmosis/pull/3737) Change FilteredPools MinLiquidity field from sdk.Coins struct to string.


#### Golang API breaks

* [#2160](https://github.com/osmosis-labs/osmosis/pull/2160) Clean up GAMM keeper (move `x/gamm/keeper/params.go` contents into `x/gamm/keeper/keeper.go`, replace all uses of `PoolNumber` with `PoolId`, move `SetStableSwapScalingFactors` to stableswap package, and delete marshal_bench_test.go and grpc_query_internal_test.go)
* [#1987](https://github.com/osmosis-labs/osmosis/pull/1987) Remove `GammKeeper.GetNextPoolNumberAndIncrement` in favor of the non-mutative `GammKeeper.GetNextPoolNumber`.
* [#1667](https://github.com/osmosis-labs/osmosis/pull/1673) Move wasm-bindings code out of app package into its own root level package.
* [#2013](https://github.com/osmosis-labs/osmosis/pull/2013) Make `SetParams`, `SetPool`, `SetTotalLiquidity`, and `SetDenomLiquidity` GAMM APIs private
* [#1857](https://github.com/osmosis-labs/osmosis/pull/1857) x/mint rename GetLastHalvenEpochNum to GetLastReductionEpochNum
* [#2133](https://github.com/osmosis-labs/osmosis/pull/2133) Add `JoinPoolNoSwap` and `CalcJoinPoolNoSwapShares` to GAMM pool interface and route `JoinPoolNoSwap` in pool_service.go to new method in pool interface
* [#2353](https://github.com/osmosis-labs/osmosis/pull/2353) Re-enable stargate query via whitelsit
* [#2394](https://github.com/osmosis-labs/osmosis/pull/2394) Remove unused interface methods from expected keepers of each module
* [#2390](https://github.com/osmosis-labs/osmosis/pull/2390) x/mint remove unused mintCoins parameter from AfterDistributeMintedCoin
* [#2418](https://github.com/osmosis-labs/osmosis/pull/2418) x/mint remove SetInitialSupplyOffsetDuringMigration from keeper
* [#2417](https://github.com/osmosis-labs/osmosis/pull/2417) x/mint unexport keeper `SetLastReductionEpochNum`, `getLastReductionEpochNum`, `CreateDeveloperVestingModuleAccount`, and `MintCoins`
* [#2587](https://github.com/osmosis-labs/osmosis/pull/2587) remove encoding config argument from NewOsmosisApp
x

### Features

* [#2387](https://github.com/osmosis-labs/osmosis/pull/2387) Upgrade to IBC v3.2.0, which allows for sending/receiving IBC tokens with slashes.
* [#1312] Stableswap: Createpool logic 
* [#1230] Stableswap CFMM equations
* [#1429] solver for multi-asset CFMM
* [#1539] Superfluid: Combine superfluid and staking query on querying delegation by delegator
* [#2223] Tokenfactory: Add SetMetadata functionality

### Bug Fixes

* [#2086](https://github.com/osmosis-labs/osmosis/pull/2086) `ReplacePoolIncentivesProposal` ProposalType() returns correct value of `ProposalTypeReplacePoolIncentives` instead of `ProposalTypeUpdatePoolIncentives`
* [1930](https://github.com/osmosis-labs/osmosis/pull/1930) Ensure you can't `JoinPoolNoSwap` tokens that are not in the pool
* [2186](https://github.com/osmosis-labs/osmosis/pull/2186) Remove liquidity event that was emitted twice per message.

### Improvements
* [#2515](https://github.com/osmosis-labs/osmosis/pull/2515) Emit events from functions implementing epoch hooks' `panicCatchingEpochHook` cacheCtx
* [#2526](https://github.com/osmosis-labs/osmosis/pull/2526) EpochHooks interface methods (and hence modules implementing the hooks) return error instead of panic

## v11.0.1

#### Golang API breaks
* [#1893](https://github.com/osmosis-labs/osmosis/pull/1893) Change `EpochsKeeper.SetEpochInfo` to `AddEpochInfo`, which has more safety checks with it. (Makes it suitable to be called within upgrades)
* [#2396](https://github.com/osmosis-labs/osmosis/pull/2396) x/mint remove unused mintCoins parameter from AfterDistributeMintedCoin
* [#2399](https://github.com/osmosis-labs/osmosis/pull/2399) Remove unused interface methods from expected keepers of each module
* [#2401](https://github.com/osmosis-labs/osmosis/pull/2401) Update Go import paths to v11

#### Bug Fixes
* [2291](https://github.com/osmosis-labs/osmosis/pull/2291) Remove liquidity event that was emitted twice per message
* [2288](https://github.com/osmosis-labs/osmosis/pull/2288) Fix swagger docs and swagger generation

## v11

#### Improvements
* [#2237](https://github.com/osmosis-labs/osmosis/pull/2237) Enable charging fee in base denom for `CreateGauge` and `AddToGauge`.

#### SDK Upgrades
* [#2245](https://github.com/osmosis-labs/osmosis/pull/2245) Upgrade SDK for to v0.45.0x-osmo-v9.2. Major changes:
   * Minimum deposit on proposer at submission time: <https://github.com/osmosis-labs/cosmos-sdk/pull/302>

## v10.1.1

#### Improvements
* [#2214](https://github.com/osmosis-labs/osmosis/pull/2214) Speedup epoch distribution, superfluid component

## v10.1.0

#### Bug Fixes
* [2011](https://github.com/osmosis-labs/osmosis/pull/2011) Fix bug in TokenFactory initGenesis, relating to denom creation fee param.

#### Improvements
* [#2130](https://github.com/osmosis-labs/osmosis/pull/2130) Introduce errors in mint types.
* [#2000](https://github.com/osmosis-labs/osmosis/pull/2000) Update import paths from v9 to v10.

#### Golang API breaks
* [#1937](https://github.com/osmosis-labs/osmosis/pull/1937) Change `lockupKeeper.ExtendLock` to take in lockID instead of the direct lock struct.
* [#2030](https://github.com/osmosis-labs/osmosis/pull/2030) Rename lockup keeper `ResetAllLocks` to `InitializeAllLocks` and `ResetAllSyntheticLocks` to `InitializeAllSyntheticLocks`.

#### SDK Upgrades
* [#2146](https://github.com/osmosis-labs/osmosis/pull/2146) Upgrade SDK for to v0.45.0x-osmo-v9.1. Major changes:
   * Concurrency query client option: <https://github.com/osmosis-labs/cosmos-sdk/pull/281>
   * Remove redacted message fix: <https://github.com/osmosis-labs/cosmos-sdk/pull/284>
   * Reduce commit store logs (change to Debug): <https://github.com/osmosis-labs/cosmos-sdk/pull/282>
   * Bring back the cliff vesting command: <https://github.com/osmosis-labs/cosmos-sdk/pull/272>
   * Allow ScheduleUpgrade to come from same block: <https://github.com/osmosis-labs/cosmos-sdk/pull/261>


## v10.0.1

This release contains minor CLI bug fixes.
* Restores vesting by duration command
* Fixes pagination in x/incentives module queries

## v10.0.0


## v9.0.1

### Breaking Changes

* [#1699](https://github.com/osmosis-labs/osmosis/pull/1699) Fixes bug in sig fig rounding on spot price queries for small values
* [#1671](https://github.com/osmosis-labs/osmosis/pull/1671) Remove methods that constitute AppModuleSimulation APIs for several modules' AppModules, which implemented no-ops
* [#1671](https://github.com/osmosis-labs/osmosis/pull/1671) Add hourly epochs to `x/epochs` DefaultGenesis.
* [#1665](https://github.com/osmosis-labs/osmosis/pull/1665) Delete app/App interface, instead use simapp.App
* [#1630](https://github.com/osmosis-labs/osmosis/pull/1630) Delete the v043_temp module, now that we're on an updated SDK version.

### Bug Fixes

* [1700](https://github.com/osmosis-labs/osmosis/pull/1700) Upgrade sdk fork with missing snapshot manager fix.
* [1716](https://github.com/osmosis-labs/osmosis/pull/1716) Fix secondary over-LP shares bug with uneven swap amounts in `CalcJoinPoolShares`.
* [1759](https://github.com/osmosis-labs/osmosis/pull/1759) Fix pagination filter in incentives query.
* [1698](https://github.com/osmosis-labs/osmosis/pull/1698) Register wasm snapshotter extension.
* [1931](https://github.com/osmosis-labs/osmosis/pull/1931) Add explicit check for input denoms to `CalcJoinPoolShares`

## [v9.0.0 - Nitrogen](https://github.com/osmosis-labs/osmosis/releases/tag/v9.0.0)

The Nitrogen release brings with it a number of features enabling further cosmwasm development work in Osmosis.
It including breaking changes to the GAMM API's, many developer and node operator improvements for Cosmwasm & IBC, along with new txfee and governance features. In addition to various bug fixes and code quality improvements.

#### GAMM API changes

API changes were done to enable more CFMM's to be implemented within the existing framework.
Integrators will have to update their messages and queries to adapt, please see <https://github.com/osmosis-labs/osmosis/blob/main/x/gamm/breaking_changes_notes.md>

#### Governance Changes

* [#1191](https://github.com/osmosis-labs/osmosis/pull/1191), [#1555](https://github.com/osmosis-labs/osmosis/pull/1555) Superfluid stakers now have their votes override their validators votes
* [sdk #239](https://github.com/osmosis-labs/cosmos-sdk/pull/239) Governance can set a distinct voting period for every proposal type.

#### IBC

* [#1535](https://github.com/osmosis-labs/osmosis/pull/1535) Upgrade to [IBC v3](https://github.com/cosmos/ibc-go/releases/tag/v3.0.0)
* [#1564](https://github.com/osmosis-labs/osmosis/pull/1564) Enable Interchain account host module
  * See [here](https://github.com/osmosis-labs/osmosis/blob/main/app/upgrades/v9/upgrades.go#L49-L71) for the supported messages

#### Txfees

[#1145](https://github.com/osmosis-labs/osmosis/pull/1145) Non-osmo txfees now get swapped into osmo everyday at epoch, and then distributed to stakers.

#### Cosmwasm

Upgrade from wasmd v0.23.x to [v0.27.0](https://github.com/CosmWasm/wasmd/releases/tag/v0.27.0). This has the following features:
  * State sync now works for cosmwasm state
  * Cosmwasm builds on M1 macs
  * Many security fixes

The TokenFactory module is added to the chain, making it possible for users and contracts to make new native tokens.
Cosmwasm bindings have been added, to make swapping and creating these new tokens easier within the contract ecosystem.

* [#1640](https://github.com/osmosis-labs/osmosis/pull/1640) fix: localosmosis to work for testing cosmwasm contracts

### Other Features

* [#1629](https://github.com/osmosis-labs/osmosis/pull/1629) Fix bug in the airdrop claim script
* [#1570](https://github.com/osmosis-labs/osmosis/pull/1570) upgrade sdk with app version fix for state-sync
* [#1554](https://github.com/osmosis-labs/osmosis/pull/1554) local dev environment
* [#1541](https://github.com/osmosis-labs/osmosis/pull/1541) Add arm64 support to Docker
* [#1535](https://github.com/osmosis-labs/osmosis/pull/1535) upgrade wasmd to v0.27.0.rc3-osmo and ibc-go to v3
  * State sync now works for cosmwasm state
  * Cosmwasm builds on M1 macs
* [#1435](https://github.com/osmosis-labs/osmosis/pull/1435) `x/tokenfactory` create denom fee for spam resistance 
* [#1253](https://github.com/osmosis-labs/osmosis/pull/1253) Add a message to increase the duration of a bonded lock.
* [#1656](https://github.com/osmosis-labs/osmosis/pull/1656) Change camelCase to snake_case in proto.
* [#1632](https://github.com/osmosis-labs/osmosis/pull/1632) augment SuperfluidDelegationsByDelegator query, return osmo equivilent is staked via superfluid
* [#1723](https://github.com/osmosis-labs/osmosis/pull/1723) fix number of LP shares returned from stableswap pool

## [v8.0.0 - Emergency proposals upgrade](https://github.com/osmosis-labs/osmosis/releases/tag/v8.0.0)

This upgrade is a patch that must be hard forked in, as on-chain governance of Osmosis approved proposal [227](https://www.mintscan.io/osmosis/proposals/227) and proposal [228](https://www.mintscan.io/osmosis/proposals/228).

This upgrade includes:

* Adding height-gated AnteHandler message filter to filter unpooling tx pre-upgrade.
* At block height 4402000 accelerates prop 225, which in turn moves incentives from certain pools according to props 222-224
* Adds a msg allowing unpooling of UST pools. 
  * This procedure is initiated by whitelisting pools 560, 562, 567, 578, 592, 610, 612, 615, 642, 679, 580, 635. 
  * Unpooling allows exiting whitelisted pools directly, finish unbonding duration with the exited tokens instead of having to wait unbonding duration to swap LP shares back to collaterals. 
  * This procedure also includes locks that were already unbonding pre-upgrade and locks that were superfluid delegated.

Every node should upgrade their software version to v8.0.0 before the upgrade block height 4402000. If you use cosmovisor, simply swap out the binary at upgrades/v7/bin to be v8.0.0, and restart the node. Do check cosmovisor version returns v8.0.0

### Features 
* {Across many PRs} Initiate emergency upgrade 
* [#1481] Emergency upgrade as of prop [226] (<https://www.mintscan.io/osmosis/proposals/226>) 
* [#1482] Checking Whitelisted Pools contain UST 
* [#1486] Update whitelisted pool IDs
* [#1262] Add a forceprune command to the binaries, that prunes golevelDB data better
* [#1154] Database stability improvements
* [#840] Move lock.go functions into iterator.go, lock_refs.go and store.go
* [#916] And a fn for Unbond and Burn tokens
* [#908] Superfluid slashing code
* [#904] LockAndSuperfluidDelegate

### Minor improvements & Bug Fixes

* [#1428] fix: pool params query (backport #1315)
* [#1390] upgrade sdk to v0.45.0x-osmo-v7.9
* [#1087] Test improvisation for Superfluid (backport #1070)
* [#1022] upgrade iavl to v0.17.3-osmo-v4

### Features

* [#1378](https://github.com/osmosis-labs/osmosis/pull/1378) add .gitpod.yml
* [#1262](https://github.com/osmosis-labs/osmosis/pull/1262) Add a `forceprune` command to the binaries, that prunes golevelDB data better.
* [#1244](https://github.com/osmosis-labs/osmosis/pull/1244) Refactor `x/gamm`'s `ExitSwapExternAmountOut`.
* [#1107](https://github.com/osmosis-labs/osmosis/pull/1107) Update to wasmvm v0.24.0, re-enabling building on M1 macs!
* [#1292](https://github.com/osmosis-labs/osmosis/pull/1292) CLI account-locked-duration

### Minor improvements & Bug Fixes

* [#1442](https://github.com/osmosis-labs/osmosis/pull/1442) Use latest tm-db release for badgerdb and rocksdb improvments
* [#1379](https://github.com/osmosis-labs/osmosis/pull/1379) Introduce `Upgrade` and `Fork` structs, to simplify upgrade logic.
* [#1363](https://github.com/osmosis-labs/osmosis/pull/1363) Switch e2e test setup to create genesis and configs via Dockertest
* [#1335](https://github.com/osmosis-labs/osmosis/pull/1335) Add utility for deriving total orderings from partial orderings.
* [#1308](https://github.com/osmosis-labs/osmosis/pull/1308) Make panics inside of epochs no longer chain halt by default.
* [#1286](https://github.com/osmosis-labs/osmosis/pull/1286) Fix release build scripts.
* [#1203](https://github.com/osmosis-labs/osmosis/pull/1203) cleanup Makefile and ci workflows
* [#1177](https://github.com/osmosis-labs/osmosis/pull/1177) upgrade to go 1.18
* [#1193](https://github.com/osmosis-labs/osmosis/pull/1193) Setup e2e tests on a single chain; add balances query test
* [#1095](https://github.com/osmosis-labs/osmosis/pull/1095) Fix authz being unable to use lockup & superfluid types.
* [#1105](https://github.com/osmosis-labs/osmosis/pull/1105) Add GitHub Actions to automatically push the osmosis Docker image
* [#1114](https://github.com/osmosis-labs/osmosis/pull/1114) Improve CI: remove duplicate runs of test worflow
* [#1127](https://github.com/osmosis-labs/osmosis/pull/1127) Stricter Linting:  bump golangci-lint version and enable additional linters.
* [#1184](https://github.com/osmosis-labs/osmosis/pull/1184) Fix endtime event output on BeginUnlocking

## [v7.1.0](https://github.com/osmosis-labs/osmosis/releases/tag/v7.1.0)

### Minor improvements & Bug Fixes

* [#1052](https://github.com/osmosis-labs/osmosis/pull/1052) Eugen/cherry pick superfluid test scaffolding updates
* [#1070](https://github.com/osmosis-labs/osmosis/pull/1070) Test improvisation for Superfluid
* [#1084](https://github.com/osmosis-labs/osmosis/pull/1084) Superfluid Misc: Improve grpc_query
* [#1081](https://github.com/osmosis-labs/osmosis/pull/1081) Genesis upgrade and add invariant cherry pick
* [#1088](https://github.com/osmosis-labs/osmosis/pull/1088) Genesis import export check for superfluid
* [#1101](https://github.com/osmosis-labs/osmosis/pull/1101) Minor PR adding some code comments
* [#1154](https://github.com/osmosis-labs/osmosis/pull/1154) Database stability improvements

### SDK fork updates

* [sdk-#136](https://github.com/osmosis-labs/iavl/pull/136) add after validator slash hook
* [sdk-#137](https://github.com/osmosis-labs/iavl/pull/137) backport feat: Modify grpc gateway to be concurrent
* [sdk-#146](https://github.com/osmosis-labs/cosmos-sdk/pull/146) extra logs during commit
* [sdk-#151](https://github.com/osmosis-labs/cosmos-sdk/pull/151) fix logs related to store keys and commit hash
* [sdk-#140](https://github.com/osmosis-labs/cosmos-sdk/pull/140) refactor: snapshot and pruning functionality
* [sdk-#156](https://github.com/osmosis-labs/cosmos-sdk/pull/156) feat: implement querying for commit hash and proofs
* [sdk-#155](https://github.com/osmosis-labs/cosmos-sdk/pull/155) fix: commit info data race
* [sdk-#158](https://github.com/osmosis-labs/cosmos-sdk/pull/158) Fixes the go race tests
* [sdk-#160](https://github.com/osmosis-labs/cosmos-sdk/pull/160) increase setupBaseAppWithSnapshots timeout to 90 seconds
* [sdk-#161](https://github.com/osmosis-labs/cosmos-sdk/pull/155) upgrade iavl to v0.17.3-osmo-v7 with lowered fast node cache size

### IAVL fork updates

* [iavl-35](https://github.com/osmosis-labs/iavl/pull/35) avoid clearing fast node cache during pruning
* [iavl-36](https://github.com/osmosis-labs/iavl/pull/36) fix data race related to VersionExists
* [iavl-37](https://github.com/osmosis-labs/iavl/pull/36) hardcode fast node cache size to 100k

## [v7.0.4](https://github.com/osmosis-labs/osmosis/releases/tag/v7.0.4)

### Minor improvements & Bug Fixes

* [#1061](https://github.com/osmosis-labs/osmosis/pull/1061) upgrade iavl to v0.17.3-osmo-v5 with concurrent map write fix
* [#1071](https://github.com/osmosis-labs/osmosis/pull/1071) improve Dockerfile

### SDK fork updates

* [sdk-#135](https://github.com/osmosis-labs/cosmos-sdk/pull/135) upgrade iavl to v0.17.3-osmo-v5 with concurrent map write fix

### IAVL fork updates

* [iavl-34](https://github.com/osmosis-labs/iavl/pull/34) fix concurrent map panic when querying and committing

## [v7.0.3](https://github.com/osmosis-labs/osmosis/releases/tag/v7.0.3)

### Minor improvements & Bug Fixes

* [#1022](https://github.com/osmosis-labs/osmosis/pull/1022) upgrade iavl to v0.17.3-osmo-v4 - fix state export at an old height
* [#988](https://github.com/osmosis-labs/osmosis/pull/988) Make `SuperfluidUndelegationsByDelegator` query also return synthetic locks
* [#984](https://github.com/osmosis-labs/osmosis/pull/984) Add wasm support to Dockerfile

## [v7.0.2 - Carbon](https://github.com/osmosis-labs/osmosis/releases/tag/v7.0.2)

This release fixes an instance of undefined behaviour present in v7.0.0.
Parts of the code use a function called [`ApplyFuncIfNoErr`]() whose purpose is to catch errors, and if found undo state updates during its execution.
It is intended to also catch panics and undo the problematic code's execution.
Right now a panic in this code block would halt the node, as it would not know how to proceed.
(But no state change would be committed)

## [v7.0.0 - Carbon](https://github.com/osmosis-labs/osmosis/releases/tag/v7.0.0)

The Osmosis Carbon Release! The changes are primarily

The large features include:

* Superfluid Staking - Allowing LP shares be staked to help secure the network
* Adding permissioned cosmwasm to the chain
* IAVL speedups, greatly improving epoch and query performance
* Local mempool filters to charge higher gas for arbitrage txs
* Allow partial unlocking of non-superfluid'd locks

Upgrade instructions for node operators can be found [here](https://github.com/osmosis-labs/osmosis/blob/main/networks/osmosis-1/upgrades/v7/guide.md)

The v7 release introduces Superfluid Staking! This allows governance-approved LP shares to be staked to help secure the network.

### Features

* {Across many PRs} Add superfluid staking
* [#893](https://github.com/osmosis-labs/osmosis/pull/893/) Allow (non-superfluid'd) locks to be partially unlocked.
* [#828](https://github.com/osmosis-labs/osmosis/pull/828) Move docs to their own repository, <https://github.com/osmosis-labs/docs>
* [#804](https://github.com/osmosis-labs/osmosis/pull/804/) Make the Osmosis repo use proper golang module versioning in self-package imports. (Enables other go projects to easily import Osmosis tags)
* [#782](https://github.com/osmosis-labs/osmosis/pull/782) Upgrade to cosmos SDK v0.45.0
* [#777](https://github.com/osmosis-labs/osmosis/pull/777) Add framework for mempool filters for charging different gas rates, add mempool filter for higher gas txs.
* [#772](https://github.com/osmosis-labs/osmosis/pull/772) Fix SDK bug where incorrect sequence number txs wouldn't get removed from blocks.
* [#769](https://github.com/osmosis-labs/osmosis/pull/769/) Add governance permissioned cosmwasm module
* [#680](https://github.com/osmosis-labs/osmosis/pull/680/),[#697](https://github.com/osmosis-labs/osmosis/pull/697/) Change app.go file structure to mitigate risk of keeper reference vs keeper struct bugs. (What caused Osmosis v5 -> v6)

### Minor improvements & Bug Fixes

* [#924](https://github.com/osmosis-labs/osmosis/pull/923) Fix long standing problems with total supply query over-reporting the number of osmo.
* [#872](https://github.com/osmosis-labs/osmosis/pull/872) Add a helper for BeginBlock/EndBlock code to have code segments that atomically revert state if any part errors.
* [#869](https://github.com/osmosis-labs/osmosis/pull/869) Update Dockerfile to use distroless base image.
* [#855](https://github.com/osmosis-labs/osmosis/pull/855) Ensure gauges can only be created for assets that exist on chain.
* [#766](https://github.com/osmosis-labs/osmosis/pull/766) Consolidate code between InitGenesis and CreateGauge
* [#763](https://github.com/osmosis-labs/osmosis/pull/763) Add rocksDB options to Makefile.
* [#740](https://github.com/osmosis-labs/osmosis/pull/740) Simplify AMM swap math / file structure.
* [#731](https://github.com/osmosis-labs/osmosis/pull/731) Add UpdateFeeToken proposal handler to app.go
* [#686](https://github.com/osmosis-labs/osmosis/pull/686) Add silence usage to cli to surpress unnecessary help logs
* [#652](https://github.com/osmosis-labs/osmosis/pull/652) Add logic for deleting a pool
* [#541](https://github.com/osmosis-labs/osmosis/pull/541) Start generalizing the AMM infrastructure

### SDK fork updates

* [sdk-#119](https://github.com/osmosis-labs/cosmos-sdk/pull/119) Add bank supply offsets to let applications have some minted tokens not count in total supply.
* [sdk-#117](https://github.com/osmosis-labs/cosmos-sdk/pull/117) Add an instant undelegate method to staking, for use in superfluid.
* [sdk-#116](https://github.com/osmosis-labs/cosmos-sdk/pull/116) Fix the slashing hooks to be correct.
* [sdk-#108](https://github.com/osmosis-labs/cosmos-sdk/pull/108) upgrade to IAVL fast storage on v0.45.0x-osmo-v7-fast

### Wasmd fork updates

* [wasmd-v.022.0-osmo-v7.2](https://github.com/osmosis-labs/wasmd/releases/tag/v0.22.0-osmo-v7.2) Upgrade SDK and IAVL dependencies to use fast storage

## [v6.4.0](https://github.com/osmosis-labs/osmosis/releases/tag/v6.4.0)

### Minor improvements & Bug Fixes

-[#907](https://github.com/osmosis-labs/osmosis/pull/907) Upgrade IAVL and SDK with RAM improvements and bug fixes for v6.4.0

### SDK fork updates

* [sdk-#114](https://github.com/osmosis-labs/cosmos-sdk/pull/114) upgrading iavl with ram optimizations during migration, and extra logs and fixes for "version X was already saved to a different hash" and "insufficient funds" bugs

### IAVL fork updates

* [iavl-19](https://github.com/osmosis-labs/iavl/pull/19) force GC, no cache during migration, auto heap profile

## [v6.3.1](https://github.com/osmosis-labs/osmosis/releases/tag/v6.3.1)

* [#859](https://github.com/osmosis-labs/osmosis/pull/859) CLI, update default durations to be in better units.

* [#Unknown](https://github.com/osmosis-labs/osmosis/commit/3bf63f1d3b7efee503106a008e84129489bdba8d) Switch to SDK branch with vesting by duration

## Minor improvements & Bug Fixes

* [#795](https://github.com/osmosis-labs/osmosis/pull/795) Annotate app.go
* [#791](https://github.com/osmosis-labs/osmosis/pull/791) Change to dependabot config to only upgrade patch version of tendermint
* [#766](https://github.com/osmosis-labs/osmosis/pull/766) Consolidate code between InitGenesis and CreateGauge

## [v6.3.0](https://github.com/osmosis-labs/osmosis/releases/tag/v6.3.0)

## Features

* [#845](https://github.com/osmosis-labs/osmosis/pull/846) Upgrade iavl and sdk with fast storage
* [#724](https://github.com/osmosis-labs/osmosis/pull/724) Make an ante-handler filter for recognizing High gas txs, and having a min gas price for them.

## Minor improvements & Bug Fixes

* [#795](https://github.com/osmosis-labs/osmosis/pull/795) Annotate app.go
* [#791](https://github.com/osmosis-labs/osmosis/pull/791) Change to dependabot config to only upgrade patch version of tendermint
* [#766](https://github.com/osmosis-labs/osmosis/pull/766) Consolidate code between InitGenesis and CreateGauge

### SDK fork updates

* [sdk-#100](https://github.com/osmosis-labs/cosmos-sdk/pull/100) Upgrade iavl with fast storage

### IAVL fork updates

* [iavl-5](https://github.com/osmosis-labs/iavl/pull/5) Fast storage optimization for queries and iterations

## [v6.2.0](https://github.com/osmosis-labs/osmosis/releases/tag/v6.2.0)

### SDK fork updates

* [sdk-#58](https://github.com/osmosis-labs/cosmos-sdk/pull/58) Fix a bug where recheck would not remove txs with invalid sequence numbers

## Minor improvements & Bug Fixes

* [#765](https://github.com/osmosis-labs/osmosis/pull/765) Fix a bug in `Makefile` regarding the location of localtestnet docker image.

## [v6.1.0](https://github.com/osmosis-labs/osmosis/releases/tag/v6.1.0)

## Features

* Update to Tendermint v0.34.15
* Increase p2p timeouts to alleviate p2p network breaking at epoch
* [#741](https://github.com/osmosis-labs/osmosis/pull/741) Allow node operators to set a second min gas price for arbitrage txs.
* [#623](https://github.com/osmosis-labs/osmosis/pull/623) Use gosec for staticly linting for common non-determinism issues in SDK applications.

## Minor improvements & Bug Fixes

* [#722](https://github.com/osmosis-labs/osmosis/issues/722) reuse code for parsing integer slices from string
* [#704](https://github.com/osmosis-labs/osmosis/pull/704) fix rocksdb
* [#666](https://github.com/osmosis-labs/osmosis/pull/666) Fix the `--log-level` and `--log-format` commands on `osmosisd start`
* [#655](https://github.com/osmosis-labs/osmosis/pull/655) Make the default genesis for pool-incentives work by default
* [97ac2a8](https://github.com/osmosis-labs/osmosis/commit/97ac2a86303fc8966a4c169107e0945775107e67) Fix InitGenesis bug for gauges

### SDK fork updates

* [sdk-#52](https://github.com/osmosis-labs/cosmos-sdk/pull/52) Fix inconsistencies in default pruning config, and change defaults. Fix pruning=everything defaults.
  * previously default was actually keeping 3 weeks of state, and every 100th state. (Not that far off from archive nodes)
  * pruning=default now changed to 1 week of state (100k blocks), and keep-every=0. (So a constant number of states stored)
  * pruning=everything now stores the last 10 states, to avoid db corruption errors plaguing everyone who used it. This isn't a significant change, because the pruning interval was anyways 10 blocks, so your node had to store 10 blocks of state anyway.
* [sdk-#51](https://github.com/osmosis-labs/cosmos-sdk/pull/51) Add hooks for superfluid staking
* [sdk-#50](https://github.com/osmosis-labs/cosmos-sdk/pull/50) Make it possible to better permission the bank keeper's minting ability

## [v6.0.0](https://github.com/osmosis-labs/osmosis/releases/tag/v6.0.0)

This upgrade fixes a bug in the v5.0.0 upgrade's app.go, which prevents new IBC channels from being created.
This binary is compatible with v5.0.0 until block height `2464000`, estimated to be at 4PM UTC Monday December 20th.

* [Patch](https://github.com/osmosis-labs/osmosis/commit/907001b08686ed980e0afa3d97a9c5e2f095b79f#diff-a172cedcae47474b615c54d510a5d84a8dea3032e958587430b413538be3f333) - Revert back to passing in the correct staking keeper into the IBC keeper constructor.
* [Height gating change](https://github.com/osmosis-labs/ibc-go/pull/1) - Height gate the change in IBC, to make the v6.0.0 binary compatible until upgrade height.

## [v5.0.0](https://github.com/osmosis-labs/osmosis/releases/tag/v5.0.0) - Boron upgrade

The Osmosis Boron release is made!

Notable features include:

* Upgrading from SDK v0.42 to [SDK v0.44](https://github.com/cosmos/cosmos-sdk/blob/v0.43.0/RELEASE_NOTES.md), bringing efficiency improvements, integrations and Rosetta support.
* Bringing in the new modules [Bech32IBC](https://github.com/osmosis-labs/bech32-ibc/), [Authz](https://github.com/cosmos/cosmos-sdk/tree/master/x/authz/spec), [TxFees](https://github.com/osmosis-labs/osmosis/tree/main/x/txfees)
* Upgrading to IBC v2, allowing for improved Ethereum Bridge and CosmWasm support
* Implementing Osmosis chain governance's [Proposal 32](https://www.mintscan.io/osmosis/proposals/32)
* Large suite of gas bugs fixed. (Including several that we have not seen on chain)
* More queries exposed to aid node operators.
* Blocking the OFAC banned Ethereum addresses.
* Several (linear factor) epoch time improvements. (Most were present in v4.2.0)

Upgrade instructions for node operators can be found [here](https://github.com/osmosis-labs/osmosis/blob/v5.x/networks/osmosis-1/upgrades/v5/guide.md)

## Features

* [\#637](https://github.com/osmosis-labs/osmosis/pull/637) Add [Bech32IBC](https://github.com/osmosis-labs/bech32-ibc/)
* [\#610](https://github.com/osmosis-labs/osmosis/pull/610) Upgrade to Cosmos SDK v0.44.x
  * Numerous large updates, such as making module accounts be 32 bytes, Rosetta support, etc.
  * Adds & integrates the [Authz module](https://github.com/cosmos/cosmos-sdk/tree/master/x/authz/spec)
    See: [SDK v0.43.0 Release Notes](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.43.0) For more details
* [\#610](https://github.com/osmosis-labs/osmosis/pull/610) Upgrade to IBC-v2
* [\#560](https://github.com/osmosis-labs/osmosis/pull/560) Implements Osmosis [prop32](https://www.mintscan.io/osmosis/proposals/32) -- clawing back the final 20% of unclaimed osmo and ion airdrop.
* [\#394](https://github.com/osmosis-labs/osmosis/pull/394) Allow whitelisted tx fee tokens based on conversion rate to OSMO
* [Commit db450f0](https://github.com/osmosis-labs/osmosis/commit/db450f0dce8c595211d920f9bca7ed0f3a136e43) Add blocking of OFAC banned Ethereum addresses

## Minor improvements & Bug Fixes

* {In the Osmosis-labs SDK fork}
  * Increase default IAVL cache size to be in the hundred megabyte range
  * Significantly improve CacheKVStore speed problems, reduced IBC upgrade time from 2hrs to 5min
  * Add debug info to make it clear whats happening during upgrade
* (From a series of commits) Fixes to the claims module to only do the reclaim logic once, not every block.
* (From a series of commits) More logging to the claims module.
* [\#563](https://github.com/osmosis-labs/osmosis/pull/563) Allow zero-weight pool-incentive distribution records
* [\#562](https://github.com/osmosis-labs/osmosis/pull/562) Store block height in epochs module for easier debugging
* [\#544](https://github.com/osmosis-labs/osmosis/pull/544) Update total liquidity tracking to be denom basis, lowering create pool and join pool gas.
* [\#540](https://github.com/osmosis-labs/osmosis/pull/540) Fix git lfs links
* [\#517](https://github.com/osmosis-labs/osmosis/pull/517) Linear time improvement for epoch time
* [\#515](https://github.com/osmosis-labs/osmosis/pull/515) Add debug command for converting secp pubkeys
* [\#510](https://github.com/osmosis-labs/osmosis/pull/510) Performance improvement for gauge distribution
* [\#505](https://github.com/osmosis-labs/osmosis/pull/505) Fix bug in incentives epoch distribution events, used to use raw address, now uses bech32 addr
* [\#464](https://github.com/osmosis-labs/osmosis/pull/464) Increase maximum outbound peers for validator nodes
* [\#444](https://github.com/osmosis-labs/osmosis/pull/444) Add script for state sync
* [\#409](https://github.com/osmosis-labs/osmosis/pull/409) Reduce epoch time growth rate for re-locking assets

## [v4.0.0]

* Significantly speedup epoch times
* Fix bug in the lockup module code that caused it to take a linear amount of gas.
* Make unbonding tokens from the lockup module get automatically claimed when unbonding is done.
* Add events for all tx types in the gamm module.
* Add events for adding LP rewards.
* Make queries to bank total chain balance account for developer vesting correctly.
* Add ability for nodes to query the total amount locked for each denomination.
* Embedded seeds in init.go
* Added changelog and info about changelog format.
* Fix accumulation store only counting bonded tokens, not unbonding tokens, that prevented the front-end from using more correct APY estimates. (Previously, the front-end could only underestimate rewards)

## [v3.2.0](https://github.com/osmosis/osmosis-labs/releases/tag/v2.0.0) - 2021-06-28

* Update the cosmos-sdk version we modify to v0.42.9
* Fix a bug in the min commission rate code that allows validators to be created with commission rates less than the minimum.
* Automatically upgrade any validator with less than the minimum comission rate to the minimum at upgrade time.
* Unbrick on-chain governance, by fixing the deposit parameter to use `uosmo` instead of `osmo`.

## [v1.0.2](https://github.com/osmosis/osmosis-labs/releases/tag/v1.0.2) - 2021-06-18

This release improves the CLI UX of creating and querying gauges.

## [v1.0.1](https://github.com/osmosis/osmosis-labs/releases/tag/v1.0.1) - 2021-06-17

This release fixes a bug in `osmosisd version` always displaying 0.0.1.

## [v1.0.0](https://github.com/osmosis/osmosis-labs/releases/tag/v1.0.0) - 2021-06-16

Initial Release!
