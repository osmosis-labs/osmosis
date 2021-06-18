# Incentives Module

The incentives module provides users the functionality to create gauges, which
distributes reward tokens to the qualified lockups. Each lockup has designated
lockup duration that indicates how much time that the user have to wait until
the token release after they request to unlock the tokens. 

## Creating Gauges

To initialize a gauge, the creator should decide the following parameters:
- Distribution condition: qualified denom and minimum lockup duration.
- Rewards: tokens to be distributed to the lockup owners.
- Start time: time when the distribution will begin.
- Total epochs: period of distributions in epochs.

Making transaction is done in the following format:

```bash
osmosisd tx incentives create-gauge [denom] [reward] 
  --duration [minimum duration for lockups, nullable]
  --start-time [start time in RFC3339 or unix format, nullable]
  # one of --perpetual, --epochs or --epochs-duration
  --epochs [total distribution epoch]
  --epochs-duration [total distribution duration]
  --perpetual
```

### Examples

#### Case 1

You want to airdrop 1000 MyTokens to the atom holders. It doesn't matter for you
how long are they are committed to lockup the tokens, so you will distribute
MyTokens to anyone who locked the atoms regardless of their duration. The
distribution will start from 2022 Jan 01, and happens during the whole year.

MsgCreateGauge:
- Distribution condition: denom "atom", 0 duration.
- Rewards: 1000 MyTokens
- Start time: 2022-01-01T00:00:00Z (in RFC3339 format)
- Total epochs: 52 (weeks)

```bash
osmosisd tx incentives create-gauge atom 1000MyToken \
  --start-time 2022-01-01T00:00:00Z \
  --epochs 52 # or --epochs-duration 8736h
```

#### Case 2

You want to distribute tokens generated from an external source. The tokens are
continuously provided, so you want the gauge to be perpetual. In this case, the
tokens has to be distribued to the lockups only with a lockup duration more than 
a month. Distribution will start immedietly.

MsgCreateGauge:
- Distribution condition: denom "atom", 720 hours.
- Rewards: 200 MyTokens
- Start time: empty(immedietly)
- Total epochs: 1 (perpetual)

```bash
osmosisd tx incentives create-gauge atom 200MyToken
  --perpetual \  
  --duration 720h 
```

Perpetual gauges distribute all of the remaning rewards at the end of each
epochs. Add rewards to the gauge to keep distribute tokens.

MsgAddToGauge:
- Gauge ID: (id of the created gauge)
- Rewards: 500 MyTokens

```bash
osmosisd tx incentives add-to-gauge $GAUGE_ID 500MyToken
```
