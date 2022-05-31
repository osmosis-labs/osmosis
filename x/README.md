# Osmosis modules

Osmosis implements the following custom modules:
* `epochs` - Makes on-chain timers which other modules can execute code during.
* `gamm` - Generalized AMM infrastructure, which includes balancer and stableswap
* `incentives` - Controls specification and distribution of rewards to lockups
* `lockup` - Enables time-lock escrowing of tokens. (Often called Locking or Bonding)
* `mint` - Controls token supply emissions, and what modules they are directed to.
* `pool-incentives` - Controls how incentives allocated towards "Liquidity Providing" are directed
  * These go towards gauges defined by the `incentives` module
* `superfluid` - Defines superfluid staking, allowing DeFi assets to have their osmo-backing be staked.
* `tokenfactory` - Allows minting of new tokens of the form `factory/{creator address}/{subdenom}` for user-defined subdenoms. 
* `txfees` - Contains logic for whitelisting txfee tokens, making them easily priceable in osmo, and auto-swapping to osmo.
  * Also contains logic for custom Osmosis mempool logic, though this should perhaps relocate.

TODO: Make visual diagram of the dependency graph between these modules.

This is done in addition to updates to several modules within the SDK.

* `gov` - {Voting period changes}
* `vesting` - {vesting changes}
* Various binding & performance improvements to other modules