#!/usr/bin/env bash
# Checks that the live Osmosis mainnet state still matches the preconditions of the v32
# HUAHUA recovery (app/upgrades/v32). Read-only; needs curl and jq.
#
#   LCD=https://lcd.osmosis.zone ./scripts/huahua-recovery/check_mainnet_state.sh
set -euo pipefail

LCD="${LCD:-https://lcd.osmosis.zone}"
HUAHUA="ibc/B9E0A1A524E98BB407D3CED8720EFEFD186002F90C1B1B7964811DD0CCC12228"
RECIPIENT="osmo14fketv99hlrlk80mkggw643spsj3yyf7t2pjhr"
declare -A DENOM=([1954]=GAMM605 [1955]=GAMM605 [1956]=GAMM605 [1957]=GAMM606 [1958]=GAMM606 [1959]=GAMM606)

fail() { echo "FAIL: $*"; exit 1; }

height=$(curl -sf "$LCD/cosmos/base/tendermint/v1beta1/blocks/latest" | jq -r .block.header.height)
echo "Osmosis mainnet height $height"

active=$(curl -sf "$LCD/osmosis/incentives/v1beta1/active_gauges?pagination.limit=100000" | jq -r '.data[].id')

total=0
for id in 1954 1955 1956 1957 1958 1959; do
  g=$(curl -sf "$LCD/osmosis/incentives/v1beta1/gauge_by_id/$id" | jq .gauge)
  denom=$(jq -r .distribute_to.denom <<<"$g")
  amount=$(jq -r --arg d "$HUAHUA" '[.coins[] | select(.denom == $d) | .amount][0] // "0"' <<<"$g")
  filled=$(jq -r .filled_epochs <<<"$g")
  paid_over=$(jq -r .num_epochs_paid_over <<<"$g")
  perpetual=$(jq -r .is_perpetual <<<"$g")

  [[ "$denom" == "${DENOM[$id]}" ]] || fail "gauge $id distributes to $denom, expected ${DENOM[$id]}"
  [[ "$amount" != "0" ]] || fail "gauge $id holds no HUAHUA"
  [[ "$perpetual" == "false" ]] || fail "gauge $id is perpetual"
  (( filled < paid_over )) || fail "gauge $id is already finished"
  grep -qx "$id" <<<"$active" || fail "gauge $id is not in the active gauges"

  echo "gauge $id: denom=$denom uhuahua=$amount filled=$filled/$paid_over active=yes"
  total=$(python3 -c "print($total + $amount)")
done

echo "total uhuahua in gauges: $total ($(python3 -c "print($total / 10**6)") HUAHUA)"
[[ "$total" == "6000000000000000" ]] || echo "NOTE: total differs from 6,000,000,000 HUAHUA (someone added to a gauge); the upgrade recovers whatever is there"

module=$(curl -sf "$LCD/cosmos/auth/v1beta1/module_accounts/incentives" | jq -r .account.base_account.address)
balance=$(curl -sf "$LCD/cosmos/bank/v1beta1/balances/$module/by_denom?denom=$HUAHUA" | jq -r .balance.amount)
echo "incentives module ($module) uhuahua balance: $balance"
python3 -c "import sys; sys.exit(0 if $balance >= $total else 1)" || fail "module balance below gauge total"

recipient=$(curl -sf "$LCD/cosmos/bank/v1beta1/balances/$RECIPIENT/by_denom?denom=$HUAHUA" | jq -r .balance.amount)
echo "recipient $RECIPIENT current uhuahua: $recipient"

echo "OK: mainnet state matches the v32 HUAHUA recovery preconditions"
