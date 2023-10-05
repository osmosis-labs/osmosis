# Concentrated Liquidity Go Client

This Go-client allows connecting to an Osmosis chain via Ignite CLI and
setting up a concentrated liquidity pool with positions.

## General Setup FAQ

- Update constants at the top of the file accordingly.
   * Make sure keyring is set up.
   * Client home is pointing to the right place.

## LocalOsmosis Setup

Make sure that you run `localosmosis` in the background and have keys
added to your keyring with:

```bash
make set-env localosmosis # sets environment to $HOME/.osmosisd-local

make localnet-start

make localnet-keys
```

See `tests/localosmosis` for more info.

## Testnet Setup

Configure a different `osmosisd` environment with configs.

```bash
make set-env .osmosisd-testnet-script

osmosisd init test-script

cd $HOME/.osmosisd-testnet-script/config

nano client.toml
```

Replace node RPC with the testnet value and save,

Next, manually edit the `localosmosisFromHomePath` variable in the script:
<https://github.com/osmosis-labs/osmosis/blob/98025f185ab2ee1b060511ed22679112abcc08fa/tests/cl-go-client/main.go#L28>

Set the value to `.osmosisd-testnet-script` and save.

Now, you are able to run this script on testnet. This assummes that
testnet accounts have been set up with the default test accounts
and balances. By default, we mean accounts created with
`make localnet-keys`.

## Running

### Crete Positions

```bash
make localnet-cl-create-positions
```

In the current state, it does the following:
- Queries status of the chain to make sure it's running.
- Queries pool with id 1. If does not exist, creates it
- Sets up 100 CL positions (count configured at the top of the file)

### Make Small Randomized Swaps

```bash
make localnet-cl-small-swap
```

In the current state, it does the following:
- Queries status of the chain to make sure it's running.
- Queries pool with id 1.
- Performs 100 randomized swaps against the pool with id 1.

Note that this script does not set up positions, assumming they are
already set up.

### Create Positions and Make Small Randomized Swaps

```bash
make localnet-cl-positions-small-swaps
```

This script runs "Create Positions" and "Make Small Randomized Swaps" scripts in sequence.

### Make Large Invertible Swaps

```bash
make localnet-cl-large-swap
```

In the current state, it does the following:
- Queries status of the chain to make sure it's running.
- Queries pool with id 1.
- Performs 100 large swaps where the output of the previous swap is swapped back at the
next swap.

In other words, it takes one large amount and swaps it into the pool. Then, takes output token
and swaps it back while accounting for the spread factor. This is done to
ensure that we cross ticks while minimizing the chance of running out of funds or liquidity.

Note that this script does not set up positions, assumming they are
already set up.

### Create Positions and Make Large Invertible Swaps

```bash
make localnet-cl-positions-large-swaps
```

This script runs "Create Positions" and "Make Large Invertible Swaps" scripts in sequence.

### Create Incentive

Creates a new gauge and waits for one epoch so that the gauge
is converted into an incentive record for pool id 1.

```bash
make localnet-cl-external-incentive
```

### Create Pool

Attempts to create a CL pool at id 1. If pool at id 1 already exists, this is a no-op.
If pool with different id is desired, tweak expectedPoolId in the script.

```bash
make localnet-cl-create-pool
```

