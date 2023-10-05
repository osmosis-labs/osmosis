# Create-lbp-pool

The below is an example of the pool.json file for a [liquidity
bootstrapping
pool](https://balancer.gitbook.io/balancer/guides/crp-tutorial/liquidity-bootstrapping).

A liquidity bootstrapping pool's weight begins at the weight set in the
`weights` parameter and linearly shifts the weights until
`target-pool-weights` is reached over a time period set by the
`duration` parameter upon pool creation.

Typically, weights begin at an unbalanced ratio with more weight given
to the token that is to be sold and shifts to a 1:1 weight (or a weight
favoring the counterparty token that the pool is aiming to accrue). The
changing of the weight affects the exchange price of the tokens even
when the tokens within the pools remain the same. Note that linear
change in weight does **not** mean linear change in price (it is highly
recommend to play around with the various parameters on this [basic LBP
simulator](https://docs.google.com/spreadsheets/d/1t6VsMJF8lh4xuH_rfPNdT5DM3nY4orF9KFOj2HdMmuY/edit#gid=1392289526)
to make sure you understand how the pool will act with different
parameters and market demand).

The pool creator can designate when the weight change begins by setting
the `start-time`. While the pool will be live and available for trade at
the initial `weights`, pool weight shift will not begin until
`start-time` is reached.

## Example Pool Files

The following is an example of a liquidity bootstrapping pool. The
weights linearly change between the initial weights provided, and the
target weights over 72 hrs (3 days) No start time, it defaults to time
the tx was succesfully executed on chain.

pool.json

``` {.json}
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

``` {.json}
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

## Example CLI tx

`osmosisd tx gamm create-pool --pool-file="path/to/lbp-pool.json" --from myKey`

NOTE: The command to create a liquidity bootstrapping pool is the same
as creating a normal pool. However, if the pool has valid `lbp-params`
in the pool file (json), it will be created as a liquidity bootstrapping
pool.
