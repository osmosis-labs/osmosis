#!/bin/bash
# Goal: validate that the concentrated liquidity pool state is valid
# Since the concentrated liquidity pool is constructed from ticks and pool state, we first validate
# that what pools and ticks queries return are consistent
#
# Next, we confirm that in sqs the current bucket index points at bucket with liquidity that is equal
# to the pool's current tick liquidity in SQS.
#
# Finally, we compare the current bucket liquidity between chain and sqs.
#
# Note that we do not validate other pool state but this passing should give a good enough indication
# that the pool state is in-sync.

# Assumes running from project root.
script_dir=$(dirname "$(readlink -f "$0")")
source $script_dir/healthcheck.sh

# url=http://localhost:9092
url=$1

# First, make sure that SQS is healthy. Otherwise, these tests make no sense
perform_health_check $url

validate_cl_pool_state() {
    local pool_id=$1
    local url=$2

    ###############################################
    # 1. Query chain for the curren ticks
    chain_tick_query_resp=$(osmosisd q concentratedliquidity liquidity-per-tick-range $pool_id --output=json)
    bucket_index_resp=$(echo $chain_tick_query_resp | jq .bucket_index)
    echo "chain_bucket_index: $bucket_index_resp"

    # Parse bucket index
    unquoted_bucket_idx="${bucket_index_resp//\"}"
    bucket_index=$(expr "$unquoted_bucket_idx" + 0)

    # Get current bucket
    query_current_bucket_liquidity=$(echo $chain_tick_query_resp | jq .liquidity[$bucket_index].liquidity_amount)

    echo "query_current_bucket_liquidity: $query_current_bucket_liquidity"

    ###############################################
    # 2. Get current bucket liqudity for pool from chain
    pools_resp=$(osmosisd q poolmanager pool $pool_id --output=json)
    pool_current_bucket_liquidity=$(echo $pools_resp | jq .pool.current_tick_liquidity)

    echo "pool_current_bucket_liquidity: $pool_current_bucket_liquidity"

    ###############################################
    # 3. Get SQS current bucket liquidity
    ticks_url="$url/pools/ticks/$pool_id"
    echo "ticks_url" $ticks_url
    sqs_ticks_resp=$(curl $ticks_url)
    sqs_bucket_index=$(echo $sqs_ticks_resp | jq .current_tick_index)

    echo "sqs_bucket_index" $sqs_bucket_index

    sqs_ticks_current_bucket_liquidity=$(echo $sqs_ticks_resp | jq .ticks[$sqs_bucket_index].liquidity_amount)

    ###############################################
    # 4. Get current bucket liquidity from pool on SQS

    sqs_pools_resp=$(curl "$url/pools/$pool_id")

    sqs_pools_current_bucket_liquidity=$(echo $sqs_pools_resp | jq .underlying_pool.current_tick_liquidity)

    echo "sqs_pools_current_bucket_liquidity: $sqs_pools_current_bucket_liquidity"

    ###############################################
    # 5. Compare the results

    # 5.1 Compare chain pool and chain ticks current bucket liquidity
    if [ "$query_current_bucket_liquidity" == "$pool_current_bucket_liquidity" ]; then
        echo "chain pool and chain ticks current bucket liquidity match - PASS -" $query_current_bucket_liquidity
    else
        echo "chain pool and chain ticks current bucket liquidity do not match - FAIL - " $query_current_bucket_liquidity $pool_current_bucket_liquidity
        exit 1
    fi

    # 5.2 Compare SQS pool and SQS ticks current bucket liquidity

    if [ "$sqs_ticks_current_bucket_liquidity" == "$sqs_pools_current_bucket_liquidity" ]; then
        echo "sqs pool and sqs ticks current bucket liquidity match - PASS -" $sqs_ticks_current_bucket_liquidity
    else
        echo "sqs pool and sqs ticks current bucket liquidity do not match - FAIL - " $sqs_ticks_current_bucket_liquidity $sqs_pools_current_bucket_liquidity
        exit 1
    fi

    # 5.3 Compare chain and SQS current bucket liquidity

    if [ "$query_current_bucket_liquidity" == "$sqs_ticks_current_bucket_liquidity" ]; then
        echo "chain and sqs ticks current bucket liquidity match - PASS -" $query_current_bucket_liquidity
    else
        echo "chain and sqs ticks current bucket liquidity do not match - FAIL - " $query_current_bucket_liquidity $sqs_ticks_current_bucket_liquidity
        exit 1
    fi
}

# Get all concentrated pools
pools=$(osmosisd q concentratedliquidity pools --output=json | jq .pools)

# Iterate over all pools and validate
echo $pools | jq -c '.[]' | while read -r entry; do
    # echo $entry

    pool_id=$(echo $entry | jq .id)

    # pool_id=$((pool_id_str))

    pool_id_no_quotes=${pool_id//\"/}
    validate_cl_pool_state $pool_id_no_quotes $url
done
