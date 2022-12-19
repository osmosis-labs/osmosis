#!/bin/sh

# Workaround due to config being hardcoded in frontend
# Replace the endpoints and chain id 
OLD_CHAIN_ID=osmo-test-4
OLD_RPC_ENDPOINT=rpc.testnet.osmosis.zone
OLD_REST_ENDPOINT=lcd.testnet.osmosis.zone

echo "Replacing $OLD_RPC_ENDPOINT with $RPC_ENDPOINT"
grep -rl $OLD_RPC_ENDPOINT /app/packages/web/.next/ | xargs sed -i "s#$OLD_RPC_ENDPOINT#$RPC_ENDPOINT#g"
grep -rl $OLD_RPC_ENDPOINT /app/node_modules/@osmosis-labs/web/ | xargs sed -i "s#$OLD_RPC_ENDPOINT#$RPC_ENDPOINT#g"

echo "Replacing $OLD_REST_ENDPOINT with $REST_ENDPOINT"
grep -rl $OLD_REST_ENDPOINT /app/packages/web/.next/ | xargs sed -i "s#$OLD_REST_ENDPOINT#$REST_ENDPOINT#g"
grep -rl $OLD_REST_ENDPOINT /app/node_modules/@osmosis-labs/web/ | xargs sed -i "s#$OLD_REST_ENDPOINT#$REST_ENDPOINT#g"

echo "Replacing $OLD_CHAIN_ID with $CHAIN_ID"
grep -rl $OLD_CHAIN_ID /app/packages/web/.next/ | xargs sed -i "s#$OLD_CHAIN_ID#$CHAIN_ID#g"
grep -rl $OLD_CHAIN_ID /app/node_modules/@osmosis-labs/web/ | xargs sed -i "s#$OLD_CHAIN_ID#$CHAIN_ID#g"

yarn start:testnet