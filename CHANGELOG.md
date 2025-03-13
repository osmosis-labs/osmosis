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

### State Breaking

* [#8996](https://github.com/osmosis-labs/osmosis/pull/8996) chore: upgrade pfm to v8.1.1 
* [#8880](https://github.com/osmosis-labs/osmosis/pull/8880) chore: bump wasmd from 0.53.0 to 0.53.2
* [#9005](https://github.com/osmosis-labs/osmosis/pull/9005) fix: bump ibc-go to v8.6.1
* [#9011](https://github.com/osmosis-labs/osmosis/pull/9011) feat: Delete superfluid balancer -> CL migration code. Makes the message now just return an error if called.

### State Compatible

* [#8985](https://github.com/osmosis-labs/osmosis/pull/8985) build: block-sdk import now uses forked version
* [#9006](https://github.com/osmosis-labs/osmosis/pull/9006) feat: CosmWasm Pool raw state query
* [#9021](https://github.com/osmosis-labs/osmosis/pull/9021) chore: slightly increase blocktime 
* [#9050](https://github.com/osmosis-labs/osmosis/pull/9050) chore: bump ibc-go to v8.7.0 
* [#9061](https://github.com/osmosis-labs/osmosis/pull/9061) fix: split-route-swap-exact-amount-in/out cli commands 

## v28.0.4

### State Breaking

### State Compatible

* [#8978](https://github.com/osmosis-labs/osmosis/pull/8978) fix: bump comebft to v0.38.17
* [#8985](https://github.com/osmosis-labs/osmosis/pull/8985) build: block-sdk import now uses forked version
* [#8983](https://github.com/osmosis-labs/osmosis/pull/8983) fix: forceprune logic

## v28.0.3

### State Breaking

### State Compatible

* [#8968](https://github.com/osmosis-labs/osmosis/pull/8968) fix: use forked block-sdk patch
* [#8895](https://github.com/osmosis-labs/osmosis/pull/8895) (xcs) feat: support XCS final hop memos

## v28.0.2

### State Breaking

### State Compatible

* [#8940](https://github.com/osmosis-labs/osmosis/pull/8940) chore: Bump store dependency version to v1.1.1-v0.50.11-v28-osmo-2
* [#8886](https://github.com/osmosis-labs/osmosis/pull/8886) chore: decouple Node releases from SQS

### store v1.1.1-v0.50.11-v28-osmo-2

* [#1026](https://github.com/cosmos/iavl/pull/1026) perf: remove duplicated rootkey fetch inpruning (9% pruning speedup on osmosis)

## v28.0.1

### State Breaking

### State Compatible

* [#8858](https://github.com/osmosis-labs/osmosis/pull/8858) chore: fix event emission for smart account module
* [#8906](https://github.com/osmosis-labs/osmosis/pull/8906) chore: bump cosmos-sdk v0.50.11
* [#8923](https://github.com/osmosis-labs/osmosis/pull/8923) fix: tag osmosis-labs/cosmos-sdk store v1.1.1 to enable async pruning

## v28.0.0

### State Breaking

### State Compatible

* [#8857](https://github.com/osmosis-labs/osmosis/pull/8857) chore: update math package

## v27.0.1

### State Compatible

* [#8831](https://github.com/osmosis-labs/osmosis/pull/8831) chore: bump cometbft
* [#8876](https://github.com/osmosis-labs/osmosis/pull/8876) (xcs) fix: XCS incorrect channel
* [#8922](https://github.com/osmosis-labs/osmosis/pull/8922) fix: untracked tokenfactory ibc-rate-limit inflow


## v27.0.0

### State Breaking

* [#8682](https://github.com/osmosis-labs/osmosis/pull/8682) chore: bump cosmwasm-optimizer
* [#8734](https://github.com/osmosis-labs/osmosis/pull/8734) chore: update cosmwasm vm
* [#8777](https://github.com/osmosis-labs/osmosis/pull/8777) fix: state export for gov module constitution
* [#8751](https://github.com/osmosis-labs/osmosis/pull/8751) fix: supply offsets for osmo token
* [#8764](https://github.com/osmosis-labs/osmosis/pull/8764) chore: add cosmwasm 1_3 feature
* [#8779](https://github.com/osmosis-labs/osmosis/pull/8779) chore: bump cometbft/cosmos-sdk versions
* [#8801](https://github.com/osmosis-labs/osmosis/pull/8801) chore: update tagged submodules for v27
* [#8807](https://github.com/osmosis-labs/osmosis/pull/8807) chore: bump cometbft version
* [#8818](https://github.com/osmosis-labs/osmosis/pull/8818) chore: bump sqsdomain to v0.27.0

### Config


### State Compatible

* [#8754](https://github.com/osmosis-labs/osmosis/pull/8754) Add missing proto files for indexing
* [#8563](https://github.com/osmosis-labs/osmosis/pull/8563) Add additional queries in x/gauges
* [#8726](https://github.com/osmosis-labs/osmosis/pull/8726) fix: multiple temp directories on command executions
* [#8731](https://github.com/osmosis-labs/osmosis/pull/8731) fix: in place testnet logs
* [#8728](https://github.com/osmosis-labs/osmosis/pull/8728) fix unsupported sign-mode issue
* [#8743](https://github.com/osmosis-labs/osmosis/pull/8743) chore: bump sdk and cometbft
* [#8765, 8768](https://github.com/osmosis-labs/osmosis/pull/8765) fix concurrency issue in go test(x/lockup)
* [#8563](https://github.com/osmosis-labs/osmosis/pull/8755) [x/concentratedliquidity]: Fix Incorrect Event Emission
* [#8765](https://github.com/osmosis-labs/osmosis/pull/8765) fix concurrency issue in go test(x/lockup)
* [#8791](https://github.com/osmosis-labs/osmosis/pull/8791) fix: superfluid log for error that should be ignored

### State Machine Breaking

* [#8732](https://github.com/osmosis-labs/osmosis/pull/8732) fix: iterate delegations continue instead of erroring

## v26.0.1

### State Machine Breaking

* [#8732](https://github.com/osmosis-labs/osmosis/pull/8732) fix: iterate delegations continue instead of erroring

## v26.0.0

### State Breaking

* [#8274](https://github.com/osmosis-labs/osmosis/pull/8274) SDK v50 and Comet v0.38 upgrade
* [#8375](https://github.com/osmosis-labs/osmosis/pull/8375) Enforce sub-authenticator to be greater than 1
* [#8509](https://github.com/osmosis-labs/osmosis/pull/8509) Change LiquidityNetInDirection return type to sdk math
* [#8535](https://github.com/osmosis-labs/osmosis/pull/8535) Prevent Setting Invalid Before Send Hook
* [#8310](https://github.com/osmosis-labs/osmosis/pull/8310) Taker fee share
* [#8494](https://github.com/osmosis-labs/osmosis/pull/8494) Add additional events in x/lockup, x/superfluid, x/concentratedliquidity
* [#8581](https://github.com/osmosis-labs/osmosis/pull/8581) feat: add ledger signing to smart account module 
* [#8573](https://github.com/osmosis-labs/osmosis/pull/8573) fix: increase unauthenticated gas to fix fee token issue
* [#8598](https://github.com/osmosis-labs/osmosis/pull/8598) feat: param changes for block and cost per byte 
* [#8609](https://github.com/osmosis-labs/osmosis/pull/8609) Exempt `UnrestrictedPoolCreatorWhitelist` addresses from pool creation fee
* [#8615](https://github.com/osmosis-labs/osmosis/pull/8615) chore: add tagged cosmos-sdk version: v0.50.6-v26-osmo-1
* [#8616](https://github.com/osmosis-labs/osmosis/pull/8616) chore: upgrade wasmd to v0.53.0 and wasmvm to v2.1.2
* [#8628](https://github.com/osmosis-labs/osmosis/pull/8628) chore: add tagged cometbft version: v0.38.11-v26-osmo-1
* [#8649](https://github.com/osmosis-labs/osmosis/pull/8649) chore: update to tagged submodules
* [#8663](https://github.com/osmosis-labs/osmosis/pull/8663) fix: protorev throws a nil pointer
* [#8676](https://github.com/osmosis-labs/osmosis/pull/8676) fix: update enforce sub-authenticator to be greater than 1 error message
* [#8682](https://github.com/osmosis-labs/osmosis/pull/8682) chore: bump cosmwasm-optimizer

### Config

* [#8548](https://github.com/osmosis-labs/osmosis/pull/8548) chore: disable sqs by default in app.toml


### State Compatible

* [#8494](https://github.com/osmosis-labs/osmosis/pull/8494) Add additional events in x/lockup, x/superfluid, x/concentratedliquidity
* [#8543](https://github.com/osmosis-labs/osmosis/pull/8543) Add OTEL wiring and new configs in app.toml
* [#8566](https://github.com/osmosis-labs/osmosis/pull/8566) Minor speedup to CalcExitCFMM shares
* [#8665](https://github.com/osmosis-labs/osmosis/pull/8665) fix: smart account signing checktx error

## v25.2.1
* [#8546](https://github.com/osmosis-labs/osmosis/pull/8546) feat: reduce commit timeout to 500ms to enable faster blocks, and timeout propose to 1.8s

### CosmosSDK 1f1e8bb04f062250af732b6df98e8581e0e0b77b

* [#612](https://github.com/osmosis-labs/cosmos-sdk/pull/612) OTEL wiring in grpcserver interceptor (DataDog POC) (#612)

### CometBFT v0.37.4-v25-osmo-12

* [#128](https://github.com/osmosis-labs/cometbft/pull/128) feat(p2p): render HasChannel(chID) is a public p2p.Peer method (#3510)
* [#126]() Remove p2p allocations for wrapping outbound packets
* [#125]() Fix marshalling and concurrency overhead within broadcast routines
* perf(p2p): Only update send monitor once per batch packet msg send (#3382)
* [#124]() Secret connection read buffer
* [#123](https://github.com/osmosis-labs/cometbft/pull/123) perf(p2p/conn): Remove unneeded global pool buffers in secret connection #3403
* perf(p2p): Delete expensive debug log already slated for deletion #3412
* perf(p2p): Reduce the p2p metrics overhead. #3411
* commit f663bd35153b0b366c1e1e6b41e7f2dcff7963fd : one more debug log deletion
* [#120](https://github.com/osmosis-labs/cometbft/pull/120) perf(consensus): Use TrySend for hasVote/HasBlockPart messages #3407

* [#8504](https://github.com/osmosis-labs/osmosis/pull/8504) Add missing module params query to CLI

## v25.2.0

* [#8455](https://github.com/osmosis-labs/osmosis/pull/8455) Further comet mempool improvements

## v25.1.3

* [#8420](https://github.com/osmosis-labs/osmosis/pull/8420) Remove further unneeded IBC acknowledgements time from CheckTx/RecheckTx
* [#8421](https://github.com/osmosis-labs/osmosis/pull/8421) Block sdk perf improvements by replacing checkTx with cached tx decoder

### Config

* [#8431](https://github.com/osmosis-labs/osmosis/pull/8431) Fix osmosis-indexer config bug that read osmosis-sqs value
* [#8436](https://github.com/osmosis-labs/osmosis/pull/8436) Added default indexer config on `osmosisd init`

## v25.1.2

* [#8415](https://github.com/osmosis-labs/osmosis/pull/8415) Reset cache on pool creation
* [#8417](https://github.com/osmosis-labs/osmosis/pull/8417) Comet bump to fix edge case with block part selection introduced in v25.1.1

## v25.1.1

* [#8394](https://github.com/osmosis-labs/osmosis/pull/8394) Lower timeout commit from 1s to 600ms
* [#8409](https://github.com/osmosis-labs/osmosis/pull/8409) Add mempool filter to combat IBC spam
* [#8408](https://github.com/osmosis-labs/osmosis/pull/8408) Bump block sdk to lower recheck times by 25%

## v25.1.0

* [#8398](https://github.com/osmosis-labs/osmosis/pull/8398) Lower JSON unmarshalling overhead in IBC packet logic

## v25.0.3

* [#8329](https://github.com/osmosis-labs/osmosis/pull/8329) Overwrite `flush_throttle_timeout` to 80ms instead of 10ms

## v25.0.2

### Osmosis

* [#8312](https://github.com/osmosis-labs/osmosis/pull/8312) Overwrite `flush_throttle_timeout` and `peer_gossip_sleep_duration` to 10ms and 50ms respectively
* [#8308](https://github.com/osmosis-labs/osmosis/pull/8308) Remove IBC rate limit and wasm hook times from mempool checkTx's

### CometBFT

* [#73](https://github.com/osmosis-labs/cometbft/pull/73) perf(consensus/blockexec): Add simplistic block validation cache
* [#74](https://github.com/osmosis-labs/cometbft/pull/74) perf(consensus): Minor speedup to mark late vote metrics
* [#75](https://github.com/osmosis-labs/cometbft/pull/75) perf(p2p): 4% speedup to readMsg by removing one allocation
* [#76](https://github.com/osmosis-labs/cometbft/pull/76) perf(consensus): Add LRU caches for blockstore operations used in gossip
* [#77](https://github.com/osmosis-labs/cometbft/pull/77) perf(consensus): Make every gossip thread use its own randomness instance, reducing mutex contention

## v25.0.1

### Osmosis

* [#8293](https://github.com/osmosis-labs/osmosis/pull/8293) Upgrade v25.x to IBC v7.4.1
* [#8128](https://github.com/osmosis-labs/osmosis/pull/8128) Cache the result for poolmanager.GetPoolModule
* [#8253](https://github.com/osmosis-labs/osmosis/pull/8253) Update gogoproto to v1.4.11 and golang.org/x/exp to
* [#8148](https://github.com/osmosis-labs/osmosis/pull/8148) Remove the deserialization time for GetDefaultTakerFee()
* [#8231](https://github.com/osmosis-labs/osmosis/issues/8231) Deprecate authorized quote denom in CL module
* [#8298](https://github.com/osmosis-labs/osmosis/pull/8298) Overwrite timeout commit from 1.5s to 1s

### CometBFT

* [#61](https://github.com/osmosis-labs/cometbft/pull/61) refactor(p2p/connection): Slight refactor to sendManyPackets that helps highlight performance improvements (backport #2953) (#2978)
* [#62](https://github.com/osmosis-labs/cometbft/pull/62) perf(consensus/blockstore): Remove validate basic call from LoadBlock
* [#69](https://github.com/osmosis-labs/cometbft/pull/69) perf: Make mempool update async from block.Commit (#3008)
* [#67](https://github.com/osmosis-labs/cometbft/pull/67) fix: TimeoutTicker returns wrong value/timeout pair when timeouts are scheduled at the same time

### IBC-go

* Upgrade to v7.4.1 which fixes the significant mempool overhead caused by RedundantRelayDecorator

## v25.0.0

### State Breaking

* [#7935](https://github.com/osmosis-labs/osmosis/pull/7935) Add block sdk and top of block auction from skip-mev
* [#7876](https://github.com/osmosis-labs/osmosis/pull/7876) Migrate subset of spread reward accumulators
* [#7005](https://github.com/osmosis-labs/osmosis/pull/7005) Adding deactivated smart account module
* [#8106](https://github.com/osmosis-labs/osmosis/pull/8106) Enable ProtoRev distro on epoch
* [#8053](https://github.com/osmosis-labs/osmosis/pull/8053) Reset validator signing info missed blocks counter
* [#8073](https://github.com/osmosis-labs/osmosis/pull/8073) Speedup CL spread factor calculations, but mildly changes rounding behavior in the final decimal place
* [#8125](https://github.com/osmosis-labs/osmosis/pull/8125) When using smart accounts, fees are deducted directly after the feePayer is authenticated. Regardless of the authentication of other signers
* [#8136](https://github.com/osmosis-labs/osmosis/pull/8136) Don't allow gauge creation/addition with rewards that have no protorev route (i.e. no way to determine if rewards meet minimum epoch value distribution requirements)
* [#8144](https://github.com/osmosis-labs/osmosis/pull/8144) IBC wasm clients can now make stargate queries and support abort
* [#8147](https://github.com/osmosis-labs/osmosis/pull/8147) Process unbonding locks once per minute, rather than every single block
* [#8157](https://github.com/osmosis-labs/osmosis/pull/8157) Speedup protorev by not unmarshalling pools in cost-estimation phase
* [#8177](https://github.com/osmosis-labs/osmosis/pull/8177) Change consensus params to match unbonding period
* [#8189](https://github.com/osmosis-labs/osmosis/pull/8189) Perf: Use local cache for isSmartAccountActive param
* [#8216](https://github.com/osmosis-labs/osmosis/pull/8216) Chore: Add circuit breaker controller to smart account params

### State Compatible

* [#35](https://github.com/osmosis-labs/cometbft/pull/35) Handle last element in PickRandom
* [#38](https://github.com/osmosis-labs/cometbft/pull/38) Remove expensive Logger debug call in PublishEventTx
* [#39](https://github.com/osmosis-labs/cometbft/pull/39) Change finalizeCommit to use applyVerifiedBlock
* [#40](https://github.com/osmosis-labs/cometbft/pull/40) Speedup NewDelimitedWriter
* [#41](https://github.com/osmosis-labs/cometbft/pull/41) Remove unnecessary atomic read
* [#42](https://github.com/osmosis-labs/cometbft/pull/42) Remove a minint call that was appearing in write packet delays
* [#43](https://github.com/osmosis-labs/cometbft/pull/43) Speedup extended commit.BitArray()
* [#8226](https://github.com/osmosis-labs/osmosis/pull/8226) Overwrite timeoutPropose from 3s to 2s

## v24.0.4

* [#8142](https://github.com/osmosis-labs/osmosis/pull/8142) Add query for getting single authenticator and add stargate whitelist for the query
* [#8149](https://github.com/osmosis-labs/osmosis/pull/8149) Default timeoutCommit to 1.5s

## v24.0.3

* [#21](https://github.com/osmosis-labs/cometbft/pull/21) Move websocket logs to Debug
* [#22](https://github.com/osmosis-labs/cometbft/pull/22) Fix txSearch pagination performance issues
* [#25](https://github.com/osmosis-labs/cometbft/pull/25) Optimize merkle tree hashing

## v24.0.2

* [#8006](https://github.com/osmosis-labs/osmosis/pull/8006), [#8014](https://github.com/osmosis-labs/osmosis/pull/8014) Speedup many BigDec operations
* [#8030](https://github.com/osmosis-labs/osmosis/pull/8030) Delete legacy behavior where lockups could not unbond at very small block heights on a testnet
* [#8118](https://github.com/osmosis-labs/osmosis/pull/8118) Config.toml and app.toml overwrite improvements
* [#8131](https://github.com/osmosis-labs/osmosis/pull/8131) Overwrite min-gas-prices to 0uosmo (defaulting to the EIP 1559 calculated value) to combat empty blocks

## v24.0.1

* [#7994](https://github.com/osmosis-labs/osmosis/pull/7994) Async pruning for IAVL v1

## v24.0.0

### Osmosis

* [#7250](https://github.com/osmosis-labs/osmosis/pull/7250) Further filter spam gauges from epoch distribution
* [#7472](https://github.com/osmosis-labs/osmosis/pull/7472) Refactor TWAP keys to only require a single key format. Significantly lowers TWAP-caused writes
* [#7499](https://github.com/osmosis-labs/osmosis/pull/7499) Slight speed/gas improvements to CL CreatePosition and AddToPosition
* [#7564](https://github.com/osmosis-labs/osmosis/pull/7564) Move protorev dev account bank sends from every backrun to once per epoch
* [#7508](https://github.com/osmosis-labs/osmosis/pull/7508) Improve protorev performance by removing iterator and storing base denoms as a single object rather than an array
* [#7509](https://github.com/osmosis-labs/osmosis/pull/7509) Distributing ProtoRev profits to the community pool and burn address
* [#7524](https://github.com/osmosis-labs/osmosis/pull/7524) Poolmanager: ListPoolsByDenom will now skip pools that cannot correctly return their constituent denoms
* [#7550](https://github.com/osmosis-labs/osmosis/pull/7550) Speedup small CL swaps, by only fetching CL uptime accumulators if there is a tick crossing
* [#7555](https://github.com/osmosis-labs/osmosis/pull/7555) Refactor taker fees, distribute via a single module account, track once at epoch
* [#7562](https://github.com/osmosis-labs/osmosis/pull/7562) Speedup Protorev estimation logic by removing unnecessary taker fee simulations
* [#7595](https://github.com/osmosis-labs/osmosis/pull/7595) Fix cosmwasm pool model code ID migration
* [#7615](https://github.com/osmosis-labs/osmosis/pull/7615) Min value param for epoch distribution
* [#7619](https://github.com/osmosis-labs/osmosis/pull/7619) Slight speedup/gas improvement to CL GetTotalPoolLiquidity queries
* [#7622](https://github.com/osmosis-labs/osmosis/pull/7622) Create/remove tick events
* [#7623](https://github.com/osmosis-labs/osmosis/pull/7623) Add query for querying all before send hooks
* [#7622](https://github.com/osmosis-labs/osmosis/pull/7622) Remove duplicate CL accumulator update logic
* [#7665](https://github.com/osmosis-labs/osmosis/pull/7665) feat(x/protorev): Use Transient store to store swap backruns
* [#7685](https://github.com/osmosis-labs/osmosis/pull/7685) Speedup CL actions by only marshalling for CL hooks if they will be used
* [#7503](https://github.com/osmosis-labs/osmosis/pull/7503) Add IBC wasm light clients module
* [#7689](https://github.com/osmosis-labs/osmosis/pull/7689) Make CL price estimations not cause state writes (speed and gas improvements)
* [#7745](https://github.com/osmosis-labs/osmosis/pull/7745) Add gauge id query to stargate whitelist
* [#7747](https://github.com/osmosis-labs/osmosis/pull/7747) Remove redundant call to incentive collection in CL position withdrawal logic
* [#7768](https://github.com/osmosis-labs/osmosis/pull/7768) Allow governance module account to transfer any CL position
* [#7746](https://github.com/osmosis-labs/osmosis/pull/7746) Make forfeited incentives redeposit into the pool instead of sending to community pool
* [#7785](https://github.com/osmosis-labs/osmosis/pull/7785) Remove reward claiming during position transfers
* [#7833](https://github.com/osmosis-labs/osmosis/pull/7833) Bump max gas wanted per tx to 60 mil
* [#7839](https://github.com/osmosis-labs/osmosis/pull/7839) Add ICA controller
* [#7527](https://github.com/osmosis-labs/osmosis/pull/7527) Add 30M gas limit to CW pool contract calls
* [#7855](https://github.com/osmosis-labs/osmosis/pull/7855) Whitelist address parameter for setting fee tokens
* [#7857](https://github.com/osmosis-labs/osmosis/pull/7857) SuperfluidDelegationsByValidatorDenom query now returns equivalent staked amount
* [#7912](https://github.com/osmosis-labs/osmosis/pull/7912) Default timeoutCommit to 2s
* [#7951](https://github.com/osmosis-labs/osmosis/pull/7951) Only migrate selected cl incentives
* [#7938](https://github.com/osmosis-labs/osmosis/pull/7938) Add missing swap events for missing swap event for cw pools
* [#7957](https://github.com/osmosis-labs/osmosis/pull/7957) Update to the latest version of ibc-go
* [#7966](https://github.com/osmosis-labs/osmosis/pull/7966) Update all governance migrated white whale pools to code id 641

### SDK

* [#525](https://github.com/osmosis-labs/cosmos-sdk/pull/525) CacheKV speedups
* [#548](https://github.com/osmosis-labs/cosmos-sdk/pull/548) Implement v0.50 slashing bitmap logic
* [#543](https://github.com/osmosis-labs/cosmos-sdk/pull/543) Make slashing not write sign info every block
* [#513](https://github.com/osmosis-labs/cosmos-sdk/pull/513) Limit expired authz grant pruning to 200 per block
* [#514](https://github.com/osmosis-labs/cosmos-sdk/pull/514) Let gov hooks return an error
* [#580](https://github.com/osmosis-labs/cosmos-sdk/pull/580) Less time intensive slashing migration

### CometBFT

* [#5](https://github.com/osmosis-labs/cometbft/pull/5) Batch verification
* [#11](https://github.com/osmosis-labs/cometbft/pull/11) Skip verification of commit sigs
* [#13](https://github.com/osmosis-labs/cometbft/pull/13) Avoid double-saving ABCI responses
* [#20](https://github.com/osmosis-labs/cometbft/pull/20) Fix the rollback command

## v23.0.12-iavl-v1

* [#7994](https://github.com/osmosis-labs/osmosis/pull/7994) Async pruning for IAVL v1

## v23.0.11-iavl-v1 & v23.0.11

* [#7987](https://github.com/osmosis-labs/osmosis/pull/7987) Added soft-forked-ibc for ASA-2024-007

## v23.0.8-iavl-v1 & v23.0.8

* [#7769](https://github.com/osmosis-labs/osmosis/pull/7769) Set and default timeout commit to 3s. Add flag to prevent custom overrides if not desired.

## v23.0.7-iavl-v1

* [#7750](https://github.com/osmosis-labs/osmosis/pull/7750) IAVL bump to improve pruning

## v23.0.6-iavl-v1 (contains everything in v23.0.6)

* [#558](https://github.com/osmosis-labs/cosmos-sdk/pull/558) Gracefully log when there is a pruning error instead of panic

## v23.0.6

* [#7716](https://github.com/osmosis-labs/osmosis/pull/7716) Unblock WW Pools
* [#7726](https://github.com/osmosis-labs/osmosis/pull/7726) Bump comet for improved P2P logic
* [#7599](https://github.com/osmosis-labs/osmosis/pull/7599) Reduce sqrt calls in TickToSqrtPrice
* [#7692](https://github.com/osmosis-labs/osmosis/pull/7692) Make CL operation mutative
* [#530](https://github.com/osmosis-labs/cosmos-sdk/pull/530) Bump sdk fork Go to 1.20
* [#540](https://github.com/osmosis-labs/cosmos-sdk/pull/540) Slashing speedup with getting params
* [#546](https://github.com/osmosis-labs/cosmos-sdk/pull/546) Speedup to UnmarshalBalanceCompat
* [#560](https://github.com/osmosis-labs/cosmos-sdk/pull/560) Enable fast nodes at a per module level

## v23.0.3-iavl-v1 (contains everything in v23.0.3)

* [#7582](https://github.com/osmosis-labs/osmosis/pull/7582) IAVL v1
* [#537](https://github.com/osmosis-labs/cosmos-sdk/pull/537) Fix IAVL db grow issue

## v23.0.3

* [#7498](https://github.com/osmosis-labs/osmosis/pull/7498) Protorev mutation speedup
* [#7497](https://github.com/osmosis-labs/osmosis/pull/7497) Better key formatting
* [#7541](https://github.com/osmosis-labs/osmosis/pull/7541) Use more mutatitve operations in uptime accumulator operations
* [#7538](https://github.com/osmosis-labs/osmosis/pull/7538) BigDec speedup
* [#7539](https://github.com/osmosis-labs/osmosis/pull/7539) Speedup CL tickToPrice
* [#7535](https://github.com/osmosis-labs/osmosis/pull/7535) Txfees speedup
* [#7563](https://github.com/osmosis-labs/osmosis/pull/7563) No longer emit meaningless superfluid error at epoch
* [#7590](https://github.com/osmosis-labs/osmosis/pull/7590) fix cwpool migration prop disallowing only one of code id or bytecode.
* [#7577](https://github.com/osmosis-labs/osmosis/pull/7577) Update to sdk math v1.3.0
* [#7598](https://github.com/osmosis-labs/osmosis/pull/7598) Remove extra code path in tickToPrice
* [#3](https://github.com/osmosis-labs/cometbft/pull/3) Avoid double-calling types.BlockFromProto
* [#4](https://github.com/osmosis-labs/cometbft/pull/4) Do not validatorBlock twice

## v23.0.0

### State Breaking

* [#7181](https://github.com/osmosis-labs/osmosis/pull/7181) Improve errors for out of gas
* [#7357](https://github.com/osmosis-labs/osmosis/pull/7357) Fix: Ensure rate limits are not applied to packets that aren't ics20s

### State Compatible

* [#7400](https://github.com/osmosis-labs/osmosis/pull/7400) Update CometBFT to v0.37.4 and IBC to v7.3.2

### Bug Fixes

* [#7360](https://github.com/osmosis-labs/osmosis/pull/7360) fix: use gov type for SetScalingFactorController
* [#7341](https://github.com/osmosis-labs/osmosis/pull/7341) fix: support CosmWasm pools in ListPoolsByDenom method

### Misc Improvements

* [#7360](https://github.com/osmosis-labs/osmosis/pull/7360) Bump cometbft-db from 0.8.0 to 0.10.0
* [#7376](https://github.com/osmosis-labs/osmosis/pull/7376) Add uptime validation logic for `NoLock` (CL) gauges and switch CL gauge to pool ID links to be duration-based
* [#7385](https://github.com/osmosis-labs/osmosis/pull/7385) Add missing protobuf interface
* [#7409](https://github.com/osmosis-labs/osmosis/pull/7409) Scaling factor for pool uptime accumulator to avoid truncation
* [#7417](https://github.com/osmosis-labs/osmosis/pull/7417) Update CL gauges to use gauge duration as uptime, falling back to default if unauthorized or invalid
* [#7419](https://github.com/osmosis-labs/osmosis/pull/7419) Use new module param for internal incentive uptimes
* [#7427](https://github.com/osmosis-labs/osmosis/pull/7427) Prune TWAP records over multiple blocks, instead of all at once at epoch

## v22.0.5

### Logging

* [#7395](https://github.com/osmosis-labs/osmosis/pull/7395) Adds logging to track incentive accumulator truncation.

### Misc Improvements

* [#7374](https://github.com/osmosis-labs/osmosis/pull/7374) In place testnet creation CLI
* [#7411](https://github.com/osmosis-labs/osmosis/pull/7411) De-duplicate fetching intermediate accounts in epoch.
* [#7415](https://github.com/osmosis-labs/osmosis/pull/7415) Speed up TWAP pruning logic.

## v22.0.3

### Config

* [#7368](https://github.com/osmosis-labs/osmosis/pull/7368) Overwrite ArbitrageMinGasPriceconfig from .005 to 0.1.

### Misc Improvements

* [#7266](https://github.com/osmosis-labs/osmosis/pull/7266) Remove an iterator call in updating TWAP records

## v22.0.1

### Bug Fixes

* [#7346](https://github.com/osmosis-labs/osmosis/pull/7346) Prevent heavy gRPC load from app hashing nodes

## v22.0.0

### Fee Market Parameter Updates
* [#7285](https://github.com/osmosis-labs/osmosis/pull/7285) The following updates are applied:
   * Dynamic recheck factor based on current base fee value. Under 0.01, the recheck factor is 3.
    In face of continuous spam, will take ~19 blocks from base fee > spam cost, to mempool eviction.
    Above 0.01, the recheck factor is 2.3. In face of continuous spam, will take ~15 blocks from base fee > spam cost, to mempool eviction.
   * Reset interval set to 6000 which is approximately 8.5 hours.
   * Default base fee is reduced by 2 to 0.005.
   * Set target gas to .625 * block_gas_limt = 187.5 million

### State Breaking

### API
* [#6991](https://github.com/osmosis-labs/osmosis/pull/6991) Fix: total liquidity poolmanager grpc gateway query
* [#7237](https://github.com/osmosis-labs/osmosis/pull/7237) Removes tx_fee_tracker from the proto rev tracker, no longer tracks in state.
* [#7240](https://github.com/osmosis-labs/osmosis/pull/7240) Protorev tracker now tracks a coin array to improve gas efficiency.

### Features
* [#6847](https://github.com/osmosis-labs/osmosis/pull/6847) feat: allow sending denoms with URL encoding
* [#7270](https://github.com/osmosis-labs/osmosis/pull/7270) feat: eip target gas from consensus params

### Bug Fixes
* [#7120](https://github.com/osmosis-labs/osmosis/pull/7120) fix: remove duplicate `query gamm pool` subcommand
* [#7139](https://github.com/osmosis-labs/osmosis/pull/7139) fix: add amino signing support to tokenfactory messages
* [#7245](https://github.com/osmosis-labs/osmosis/pull/7245) fix: correcting json tag value for `SwapAmountOutSplitRouteWrapper.OutDenom`
* [#7267](https://github.com/osmosis-labs/osmosis/pull/7267) fix: support CL pools in tx fee module
* [#7220](https://github.com/osmosis-labs/osmosis/pull/7220) Register consensus params; Set MaxGas to 300m and MaxBytes to 5mb.
* [#7300](https://github.com/osmosis-labs/osmosis/pull/7300) fix: update wasm vm as per CWA-2023-004

### Misc Improvements
* [#6993](https://github.com/osmosis-labs/osmosis/pull/6993) chore: add mutative api for BigDec.BigInt()
* [#7074](https://github.com/osmosis-labs/osmosis/pull/7074) perf: don't load all poolmanager params every swap
* [#7243](https://github.com/osmosis-labs/osmosis/pull/7243) chore: update gov metadata length from 256 to 10200
* [#7258](https://github.com/osmosis-labs/osmosis/pull/7258) Remove an iterator call in CL swaps and spot price calls.
* [#7259](https://github.com/osmosis-labs/osmosis/pull/7259) Lower gas and CPU overhead of chargeTakerFee (in every swap)
* [#7249](https://github.com/osmosis-labs/osmosis/pull/7249) Double auth tx size cost per byte from 10 to 20
* [#7272](https://github.com/osmosis-labs/osmosis/pull/7272) Upgrade go 1.20 -> 1.21
* [#7282](https://github.com/osmosis-labs/osmosis/pull/7282) perf:Update sdk fork to no longer utilize reverse denom mapping, reducing gas costs.
* [#7203](https://github.com/osmosis-labs/osmosis/pull/7203) Make a maximum number of pools of 100 billion.
* [#7282](https://github.com/osmosis-labs/osmosis/pull/7282) Update sdk fork to no longer utilize reverse denom mapping, reducing gas costs.
* [#7291](https://github.com/osmosis-labs/osmosis/pull/7291) Raise mempool config's default max gas per tx configs.

## v21.2.2
### Features
* [#7238](https://github.com/osmosis-labs/osmosis/pull/7238) re-add clawback vesting command
* [#7253](https://github.com/osmosis-labs/osmosis/pull/7253) feat: extended app hash logs

### Bug Fixes
* [#7233](https://github.com/osmosis-labs/osmosis/pull/7233) fix: config overwrite ignores app.toml values
* [#7246](https://github.com/osmosis-labs/osmosis/pull/7246) fix: config overwrite fails with exit code 1 if wrong permissions

### Misc Improvements
* [#7254](https://github.com/osmosis-labs/osmosis/pull/7254) chore: remove cl test modules
* [#7269](https://github.com/osmosis-labs/osmosis/pull/7269) chore: go mod dependency updates
* [#7126](https://github.com/osmosis-labs/osmosis/pull/7126) refactor: using coins.Denoms() from sdk instead of osmoutils
* [#7127](https://github.com/osmosis-labs/osmosis/pull/7127) refactor: replace MinCoins with sdk coins.Min()
* [#7214](https://github.com/osmosis-labs/osmosis/pull/7214) Speedup more stable swap math operations

## v21.2.1

* [#7233](https://github.com/osmosis-labs/osmosis/pull/7233) fix: config overwrite ignores app.toml values

## v21.1.5

* [#7210](https://github.com/osmosis-labs/osmosis/pull/7210) Arb filter for new authz exec swap.

## v21.1.4

* [#7180](https://github.com/osmosis-labs/osmosis/pull/7180) Change `consensus.timeout-commit` from 5s to 4s in `config.toml`. Overwrites the existing value on start-up. Default is set to 4s.

## v21.1.3

Epoch and CPU time optimizations

* [#7093](https://github.com/osmosis-labs/osmosis/pull/7093),[#7100](https://github.com/osmosis-labs/osmosis/pull/7100),[#7172](https://github.com/osmosis-labs/osmosis/pull/7172),[#7174](https://github.com/osmosis-labs/osmosis/pull/7174),[#7186](https://github.com/osmosis-labs/osmosis/pull/7186), [#7192](https://github.com/osmosis-labs/osmosis/pull/7192)   Lower CPU overheads of the Osmosis epoch.
* [#7106](https://github.com/osmosis-labs/osmosis/pull/7106) Halve the time of log2 calculation (speeds up TWAP code)

## v21.1.2

* [#7170](https://github.com/osmosis-labs/osmosis/pull/7170) Update mempool-eip1559 params to cause less base fee spikes on mainnet.
* [#7093](https://github.com/osmosis-labs/osmosis/pull/7093),[#7100](https://github.com/osmosis-labs/osmosis/pull/7100),[#7172](https://github.com/osmosis-labs/osmosis/pull/7172),[#7174](https://github.com/osmosis-labs/osmosis/pull/7174),[#7186](https://github.com/osmosis-labs/osmosis/pull/7186), [#7192](https://github.com/osmosis-labs/osmosis/pull/7186)   Lower CPU overheads of the Osmosis epoch.
* [#7106](https://github.com/osmosis-labs/osmosis/pull/7106) Halve the time of log2 calculation (speeds up TWAP code)

## v21.1.1

Epoch optimizations are in this release, see a subset of PR links in v21.1.3 section.

### Bug Fixes

* [#7209](https://github.com/osmosis-labs/osmosis/pull/7209) Charge gas on input context when querying cw contracts.

## v21.0.0

### API

* [#6939](https://github.com/osmosis-labs/osmosis/pull/6939) Fix taker fee GRPC gateway query path in poolmanager.

### Features

* [#6804](https://github.com/osmosis-labs/osmosis/pull/6804) feat: track and query protocol rev across all modules
* [#7139](https://github.com/osmosis-labs/osmosis/pull/7139) feat: add amino signing support to tokenfactory messages

### Fix Localosmosis docker-compose with state.

* Updated the docker-compose for localosmosis with state to be inline with Operations updated process.

### State Breaks

* [#6758](https://github.com/osmosis-labs/osmosis/pull/6758) Add codec for MsgUndelegateFromRebalancedValidatorSet
* [#6836](https://github.com/osmosis-labs/osmosis/pull/6836) Add DenomsMetadata to stargate whitelist and fixs the DenomMetadata response type
* [#6814](https://github.com/osmosis-labs/osmosis/pull/6814) Add EstimateTradeBasedOnPriceImpact to stargate whitelist
* [#6886](https://github.com/osmosis-labs/osmosis/pull/6886) Add Err handling for ABCI Query Route for wasm binded query
* [#6859](https://github.com/osmosis-labs/osmosis/pull/6859) Add hooks to core CL operations (position creation/withdrawal and swaps)
* [#6932](https://github.com/osmosis-labs/osmosis/pull/6932) Allow protorev module to receive tokens
* [#6937](https://github.com/osmosis-labs/osmosis/pull/6937) Update wasmd to v0.45.0 and wasmvm to v1.5.0
* [#6949](https://github.com/osmosis-labs/osmosis/pull/6949) Valset withdraw rewards now considers all validators user is delegated to instead of valset

### Misc Improvements

* [#7147](https://github.com/osmosis-labs/osmosis/pull/7147) Add poolID to collect CL rewards and incentives events.
* [#6788](https://github.com/osmosis-labs/osmosis/pull/6788) Improve error message when CL LP fails due to slippage bound hit.
* [#6858](https://github.com/osmosis-labs/osmosis/pull/6858) Merge mempool improvements from v20
* [#6861](https://github.com/osmosis-labs/osmosis/pull/6861) Protorev address added to reduced taker fee whitelist
* [#6884](https://github.com/osmosis-labs/osmosis/pull/6884) Improve ListPoolsByDenom function filter denom logic
* [#6890](https://github.com/osmosis-labs/osmosis/pull/6890) Enable arb filter for affiliate swap contract
* [#6884](https://github.com/osmosis-labs/osmosis/pull/6914) Update ListPoolsByDenom function by using pool.GetPoolDenoms to filter denom directly
* [#6959](https://github.com/osmosis-labs/osmosis/pull/6959) Increase high gas threshold to 2m from 1m

### API Breaks

* [#6805](https://github.com/osmosis-labs/osmosis/pull/6805) return bucket index of the current tick from LiquidityPerTickRange query
* [#6530](https://github.com/osmosis-labs/osmosis/pull/6530) Improve error message when CL LP fails due to slippage bound hit.

### Bug Fixes

* [#6840](https://github.com/osmosis-labs/osmosis/pull/6840) fix: change TypeMsgUnbondConvertAndStake value to "unbond_convert_and_stake" and improve error message when epoch currentEpochStartHeight less than zero
* [#6769](https://github.com/osmosis-labs/osmosis/pull/6769) fix: improve dust handling in EstimateTradeBasedOnPriceImpact
* [#6841](https://github.com/osmosis-labs/osmosis/pull/6841) fix: fix receive_ack response field and improve error message of InvalidCrosschainSwapsContract and NoDenomTrace

## v20.4.0

### Bug Fixes

* [#6906](https://github.com/osmosis-labs/osmosis/pull/6906) Fix issue with the affiliate swap contract mempool check.

### Misc Improvements

* [#6863](https://github.com/osmosis-labs/osmosis/pull/6863) GetPoolDenoms method on PoolI interface in poolmanager

## v20.3.0

### Configuration Changes

* [#6897](https://github.com/osmosis-labs/osmosis/pull/6897) Enable 1559 mempool by default.

## v20.2.2

### Features
* [#6890](https://github.com/osmosis-labs/osmosis/pull/6890) Enable arb filter for affiliate swap contract

### Misc Improvements

* [#6847](https://github.com/osmosis-labs/osmosis/pull/6847) feat: allow sending denoms with URL encoding
* [#6788](https://github.com/osmosis-labs/osmosis/pull/6788) Improve error message when CL LP fails due to slippage bound hit.

### API Breaks

* [#6805](https://github.com/osmosis-labs/osmosis/pull/6805) return bucket index of the current tick from LiquidityPerTickRange query
* [#6863](https://github.com/osmosis-labs/osmosis/pull/6863) GetPoolDenoms method on PoolI interface in poolmanager

## v20.0.0

### Features

* [#6847](https://github.com/osmosis-labs/osmosis/pull/6847) feat: allow sending denoms with URL encoding
* [#6766](https://github.com/osmosis-labs/osmosis/pull/6766) CLI: Query pool by coin denom
* [#6468](https://github.com/osmosis-labs/osmosis/pull/6468) feat: remove osmo multihop discount
* [#6420](https://github.com/osmosis-labs/osmosis/pull/6420) feat[CL]: Creates a governance set whitelist of addresses that can bypass the normal pool creation restrictions on concentrated liquidity pools
* [#6623](https://github.com/osmosis-labs/osmosis/pull/6420) feat: transfer cl positions to new owner
* [#6632](https://github.com/osmosis-labs/osmosis/pull/6632) Taker fee bypass whitelist
* [#6709](https://github.com/osmosis-labs/osmosis/pull/6709) CLI: Add list-env, all Environment for CLI

### State Breaking

* [#6413](https://github.com/osmosis-labs/osmosis/pull/6413) feat: update sdk to v0.47x
* [#6344](https://github.com/osmosis-labs/osmosis/pull/6344) fix: set name, display and symbol of denom metadata in tokenfactory's CreateDenom
* [#6279](https://github.com/osmosis-labs/osmosis/pull/6279) fix prop-597 introduced issue
* [#6282](https://github.com/osmosis-labs/osmosis/pull/6282) Fix CreateCanonicalConcentratedLiquidityPoolAndMigrationLink overriding migration records
* [#6309](https://github.com/osmosis-labs/osmosis/pull/6309) Add  Cosmwasm Pool Queries to Stargate Query
* [#6493](https://github.com/osmosis-labs/osmosis/pull/6493) Add PoolManager Params query to Stargate Whitelist
* [#6421](https://github.com/osmosis-labs/osmosis/pull/6421) Moves ValidatePermissionlessPoolCreationEnabled out of poolmanager module
* [#5967](https://github.com/osmosis-labs/osmosis/pull/5967) fix ValSet undelegate API out of sync with existing staking
* [#6627](https://github.com/osmosis-labs/osmosis/pull/6627) Limit pow iterations in osmomath.
* [#6586](https://github.com/osmosis-labs/osmosis/pull/6586) add auth.moduleaccounts to the stargate whitelist
* [#6680](https://github.com/osmosis-labs/osmosis/pull/6680) Add Taker Fee query and add it to stargate whitelist
* [#6699](https://github.com/osmosis-labs/osmosis/pull/6699) fix ValSet undelegate API to work with tokens instead of shares

### Bug Fixes
* [#6644](https://github.com/osmosis-labs/osmosis/pull/6644) fix: genesis bug in pool incentives linking NoLock gauges and PoolIDs
* [#6666](https://github.com/osmosis-labs/osmosis/pull/6666) fix: cosmwasmpool state export bug
* [#6674](https://github.com/osmosis-labs/osmosis/pull/6674) fix: remove dragonberry replace directive
* [#6692](https://github.com/osmosis-labs/osmosis/pull/6692) chore: add cur sqrt price to LiquidityNetInDirection return value
* [#6757](https://github.com/osmosis-labs/osmosis/pull/6757) fix: add gas metering to block before send for token factory bank hook
* [#6710](https://github.com/osmosis-labs/osmosis/pull/6710) fix: `{overflow}` bug when querying cosmwasmpool spot price
* [#6734](https://github.com/osmosis-labs/osmosis/pull/6734) fix: PFM serialization error
* [#6767](https://github.com/osmosis-labs/osmosis/pull/6767) fix: typo in ibc lifecycle message in crosschain swap contract

## v19.2.0

### Misc Improvements

* [#6476](https://github.com/osmosis-labs/osmosis/pull/6476) band-aid state export fix for cwpool gauges
* [#6492](https://github.com/osmosis-labs/osmosis/pull/6492) bump IAVL version to v0.19.7


### Features

* [#6427](https://github.com/osmosis-labs/osmosis/pull/6427) sdk.Coins Mul and Quo helpers in osmoutils
* [#6437](https://github.com/osmosis-labs/osmosis/pull/6437) mutative version for QuoRoundUp
* [#6261](https://github.com/osmosis-labs/osmosis/pull/6261) mutative and efficient BigDec truncations with arbitrary decimals
* [#6416](https://github.com/osmosis-labs/osmosis/pull/6416) feat[CL]: add num initialized ticks query
* [#6488](https://github.com/osmosis-labs/osmosis/pull/6488) v2 SpotPrice CLI and GRPC query with 36 decimals in poolmanager

### API Breaks

* [#6487](https://github.com/osmosis-labs/osmosis/pull/6487) make PoolModuleI CalculateSpotPrice API return BigDec
* [#6511](https://github.com/osmosis-labs/osmosis/pull/6511) remove redundant param from CreateGaugeRefKeys in incentives
* [#6510](https://github.com/osmosis-labs/osmosis/pull/6510) remove redundant ctx param from DeleteAllKeysFromPrefix in osmoutils

## v19.1.0

### Features

* [#6427](https://github.com/osmosis-labs/osmosis/pull/6427) sdk.Coins Mul and Quo helpers in osmoutils
* [#6428](https://github.com/osmosis-labs/osmosis/pull/6428) osmomath: QuoTruncateMut
* [#6437](https://github.com/osmosis-labs/osmosis/pull/6437) mutative version for QuoRoundUp
* [#6261](https://github.com/osmosis-labs/osmosis/pull/6261) mutative and efficient BigDec truncations with arbitrary decimals
* [#6416](https://github.com/osmosis-labs/osmosis/pull/6416) feat[CL]: add num initialized ticks query

### Misc Improvements

* [#6392](https://github.com/osmosis-labs/osmosis/pull/6392) Speedup fractional exponentiation

### Bug Fixes

* [#6334](https://github.com/osmosis-labs/osmosis/pull/6334) fix: enable taker fee cli
* [#6352](https://github.com/osmosis-labs/osmosis/pull/6352) Reduce error blow-up in CalcAmount0Delta by changing the order of math operations.
* [#6368](https://github.com/osmosis-labs/osmosis/pull/6368) Stricter rounding behavior in CL math's CalcAmount0Delta and GetNextSqrtPriceFromAmount0InRoundingUp
* [#6409](https://github.com/osmosis-labs/osmosis/pull/6409) CL math: Convert Int to BigDec

### API Breaks

* [#6256](https://github.com/osmosis-labs/osmosis/pull/6256) Refactor CalcPriceToTick to operate on BigDec price to support new price range.
* [#6317](https://github.com/osmosis-labs/osmosis/pull/6317) Remove price return from CL `math.TickToSqrtPrice`
* [#6368](https://github.com/osmosis-labs/osmosis/pull/6368) Convert priceLimit API in CL swaps to BigDec
* [#6371](https://github.com/osmosis-labs/osmosis/pull/6371) Change PoolI.SpotPrice API from Dec (18 decimals) to BigDec (36 decimals), maintain state-compatibility.
* [#6388](https://github.com/osmosis-labs/osmosis/pull/6388) Make cosmwasmpool's create pool cli generic
* [#6238] switch osmomath to sdkmath types and rename BigDec constructors to contain "Big" in the name.

Note: with the update, the Dec and Int do not get initialized to zero values by default in proto marhaling/unmarshaling. Instead, they get set to nil values.
maxDecBitLen has changed by one bit so overflow panic can be triggered sooner.

## v19.0.0

### Features

* [#6034](https://github.com/osmosis-labs/osmosis/pull/6034) Taker fee

### Bug Fixes
* [#6190](https://github.com/osmosis-labs/osmosis/pull/6190) v19 upgrade handler superfluid fix
* [#6195](https://github.com/osmosis-labs/osmosis/pull/6195) (x/tokenfactory) Fix events for `mintTo` and `burnFrom`
* [#6195](https://github.com/osmosis-labs/osmosis/pull/6195) Fix panic edge case in superfluid AfterEpochEnd hook by surrounding CL multiplier update with ApplyFuncIfNoError

### Misc Improvements

### Minor improvements & Bug Fixes

### Security

## v18.0.0

### Misc Improvements

* [#6161](https://github.com/osmosis-labs/osmosis/pull/6161) Reduce CPU time of epochs

### Bug Fixes

* [#6162](https://github.com/osmosis-labs/osmosis/pull/6162) allow zero qualifying balancer shares in CL incentives

### Features

* [#6034](https://github.com/osmosis-labs/osmosis/pull/6034) feat(spike): taker fee
## v18.0.0

Fixes mainnet bugs w/ incorrect accumulation sumtrees, and CL handling for a balancer pool with 0 bonded shares.

### Improvements

* [#6144](https://github.com/osmosis-labs/osmosis/pull/6144) perf: Speedup compute time of Epoch
* [#6144](https://github.com/osmosis-labs/osmosis/pull/6144) misc: Move many Superfluid info logs to debug logs

### API breaks

* [#6167](https://github.com/osmosis-labs/osmosis/pull/6167) add EstimateTradeBasedOnPriceImpact query to x/poolmanager.
* [#6238](https://github.com/osmosis-labs/osmosis/pull/6238) switch osmomath to sdkmath types and rename BigDec constructors to contain "Big" in the name.
   * Note: with the update, the Dec and Int do not get initialized to zero values
   by default in proto marhaling/unmarshaling. Instead, they get set to nil values.
   * maxDecBitLen has changed by one bit so overflow panic can be triggered sooner.
* [#6071](https://github.com/osmosis-labs/osmosis/pull/6071) reduce number of returns for UpdatePosition and TicksToSqrtPrice functions
* [#5906](https://github.com/osmosis-labs/osmosis/pull/5906) Add `AccountLockedCoins` query in lockup module to stargate whitelist.
* [#6053](https://github.com/osmosis-labs/osmosis/pull/6053) monotonic sqrt with 36 decimals

## v17.0.0

### API breaks

* [#6014](https://github.com/osmosis-labs/osmosis/pull/6014) refactor: reduce the number of returns in superfluid migration
* [#5983](https://github.com/osmosis-labs/osmosis/pull/5983) refactor(CL): 6 return values in CL CreatePosition with a struct
* [#6004](https://github.com/osmosis-labs/osmosis/pull/6004) reduce number of returns for creating full range position
* [#6018](https://github.com/osmosis-labs/osmosis/pull/6018) golangci: add unused parameters linter
* [#6033](https://github.com/osmosis-labs/osmosis/pull/6033) change tick API from osmomath.Dec to osmomath.BigDec

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
* [#5841](https://github.com/osmosis-labs/osmosis/pull/5841) Fix protorev's out of gas erroring of the user's transaction.
* [#5930](https://github.com/osmosis-labs/osmosis/pull/5930) Updating Protorev Binary Search Range Logic with CL Pools
* [#5950](https://github.com/osmosis-labs/osmosis/pull/5950) fix: spot price for cosmwasm pool types

### Misc Improvements

* [#5534](https://github.com/osmosis-labs/osmosis/pull/5534) fix: fix the account number of x/tokenfactory module account
* [#5750](https://github.com/osmosis-labs/osmosis/pull/5750) feat: add cli command for converting proto structs to proto marshalled bytes
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
* [#6005](https://github.com/osmosis-labs/osmosis/pull/6005) osmocli: parse Use field's arguments dynamically
* [#5953] (https://github.com/osmosis-labs/osmosis/pull/5953) Supporting two pool routes in ProtoRev
* [#6012](https://github.com/osmosis-labs/osmosis/pull/6012) chore: add autocomplete to makefile
* [#6085](https://github.com/osmosis-labs/osmosis/pull/6085) (v18: feat) Volume-Split, setup gauges to split evenly

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
  * [#5045] (https://github.com/osmosis-labs/osmosis/pull/5045) Implement hook-based backrunning logic for ProtoRev
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

This release contains the following new modules:
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
  * Add message responses to gamm messages, to remove the necessity of bindings
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
* [#1632](https://github.com/osmosis-labs/osmosis/pull/1632) augment SuperfluidDelegationsByDelegator query, return osmo equivalent is staked via superfluid
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

* [#1442](https://github.com/osmosis-labs/osmosis/pull/1442) Use latest tm-db release for badgerdb and rocksdb improvements
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
* [#1114](https://github.com/osmosis-labs/osmosis/pull/1114) Improve CI: remove duplicate runs of test workflow
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

* [sdk-#136](https://github.com/osmosis-labs/cosmos-sdk/pull/136) add after validator slash hook
* [sdk-#137](https://github.com/osmosis-labs/cosmos-sdk/pull/137) backport feat: Modify grpc gateway to be concurrent
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
* [#686](https://github.com/osmosis-labs/osmosis/pull/686) Add silence usage to cli to suppress unnecessary help logs
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
* [#623](https://github.com/osmosis-labs/osmosis/pull/623) Use gosec for statically linting for common non-determinism issues in SDK applications.

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
* Bringing in the new modules [Bech32IBC](https://github.com/osmosis-labs/bech32-ibc/), [Authz](https://github.com/cosmos/cosmos-sdk/tree/main/x/authz), [TxFees](https://github.com/osmosis-labs/osmosis/tree/main/x/txfees)
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
  * Adds & integrates the [Authz module](https://github.com/cosmos/cosmos-sdk/tree/master/x/authz)
    See: [SDK v0.43.0 Release Notes](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.43.0) For more details
* [\#610](https://github.com/osmosis-labs/osmosis/pull/610) Upgrade to IBC-v2
* [\#560](https://github.com/osmosis-labs/osmosis/pull/560) Implements Osmosis [prop32](https://www.mintscan.io/osmosis/proposals/32) -- clawing back the final 20% of unclaimed osmo and ion airdrop.
* [\#394](https://github.com/osmosis-labs/osmosis/pull/394) Allow whitelisted tx fee tokens based on conversion rate to OSMO
* [Commit db450f0](https://github.com/osmosis-labs/osmosis/commit/db450f0dce8c595211d920f9bca7ed0f3a136e43) Add blocking of OFAC banned Ethereum addresses

## Minor improvements & Bug Fixes

* {In the Osmosis-labs SDK fork}
  * Increase default IAVL cache size to be in the hundred megabyte range
  * Significantly improve CacheKVStore speed problems, reduced IBC upgrade time from 2hrs to 5min
  * Add debug info to make it clear what's happening during upgrade
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

## [v3.1.0](https://github.com/osmosis-labs/osmosis/releases/tag/v3.1.0) - 2021-06-28

* Update the cosmos-sdk version we modify to v0.42.9
* Fix a bug in the min commission rate code that allows validators to be created with commission rates less than the minimum.
* Automatically upgrade any validator with less than the minimum commission rate to the minimum at upgrade time.
* Unbrick on-chain governance, by fixing the deposit parameter to use `uosmo` instead of `osmo`.

## [v1.0.2](https://github.com/osmosis-labs/osmosis/releases/tag/v1.0.2) - 2021-06-18

This release improves the CLI UX of creating and querying gauges.

## [v1.0.1](https://github.com/osmosis-labs/osmosis/releases/tag/v1.0.1) - 2021-06-17

This release fixes a bug in `osmosisd version` always displaying 0.0.1.

## [v1.0.0](https://github.com/osmosis-labs/osmosis/releases/tag/v1.0.0-rc0) - 2021-06-16

Initial Release!
