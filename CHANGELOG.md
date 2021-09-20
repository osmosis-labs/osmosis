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

## [v4.0.0]

* Significantly speedup epoch times
* Fix bug in the lockup module code that caused it to take a linear amount of gas.
* Make unbonding tokens from the lockup module get automatically claimed when unbonding is done.
* Add events for all tx types in the gamm module.
* Add events for adding LP rewards.
* Make queries to bank total chain balance account for developer vesting correctly.
* Embedded seeds in init.go
* Many SDK performance improvements on the epoch's critical path
* Many accumulation store performance improvements
* Disable the accumulation store until v5
* (Disabled until v5) Add ability for nodes to query the total amount locked for each denomination.

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
