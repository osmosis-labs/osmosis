#!/usr/bin/env bash
# Runs the v32 upgrade handler on a copy of osmosis-1 mainnet state and checks the HUAHUA recovery.
#
# Prerequisites:
#   - NODE_HOME contains a recent osmosis-1 snapshot (data/, wasm/) plus config/ with the osmosis-1
#     genesis. The node must not have peers configured: it should never talk to mainnet.
#   - OLD_BIN is the osmosisd release currently running on mainnet (v31.x).
#   - BIN is osmosisd built from this branch (make build).
#   - The API server is enabled in app.toml and reachable at API.
#
# The script follows the same path as a real upgrade:
#   1. starts OLD_BIN on the untouched snapshot and records the mainnet state,
#   2. forks it into a single-validator chain with `OLD_BIN in-place-testnet --trigger-testnet-upgrade v32`;
#      this schedules v32 ten blocks later, and OLD_BIN halts there with "UPGRADE "v32" NEEDED".
#      The state at the halt height is recorded as the pre-upgrade state,
#   3. starts BIN, which runs the v32 upgrade handler, records the post-upgrade state and compares.
#
# Usage:
#   OLD_BIN=./osmosisd-v31 BIN=./build/osmosisd NODE_HOME=~/osmosis-fork/home \
#   API=http://127.0.0.1:1317 RPC=http://127.0.0.1:26657 ./scripts/huahua-recovery/mainnet_fork_test.sh
#
# Set KEEP_RUNNING=1 to leave the forked chain running afterwards (for example to watch the next epoch).
set -euo pipefail

OLD_BIN="${OLD_BIN:?set OLD_BIN to the osmosisd release running on mainnet (v31.x)}"
BIN="${BIN:?set BIN to the osmosisd binary built from this branch}"
NODE_HOME="${NODE_HOME:?set NODE_HOME to the node home holding the mainnet snapshot}"
API="${API:-http://127.0.0.1:1317}"
RPC="${RPC:-http://127.0.0.1:26657}"
OUT="${OUT:-$NODE_HOME/../huahua-fork-test}"
KEEP_RUNNING="${KEEP_RUNNING:-0}"

HUAHUA="ibc/B9E0A1A524E98BB407D3CED8720EFEFD186002F90C1B1B7964811DD0CCC12228"
RECIPIENT="osmo14fketv99hlrlk80mkggw643spsj3yyf7t2pjhr"
GAUGES=(1954 1955 1956 1957 1958 1959)

mkdir -p "$OUT"
log() { echo "[$(date -u +%H:%M:%S)] $*"; }

wait_for_api() {
  for _ in $(seq 1 600); do
    if curl -sf "$API/osmosis/incentives/v1beta1/gauge_by_id/1954" >/dev/null 2>&1; then return 0; fi
    sleep 2
  done
  log "API did not come up"; return 1
}

wait_for_height_above() {
  local target=$1
  for _ in $(seq 1 900); do
    h=$(curl -sf "$RPC/status" 2>/dev/null | jq -r .result.sync_info.latest_block_height 2>/dev/null || echo 0)
    if [[ "$h" =~ ^[0-9]+$ ]] && (( h > target )); then return 0; fi
    sleep 2
  done
  log "chain did not pass height $target"; return 1
}

stop_node() {
  local pid=$1
  kill -INT "$pid" 2>/dev/null || true
  for _ in $(seq 1 120); do kill -0 "$pid" 2>/dev/null || return 0; sleep 1; done
  kill -KILL "$pid" 2>/dev/null || true
}

