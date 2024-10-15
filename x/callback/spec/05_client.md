# Client

Section describes interaction with the module by the user

## CLI

### Query

The `query` commands alllows a user to query the module state

Use the `-h`/`--help` flag to get a help description of a command.

`osmosisd q callback -h`

> You can add the `-o json` for the JSON output format

#### params

Get the current module parameters

Usage:

`osmosisd q callback params [flags]`

Example output:

```yaml
block_reservation_fee_multiplier: "1.000000000000000000"
callback_gas_limit: "1000000"
future_reservation_fee_multiplier: "1.000000000000000000"
max_block_reservation_limit: "3"
max_future_reservation_limit: "10000"
```

#### callbacks

List all the callbacks for the given height

Usage:

`osmosisd q callback callbacks [block-height]`

Example:

`osmosisd q callback callbacks 1234`

Example output:

```yaml
callbacks:
- callback_height: "400"
  contract_address: cosmos1wug8sewp6cedgkmrmvhl3lf3tulagm9hnvy8p0rppz9yjw0g4wtqukxvuk
  fee_split:
    block_reservation_fees:
      amount: "1"
      denom: stake
    future_reservation_fees:
      amount: "80"
      denom: stake
    surplus_fees:
      amount: "88"
      denom: stake
    transaction_fees:
      amount: "0"
      denom: stake
  job_id: "5"
  reserved_by: cosmos1x394ype3x8nt9wz0j78m8c8kcezpslrcnvs6ef
```

#### estimate-callback-fees

Estimate the minimum fees to be paid to register a callback based on the requested height

Usage:

`osmosisd q callback estimate-callback-fees [block-height]`

Example:

`osmosisd q calback estimate-callback-fees 1234`

Example output:

```yaml
fee_split:
  block_reservation_fees:
    amount: "2000"
    denom: stake
  future_reservation_fees:
    amount: "1000"
    denom: stake
  surplus_fees: null
  transaction_fees:
    amount: "5000"
    denom: stake
total_fees:
  amount: "7000"
  denom: stake
```

### TX

The `tx` commands allows a user to interact with the module.

Use the `-h`/`--help` flag to get a help description of a command.

`osmosisd tx callback -h`

#### request-callback

Create a new callback for the given contract at specified height and given job id by paying the mentioned fees

Usage:

`osmosisd tx callback request-callback [contract-address] [job-id] [callback-height] [fee-amount] [flags]`

Example:

`osmosisd tx callback request-callback cosmos1wug8sewp6cedgkmrmvhl3
lf3tulagm9hnvy8p0rppz9yjw0g4wtqukxvuk 1 1234 7000stake --from myAccountKey`

#### cancel-callback

Cancel an existing callback for the given contract at specified height and given job id

Usage:

`osmosisd tx callback cancel-callback [contract-address] [job-id] [callback-height] [flags]`

Example:

`osmosisd tx callback cancel-callback cosmos1wug8sewp6cedgkmrmvhl3 1 1234  --from myAccountKey`
