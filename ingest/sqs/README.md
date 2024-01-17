# Sidecar Query Server

This is a sidecar query server that is used for performing query tasks outside of the main chain.
The high-level architecture is that the chain reads data at the end of the block, parses it
and then writes into a Redis instance.

The sidecar query server then reads the parsed data from Redis and serves it to the client
via HTTP endpoints.

The use case for this is performing certain data and computationally intensive tasks outside of
the chain node or the clients. For example, routing falls under this category because it requires
all pool data for performing the complex routing algorithm.

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