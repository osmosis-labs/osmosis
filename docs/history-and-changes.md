# History of Changes

This is a history of changes to the Osmosis repository.

## v6.0.0 (Carbon) - Dec 16, 2021

This upgrade fixes a bug in the v5.0.0 upgrade's app.go, which prevents new IBC channels from being created. All existing IBC channels are believed to be fine.

This binary is compatible with v5.0.0 until block height `2464000`, estimated to be at 4PM UTC Monday December 20th. All nodes must upgrade to this binary prior to that height. This upgrade is intended to be voted in by on-chain governance, but to otherwise be used in place of v5.0.0 at all times.

### Testing methodology

This upgrade has been tested to be compatible with v5.0.0 until the upgrade height on a testnet. This was done by having a v6.0.0 validator and a v5.0.0 full node peered to each other. Prior to upgrade height, both would reject channel open txs. Past upgrade height, the validator would accept channel opens, and the v5.0.0 full node would crash with a conflicting state hash (as expected). The v6.0.0 node could then receive IBC sends/receives.

### Changelog lines

- [Patch](https://github.com/osmosis-labs/osmosis/commit/907001b08686ed980e0afa3d97a9c5e2f095b79f#diff-a172cedcae47474b615c54d510a5d84a8dea3032e958587430b413538be3f333) - Revert back to passing in the correct staking keeper into the IBC keeper constructor.
- [Height gating change](https://github.com/osmosis-labs/ibc-go/pull/1) - Height gate the change in IBC, to make the v6.0.0 binary compatible until upgrade height.


## v5.0.0 (Boron) - Dec 10, 2021

This upgrade is primarily a maintenance upgrade to Osmosis. It updates many of the libraries, brings in the modules [Bech32IBC](https://github.com/osmosis-labs/bech32-ibc), [Authz](https://github.com/cosmos/cosmos-sdk/tree/v0.44.3/x/authz/spec), [Txfees], prepares the chain for [Proposal 32](https://www.mintscan.io/osmosis/proposals/32), and has numerous bug fixes.

If you are building this release from source, you must use go 1.17.

This upgrade adds features such as:

- Upgrade Cosmos-SDK to [SDK v0.44](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.44.3) from SDK v0.42 For a full list of updates in Cosmos-SDK v0.44.3 please see its [changelog](https://github.com/cosmos/cosmos-sdk/blob/release/v0.44.x/CHANGELOG.md#v0443---2021-10-21).
- New modules:
    - [Authz](https://github.com/cosmos/cosmos-sdk/tree/v0.44.3/x/authz/spec) - allows granting arbitrary privileges from one account (the granter) to another account (the grantee). Authorizations must be granted for a particular Msg service method one by one using an implementation of the Authorization interface.
    - [Bech32IBC](https://github.com/osmosis-labs/bech32-ibc) - Allows auto-routing of send msgs to addresses on other chains, once configured by governance. Allows you to do a bank send on Osmosis to a cosmos1... address, and it automatically gets IBC'd there.
    - [TxFees](https://github.com/osmosis-labs/osmosis/tree/main/x/txfees) - Enables validators to easily accept txfees in multiple assets
- Implements [Proposal 32](https://www.mintscan.io/osmosis/proposals/32) - Clawback of unclaimed uosmo and uion on airdrop end date. (December 15th, 5PM UTC)
- Upgrade IBC from a standalone module in the SDK to [IBC v2](https://github.com/cosmos/ibc-go/releases/tag/v2.0.0). This improves the utility of Ethereum Bridges and Cosmwasm bridges.
- Blocking OFAC banned ETH addresses
- Numerous bug fixes, gas fixes, and speedups.

See more in the [changelog](https://github.com/osmosis-labs/osmosis/blob/v5.0.0/CHANGELOG.md)

## v5.0.0-rc2 - Dec 10, 20201

Release candidate #2 of v5.0.0, for use on the public testnet.

High level overview of upgrades included in this:

- Upgrade to SDK v0.44, include authz module. Also upgrade to IBC v2
- Add Bech32IBC, so once governance approves certain IBC channels, you can just send to other chains addresses like cosmos1..., and the chain will handle IBC'ing it out for you
- Allow whitelisted tx fee tokens based on conversion rate to OSMO. This means that a single min-fee can be set in osmo, and a full node will accept fees in atoms, etc. at an equivalent spot price derived from the AMMs. (The tx fees are not auto-converted)
- Reduce growth rate of epoch time due to common user actions
- {Minor bug fixes}


## v5.0.0-rc1 - Dec 8, 2021

rc1 of v5.0.0, for testing.

High level overview of upgrades included in this:
* Upgrade to SDK v0.44, include authz module. Also upgrade to IBC v2
* Add [Bech32IBC](https://github.com/osmosis-labs/bech32-ibc/), so once governance approves certain IBC channels, you can just send to other chains addresses like cosmos1..., and the chain will handle IBC'ing it out for you
* Allow whitelisted tx fee tokens based on conversion rate to OSMO. This means that a single min-fee can be set in osmo, and a full node will accept fees in atoms, etc. at an equivalent spot price derived from the AMMs. (The tx fees are not auto-converted)
* Reduce growth rate of epoch time due to common user actions
* {Minor bug fixes}


## v4.2.0-relayer - Dec 3, 2021

Increases IAVL cache for improved relayer performance


## v4.2.0 - Oct 29, 2021

The v4.2.0 release includes significant reductions in the I/O time used at the Osmosis epoch, and mempool improvements.
The prior release, v4.1.0, improved the CPU time taken by the epoch significantly, but did not change the I/O time which thanks to very detailed profiling from @blockpane , was determined to be the bottleneck.

This is not a permanent fix for I/O time, but instead a constant factor improvement.

As a headsup, there has been two full nodes who tried the new software version and had an app hash issue. This has not been seen on any other nodes using this software version. It is suspected that there was some unfortunate db corruption unrelated to update (or perhaps the new Tendermint version), but please do exercise caution / gradual rollouts.

### What's Changed

* Lower epoch I/O time by an expected 2-3x https://github.com/osmosis-labs/osmosis/pull/561
* Add local mempool filter to block txs that take > 25M gas. (This in large part fixed the chain congestion issue of 10/28)
* Add local mempool filter to block redundant IBC relays  https://github.com/osmosis-labs/osmosis/pull/556
* Upgrade to Tendermint v0.34.14 by @faddat  in https://github.com/osmosis-labs/osmosis/pull/529
* Add rollback to command tree by @jackzampolin in https://github.com/osmosis-labs/osmosis/pull/555

Also huge thanks to @blockpane who was instrumental in diagnosing the I/O time issues, and @jackzampolin @faddat @clemensgg @UnityChaos @imperator-co @wolfcontract for the work in testing the various new versions here!


## v4.1.0 - Oct 14, 2021

This release provides large speedups to the Osmosis epoch time. It does by reducing the amount of events emitted, to have less redundant data. It now only emits one event per address receiving LP rewards.

This upgrade is state-compatible with v4.0.0, and has been tested to be so with many different node operators now. It is encouraged for full nodes & validators to upgrade, in order to significantly reduce their load every epoch.



## v4.0.0 (Berylium) - Sep 19, 2021

This upgrade is a large stability upgrade to Osmosis. It brings with it faster epochs, and improved computation time for various on-chain operations, and fixes to the high gas amounts needed for bonding and unbonding txs.

The features of this upgrade are:

* Fixing gas issues for bonding and unbonding tokens (NOTE: issues at epoch of there just being super high amounts of activity may still persist, with it taking seconds for txs to get into a block)
* Removing the need for users to withdraw locked tokens once they are finished unlocking
* Adding a governance parameter for a minimum fee to create a pool.
* Implements prop 12

See more in the [changelog](https://github.com/osmosis-labs/osmosis/blob/v4.0.0/CHANGELOG.md)


## v4.0.0-rc1 - Sep 18, 2021

Release candidate #1 for osmosis v4

This change primarily brings numerous stability improvements to the chain. It brings with it faster epochs, and improved computation time for various on-chain operations, and fixes to the high gas amounts needed for bonding and unbonding txs.

The features of this upgrade are:
* Fixing gas issues for bonding and unbonding tokens (NOTE: issues at epoch of there just being super high amounts of activity may still persist, with it taking seconds for txs to get into a block)
* Removing the need for users to withdraw locked tokens once they are finished unlocking
* Adding a governance parameter for a minimum fee to create a pool.

If no bugs are found, this state machine will be what is on Osmosis v4.

More thorough changelog [here](https://github.com/osmosis-labs/osmosis/blob/v4.0.0-rc1/CHANGELOG.md)


## v3.1.0 - Aug 7, 2021

This upgrade is meant as a patch that must be hard forked in, due to a bug in proposal 16 breaking on-chain governance of Osmosis. It prevents governance proposals from moving into voting period. Details of the bug are at the bottom. This is the version that should be used, not `v2.0.0` or `v3.0.0`.

This upgrade includes:
* Update to Cosmos-SDK v0.42.9, which fixes state syncing.
* At block height 712000
  * Fixing the immediate governance issue, by changing the min_deposit parameter to what was intended
  * Fixing the bug in min_commission_rate, that allowed validators to create a validator with a lower rate than the minimum. It also bumps up all validators to the minimum.

### Proposed Upgrade Process

* Every node should upgrade their software version from `v1.0.x` to `v3.1.0` before the upgrade block height 712000. If you use cosmovisor, simply swap out the binary at genesis/bin to be v3.1.0, and restart the node.
* Upon upgrading their setup, every validator should place a deposit on a signalling proposal, to signal readiness for the upgrade. (1uosmo suffices)
* Every node should check in between August 10th 1:00AM UTC and 2:00PM UTC, and see if 2/3rds of validators have put a non-zero deposit on the proposal. If so, no further action needed (unless they didn't upgrade yet, in which case they should). If 2/3rds of validators have not signalled readiness by this time, then the upgrade is considered to have not reached agreement, and all nodes should downgrade their binary back to `v1.0.x` for further coordination.

### Governance Bug

In proposal 16, the 'min_deposit' value on the proposal was set to 500osmo and not the intended 500000000uosmo. On chain, the denomination "osmo" doesn't exist, there is only "uosmo" (Similar to how on Bitcoin there are only sats).

Due to this parameter change, a sufficient governance deposit to enter on-chain voting must be in Osmo, which is a denomination that does not exist on chain. Thus no new governance proposals can enter a voting period and get decided on chain.


## v3.0.0 (Lithium) - Aug 6, 2021

This upgrade is meant as a patch that must be hard forked in, due to a bug in proposal 16 breaking on-chain governance of Osmosis. It prevents governance proposals from moving into voting period. Details of the bug are at the bottom. This is the version that should be used, not `v2.0.0`. 

This upgrade includes:
* Update to Cosmos-SDK v0.42.9, which fixes state syncing.
* At block height 712000
  * Fixing the immediate governance issue, by changing the min_deposit parameter to what was intended
  * Fixing the bug in min_commission_rate, that allowed validators to create a validator with a lower rate than the minimum. It also bumps up all validators to the minimum.

### Proposed Upgrade Process

* Every node should upgrade their software version from `v1.0.x` to `v3.0.x` before the upgrade block height 712000. If you use cosmovisor, simply swap out the binary at genesis/bin to be v3.0.0, and restart the node.
* Upon upgrading their setup, every validator should place a deposit on a signalling proposal, to signal readiness for the upgrade. (1uosmo suffices)
* Every node should check in between August 10th 1AM UTC and 1PM UTC, and see if 2/3rds of validators have put a non-zero deposit on the proposal. If so, no further action needed (unless they didn't upgrade yet, in which case they should). If 2/3rds of validators have not signalled readiness by this time, then the upgrade is considered to have not reached agreement, and all nodes should downgrade their binary back to `v1.0.x` for further coordination.

### Governance Bug

In proposal 16, the 'min_deposit' value on the proposal was set to 500osmo and not the intended 500000000uosmo. On chain, the denomination "osmo" doesn't exist, there is only "uosmo" (Similar to how on Bitcoin there are only sats).

Due to this parameter change, a sufficient governance deposit to enter on-chain voting must be in Osmo, which is a denomination that does not exist on chain. Thus no new governance proposals can enter a voting period and get decided on chain.



## v2.0.0 (Helium) - Aug 3, 2021

This upgrade is meant as a patch that must be hard forked in, due to a bug in [proposal 16](https://www.mintscan.io/osmosis/proposals/16) breaking on-chain governance of Osmosis. Details of the bug are at the bottom.

UPDATE: The version that will be used on-chain will not be this version, due to a bug in the cosmos-sdk version v0.42.7

This upgrade includes
* Fixing the immediate governance issue, by changing the min_deposit parameter to what was intended
* Fixing the bug in min_commission_rate, that allowed validators to create a validator with a lower rate than the minimum. It also bumps up all validators to the minimum.
* Update to Cosmos-SDK v0.42.7, which fixes state syncing.

### Governance Bug

In proposal 16, the 'min_deposit' value on the proposal was set to 500**osmo** and not the intended 500000000uosmo. On chain, the denomination "osmo" doesn't exist, there is only "uosmo" (Similar to how on Bitcoin there are only sats).

Due to this parameter change, a sufficient governance deposit to enter on-chain voting must be in Osmo, which is a denomination that does not exist on chain. Thus no new governance proposals can enter a voting period and get decided on chain.



## v2.0.0-rc1 - Jul 9, 2021

Upgrade to SDK version v0.42.7 which fixes state sync


## v2.0.0-rc1 - Jun 28, 2021

This release contains a release candidate for the v2.0.0 upgrade for Osmosis.

We are using this to test if cosmovisor auto-downloading of binaries works as wel


## v1.0.0 (Hydrogen) - Jun 17, 2021

This version is the version of the binary for Osmosis launch.

It is fully compatible with tag v1.0.0, it just fixes a bug where `osmosisd version` didn't show the correct version.


## v1.0.0-rc1 - Jun 16, 2021

Release candidate 0 for Osmosis mainnet!


# Medium archives

- [Medium Archives](https://medium.com/Osmosis/archive/)

