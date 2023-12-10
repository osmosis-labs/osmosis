# Sidecar Query Server

This is a sidecar query server that is used for performing query tasks outside of the main chain.
The high-level architecture is that the chain reads data at the end of the block, parses it
and then writes into a Redis instance.

The sidecar query server then reads the parsed data from Redis and serves it to the client
via HTTP endpoints.

The use case for this is performing certain data and computationally intensive tasks outside of
the chain node or the clients. For example, routing falls under this category because it requires
all pool data for performing the complex routing algorithm.

## Supported Endpoints

## Pools Resource

1. GET `/pools/all`

Description: returns all pools in the chain state instrumented with denoms and TVL if available

Parameters: none

Response example:
```bash
curl "https://sqs.osmosis.zone/pools/all" | jq .
[
  {
    "underlying_pool": {
      "address": "osmo164pg0key096kxe7940h45csw5w9cmf4nc2a83m73e8tc0u2eymeqj7t0rm",
      "incentives_address": "osmo1z6ae7saxlwnnd6whl2svw0k3zhsjdvzaap9ya2lcj8m3c0r02a6svtuzn2",
      "spread_rewards_address": "osmo1jhzypnueeg75cm8tps7srn7ujva9qe4msjmtkajwaqpj0kyvdjjqu0xule",
      "id": 1323,
      "current_tick_liquidity": "1000000.000000000000100000",
      "token0": "ibc/ECBE78BF7677320A93E7BA1761D144BCBF0CBC247C290C049655E106FE5DC68E",
      "token1": "uosmo",
      "current_sqrt_price": "1.000000000000000000000000000000000000",
      "tick_spacing": 100,
      "exponent_at_price_one": -6,
      "spread_factor": "0.000500000000000000",
      "last_liquidity_update": "2023-12-06T14:36:29.772040341Z"
    },
    "sqs_model": {
      "total_value_locked_uosmo": "1000000",
      "total_value_locked_error": "error getting token precision ibc/ECBE78BF7677320A93E7BA1761D144BCBF0CBC247C290C049655E106FE5DC68E",
      "balances": [
        {
          "denom": "ibc/ECBE78BF7677320A93E7BA1761D144BCBF0CBC247C290C049655E106FE5DC68E",
          "amount": "1000000"
        },
        {
          "denom": "uosmo",
          "amount": "1000000"
        }
      ],
      "pool_denoms": [
        "ibc/ECBE78BF7677320A93E7BA1761D144BCBF0CBC247C290C049655E106FE5DC68E",
        "uosmo"
      ],
      "spread_factor": "0.000500000000000000"
    }
  },
  ...
]
```

## Router Resource


1. GET `/router/quote?tokenIn=<tokenIn>&tokenOutDenom=<tokenOutDenom>`

Description: returns the best quote it can compute for the given tokenIn and tokenOutDenom

Parameters:
- `tokenIn` the string representation of the sdk.Coin for the token in
- `tokenOutDenom` the string representing the denom of the token out

Response example:

```bash
curl "https://sqs.osmosis.zone/router/quote?tokenIn=1000000uosmo&tokenOutDenom=uion" | jq .
{
  "amount_in": {
    "denom": "uosmo",
    "amount": "1000000"
  },
  "amount_out": "1803",
  "route": [
    {
      "pools": [
        {
          "id": 2,
          "type": 0,
          "balances": [],
          "spread_factor": "0.005000000000000000",
          "token_out_denom": "uion",
          "taker_fee": "0.001000000000000000"
        }
      ],
      "out_amount": "1803",
      "in_amount": "1000000"
    }
  ],
  "effective_fee": "0.006000000000000000"
}
```

2. GET `/router/single-quote?tokenIn=<tokenIn>&tokenOutDenom=<tokenOutDenom>`

Description: returns the best quote it can compute w/o performing route splits,
performing single direct route estimates only.

Parameters:
- `tokenIn` the string representation of the sdk.Coin for the token in
- `tokenOutDenom` the string representing the denom of the token out

Response example:
```bash
curl "https://sqs.osmosis.zone/router/single-quote?tokenIn=1000000uosmo&tokenOutDenom=uion" | jq .
{
  "amount_in": {
    "denom": "uosmo",
    "amount": "1000000"
  },
  "amount_out": "1803",
  "route": [
    {
      "pools": [
        {
          "id": 2,
          "type": 0,
          "balances": [],
          "spread_factor": "0.005000000000000000",
          "token_out_denom": "uion",
          "taker_fee": "0.001000000000000000"
        }
      ],
      "out_amount": "1803",
      "in_amount": "1000000"
    }
  ],
  "effective_fee": "0.006000000000000000"
}
```

3. GET `/router/routes?tokenIn=<tokenIn>&tokenOutDenom=<tokenOutDenom>`

