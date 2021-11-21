# Token Listings

## How to create a new pool with IBC assets

Osmosis is a automated market maker blockchain. This means any IBC-enabled zone can add its token as an asset to be traded on Osmosis AMM completely permissionlessly. Because Osmosis is fundamentally designed as an IBC-native AMM that trades IBC tokens, rather than tokens issued on the Osmosis zone, there are additional nuances to understand and steps to be taken in order to ensure your asset is supported by Osmosis.

This document lays out the prerequisites and the process process that's needed to ensure that your token meets the interchain UX standards set by Osmosis.

### Prerequisites
1. Zone must have IBC token transferred enabled (ICS20 standard).
2. Assets to be traded should be a fungible `sdk.Coins` asset.
3. Highly reliable, highly available altruistic (as in relay tx fees paid on behalf of user) relayer service.
4. Highly reliable, highly available, and scalable RPC/REST endpoint infrastructure.


### 0. Enabling IBC transfers
Because only IBC assets that have been transferred to Osmosis can be traded on Osmosis, the native chain of the asset must have IBC transfers enabled. Cosmos defines the fungible IBC token transfer standard in [ICS20](https://github.com/cosmos/ibc/tree/master/spec/app/ics-020-fungible-token-transfer) specification.

At this time, only chains using Cosmos-SDK v0.40+ (aka Stargate) can support IBC transfers.

Note that IBC transfers can be enabled via:
1. as part of a software upgrade, or
2. a `ParameterChange` governance proposal

To ensure a smooth user experience, Osmosis assumes all tokens will be transferred through a single designated IBC channel between Osmosis and the counterparty zone.

Recommended readings:
* [IBC Overview](https://docs.cosmos.network/v0.43/ibc/overview.html) - To understand IBC clients, connections, 
* [How to Upgrade IBC Chains and their Clients](https://docs.cosmos.network/v0.43/ibc/upgrades/quick-guide.html)

### 1. Add your chain to cosmos/chain-registry and SLIP73

#### Cosmos Chain Registry
Make a PR to add your chain's entry to the [Cosmos Chain Registry](https://github.com/cosmos/chain-registry). This allows Osmosis frontend to suggest your chain for asset deposit/withdrawals(IBC transfers).

Make sure to include at least one reliable RPC, gRPC, REST endpoint behind https. Refer to the [Osmosis entry](https://github.com/cosmos/chain-registry/blob/master/osmosis/chain.json) as an example.

### 2. Setting up and operating relayer to Osmosis
Relayers are responsible of transferring IBC packets between Osmosis chain and the native chain of an asset. All Osmosis 'deposits' and 'withdrawals' are IBC transfers which dedicated relayers process.

To ensure fungibility amongst IBC assets, the frontend will assume social consensus have been achieved and designate one specific channel between Osmosis and the native chain as the primary channel for all IBC token transfers. Multiple relayers can be active on the same channel, and for the sake of redundancy and increased resilience we recommend having multiple relayers actively relaying packets. It is recommended to initialize the channel as an unordered IBC channel, rather than an ordered IBC channel.

Currently, there are three main Cosmos-SDK IBC relayer implementations:
* [Go relayer](https://github.com/cosmos/relayer): A Golang implementation of IBC relayer.
* [Hermes](https://hermes.informal.systems/): A Rust implementation of IBC relayer.
* [ts-relayer](https://github.com/confio/ts-relayer): A TypeScript implementation of IBC relayer.
* 
**Note: We are actively investigating issues regarding ts-relayer not working with Osmosis. In the meantime, we recommend using Hermes/Go relayer**

All relayers are compatible with IBC token transfers on the same channel. Each relayer implementation may have different configuration requirements, and have various configuration customizability.

At this time, Osmosis requires that all relayers to pay for the transaction fees for IBC relay transactions, and not the user.

If you prefer not to run your own chain's relayer to Osmosis, there may be various entities ([Cephalopod Equipment Corp.](https://cephalopod.equipment/), [Vitwit](https://www.vitwit.com/), etc) that provide relayers-as-a-service, or you may reach out to various validators in your ecosystem that may be able to operate a relayer. The Osmosis team does **not** provide relayer services for IBC assets.

#### SLIP73 bech32 prefix
Add your chain's bech32 prefix to the [SLIP73 repo](https://github.com/satoshilabs/slips/blob/master/slip-0173.md). The bech32 prefix should be a unix prefix, and only mainnet prefixes should be included.


### 3. Making a PR to Osmosis/assetlists
Due to the permissionless nature of IBC protocol, the same base asset transferred over two different IBC channels will result in two different token denominations.

Example:
* `footoken` transferred to `barchain` through `channel-1`: `ibc/1b3d5f...`
* `footoken` transferred to `barchain` through `channel-2`: `ibc/a2c4e6...`

In order to reduce user confusion and prevent token non-fungibility, Osmosis frontends are recommended to designate one specific channel as the primary channel for the chain's assets. The Osmosis will only show the IBC token denomination of the designated channel as with the original denomination (i.e. ATOM, AKT, etc).

Therefore, Osmosis uses [Assetlists](https://github.com/osmosis-labs/assetlists) as a way to designate and manage token denominations of IBC tokens.

Please create a pull request with the necessary information to allow your token to be shown in its original denomination, rather than as an IBC token denomination.

If you need to verify the base denom of an IBC asset, you can use `{REST Endpoint Address}
/ibc/applications/transfer/v1beta1/denom_traces` for all IBC denoms or `{REST Endpoint Address}
/ibc/applications/transfer/v1beta1/denom_traces/{hash}` for one specific IBC denom. (If you need an RPC/REST endpoint for Osmosis, [Figment DataHub](https://datahub.figment.io) provides a free service for up to 100k requests/day.)

### 4. Creating a pool on Osmosis
Please refer to the [`create-pool` transaction example on the Osmosis repository](https://github.com/osmosis-labs/osmosis/tree/main/x/gamm#create-pool) on how to create a pool using your IBC tokens.

Recommended are:
* 50:50 OSMO-Token pool with 0.2% swap fee and 0% exit fee
* 50:50 ATOM-Token pool with 0.3% swap fee and 0% exit fee



Guide created by dogemos.
