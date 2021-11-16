# GAMM Module
This document introduces the [Queries](#queries) and [Transactions](#transactions) of the **G**eneralized **A**utomated **M**arket **M**aker (GAMM) module. The GAMM module provides the logic to create and interact with liquidity pools on the Osmosis DEX.



# Queries
The **Query** submodule of the GAMM module provides the logic to request information from the liquidity pools. It contains the following functions:
- [Estimate Swap Exact Amount In](#estimate-swap-exact-amount-in)
- [Estimate Swap Exact Amount Out](#estimate-swap-exact-amount-out)
- [Num Pools](#num-pools)
- [Pool](#pool)
- [Pool Assets](#pool-assets)
- [Pool Params](#pool-params)
- [Pools](#pools)
- [Spot Price](#spot-price)
- [Total Liquidity](#total-liquidity)
- [Total Share](#total-share)

## Estimate Swap Exact Amount In
Query the estimated result of the [Swap Exact Amount In](#swap-exact-amount-in) transaction. Note that the flags *swap-route-pool* and *swap-route-denoms* are required.
### Usage
```sh
osmosisd query gamm estimate-swap-exact-amount-in <poolID> <sender> <tokenIn> [flags]
```
### Example
Query the amount of ATOM the sender would receive for swapping 1 OSMO in pool 1.
```sh
osmosisd query gamm estimate-swap-exact-amount-in 1 osmo123nfq6m8f88m4g3sky570unsnk4zng4uqv7cm8 1000000uosmo --swap-route-pool-ids 1 --swap-route-denoms ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2 
```


## Estimate Swap Exact Amount Out
Query the estimated result of the [Swap Exact Amount Out](#swap-exact-amount-out) transaction. Note that the flags *swap-route-pool* and *swap-route-denoms* are required.
### Usage
```sh
osmosisd query gamm estimate-swap-exact-amount-out <poolID> <sender> <tokenOut> [flags]
```
### Example
Query the amount of OSMO the sender would require to swap 1 ATOM out of pool 1.
```sh
osmosisd query gamm estimate-swap-exact-amount-out 1 osmo123nfq6m8f88m4g3sky570unsnk4zng4uqv7cm8 1000000ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2 --swap-route-pool-ids 1 --swap-route-denoms uosmo
```


## Num Pools
Query the number of active pools.
### Usage
```sh
osmosisd query gamm num-pools
```


## Pool
Query the parameter and assets of a specific pool. 
### Usage
```sh
osmosisd query gamm pool <poolID> [flags]
```
### Example
Query parameters and assets from pool 1.
```sh
osmosisd query gamm pool 1
```


## Pool Assets
Query the assets of a specific pool. This query is a reduced form of the [Pool](#pool) query.
### Usage
```sh
osmosisd query gamm pool-assets <poolID> [flags]
```
Query the assets from pool 1.
### Example
```sh
osmosisd query gamm pool-assets 1
```


## Pool Params
Query the parameters of a specific pool. This query is a reduced form of the [Pool](#pool) query.
### Usage
```sh
osmosisd query gamm pool-params <poolID> [flags]
```
Query the parameters from pool 1.
### Example
```sh
osmosisd query gamm pool-params 1
```


## Pools
Query parameters and assets of all active pools.
### Usage
```sh
osmosisd query gamm pools
```


## Spot Price
Query the spot price of a pool asset based on a specific pool it is in.
### Usage
```sh
osmosisd query gamm spot-price <poolID> <tokenInDenom> <tokenOutDenom> [flags]
```
### Example
Query the price of OSMO based on the price of ATOM in pool 1.
```sh
osmosisd query gamm spot-price 1 uosmo ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2
```


## Total Liquidity
Query the total liquidity of all active pools.
### Usage
```sh
osmosisd query gamm total-liquidity
```


## Total Share
Query the total amount of GAMM shares of a specific pool.
### Usage
```sh
osmosisd query gamm total-share <poolID> [flags]
```
### Example
Query the total amount of GAMM shares of pool 1.
```sh
osmosisd query gamm total-share 1
```




# Transactions
The **Transaction** submodule of the GAMM module provides the logic to create and interact with the liquidity pools. It contains the following functions:
- [Create Pool](#create-pool)
- [Join Pool](#join-pool)
- [Exit Pool](#exit-pool)
- [Join Swap Extern Amount In](#join-swap-extern-amount-in)
- [Exit Swap Extern Amount Out](#exit-swap-extern-amount-out)
- [Join Swap Share Amount Out](#join-swap-share-amount-out)
- [Exit Swap Share Amount In](#exit-swap-share-amount-in)
- [Swap Exact Amount In](#swap-exact-amount-in)
- [Swap Exact Amount Out](#swap-exact-amount-out)


## Create Pool
Create a new liquidity pool and provide the initial liquidity to it. Pool initialization parameters must be provided through a JSON file using the flag *pool-file*.
#### Usage
```sh
osmosisd tx gamm create-pool [flags]
```
The configuration file *config.json* must specify the following parameters.
```sh
{
	"weights": [list weighted denoms],
	"initial-deposit": [list of denoms with initial deposit amount],
	"swap-fee": [swap fee in percentage],
	"exit-fee": [exit fee in percentage],
	"future-governor": [number of hours]
}
```
### Example
Create a new ATOM-OSMO liquidity pool with a swap and exit fee of 1%.
```sh
tx gamm create-pool --pool-file ../public/config.json --from myKeyringWallet
```
The configuration file contains the following parameters.
```sh
{
	"weights": "5uatom,5uosmo",
	"initial-deposit": "100uatom,100uosmo",
	"swap-fee": "0.01",
	"exit-fee": "0.01",
	"future-governor": "168h"
}
```



## Join Pool
Join a specific pool with a custom amount of tokens. Note that the flags *pool-id*, *max-amounts-in* and *share-amount-out* are required.
#### Usage
```sh
osmosisd tx gamm join-pool [flags]
```
#### Example
Join pool 1 with 1 OSMO and the respective amount of ATOM, using myKeyringWallet.
```sh
osmosisd tx gamm join-pool --pool-id 2 --max-amounts-in 1000000uosmo --max-amounts-in 1000000uion --share-amount-out 1000000 --from myKeyringWallet
```


## Exit Pool
Exit a specific pool with a custom amount of tokens. Note that the flags *pool-id*, *min-amounts-out* and *share-amount-in* are required.
#### Usage
```sh
osmosisd tx gamm exit-pool [flags]
```
#### Example
Exit pool one with 1 OSMO and the respective amount of ATOM using myKeyringWallet.
```sh
osmosisd tx gamm exit-pool --pool-id 1 --min-amounts-out 1000000uosmo --share-amount-in 1000000 --from myKeyringWallet
```


## Join Swap Extern Amount In
Note that the flags *pool-id* is required.
#### Usage
```sh
osmosisd tx gamm join-swap-extern-amount-in [token-in] [share-out-min-amount] [flags]
```
#### Example
```sh
osmosisd tx gamm join-swap-extern-amount-in 1000000uosmo 1000000 --pool-id 1 --from myKeyringWallet
```


## Exit Swap Extern Amount Out
Note that the flag *pool-id* is required.
#### Usage
```sh
osmosisd tx gamm exit-swap-extern-amount-out [token-out] [share-in-max-amount] [flags]
```
#### Example
```sh
osmosisd tx gamm exit-swap-extern-amount-out 1000000uosmo 1000000 --pool-id 1 --from myKeyringWallet

```


## Join Swap Share Amount Out
Note that the flag *pool-id* is required.
#### Usage
```sh
osmosisd tx gamm join-swap-share-amount-out [token-in-denom] [token-in-max-amount] [share-out-amount] [flags]
```
#### Example
```sh
osmosisd tx gamm join-swap-share-amount-out uosmo 1000000 1000000 --pool-id 1 --from myKeyringWallet
```


## Exit Swap Share Amount In
Note that the flag *pool-id* is required.
#### Usage
```sh
osmosisd tx gamm exit-swap-share-amount-in [token-out-denom] [share-in-amount] [token-out-min-amount] [flags]
```
#### Example
```sh
osmosisd tx gamm exit-swap-share-amount-in uosmo 1000000 1000000 --pool-id 1 --from myKeyringWallet
```

## Swap Exact Amount In
Swap an exact amount of tokens into a specific pool. Note that the flags *swap-route-pool-ids* and *swap-route-denoms* are required.
#### Usage
```sh
osmosisd tx gamm swap-exact-amount-in [token-in] [token-out-min-amount] [flags]
```
#### Example
Swap 1 OSMO through pool 1 into at least 0.3 ATOM using MyKeyringWallet.
```sh
osmosisd tx gamm swap-exact-amount-in 1000000uosmo 300000 --swap-route-pool-ids 1 --swap-route-denoms ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2 --from MyKeyringWallet
```


## Swap Exact Amount Out
Swap an exact amount of tokens out of a specific pool. Note that the flags *swap-route-pool-ids* and *swap-route-denoms* are required.
#### Usage
```sh
osmosisd tx gamm swap-exact-amount-out [token-out] [token-out-max-amount] [flags]
```
#### Example
Swap 1 ATOM through pool 1 into at most 2.5 OSMO using MyKeyringWallet.
```sh
osmosisd tx gamm swap-exact-amount-out 1000000ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2 250000 --swap-route-pool-ids 1 --swap-route-denoms uosmo --from MyKeyringWallet
```

# Other resources
* [Creating a liquidity bootstrapping pool](./client/docs/create-lbp-pool.md)
* [Creating a pool with a pool file](./client/docs/create-pool.md)