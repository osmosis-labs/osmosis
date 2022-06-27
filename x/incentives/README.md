# Incentives Module

The incentives module provides users the functionality to create gauges,
which distributes reward tokens to the qualified lockups. Each lockup
has designated lockup duration that indicates how much time that the
user have to wait until the token release after they request to unlock
the tokens.

## Creating Gauges

To initialize a gauge, the creator should decide the following
parameters:

- Distribution condition: denom to incentivize and minimum lockup
    duration.
- Rewards: tokens to be distributed to the lockup owners.
- Start time: time when the distribution will begin.
- Total epochs: number of epochs to distribute over. (Osmosis epochs
    are 1 day each, ending at 5PM UTC everyday)

Making transaction is done in the following format:

``` {.bash}
osmosisd tx incentives create-gauge [denom] [reward] 
  --duration [minimum duration for lockups, nullable]
  --start-time [start time in RFC3339 or unix format, nullable]
  # one of --perpetual or --epochs
  --epochs [total distribution epoch]
  --perpetual
```

### Examples

#### Case 1

I want to make incentives for LP tokens of pool X, namely LPToken, that
have been locked up for at least 1 day. I want to reward 1000 Mytoken to
this pool over 2 days (2 epochs). (500 rewarded on each day) I want the
rewards to start disbursing at 2022 Jan 01.

MsgCreateGauge:

- Distribution condition: denom "LPToken", 1 day.
- Rewards: 1000 MyToken
- Start time: 1624000706 (in unix time format)
- Total epochs: 2 (days)

``` {.bash}
osmosisd tx incentives create-gauge LPToken 1000MyToken \
  --duration 24h \
  --start-time 2022-01-01T00:00:00Z \
  --epochs 2
```

#### Case 2

I want to make incentives for atoms that have been locked up for at
least 1 month. I want to reward 1000 MyToken to atom holders
perpetually. (Meaning I add more tokens to this gauge myself every
epoch) I want the reward to start disbursing immedietly.

MsgCreateGauge:

- Distribution condition: denom "atom", 720 hours.
- Rewards: 1000 MyTokens
- Start time: empty(immedietly)
- Total epochs: 1 (perpetual)

``` {.bash}
osmosisd tx incentives create-gauge atom 1000MyToken
  --perpetual \  
  --duration 168h 
```

I want to refill the gauge with 500 MyToken after the distribution.

MsgAddToGauge:

- Gauge ID: (id of the created gauge)
- Rewards: 500 MyTokens

``` {.bash}
osmosisd tx incentives add-to-gauge $GAUGE_ID 500MyToken
```
