# CosmWasm Pool Module

## Overview
The CosmWasm Pool Module is an extension for the Osmosis pools, aiming to create a custom module that allows users to
create and manage liquidity pools backed by CosmWasm smart contracts. The feature enables developers to build and deploy
custom smart contracts that can be integrated with the rest of the pool types on the Osmosis chain.

The module is built on top of the CosmWasm smart contracting platform, which provides a secure and efficient way to develop
and execute WebAssembly (Wasm) smart contracts on the Cosmos SDK.

Having pools in CosmWasm provides several benefits, one of which is avoiding the need for chain upgrades when introducing new functionalities
or modifying existing ones related to liquidity pools. This advantage is particularly important in the context of speed of development and
iteration.

## Key Components

- **Keeper**: The module's keeper is responsible for managing the state of the CosmWasm pools, including creating and initializing pools,
querying pool data, and executing privileged operations such as swaps using the CosmWasm sudo message. 
   * `InitializePool`: Initializes a new CosmWasm pool by instantiating a Wasm contract and storing the pool model in the keeper.
   * `Swap operations`: Swap operations like `SwapExactAmountIn` and `SwapExactAmountOut` are implemented, allowing users to perform swaps
   within the CosmWasm pools.
   * `Swap estimation`: Functions like CalcOutAmtGivenIn, and CalcInAmtGivenOut are provided to calculate prices and amounts for swap operations.
   * `Pool information`: Functions like `CalculateSpotPrice`, `GetPool`, `GetPoolAssets`, `GetPoolBalances`, `GetPoolTotalShares` allow
   for querying the state of the CosmWasm pools.


- **Query and Sudo functions**: The module includes generic functions to query CosmWasm smart contracts and execute sudo messages.
The Query and Sudo functions are used to interact with the smart contracts, while MustQuery and MustSudo variants panic if an error
occurs during the query or sudo call, respectively.

- **`poolmanager.PoolI` Interface**: The CosmWasm Pool Model implements the PoolI interface from the Pool Manager Module to enable
the creation and management of liquidity pools backed by CosmWasm smart contracts. By implementing the PoolI interface, the model
ensures compatibility with the existing Pool Manager Module's structure and functionalities and integrates seamlessly with
other modules such as `x/concentrated-liquidity` and `x/gamm`.

- **`poolmanager.PoolModule` Interface**: To integrate the CosmWasm Pool Module with the existing Pool Manager Module,
the module's keeper has to implement the PoolModule interface from `x/poolmanager` Module. By implementing the PoolModule interface,
the CosmWasm Pool Keeper can register itself as an extension to the existing Pool Manager Module and handle the creation and management
of CosmWasm-backed liquidity pools as well as receive swaps propagated from the `x/poolmanager`.