Description: returns all routes that can be used for routing from tokenIn to tokenOutDenom

Parameters:
- `tokenIn` the string representation of the denom of the token in
- `tokenOutDenom` the string representing the denom of the token out


Response example:
```bash
curl "https://sqs.osmosis.zone/routes?tokenIn=uosmo&tokenOutDenom=uion" | jq .
{
  "Routes": [
    {
      "Pools": [
        {
          "ID": 1100,
          "TokenOutDenom": "uion"
        }
      ]
    },
    {
      "Pools": [
        {
          "ID": 2,
          "TokenOutDenom": "uion"
        }
      ]
    },
    {
      "Pools": [
        {
          "ID": 1013,
          "TokenOutDenom": "uion"
        }
      ]
    },
    {
      "Pools": [
        {
          "ID": 1092,
          "TokenOutDenom": "ibc/E6931F78057F7CC5DA0FD6CEF82FF39373A6E0452BF1FD76910B93292CF356C1"
        },
        {
          "ID": 476,
          "TokenOutDenom": "uion"
        }
      ]
    },
    {
      "Pools": [
        {
          "ID": 1108,
          "TokenOutDenom": "ibc/9712DBB13B9631EDFA9BF61B55F1B2D290B2ADB67E3A4EB3A875F3B6081B3B84"
        },
        {
          "ID": 26,
          "TokenOutDenom": "uion"
        }
      ]
    }
  ],
  "UniquePoolIDs": {
    "1013": {},
    "1092": {},
    "1100": {},
    "1108": {},
    "2": {},
    "26": {},
    "476": {}
  }
}
```

4. GET `/router/custom-quote?tokenIn=<tokenIn>&tokenOutDenom=<tokenOutDenom>&poolIDs=<poolIDs>`

Description: returns the quote over route with the given poolIDs. If such route does not exist, returns error.

Parameters:
- `tokenIn` the string representation of the sdk.Coin for the token in
- `tokenOutDenom` the string representing the denom of the token out
- `poolIDs` comma-separated list of pool IDs

Response example:
```bash
curl "https://sqs.osmosis.zone/router/custom-quote?tokenIn=1000000uosmo&tokenOutDenom=uion?poolIDs=2" | jq .
{
  "amount_in": {
    "denom": "uosmo",
    "amount": "1000000"
  },
  "amount_out": "1803",
  "route": [
    {
      "pools": [
        {
          "id": 2,
          "type": 0,
          "balances": [],
          "spread_factor": "0.005000000000000000",
          "token_out_denom": "uion",
          "taker_fee": "0.001000000000000000"
        }
      ],
      "out_amount": "1803",
      "in_amount": "1000000"
    }
  ],
  "effective_fee": "0.006000000000000000"
}
```


5. GET `/router/cached-routes?tokenIn=uosmo&tokenOutDenom=uion`

Description: returns cached routes for the given tokenIn and tokenOutDenomn if cache
is enabled. If not, returns error. Contrary to `/router/routes...` endpoint, does
not attempt to compute routes if cache is not enabled.

Parameters: none

Parameters:
- `tokenIn` the string representation of the denom of the token in
- `tokenOutDenom` the string representing the denom of the token out


Response example:
```bash
curl "https://sqs.osmosis.zone/cached-routes?tokenIn=uosmo&tokenOutDenom=uion" | jq .
{
  "Routes": [
    {
      "Pools": [
        {
          "ID": 1100,
          "TokenOutDenom": "uion"
        }
      ]
    },
    {
      "Pools": [
        {
          "ID": 2,
          "TokenOutDenom": "uion"
        }
      ]
    },
    {
      "Pools": [
        {
          "ID": 1013,
          "TokenOutDenom": "uion"
        }
      ]
    },
    {
      "Pools": [
        {
          "ID": 1092,
          "TokenOutDenom": "ibc/E6931F78057F7CC5DA0FD6CEF82FF39373A6E0452BF1FD76910B93292CF356C1"
        },
        {
          "ID": 476,
          "TokenOutDenom": "uion"
        }
      ]
    },
    {
      "Pools": [
        {
          "ID": 1108,
          "TokenOutDenom": "ibc/9712DBB13B9631EDFA9BF61B55F1B2D290B2ADB67E3A4EB3A875F3B6081B3B84"
        },
        {
          "ID": 26,
          "TokenOutDenom": "uion"
        }
      ]
    }
  ],
  "UniquePoolIDs": {
    "1013": {},
    "1092": {},
    "1100": {},
    "1108": {},
    "2": {},
    "26": {},
    "476": {}
  }
}
```

6. POST `/router/store-state`

Description: stores the current state of the router in a JSON file locally. Used for debugging purposes.
This endpoint should be disabled in production.

