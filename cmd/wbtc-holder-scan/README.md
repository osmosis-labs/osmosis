# wbtc-holder-scan

Lists every holder of native WBTC
(`factory/osmo1z0qrq605sjgcqpylfl4aa6s90x738j7m58wyatt0tdzflg2ha26q67k743/wbtc`) by scanning the
bank balances store of a mainnet node directly.

The bank denom index on osmosis-1 was never backfilled (#6923), so `DenomOwners` and any
index-based holder list miss accounts. On 2026-10-27 the index listed 1,457 holders with
81.71 WBTC, against a supply of 85.96 WBTC. This tool finds the rest, so the v33 WBTC to allBTC
migration can use a complete list.

The tool only reads state. It never commits to the database.

## Runbook

1. Use a **stopped** osmosis-1 node, or a copy of a recent snapshot. The tool takes the database
   and wasm cache locks, so it cannot run beside a live node. The home directory must contain
   `data/application.db` and `wasm/`: the WBTC controller contract is queried for its merchants
   and custodian.
2. Build from this branch, using the same Go toolchain as the node build:

   ```sh
   go build -o wbtc-holder-scan ./cmd/wbtc-holder-scan
   ```

3. Run:

   ```sh
   ./wbtc-holder-scan --home /path/to/.osmosisd --out ./wbtc-scan
   ```

   - `--db-backend` defaults to `app-db-backend` from `config/app.toml`, else `goleveldb`. Pass
     `pebbledb` if the node uses pebble and app.toml does not say so.
   - `--height` reads an older retained version. The default is the latest committed height.

   The full balances store is scanned, so the run takes minutes. A progress line is printed every
   5M balance entries.

4. Check the result. The last line printed must end in `supply_matches=true`, and the exit code
   must be 0. A non-zero exit means the holders did not sum to the total supply. Do not use the
   output in that case.
5. Send back the whole `--out` directory.

## Output

| File | Contents |
|---|---|
| `summary.json` | height, app hash, supply reconciliation, indexed and unindexed counts, per-category totals, issuer merchants and custodian |
| `holders.csv` | every holder: amount, spendable amount, whether it is in the denom index, category, whether v33 would migrate it, account facts, tags |
| `unindexed.txt` | holders missing from the denom index, one `address amount` per line |

`would_migrate` applies the proposed v33 rule. The account must have a public key and spendable
WBTC. It must not be a module account, contract, ICA host account, IBC escrow, CL pool
sub-address, WBTC merchant, WBTC custodian, or the allBTC transmuter.
