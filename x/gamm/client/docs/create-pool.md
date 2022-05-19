# Create-pool

In a create-pool tx, it is important that the initial-deposit match the
intended 'actual' initial price as closely as possible.

The below is an example of the pool.json file for a pool with a given
set of weights, and initial reserves for each asset.

The following pool.json defines a `1:1:2` pool between Eth, Regen, and
Btc, with initial spot prices of `1 eth = .5 btc`, `1 regen = .5 btc`,
`1 eth = 1 regen`.

pool.json

``` {.json}
{
    "weights": "1eth,1regen,2btc",
    "initial-deposit": "100eth,100regen,100btc",
    "swap-fee": "0.001",
    "exit-fee": "0.001"
}
```
