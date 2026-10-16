# v32 HUAHUA recovery: how to reproduce and verify

Governance [proposal 609](https://www.mintscan.io/osmosis/proposals/609) (passed on 2023-09-06:
160.98M yes, 0.61M no, 0.0002M no-with-veto, 22.11M abstain) approved merging the code needed to
recover 6,000,000,000 HUAHUA stuck in misconfigured incentive gauges.
Background: [forum thread](https://forum.osmosis.zone/t/save-6bn-huahua-from-locked-pool/238).

## The problem

In January 2022 six external gauges were created for pools 605 (HUAHUA/OSMO) and 606 (HUAHUA/ATOM)
with lock denoms `GAMM605` / `GAMM606` instead of `gamm/pool/605` / `gamm/pool/606`.
No lock can match these denoms, so the gauges never distributed:

| gauge | denom   | duration | uhuahua               | filled epochs |
|-------|---------|----------|-----------------------|---------------|
| 1954  | GAMM605 | 1d       | 353,000,000,000,000   | 0 / 180       |
| 1955  | GAMM605 | 7d       | 882,500,000,000,000   | 0 / 180       |
| 1956  | GAMM605 | 14d      | 1,764,500,000,000,000 | 0 / 180       |
| 1957  | GAMM606 | 1d       | 353,000,000,000,000   | 0 / 180       |
| 1958  | GAMM606 | 7d       | 882,500,000,000,000   | 0 / 180       |
| 1959  | GAMM606 | 14d      | 1,764,500,000,000,000 | 0 / 180       |

Total: 6,000,000,000,000,000 uhuahua (`ibc/B9E0A1A5…2228`, `transfer/channel-113`), held by the
incentives module account `osmo1krxwf5e308jmclyhfd9u92kp369l083wequge6`.

Note: gauges 1954-1956 are indexed under the denom `GAMM605`, but 1957-1959 have **no** entry under
`GAMM606`. Calling the existing `moveActiveGaugeToFinishedGauge` on them would fail, so the new keeper
method tolerates a missing denom index.

## The change

- `x/incentives/keeper/gauge.go`: `ForceFinishGaugeAndSendUndistributed(ctx, gaugeID, recipient)`.
  Only callable from Go (no message, no query). For an active, non-perpetual, non-group gauge it:
  1. sends `Coins - DistributedCoins` from the incentives module to `recipient`;
  2. sets `DistributedCoins = Coins` and `FilledEpochs = NumEpochsPaidOver`, so the module accounting
     (`GetModuleToDistributeCoins` / `GetModuleDistributedCoins`) matches the balance change;
  3. moves the gauge from the active to the finished set and drops its denom index if present;
  4. emits a `force_finish_gauge` event.
- `app/upgrades/v32`: calls it for gauges 1954-1959 with recipient
  `osmo14fketv99hlrlk80mkggw643spsj3yyf7t2pjhr` (Chihuahua ecosystem fund).
  Before touching a gauge the handler checks that it distributes to the expected `GAMM60x` denom
  and holds HUAHUA. The recovery runs in a cache context: if any check fails, nothing is written,
  an error is logged, and **the upgrade continues** (the chain is never halted by this recovery).
  Anyone could add coins to these gauges before the upgrade (`MsgAddToGauge`), so amounts are not
  hard-coded: whatever the gauges hold is recovered, including any non-HUAHUA coins added to them
  (a finished gauge could never distribute those). Once finished, no one can add to them anymore.

## 1. Unit and upgrade tests

```bash
go test ./x/incentives/keeper/ -run 'TestKeeperTestSuite/TestForceFinishGaugeAndSendUndistributed' -v -count=1
go test ./app/upgrades/v32/ -v -count=1
```

`app/upgrades/v32/upgrades_test.go` writes gauges 1954-1959 into state exactly as on mainnet (IDs, denoms,
durations, amounts, start time, missing `GAMM606` index, 14382 uhuahua dust in the module) plus
an unrelated control gauge, runs the real v32 upgrade through the upgrade module's `PreBlock`, and checks:

- `TestHuahuaRecovery`: recipient +6,000,000,000 HUAHUA; dust and the control gauge's funds stay in
  the module; the six gauges are finished; the control gauge is unchanged; the next `Distribute`
  works; `AddToGaugeRewards` on a recovered gauge fails.
- `TestHuahuaRecoveryWithExtraCoinsAdded`: coins added to a gauge before the upgrade are recovered too.
- `TestHuahuaRecoverySkippedOnUnexpectedState`: wrong denom, no HUAHUA, or gauge already finished →
  upgrade succeeds, no funds move, and none of the six gauges change (including the ones processed
  before the failing one).

Keeper tests cover: full recovery, partially distributed gauge, missing denom index, perpetual,
upcoming, not yet moved to the active set, already finished, and non-existent gauges.

Regression: `go test ./x/incentives/... ./app/upgrades/...` passes.

## 2. Check live mainnet state

```bash
./scripts/huahua-recovery/check_mainnet_state.sh            # LCD=... to use another endpoint
```

This is read-only. It checks every precondition the handler relies on, and prints a note if a gauge
holds denoms other than HUAHUA (those are sent to the recipient as well). Run it again just before the
upgrade height.

After the upgrade, confirm that the recovery was applied. The handler does not halt the chain when it
skips the recovery, it only logs `v32 upgrade: HUAHUA recovery skipped`, so this check is the on-chain
confirmation:

```bash
./scripts/huahua-recovery/check_mainnet_state.sh --post-upgrade
```

It fails unless all six gauges are finished with `distributed_coins == coins` and none is still active.

## 3. Mainnet-fork upgrade test

`mainnet_fork_test.sh` runs the upgrade on a copy of mainnet state, following the same path as the real
upgrade: the current release produces blocks until the upgrade height and halts, then the new binary
takes over and runs the v32 handler.

Requirements: a recent osmosis-1 snapshot (about 55 GB extracted, for example from Polkachu), the
current mainnet release binary and a binary built from this branch. The node must have no seeds or
peers configured, and the API must be enabled in `app.toml`.

```bash
# node home with the snapshot and the osmosis-1 genesis
mkdir -p ~/osmosis-fork/home
curl -L https://snapshots.polkachu.com/snapshots/osmosis/<latest>.tar.lz4 | lz4 -c -d - | tar -x -C ~/osmosis-fork/home
osmosisd init fork --chain-id osmosis-1 --home /tmp/fork-init && cp -n /tmp/fork-init/config/* ~/osmosis-fork/home/config/
curl -L https://snapshots.polkachu.com/genesis/osmosis/genesis.json -o ~/osmosis-fork/home/config/genesis.json
# in config.toml: seeds = "", persistent_peers = "", pex = false; in app.toml: [api] enable = true

# binaries
git clone --branch v31.x https://github.com/osmosis-labs/osmosis osmosis-v31 && (cd osmosis-v31 && make build)
make build   # this branch

OLD_BIN=osmosis-v31/build/osmosisd BIN=./build/osmosisd NODE_HOME=~/osmosis-fork/home \
API=http://127.0.0.1:1317 RPC=http://127.0.0.1:26657 ./scripts/huahua-recovery/mainnet_fork_test.sh
```

The script:

1. starts the old binary on the untouched snapshot and saves the mainnet state (`mainnet.json`);
2. runs `in-place-testnet localosmosis <operator> --trigger-testnet-upgrade v32` with the old binary,
   which schedules v32 ten blocks later and halts there with `UPGRADE "v32" NEEDED`; the halted node
   still answers queries, so the pre-upgrade state is saved (`before.json`);
3. starts the new binary, which runs the upgrade, saves `after.json` and compares the three.

It checks: the six gauges on the fork match mainnet; the recipient gets exactly the undistributed
HUAHUA; the module balance drops by the same amount; HUAHUA supply is unchanged;
`module_to_distribute_coins` drops by the same amount; the six gauges are at 180/180 filled epochs with
`distributed_coins == coins`; exactly these six gauges leave the active set and nothing else changes in
it; the `GAMM605` denom index is empty. `KEEP_RUNNING=1` leaves the chain running so the next epoch
can be observed (the fork shortens the day epoch to 6 hours).