# Writes the state relevant to the recovery to $OUT/<label>.json.
capture_state() {
  local label=$1
  local file="$OUT/$label.json" module height
  module=$(curl -sf "$API/cosmos/auth/v1beta1/module_accounts/incentives" | jq -r .account.base_account.address)
  height=$(curl -sf "$API/cosmos/base/tendermint/v1beta1/blocks/latest" | jq -r .block.header.height)

  {
    echo "{"
    echo "\"height\": $height,"
    echo "\"module_address\": \"$module\","
    echo "\"recipient_huahua\": \"$(curl -sf "$API/cosmos/bank/v1beta1/balances/$RECIPIENT/by_denom?denom=$HUAHUA" | jq -r .balance.amount)\","
    echo "\"module_huahua\": \"$(curl -sf "$API/cosmos/bank/v1beta1/balances/$module/by_denom?denom=$HUAHUA" | jq -r .balance.amount)\","
    echo "\"supply_huahua\": \"$(curl -sf "$API/cosmos/bank/v1beta1/supply/by_denom?denom=$HUAHUA" | jq -r .amount.amount)\","
    echo "\"to_distribute_huahua\": \"$(curl -sf "$API/osmosis/incentives/v1beta1/module_to_distribute_coins" | jq -r --arg d "$HUAHUA" '[.coins[] | select(.denom == $d) | .amount][0] // "0"')\","
    echo "\"gamm605_denom_index\": $(curl -sf "$API/osmosis/incentives/v1beta1/active_gauges_per_denom?denom=GAMM605" | jq '[.data[].id | tonumber]'),"
    echo "\"active_gauge_ids\": $(curl -sf "$API/osmosis/incentives/v1beta1/active_gauges?pagination.limit=1000000" | jq '[.data[].id | tonumber]'),"
    echo "\"gauges\": {"
    local first=1
    for id in "${GAUGES[@]}"; do
      [[ $first == 1 ]] || echo ","
      first=0
      echo "\"$id\": $(curl -sf "$API/osmosis/incentives/v1beta1/gauge_by_id/$id" | jq .gauge)"
    done
    echo "}}"
  } | jq . > "$file"
  log "$label state at height $height saved to $file"
}

# ---------------------------------------------------------------------------------------------
log "1/3 starting $("$OLD_BIN" version 2>&1) on the untouched snapshot to record the mainnet state"
"$OLD_BIN" start --home "$NODE_HOME" > "$OUT/node-mainnet.log" 2>&1 &
PID=$!
wait_for_api
capture_state mainnet
stop_node "$PID"

# ---------------------------------------------------------------------------------------------
log "2/3 forking mainnet state into a local chain with $("$OLD_BIN" version 2>&1) and scheduling v32"
if ! "$OLD_BIN" keys show forkval --keyring-backend test --home "$NODE_HOME" >/dev/null 2>&1; then
  "$OLD_BIN" keys add forkval --keyring-backend test --home "$NODE_HOME" > "$OUT/forkval-key.txt" 2>&1
fi
OPERATOR=$("$OLD_BIN" keys show forkval -a --keyring-backend test --home "$NODE_HOME")

"$OLD_BIN" in-place-testnet localosmosis "$OPERATOR" --home "$NODE_HOME" \
  --trigger-testnet-upgrade v32 --skip-confirmation > "$OUT/node-fork-v31.log" 2>&1 &
PID=$!
for _ in $(seq 1 900); do
  grep -q 'UPGRADE "v32" NEEDED' "$OUT/node-fork-v31.log" && break
  kill -0 "$PID" 2>/dev/null || { log "old binary exited before reaching the upgrade height"; exit 1; }
  sleep 2
done
grep -o 'UPGRADE "v32" NEEDED at height: [0-9]*' "$OUT/node-fork-v31.log" | head -1
UPGRADE_HEIGHT=$(grep -o 'UPGRADE "v32" NEEDED at height: [0-9]*' "$OUT/node-fork-v31.log" | head -1 | grep -o '[0-9]*$')
# The old binary halts but keeps serving queries on the last committed state (UPGRADE_HEIGHT - 1).
wait_for_api
capture_state before
stop_node "$PID"

log "starting $("$BIN" version 2>&1) (this branch) to run the v32 upgrade at height $UPGRADE_HEIGHT"
"$BIN" start --home "$NODE_HOME" > "$OUT/node-fork-v32.log" 2>&1 &
FORK_PID=$!
echo "$FORK_PID" > "$OUT/fork.pid"
wait_for_height_above $((UPGRADE_HEIGHT + 1))
wait_for_api
capture_state after

