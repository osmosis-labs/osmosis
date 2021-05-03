# Create-lbp-pool

The below is an example of the pool.json file for a [liquidity bootstrapping pool](https://docs.balancer.finance/guides/crp-tutorial/liquidity-bootstrapping).
It creates a liquidity bootstrapping pool, which changes the weights of the pool over a certain time period.

TODO: Give more details

The following is an example of a liquidity bootstrapping pool.
The weights linearly change between the initial weights provided, and the target weights over 72 hrs (3 days)
No start time, it defaults to time the tx was succesfully executed on chain.

pool.json

```json
{
    "weights": "10akt,1atom",
    "initial-deposit": "1000akt,100atom",
    "swap-fee": "0.001",
    "exit-fee": "0.001",
    "lbp-params": {
        "duration": "72h",
        "target-pool-weights": "1akt,1atom"
    }
}
```

Start time included

```json
{
    "weights": "10akt,1atom",
    "initial-deposit": "1000akt,100atom",
    "swap-fee": "0.001",
    "exit-fee": "0.001",
    "lbp-params": {
        "duration": "72h",
        "target-pool-weights": "1akt,1atom",
        "start-time": "2006-01-02T15:04:05Z"
    }
}
```