Parameters: none

## System Resource

1. GET `/system/healthcheck`

Description: returns 200 if the server is healthy.
Validates the following conditions:
- Redis is reachable
- Node is reachable
- Node is not synching
- The latest height in Redis is within threshold of the latest height in the node
- The latest height in Redis was updated within a configurable number of seconds

2. GET `/system/metrics`

Description: returns the prometheus metrics for the server

## Development Setup

### Mainnet

To setup a development environment against mainnet, sync the node in the default
home directory and then run the following commands:

```bash
# Starts a detached redis container, to stop: 'make redis-stop'
make redis-start

# Rebuild the binary and start the node with sqs enabled in-process
make sqs-start
```

### Localosmosis

It is also possible to run the sidecar query server against a localosmosis node.

```bash
# Starts localosmosis with all services enabled and a few pools pre-created
# make localnset-start for empty state
# See localosmosis docs for more details
make localnet-start-with-state
```

## Data

### Pools

For every chain pool, its pool model is written to Redis.

Additionally, we instrument each pool model with bank balances and OSMO-denominated TVL.

Some pool models do not contain balances by default. As a result, clients have to requery balance
for each pool model directly from chain. Having the balances in Redis allows us to avoid this and serve
pools with balances directly.

The routing algorithm requires the knowledge of TVL for prioritizing pools. As a result, each pool model
is instrumented with OSMO-denominated TVL.

### Router

For routing, we must know about the taker fee for every denom pair. As a result, in the router
repository, we stote the taker fee keyed by the denom pair.

These taker fees are then read from Redis to initialize the router.

### Token Precision

The chain is agnostic to token precision. As a result, to compute OSMO-denominated TVL,
we query [chain registry file](https://github.com/osmosis-labs/assetlists/blob/main/osmosis-1/osmosis-1.assetlist.json)
parse the precision exponent and use it scaling the spot price to the right value.

The following are the tokens that are either malformed or are missing from the chain registry file:
```md
ibc/CD942F878C80FBE9DEAB8F8E57F592C7252D06335F193635AF002ACBD69139CC
ibc/FE2CD1E6828EC0FAB8AF39BAC45BC25B965BA67CCBC50C13A14BD610B0D1E2C4
ibc/4F3B0EC2FE2D370D10C3671A1B7B06D2A964C721470C305CBB846ED60E6CAA20
ibc/CD20AC50CE57F1CF2EA680D7D47733DA9213641D2D116C5806A880F508609A7A
ibc/52E12CF5CA2BB903D84F5298B4BFD725D66CAB95E09AA4FC75B2904CA5485FEB
ibc/49C2B2C444B7C5F0066657A4DBF19D676E0D185FF721CFD3E14FA253BCB9BC04
ibc/7ABF696369EFB3387DF22B6A24204459FE5EFD010220E8E5618DC49DB877047B
ibc/E27CD305D33F150369AB526AEB6646A76EC3FFB1A6CA58A663B5DE657A89D55D
factory/osmo130w50f7ta00dxkzpxemuxw7vnj6ks5mhe0fr8v/oDOGE
ibc/5BBB6F9C8ECA31508EE5B68F2E27B57532E1595C57D0AE5C8D64E1FBCB756247
ibc/00BC6883C29D45EAA021A55CFDD5884CA8EFF9D39F698A9FEF79E13819FF94F8
ibc/BCDB35B7390806F35E716D275E1E017999F8281A81B6F128F087EF34D1DFA761
ibc/020F5162B7BC40656FC5432622647091F00D53E82EE8D21757B43D3282F25424
ibc/D3A1900B2B520E45608B5671ADA461E1109628E89B4289099557C6D3996F7DAA
ibc/1271ACDB6421652A2230DECCAA365312A32770579C2B22D2B60A89FE39106611
ibc/DEA3B0BB0006C69E75D2247E8DC57878758790556487067F67748FDC237CE2AE
ibc/72D0C53912C461FC9251E3135459746380E9030C0BFDA13D45D3BAC47AE2910E
ibc/0E30775281643124D79B8670ACD3F478AC5FAB2B1CA1E32903D0775D8A8BB064
ibc/4E2A6E691D0EB60A25AE582F29D19B20671F723DF6978258F4043DA5692120AE
ibc/F2F19568D75125D7B88303ADC021653267443679780D6A0FD3E1EC318E0C51FD
factory/osmo19pw5d0jset8jlhawvkscj2gsfuyd5v524tfgek/TURKEY
```

Any pool containing these tokens would have the TVL error error set to
non-empty string, leading to the pool being deprioritized from the router.

## Open Questions

- How to handle atomicity between ticks and pools? E.g. let's say a block is written between the time initial pools are read
and the time the ticks are read. Now, we have data that is partially up-to-date.