# Create-pool

In a create-stableswap-pool tx, it is important that the initial-deposit, and scaling factors, match the 1:1 price of the pool.

So you want the scaling factors to be such that `deposit_asset_1 / scaling_factor_1` is intended to have the same economic value as `deposit_asset_2 / scaling_factor_2`.

The below is an example of the pool.json file for a pool with $1 worth of `uusdc` and an imagined asset `miliusdc`. So namely, `1 USDC = 10^6 uusdc = 10^3 miliusdc`. This implies a need for a `1000:1` scaling factor.

Here is what this would look like:

pool.json

``` {.json}
{
	"initial-deposit": "1000000uusdc,1000miliusdc",
    "scaling-factors": "1000,1",
	"swap-fee": "0.005",
	"exit-fee": "0.00",
	"future-governor": "168h",
    "scaling-factor-controller": ""
}
```

There is also an optional field called `scaling-factor-controller`,
where you give a certain address the ability to control the scaling factors.