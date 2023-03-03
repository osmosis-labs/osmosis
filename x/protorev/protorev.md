# x/protorev

# Abstract

ProtoRev is a module that:

1. Runs during the Posthandler (core trading execution) and Epoch Hook (keeper store updating)
2. In the posthandler of a tx, checks if that tx swaps (has SwapExactAmountIn or SwapExactAmountOut as Msgs)
3. If a tx swaps, generates routes related to the pool swapped against that may contain cyclic arbitrage opportunities after the user’s swap
4. For each route, determines the optimal amount of the asset to swap in that results in maximum amount of the same asset out (profit)
5. Compares profits and selects the route that generates the most profit and is greater than 0
6. Mints the optimal amount of asset to swap in from the Bank module (as determined previously)
7. Executes the MultiHopSwapExactAmountIn with the optimal input amount for the route
8. Burns the same amount of asset previously minted to execute the swap
9. Redistributes the profit captured back to the Osmosis ecosystem based on Governance.

For ecosystem context about the purpose of the module, please see the ProtoRev governance proposal discussion: [https://gov.osmosis.zone/discussion/7078-skip-x-osmosis-proposal-to-capture-mev-as-protocol-revenue-on-chain](https://gov.osmosis.zone/discussion/7078-skip-x-osmosis-proposal-to-capture-mev-as-protocol-revenue-on-chain)

# Concepts

## Cyclic Arbitrage

Cyclic arbitrage is a series of swaps that results in more of the same asset that was initially swapped in. An example of this is as follows:

Assume there exist three pools with the following asset pairs:

```bash
1. A/B
2. B/C
3. C/A
```

A user executes a multi-hop swap that swaps between pools 1, 2, and 3 with the following outcome (user inputs 10A into pool 1, and receives 15A from pool 3):

```bash
User -> 10A -> Pool 1 -> 5B -> Pool 2 -> 20C -> Pool 3 -> 15A -> User
```

This series of swaps is known as a cyclic swap because it starts and ends in the same asset. A cyclic swap is known as a cyclic arbitrage swap when the output amount is greater than the input amount, for the same asset.

## Cyclic Arbitrage Route

A Cyclic Arbitrage Route describes an ordered set of pools that need to be swapped through in consecutive order to capture a cyclic arbitrage opportunity. A Cyclic Route can be determined without knowing current reserve ratios of pools by assessing if one can swap in an asset into the series of pools and receive the same asset out. 

So for the same pools as the example above, an exhaustive list of Cyclic Routes are as follows:

```python
1. A/B
2. B/C
3. C/A

(1,2,3) # Asset A in, Asset A Out
(3,2,1) # Asset A in, Asset A Out
(3,1,2) # Asset C in, Asset C Out
(2,1,3) # Asset C in, Asset C Out
(2,3,1) # Asset B in, Asset B Out
(1,3,2) # Asset B in, Asset B Out
```

What determines if a Cyclic Route is a Cyclic Arbitrage Route at any given state of the chain (state of pool reserves) is if there exists an amount of an asset to be swapped into the route that results in more of the same asset out (10A in, 10A+ Out).

## Optimal Amount In to Swap

When given an ordered route against a specific chain state (state of pool reserves) where a cyclic arbitrage opportunity exists, one must then determine how much to swap in to capture maximum profits (where profits is defined as Asset Out Amount - Asset In Amount). 

ProtoRev uses a binary search algorithm to determine the optimal amount in to swap, using functions from the PoolManager module for calculations and swap execution.

# State

## State Object

The `x/protorev` module keeps the following objects in state:

| State Object | Description | Key | Values | Store |
| --- | --- | --- | --- | --- |
| TokenPairArbRoutes | TokenPairRoutes tracks cyclic arb routes that can be used to create a MultiHopSwap given two denoms | []byte{1} + []byte{inputDenom} +[]byte{outputDenom} | []byte{TokenPairArbRoutes} | KV |
| DenomPairToPool | Tracks the pool ids of the highest liquidity pools matched with a given denom[]byte{2} | []byte{2} + []byte{baseDenom} + []byte{denomToMatch} | []byte{poolID} | KV |
| BaseDenoms | Tracks all of the base denominations that will be used to construct arbitrage routes | []byte{3} | []byte{[]BaseDenoms{}} | KV |
| NumberOfTrades | Tracks the number of trades protorev has executed | []byte{4} | []byte{numberOfTrades} | KV |
| ProfitsByDenom | Tracks the profits protorev has made | []byte{5} + []byte{tokenDenom} | []byte{sdk.Coin} | KV |
| TradesByRoute | Tracks the number of trades the module has executed on a given route | []byte{6} + []byte{route} | []byte{numberOfTrades} | KV |
| ProfitsByRoute | Tracks the profits the module has accumulated after trading on a given route | []byte{7} + []byte{route} | []byte{sdk.Coin} | KV |
| DeveloperAccount | Tracks the developer account for protorev | []byte{8} | []byte{sdk.AccAddress} | KV |
| DaysSinceModuleGenesis | Tracks the number of days since the module was initialized. Used to track profits that can be withdrawn by the developer account | []byte{9} | []byte{uint} | KV |
| DeveloperFees | Tracks the profits that the developer account can withdraw | []byte{10} + []byte{tokenDenom} | []byte{sdk.Coin} | KV |
| MaxPoolPointsPerTx | Tracks the maximum number of pool points that can be consumed per tx | []byte{11} | []byte{uint64} | KV |
| MaxPoolPointsPerBlock | Tracks the maximum number of pool points that can be consumed per block | []byte{12} | []byte{uint64} | KV |
| PoolPointCountForBlock | Tracks the number of pool points that have been consumed in this block | []byte{13} | []byte{uint64} | KV |
| LatestBlockHeight | Tracks the latest recorded block height | []byte{14} | []byte{uint64} | KV |
| PoolWeights | Tracks the weights (pool points) of the different pool types | []byte{15} | []byte{PoolWeights} | KV |

### TokenPairArbRoutes

TokenPairArbRoutes are cyclic arbitrage routes that are not going to be captured by the highest liquidity method (described in state transitions below). If there is a cyclic arbitrage route that is frequently being utilized by searchers, `x/protorev` can manually enter this route - through the admin account - and allow it to be used for trading. Each TokenPairArbRoutes object tracks a directional swap of two assets, and associats the swap with cyclic routes. When the module sees a swap of (`token_in`, `token_out`), it will extract the `arb_routes` that should be used and will simulate trades and execute them if profitable.

```go
// TokenPairArbRoutes tracks all of the hot routes for a given pair of tokens
message TokenPairArbRoutes {
  option (gogoproto.equal) = true;

  // Stores all of the possible hot paths for a given pair of tokens
  repeated Route arb_routes = 1;
  // Token denomination of the first asset
  string token_in = 2;
  // Token denomination of the second asset
  string token_out = 3;
}

// Route is a hot route for a given pair of tokens
message Route {
  option (gogoproto.equal) = true;

  // The pool IDs that are travered in the directed cyclic graph (traversed left
  // -> right)
  repeated Trade trades = 1;
  // The step size that will be used to find the optimal swap amount in the
  // binary search
  string step_size = 2 [
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Int",
    (gogoproto.nullable) = true
  ];
}

// Trade is a single trade in a route
message Trade {
  option (gogoproto.equal) = true;

  // The pool IDs that are travered in the directed cyclic graph (traversed left
  // -> right)
  uint64 pool = 1;
  // The denom of token A that is traded
  string token_in = 2;
  // The denom of token B that is traded
  string token_out = 3;
}
```

### DenomPairToPool

DenomPairToPool takes in a base denomination (read below) – denom that is used to build routes (ex. osmo, atom, usdc) – and a denom to match (akash, juno) and returns the highest liquidity pool id between the pair of denominations. For example, an input might look like (osmo, juno) —> poolID: 5. This store is directly tied to the highest liquidity method (described in state transitions below). Each base denomination is going to have its own set of denominations it maps to.

### BaseDenoms

BaseDenoms are the denominations that are used to build the highest liquidity routes. This will be configurable by the admin account, but will always maintain at least `uosmo` as a base denom. A base denom just means the denomination that will be used to start and end a cyclic arbitrage route. Base denoms can be added on as needed basis. 

***NOTE***: BaseDenoms do have a priority that is directly tied down to the order in the list of base denoms that are used i.e. BaseDenoms that are closer to the front of the list will likely be simulated and executed more often than those later in the list. This is done by design so that we can prioritize certain denoms over others in order to simulate and execute the most profitable trades.

### NumberOfTrades

This will store the total number of arbitrage trades that `x/protorev` has executed since genesis. This gets incremented every time the module executes a trade.

### ProfitsByDenom

This will store the profits `x/protorev` has accumulated for a given denom.

### TradesByRoute & ProfitsByRoute

These stores allow users and researchers to query the number of cyclic arbitrage trades that have been executed by `x/protorev` on an cyclic arbitrage route as well as all of the profits captured on that same route. Routes are denoted by the pool ids in the route i.e. []uint64{1,2,3}.

### ProtoRevEnabled

`x/protorev` can be enabled or disabled through governance. As a proposal is a stateful change, we store whether the module is currently enabled or disabled in the module.

### AdminAccount

The admin account is set through governance and has permissions to set hot routes, the maximum number of pool points per transaction, maximum number of pool points per block, pool type weights, base denoms and the developer account. On genesis, the admin account is set to a trusted address that is stored on a ledger - currently configured to be the Skip dev team's address. Note that governance has full ability to change this live on-chain, and this admin can at most prevent `x/protorev` from working. All the admin account's controls have limits, so it can't lead to a chain halt, excess processing time or prevention of swaps.

### DeveloperAccount

The developer account is set through a MsgSetDeveloperAccount tx. This is the account that will be able to withdraw a portion of the profits from `x/protorev` as specified by the Osmosis ↔ Skip proposal. Only the admin account has permission to make this message.

### DaysSinceModuleGenesis

`x/protorev` will distribute 20% of profits to the developer account in year 1, 10% of profits in year 2, and 5% thereafter. To track how much profit can be distributed to the developer account at any given moment, we store the amount of days since module genesis.

### DeveloperFees

DeveloperFees tracks the total amount of profit that can be withdrawn by the developer account. These fees are sent to the developer account, if set, every week through the `epoch` hook. If unset, the funds are held in the module account. All `x/protorev` profits are going to be stored on the module account.

### MaxPoolPointsPerTx

A pool point roughly corresponds to a millisecond of trading simulation and execution time. In order to bound the compute time of `x/protorev` , we set a maximum number of pool points (execution time) per transaction and per block. MaxPoolPointsPerTx tracks the maximum number of pool points that can be consumed in a given transaction. This is configurable (but bounded) by the admin account. We limit the number of pool points per transaction so that all `x/protorev` execution is not limited to the top of the block.

### MaxPoolPointsPerBlock

MaxPoolPointsPerBlock tracks the maximum number of pool points that can be consumed in a given block. This is configurable (but bounded) by the admin account. We limit the number of pool points per block so that the execution time of the `x/protorev` posthandler is reasonably bounded to ensure that block time remains as is.

### PoolPointCountForBlock

PoolPointCountForBlock tracks the number of pool points that have been consumed in the current block. Used to ensure that the module is not slowing down block speed.

### LatestBlockHeight

LatestBlockHeight tracks the latest recorded block height. This is used to update and reset the pool point count within a block and after new blocks are proposed.

### PoolWeights

PoolWeights assigns each pool type to a number of pool points it will approximately consume. This tracks the pool points or weight of each pool type that can be traversed. This distinction is necessary because different pool types have different simulation and execution times.

```go
// PoolWeights contains the weights of all of the different pool types. This
// distinction is made and necessary because the execution time ranges
// significantly between the different pool types. Each weight roughly
// corresponds to the amount of time (in ms) it takes to execute a swap on that
// pool type.
type PoolWeights struct {
	// The weight of a stableswap pool
	StableWeight uint64 `protobuf:"varint,1,opt,name=stable_weight,json=stableWeight,proto3" json:"stable_weight,omitempty"`
	// The weight of a balancer pool
	BalancerWeight uint64 `protobuf:"varint,2,opt,name=balancer_weight,json=balancerWeight,proto3" json:"balancer_weight,omitempty"`
	// The weight of a concentrated pool
	ConcentratedWeight uint64 `protobuf:"varint,3,opt,name=concentrated_weight,json=concentratedWeight,proto3" json:"concentrated_weight,omitempty"`
}
```

### GenesisState

There is only one configurable parameter for the genesis state —> whether protorev is enabled or not.

```go
// GenesisState defines the protorev module's genesis state.
type GenesisState struct {
	// Module Parameters
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
}
```

# State Transitions

The `protorev` module triggers state transitions in the `postHandler` , governance proposals, and admin account transactions. After each `sdk.Tx`, the `postHandler` will determine whether there were any `MsgSwapExactAmountIn` or `MsgSwapExactAmountOut` in the transaction. If so, the module gets all of the pools that were used in the swap(s), temporarily stores the pool ids accessed along with their respective tokenIn/tokenOut denoms, and then builds cyclic arbitrage routes for each pool swapped against.

## Route Generation

There are two primary methods for route generation: **Highest Liquidity Pools** and **Hot Routes**.

### Highest Liquidity Pool Method

The highest liquidity pool method will always create cyclic arbitrage routes that have three pools. The routes that are created will always start and end with one of the denominations that are stored in BaseDenoms. The pool swapped against that the `postHandler` processes will always be the 2nd pool in the three-pool cyclic arbitrage route. 

**Highest Liquidity Pools:** Updated via the weekly epoch, the module iterates through all the pools and stores the highest liquidity pool for every asset that pairs with any of the base denominations the module stores (for example, the osmo/juno key will have a single pool id stored, that pool id having the most liquidity out of all the osmo/juno pools). New base denominations can be added or removed on an as needed basis by the admin account. A base denomination is just another way of describing the denomination we want to use for cyclic arbitrage. This store is then used to create routes at runtime after analyzing a swap. This store is updated through the `epoch` hook and when the admin account submits a `MsgSetBaseDenoms` tx.

The simplest way to conceptualize how the route is generated is by the following example. Assume we have two base denominations that `x/protorev` is currently tracking.

BaseDenoms

- Osmosis
- Atom

Lets say the `postHandler` receives a transaction that contains a swap of **Juno** —> **Akash** on pool **4**. In this case, the module will attempt to create three-pool route where a base denomination is on either side of the route. For example, a route that it might create is

- Osmosis —> Akash (on pool 1), Akash —> Juno (on pool 4), Juno —> Osmosis (on pool 2)

It does so by finding the highest liquidity pool between (Osmosis, Akash) —> pool 1 and the highest liquidity pool between (Osmosis, Juno) —> pool 2. If there is no highest liquidity pool pair between (Osmosis, Juno) or (Osmosis, Akash), no route will be generated.

**NOTE: Cyclic arbitrage routes will always go in the opposite direction of the original swap i.e. in this case we see Juno —> Akash so we know that the route must include a swap of Akash —> Juno.**

The same line of reasoning exists for Atom. `x/protorev` will attempt to find the highest liquidity pool between (Atom, Akash) and (Atom, Juno). If these pools exist, they will be added to the list of routes that can be simulated later in the pipeline. If not, the route is discarded.

In both cases, the route that is built will always surround the pool of the original swap that was made. However, we allow for more flexibility in route generation as the highest liquidity method may not be optimal, hence the additional of hot routes.

### Hot Route Method

Populated through the admin account, the module’s keeper holds a KV store that associates token pairs (for example, osmo/juno) to the routes that result in a high percentage of arbitrage profit on Osmosis (as determined by external analysis).

The purpose of storing Hot Routes is a recognition that the Highest Liquidity Pool method may not present the best arbitrage routes. As such, hot routes can be configured by the admin account to store additional routes that may be more effective at capturing arbitrage opportunities. Each hot route will store a placeholder for where the current swapped pool will fit into the trade.

### Pool Rebalancing

Now that we have a list of cyclic routes for each pool swapped by the user’s tx, we then determine if any of the routes are profitable. We determine this using a binary search algorithm that finds the amount of the asset to swap in that results in the most of that same asset out. We then calculate profits by taking the difference between the amount of the asset out and amount of the asset in. By iterating through the routes and storing the route, optimal input amount, and profit of the route with the highest profit > 0, we are left with the route and amount to execute the MultiHopSwap against.

Each swap will generate its own set of routes and `x/protorev` will execute only the most profitable route.

The module mints the optimal input amount of the coin to swap in from the `bankkeeper` to the `x/protorev` module account, executes the MultiHopSwap by interacting with the `x/poolmanager` module, burns the optimal input amount of the coin minted to execute the MultiHopSwap, and sends subsequent profits to the module account.

## Governance Proposals

`x/protorev` implements two different governance proposals. 

**SetProtoRevAdminAccountProposal**

As the landscape of pools on Osmosis evolves, an admin account will be able to add and remove routes for `x/protorev` to check for cyclic arbitrage opportunities along with several other optimization txs. Largely, the purpose of maintaining hot routes is to reduce the amount of computation that would otherwise be required to determine optimal paths at runtime. 

This proposal is put in place in case the admin account needs to be transferred over. However, as mentioned above, it will be initialized to a trusted address on genesis.

**SetProtoRevEnabledProposal**

This proposal type allows the chain to turn the module on or off. This is meant to be used as a fail safe in the case stakers and the chain decide to turn the module off. This might be used to halt the execution of trades in the case that the `x/gamm` module has significant upgrades that might produce unexpected behavior from the module.

## PostHandler

The `postHandler` extracts pools that were swapped in a transaction and determines if there is a cyclic arbitrage opportunity. If so, the handler will find an optimal route and execute it - rebalancing the pool and returning arbitrage profits to the module account.

1. Check if the module is enabled.
    1. If the module is disabled, nothing happens.
2. Extract all pools that were traded on in the transaction (`ExtractSwappedPools`) as well as the direction of the trade.
3. Create cyclic arbitrage routes for each of the swaps above (`BuildRoutes`)
4. For each feasible route, determine if there is a cyclic arbitrage opportunity (`IterateRoutes`)
    1. Determine the optimal amount to swap in and its respective profits via binary search over range of potential input amounts (`FindMaxProfitForRoute`)
    2. Compare profits of each route, keep the best route and input amount with the highest profit
5. If the best route and input amount has a profit > 0, execute the trade (`ExecuteTrade`) and rebalance the pools on-behalf of the chain through the `poolmanagerkeeper` (`MultiHopSwapExactAmountIn`)
6. Keep the profits in the module’s account for subsequent distribution.

### ExtractSwappedPools

Checks if there were any swaps made on pools in a transaction, returning the pool ids and input/output denoms for each pool that was traded on.

### BuildRoutes

BuildRoutes takes a token pair (input and output denom) as well as the pool id and returns a list of routes for that token pair that potentially contain a cyclic arbitrage opportunity, populated via the Hot Route and Highest Liquidity Pools method as described above.

### IterateRoutes

IterateRoutes iterates through a list of routes, determining the route and input amount that results in the highest cyclic arbitrage profits..

### FindMaxProfitForRoute

This will take in a route and determine the optimal amount to swap in to maximize profits, given the reserves of all of the pools that are swapped against in the route.

### ExecuteTrade

Execute trade takes the route and optimal input amount as params, mints the optimal amount of input coin, executes the swaps via `poolmanagerKeeper`’s `MultiHopSwapExactAmountIn`, and then burns the amount of coins originally minted, storing the profits in it’s own module account.

This will also update various trading statistics in the module’s store. It will update the total number of trades the module has executed, total profits captured, profits made on this specific route, share of profits the developer account can withdraw, and mor.

## Execution Guardrails

`x/protorev` is bounded and limited in the number of trades the module can execute per block. The purpose of doing so is to ensure that the current block time does not substantially change and that the module does not introduce a new attack vector. 

Execution is currently limited in the following ways

1. The binary search method for finding input amounts is bounded by some number of iterations.
2. The number of routes that can be traversed in a given transaction is bounded by some number.
3. The number of routes that can be traversed in a given block is bounded by some number.

# Hooks

The `x/protorev` module implements epoch hooks in order to trigger the recalculation of the highest liquidity pools paired with any of the base denominations, manages the distribution of developer profits over time, and updates pool point information.

## Epoch Hook

The Epoch hook allows the module to update the information listed above using the epoch identifiers `week` and `day`. 

### Highest Liquidity Pools

As described above, one method of determining cyclic arbitrage opportunities is to use the highest liquidity pools paired with any base denomination. While this calculation is done on genesis (with only Osmo configured), the pools may restructure over time and new tokens may end up being traded heavily with the base denominations. As such, it is necessary to update this over time so that the module’s logic in determining cyclic arbitrage opportunities is most optimal and updated. Using the `AfterEpochEnd` hook in combination with the `week` epoch identifier, we are able to successfully update the pool information every week. At runtime, `UpdatePools` will be executed and all of the internal pool info will be updated.

### Profit Distribution

Profits accumulated by the module will be partially distributed to the developers that built the module in accordance with the governance proposal that was passed: year 1 is 20% of profits, year 2 is 10%, and subsequent years is 5%.

In order to track how much profit the developers can withdraw at any given moment, the module tracks the number of days since module genesis. This gets incremented in the epoch hook after every day. When a trade gets executed by the module, the module will determine how much of the profit from the trade the developers can receive by using `daysSinceModuleGenesis` in a simple calculation. 

If the developer account is not set (which it is not on genesis), all funds are held in the module account. Once the developer address is set by the admin account, the developer address will start to automatically receive a share of profits every week through the epoch hook. The distribution of funds from the module account is done through `SendDeveloperFeesToDeveloperAccount`. Once the funds are distributed, the amount of profit developers can withdraw gets reset to 0 and profits will start to be accumulated and distributed on a week to week basis.

# Governance Proposals

This section defines the governance proposals that result in the state transitions defined on the previous section.

## **`SetProtoRevAdminAccountProposal`**

A gov `content` type to set the admin account which will be overseeing the selection of hot routes, developer account, and more. Governance users vote on this proposal and it automatically executes the custom handler for `SetProtoRevAdminAccountProposal` when the vote passes.

```go
// SetProtoRevAdminAccountProposal is a gov Content type to set the admin
// account that will receive permissions to alter hot routes
type SetProtoRevAdminAccountProposal struct {
	Title       string `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	Description string `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	Account     string `protobuf:"bytes,3,opt,name=account,proto3" json:"account,omitempty"`
}
```

The proposal content stateless validation fails if:

- The account entered is not a valid bech32 address.

## **`SetProtoRevEnabledProposal`**

A gov `content` type to enable or disable the `x/protorev` module. Governance users vote on this proposal and it automatically executes the custom handler for `SetProtoRevEnabledProposal` when the vote passes.

```go
// SetProtoRevEnabledProposal is a gov Content type to update the proto rev
// enabled field in the params
type SetProtoRevEnabledProposal struct {
	Title       string `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	Description string `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	Enabled     bool   `protobuf:"varint,3,opt,name=enabled,proto3" json:"enabled,omitempty"`
}
```

The proposal content stateless validation fails if:

- The entered field to `enabled` is not a boolean.

# Transactions

This section defines the `sdk.Msg` concrete types that result in the state transitions defined on the previous section.

## `MsgSetDeveloperAccount`

The admin account broadcasts a `MsgSetDeveloperAccount` to set the developer account.

```go
// MsgSetDeveloperAccount defines the Msg/SetDeveloperAccount request type.
type MsgSetDeveloperAccount struct {
	// admin is the account that is authorized to set the developer account.
	Admin string `protobuf:"bytes,1,opt,name=admin,proto3" json:"admin,omitempty"`
	// developer_account is the account that will receive a portion of the profit
	// from the protorev module.
	DeveloperAccount string `protobuf:"bytes,2,opt,name=developer_account,json=developerAccount,proto3" json:"developer_account,omitempty"`
}
```

Messsage stateless validation fails if:

- The admin is not a valid bech32 address
- The signature of the user does not match the admin account’s
- The developer account is not a valid bech32 address

Message stateful validation fails if:

- The admin is not set in state
- The admin entered in the message does not match the admin on chain
- The admin’s signatures are not the same

## `MsgSetHotRoutes`

The admin account broadcasts a `MsgSetHotRoutes` to set the hot routes.

```go
// MsgSetHotRoutes defines the Msg/SetHotRoutes request type.
type MsgSetHotRoutes struct {
	// admin is the account that is authorized to set the hot routes.
	Admin string `protobuf:"bytes,1,opt,name=admin,proto3" json:"admin,omitempty"`
	// hot_routes is the list of hot routes to set.
	HotRoutes []*TokenPairArbRoutes `protobuf:"bytes,2,rep,name=hot_routes,json=hotRoutes,proto3" json:"hot_routes,omitempty"`
}
```

Message statless validation fails if:

- The admin is not a valid bech32 address
- The signature of the user does not match the admin account’s
- The hot routes are not valid
    - The starting and ending denominations for each route must be the same
    - The routes in between must have valid swaps i.e. a → b, b → c, c → a and not a → b, c → b, b → a
    - None of the routes can be nil
    - The step size must be set for each route  - the step size is used in the binary search method
    - There must be at least two hops in each route
    - There are duplicate token pairs in the msg

Message stateful validation fails if:

- The admin is not set in state
- The admin entered in the message does not match the admin on chain
- The admin’s signatures are not the same
- `NewMsgSetHotRoutes`

## **`MsgSetMaxPoolPointsPerTx`**

The admin account broadcasts a **`MsgSetMaxPoolPointsPerTx`** to set the maximum number of pool points that can consumed per transaction.

```go
// MsgSetMaxPoolPointsPerTx defines the Msg/SetMaxPoolPointsPerTx request type.
type MsgSetMaxPoolPointsPerTx struct {
	// admin is the account that is authorized to set the max pool points per tx.
	Admin string `protobuf:"bytes,1,opt,name=admin,proto3" json:"admin,omitempty"`
	// max_pool_points_per_tx is the maximum number of pool points that can be
	// consumed per transaction.
	MaxPoolPointsPerTx uint64 `protobuf:"varint,2,opt,name=max_pool_points_per_tx,json=maxPoolPointsPerTx,proto3" json:"max_pool_points_per_tx,omitempty"`
}
```

Message statless validation fails if:

- The admin is not a valid bech32 address
- The signature of the user does not match the admin account’s
- The MaxPoolPointsPerTx is out of range of the limits we hardcode

Message stateful validation fails if:

- The admin is not set in state
- The admin entered in the message does not match the admin on chain
- The admin’s signatures are not the same

## `MsgSetMaxPoolPointsPerBlock`

The admin account broadcasts a `MsgSetMaxPoolPointsPerBlock` to set the maximum number of pool points that can consumed per block.

```go
// MsgSetMaxPoolPointsPerBlock defines the Msg/SetMaxPoolPointsPerBlock request
// type.
type MsgSetMaxPoolPointsPerBlock struct {
	// admin is the account that is authorized to set the max pool points per
	// block.
	Admin string `protobuf:"bytes,1,opt,name=admin,proto3" json:"admin,omitempty"`
	// max_pool_points_per_block is the maximum number of pool points that can be
	// consumed per block.
	MaxPoolPointsPerBlock uint64 `protobuf:"varint,2,opt,name=max_pool_points_per_block,json=maxPoolPointsPerBlock,proto3" json:"max_pool_points_per_block,omitempty"`
}
```

Message statless validation fails if:

- The admin is not a valid bech32 address
- The signature of the user does not match the admin account’s
- The MaxRoutesPerBlock is out of range of the limits we hardcode

Message stateful validation fails if:

- The admin is not set in state
- The admin entered in the message does not match the admin on chain
- The admin’s signatures are not the same

## **`MsgSetPoolWeights`**

The admin account broadcasts a **`MsgSetPoolWeights`** to set the pool weights. The pool weights roughly correspond to the execution time of a swap on that pool type (stable, balancer, concentrated).

```go
// MsgSetPoolWeights defines the Msg/SetPoolWeights request type.
type MsgSetPoolWeights struct {
	// admin is the account that is authorized to set the pool weights.
	Admin string `protobuf:"bytes,1,opt,name=admin,proto3" json:"admin,omitempty"`
	// pool_weights is the list of pool weights to set.
	PoolWeights *PoolWeights `protobuf:"bytes,2,opt,name=pool_weights,json=poolWeights,proto3" json:"pool_weights,omitempty"`
}

// PoolWeights contains the weights of all of the different pool types. This
// distinction is made and necessary because the execution time ranges
// significantly between the different pool types. Each weight roughly
// corresponds to the amount of time (in ms) it takes to execute a swap on that
// pool type.
type PoolWeights struct {
	// The weight of a stableswap pool
	StableWeight uint64 `protobuf:"varint,1,opt,name=stable_weight,json=stableWeight,proto3" json:"stable_weight,omitempty"`
	// The weight of a balancer pool
	BalancerWeight uint64 `protobuf:"varint,2,opt,name=balancer_weight,json=balancerWeight,proto3" json:"balancer_weight,omitempty"`
	// The weight of a concentrated pool
	ConcentratedWeight uint64 `protobuf:"varint,3,opt,name=concentrated_weight,json=concentratedWeight,proto3" json:"concentrated_weight,omitempty"`
}
```

Message statless validation fails if:

- The admin is not a valid bech32 address
- The signature of the user does not match the admin account’s
- Any of the pool weights is not set or is less than or equal to 0.

Message stateful validation fails if:

- The admin is not set in state
- The admin entered in the message does not match the admin on chain
- The admin’s signatures are not the same

## **`MsgSetBaseDenoms`**

The admin account broadcasts a **`MsgSetBaseDenoms`** to set the base denominations the module will use to create cyclic arbitrage routes.

```go
// MsgSetBaseDenoms defines the Msg/SetBaseDenoms request type.
type MsgSetBaseDenoms struct {
	// admin is the account that is authorized to set the base denoms.
	Admin string `protobuf:"bytes,1,opt,name=admin,proto3" json:"admin,omitempty"`
	// base_denoms is the list of base denoms to set.
	BaseDenoms []*BaseDenom `protobuf:"bytes,2,rep,name=base_denoms,json=baseDenoms,proto3" json:"base_denoms,omitempty"`
}

// BaseDenom represents a single base denom that the module uses for its
// arbitrage trades. It contains the denom name alongside the step size of the
// binary search that is used to find the optimal swap amount
type BaseDenom struct {
	// The denom i.e. name of the base denom (ex. uosmo)
	Denom string `protobuf:"bytes,1,opt,name=denom,proto3" json:"denom,omitempty"`
	// The step size of the binary search that is used to find the optimal swap
	// amount
	StepSize github_com_cosmos_cosmos_sdk_types.Int `protobuf:"bytes,2,opt,name=step_size,json=stepSize,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"step_size"`
}
```

Message statless validation fails if:

- The admin is not a valid bech32 address
- The signature of the user does not match the admin account’s
- Osmosis is not the first base denom in the list
- The step size for any of the base denoms is not set
- There are duplicate base denoms

Message stateful validation fails if:

- The admin is not set in state
- The admin entered in the message does not match the admin on chain
- The admin’s signatures are not the same

# Parameters

Tracks whether the module is enabled on genesis.

```go
// Params defines the parameters for the module.
type Params struct {
	// Boolean whether the module is going to be enabled
	Enabled bool `protobuf:"varint,1,opt,name=enabled,proto3" json:"enabled,omitempty"`
}
```

## Enabled

The `Enabled` parameters toggles all state transitions in the module. When the parameter is disabled, it will prevent all module functionality. 

# Clients

## CLI

Find below a lost of `osmosisd` commands added with the `x/protorev` module. A CLI command can look like this:

```bash
osmosisd query protorev params
```

### Queries

| Command | Subcommand | Description |
| --- | --- | --- |
| query protorev | params | Queries the parameters of the module |
| query protorev | number-of-trades | Queries the number of cyclic arbitrage trades ProtoRev has executed |
| query protorev | profits-by-denom [denom] | Queries ProtoRev profits by denom |
| query protorev | all-profits | Queries all ProtoRev profits |
| query protorev | statistics-by-route [route] where route is the list of pool ids i.e. [1,2,3] | Queries ProtoRev statistics by route |
| query protorev | all-statistics | Queries all ProtoRev statistics |
| query protorev | token-pair-arb-routes | Queries the ProtoRev token pair arb routes |
| query protorev | admin-account | Queries the ProtoRev admin account |
| query protorev | developer-account | Queries the ProtoRev developer account |
| query protorev | max-pool-points-per-tx | Queries the ProtoRev max pool points per transaction |
| query protorev | max-pool-points-per-block | Queries the ProtoRev max pool points per block |
| query protorev | base-denoms | Queries the ProtoRev base denoms used to create cyclic arbitrage routes |
| query protorev | enabled | Queries whether the ProtoRev module is currently enabled |

### Proposals

| Command | Subcommand | Description |
| --- | --- | --- |
| tx protorev | set-protorev-admin-account-proposal [sdk.AccAddress] | Submit a proposal to set the admin account for ProtoRev |
| tx protorev | set-protorev-enabled-proposal [boolean] | Submit a proposal to disable/enable the ProtoRev module |

## gRPC & REST

### Queries

| Verb | Method | Description |
| --- | --- | --- |
| gRPC | osmosis.v14.protorev.Query/Params | Queries the parameters of the module |
| gRPC | osmosis.v14.protorev.Query/GetProtoRevNumberOfTrades | Queries the number of arbitrage trades the module has executed |
| gRPC | osmosis.v14.protorev.Query/GetProtoRevProfitsByDenom | Queries the profits of the module by denom |
| gRPC | osmosis.v14.protorev.Query/GetProtoRevAllProfits | Queries all of the profits from the module |
| gRPC | osmosis.v14.protorev.Query/GetProtoRevStatisticsByRoute | Queries the number of arbitrages and profits that have been executed for a given route |
| gRPC | osmosis.v14.protorev.Query/GetProtoRevAllStatistics | Queries all of routes that the module has arbitrage against and the number of trades and profits that have been executed for each route |
| gRPC | osmosis.v14.protorev.Query/GetProtoRevTokenPairArbRoutes | Queries all of the hot routes that the module is currently arbitraging |
| gRPC | osmosis.v14.protorev.Query/GetProtoRevMaxPoolPointsPerTx | Queries the ProtoRev max pool points per transaction |
| gRPC | osmosis.v14.protorev.Query/GetProtoRevMaxPoolPointsPerBlock | Queries the ProtoRev max pool points per block |
| gRPC | osmosis.v14.protorev.Query/GetProtoRevAdminAccount | Queries the admin account of the ProtoRev |
| gRPC | osmosis.v14.protorev.Query/GetProtoRevDeveloperAccount | Queries the developer account of the ProtoRev |
| gRPC | osmosis.v14.protorev.Query/GetProtoRevBaseDenoms | Queries the ProtoRev base denoms used to create cyclic arbitrage routes |
| gRPC | osmosis.v14.protorev.Query/GetProtoRevEnabled | Queries whether the ProtoRev module is currently enabled |
| gRPC | osmosis.14.protorev.Query/GetProtoRevPoolWeights | Queries the number of pool points each pool type will consume when executing and simulating trades |
| GET | /osmosis/v14/protorev/params | Queries the parameters of the module |
| GET | /osmosis/v14/protorev/number_of_trades | Queries the number of arbitrage trades the module has executed |
| GET | /osmosis/v14/protorev/profits_by_denom | Queries the profits of the module by denom |
| GET | /osmosis/v14/protorev/all_profits | Queries all of the profits from the module |
| GET | /osmosis/v14/protorev/statistics_by_route | Queries the number of arbitrages and profits that have happened for a given route |
| GET | /osmosis/v14/protorev/all_route_statistics | Queries all of routes that the module has arbitrage against and the number of trades and profits that have happened for each route |
| GET | /osmosis/v14/protorev/token_pair_arb_routes | Queries all of the hot routes that the module is currently arbitraging |
| GET | /osmosis/v14/protorev/max_pool_points_per_tx | Queries the maximum number of pool points that can be consumed per transaction |
| GET | /osmosis/v14/protorev/max_pool_points_per_block | Queries the maximum number of pool points that can be consumed per block |
| GET | /osmosis/v14/protorev/admin_account | Queries the admin account of the ProtoRev |
| GET | /osmosis/v14/protorev/developer_account | Queries the developer account of the ProtoRev |
| GET | /osmosis/v14/protorev/base_denoms | Queries the base denominations ProtoRev is currently using to create cyclic arbitrage routes |
| GET | /osmosis/v14/protorev/enabled | Queries whether the ProtoRev module is currently enabled |
| GET | /osmosis/v14/protorev/pool_weights | Queries the number of pool points each pool type will consume when executing and simulating trades |

### Transactions

| Verb | Method | Description |
| --- | --- | --- |
| gRPC | osmosis.v14.protorev.Msg/SetHotRoutes | Sets the hot routes that will be explored when creating cyclic arbitrage routes. Can only be called by the admin account |
| gRPC | osmosis.v14.protorev.Msg/SetDeveloperAccount | Sets the account that can withdraw a portion of the profit from the ProtoRev module. Can only be called by the admin account |
| gRPC | osmosis.v14.protorev.Msg/SetMaxPoolPointsPerTx | Sets the maximum number of pool points that can be consumed per transaction |
| gRPC | osmosis.v14.protorev.Msg/SetMaxPoolPointsPerBlock | Sets the maximum number of routes that can be iterated per block |
| gRPC | osmosis.v14.protorev.Msg/SetBaseDenoms | Sets the base denominations the ProtoRev module will use to create cyclic arbitrage routes |
| gRPC | osmosis.v14.protorev.Msg/SetPoolWeights | Sets the amount of pool points each pool type will consume when executing and simulating trades |
| POST | /osmosis/v14/protorev/set_hot_routes | Sets the hot routes that will be explored when creating cyclic arbitrage routes. Can only be called by the admin account |
| POST | /osmosis/v14/protorev/set_developer_account | Sets the account that can withdraw a portion of the profit from the ProtoRev module. Can only be called by the admin account |
| POST | /osmosis/v14/protorev/set_max_pool_points_per_tx | Sets the maximum number of pool points that can be consumed per transaction |
| POST | /osmosis/v14/protorev/set_max_pool_points_per_block | Sets the maximum number of pool points that can be consumed per block |
| POST | /osmosis/v14/protorev/set_pool_weights | Sets the amount of pool points each pool type will consume when executing and simulating trades |
| POST | /osmosis/v14/protorev/set_base_denoms | Sets the base denominations that will be used by ProtoRev to construct cyclic arbitrage routes |