# ---------------------------------------------------------------------------------------------
log "3/3 checking the results"
set +e
grep -E "v32 upgrade: HUAHUA (recovered|recovery skipped)" "$OUT/node-fork-v32.log" | tail -1 | tee "$OUT/upgrade-log-line.txt"

python3 - "$OUT/mainnet.json" "$OUT/before.json" "$OUT/after.json" "$HUAHUA" <<'EOF' | tee "$OUT/result.txt"
import json, sys

mainnet, before, after = (json.load(open(p)) for p in sys.argv[1:4])
huahua = sys.argv[4]
gauges = [str(i) for i in range(1954, 1960)]
failures = []

def check(cond, msg):
    print(("PASS  " if cond else "FAIL  ") + msg)
    if not cond:
        failures.append(msg)

def amount(coins, denom):
    return sum(int(c["amount"]) for c in coins if c["denom"] == denom)

stuck = sum(amount(before["gauges"][g]["coins"], huahua) - amount(before["gauges"][g]["distributed_coins"], huahua) for g in gauges)
print(f"mainnet snapshot height {mainnet['height']}, pre-upgrade height {before['height']}, post-upgrade height {after['height']}")
check(mainnet["gauges"] == before["gauges"], "gauges 1954-1959 on the fork before the upgrade are identical to mainnet")
print(f"undistributed HUAHUA in gauges 1954-1959 before the upgrade: {stuck} uhuahua ({stuck / 1e6:,.6f} HUAHUA)")

for g in gauges:
    b = before["gauges"][g]
    check(int(b["filled_epochs"]) < int(b["num_epochs_paid_over"]) and g in map(str, before["active_gauge_ids"]),
          f"gauge {g} was active before the upgrade ({b['distribute_to']['denom']}, filled {b['filled_epochs']}/{b['num_epochs_paid_over']})")

rec_b, rec_a = int(before["recipient_huahua"]), int(after["recipient_huahua"])
mod_b, mod_a = int(before["module_huahua"]), int(after["module_huahua"])
check(rec_a - rec_b == stuck, f"recipient received {rec_a - rec_b} uhuahua (expected {stuck})")
check(mod_b - mod_a == stuck, f"incentives module sent {mod_b - mod_a} uhuahua (expected {stuck}); {mod_a} uhuahua left")
check(before["supply_huahua"] == after["supply_huahua"], f"HUAHUA total supply unchanged ({after['supply_huahua']})")

td_b, td_a = int(before["to_distribute_huahua"]), int(after["to_distribute_huahua"])
check(td_b - td_a == stuck, f"module_to_distribute_coins HUAHUA decreased by {td_b - td_a} (expected {stuck})")

for g in gauges:
    a = after["gauges"][g]
    check(a["filled_epochs"] == a["num_epochs_paid_over"], f"gauge {g} filled_epochs = {a['filled_epochs']}/{a['num_epochs_paid_over']}")
    check(a["distributed_coins"] == a["coins"], f"gauge {g} distributed_coins == coins")
    check(a["coins"] == before["gauges"][g]["coins"], f"gauge {g} coins field unchanged")

removed = set(before["active_gauge_ids"]) - set(after["active_gauge_ids"])
added = set(after["active_gauge_ids"]) - set(before["active_gauge_ids"])
check(removed == set(map(int, gauges)), f"active gauges removed: {sorted(removed)}")
check(not added, f"no gauges added to the active set ({sorted(added)})")
print(f"active gauges: {len(before['active_gauge_ids'])} -> {len(after['active_gauge_ids'])}")
check(after["gamm605_denom_index"] == [], f"GAMM605 denom index after: {after['gamm605_denom_index']} (before: {before['gamm605_denom_index']})")

print()
print("RESULT: " + ("ALL CHECKS PASSED" if not failures else f"{len(failures)} CHECK(S) FAILED"))
sys.exit(1 if failures else 0)
EOF
STATUS=${PIPESTATUS[0]}
set -e

if [[ "$KEEP_RUNNING" == "1" ]]; then
  log "forked chain left running (pid $FORK_PID, log $OUT/node-fork-v32.log)"
else
  stop_node "$FORK_PID"
fi
exit "$STATUS"
