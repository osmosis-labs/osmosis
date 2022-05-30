# Osmosis modules

Osmosis implements the following custom modules:
* `epochs` - Makes on-chain timers which other modules can execute code during.
* `gamm` - Generalized AMM infrastructure, which includes
* `incentives` - 
* `lockup` - 
* `mint` - Controls token supply emissions, and what modules they are directed to.
* `pool-incentives` - Controls how incentives allocated towards "Liquidity Providing" are directed
    * These go towards gauges defined by the `incentives` module
* `superfluid`
* `tokenfactory` - Allows minting of new tokens of the form `factory/{creator address}/{subdenom}` for user-defined subdenoms.
    * 
* `txfees`

TODO: Make visual diagram of the dependency graph between these modules.

This is done in addition to updates to several modules within the SDK.

* `gov` - {Voting period changes}
* `vesting` - {vesting changes}

