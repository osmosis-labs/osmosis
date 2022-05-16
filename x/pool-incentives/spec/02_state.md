```html
<!--
order: 2
-->
```

# State

### Genesis states

```go
type GenesisState struct {
 // params defines all the paramaters of the module.
 Params            Params          
 LockableDurations []time.Duration 
 DistrInfo         *DistrInfo      
}

type Params struct {
 // minted_denom is the denomination of the coin expected to be minted
 //  by the minting module.
 // Pool-incentives module doesn’t actually mint the coin itself, 
 // but rather manages the distribution of coins that matches the defined minted_denom.
 MintedDenom string 
 // allocation_ratio defines the proportion of the minted minted_denom 
 // that is to be allocated as pool incentives.
 AllocationRatio github_com_cosmos_cosmos_sdk_types.Dec 
}
```

Lockable durations can be set to the pool incentives module at genesis.
Every time a pool is created, the `pool incentives` module creates the
same amount of 'gauge' as there are lockable durations for the pool.

Also in regards to the `Params`, when the mint module mints new tokens
to the fee collector at Begin Block, the `pool incentives` module takes
the token which matches the 'minted denom' from the fee collector.
Tokens are taken according to the 'allocationRatio', and are distributed
to each `DistrRecord` of the DistrInfo. For example, if the fee
collector holds 1000uatom and 2000 uosmo at Begin Block, and Params'
mintedDenom is set to uosmo, and AllocationRatio is set to 0.1, 200uosmo
will be taken from the fee collector and distributed to the
`DistrRecord`s.
