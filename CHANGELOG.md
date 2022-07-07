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

## Unreleased

### Breaking Changes


* [#1889](https://github.com/osmosis-labs/osmosis/pull/1825) Add proto responses to gamm LP messages:
  * MsgJoinPoolResponse: share_out_amount and token_in fields 
  * MsgExitPoolResponse: token_out field 
* [#1825](https://github.com/osmosis-labs/osmosis/pull/1825) Fixes Interchain Accounts (host side) by adding it to AppModuleBasics
* [#1699](https://github.com/osmosis-labs/osmosis/pull/1699) Fixes bug in sig fig rounding on spot price queries for small values

#### Golang API breaks

* [#1937](https://github.com/osmosis-labs/osmosis/pull/1937) Change `lockupKeeper.ExtendLock` to take in lockID instead of the direct lock struct.
* [#1893](https://github.com/osmosis-labs/osmosis/pull/1893) Change `EpochsKeeper.SetEpochInfo` to `AddEpochInfo`, which has more safety checks with it. (Makes it suitable to be called within upgrades)
* [#1671](https://github.com/osmosis-labs/osmosis/pull/1671) Remove methods that constitute AppModuleSimulation APIs for several modules' AppModules, which implemented no-ops
* [#1671](https://github.com/osmosis-labs/osmosis/pull/1671) Add hourly epochs to `x/epochs` DefaultGenesis.
* [#1665](https://github.com/osmosis-labs/osmosis/pull/1665) Delete app/App interface, instead use simapp.App
* [#1630](https://github.com/osmosis-labs/osmosis/pull/1630) Delete the v043_temp module, now that we're on an updated SDK version.
* [#1667](https://github.com/osmosis-labs/osmosis/pull/1673) Move wasm-bindings code out of app .

### Features

* [#1312] Stableswap: Createpool logic 
* [#1230] Stableswap CFMM equations
* [#1429] solver for multi-asset CFMM
* [#1539] Superfluid: Combine superfluid and staking query on querying delegation by delegator

### Bug Fixes

* [1930](https://github.com/osmosis-labs/osmosis/pull/1930) Ensure you can't `JoinPoolNoSwap` tokens that are not in the pool
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
Integrators will have to update their messages and queries to adapt, please see https://github.com/osmosis-labs/osmosis/blob/main/x/gamm/breaking_changes_notes.md

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
* [#1481] Emergency upgrade as of prop [226] (https://www.mintscan.io/osmosis/proposals/226) 
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
