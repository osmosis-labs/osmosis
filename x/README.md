# Symbiosis - On chain decentralized marketing

Symbiosis is a sophisticated multi-leveled marketing schema integrated with Osmosis. It facilitates revenue sharing by distributing the taker fee among participants. This innovative approach aims to enhance user engagement and incentivize active participation within the Osmosis ecosystem.

Key features of Symbiosis include:

- **Revenue Sharing**: Distributes taker fees among users, promoting a fair and balanced economic model.
- **Multi-Leveled Marketing**: Implements a tiered marketing strategy to maximize reach and engagement.
- **Integration with Osmosis**: Seamlessly works with Osmosis modules to provide a cohesive user experience.

Symbiosis leverages the strengths of Osmosis to create a dynamic and rewarding environment for all participants.

# Osmosis modules

Osmosis implements the following custom modules:

- `epochs` - Makes on-chain timers which other modules can execute code during.
- `gamm` - Generalized AMM infrastructure, which includes balancer and stableswap
- `incentives` - Controls specification and distribution of rewards to lockups
- `lockup` - Enables time-lock escrowing of tokens. (Often called Locking or Bonding)
- `mint` - Controls token supply emissions, and what modules they are directed to.
- `pool-incentives` - Controls how incentives allocated towards "Liquidity Providing" are directed
  - These go towards gauges defined by the `incentives` module
- `protorev` - Cyclic arbitrage module that redistributes backrunning profits to the protocol
- `superfluid` - Defines superfluid staking, allowing DeFi assets to have their osmo-backing be staked.
- `tokenfactory` - Allows minting of new tokens of the form `factory/{creator address}/{subdenom}` for user-defined subdenoms.
- `twap` - The TWAP package is responsible for being able to serve TWAPs for every AMM pool.
- `txfees` - Contains logic for whitelisting txfee tokens, making them easily priceable in osmo, and auto-swapping to osmo.
  - Also contains logic for custom Osmosis mempool logic, though this should perhaps relocate.

See the module dependence graph below for further information:

![ModuleDependenceGraph](https://user-images.githubusercontent.com/76530366/175043735-c66c2646-6afc-4a53-9f4b-d26ec45c73d9.png)

This is done in addition to updates to several modules within the SDK.

- `gov` - {Voting period changes}
- `vesting` - {vesting changes}
- Various binding & performance improvements to other modules
