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
"State Machine Breaking" for any changes that result in a different AppState given same genesisState and txList.
Ref: https://keepachangelog.com/en/1.0.0/
-->

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
