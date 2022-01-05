# run sed

export EXPORTED_GENESIS=testnet_genesis.json

# Replace Sentinel addrs/pubkeys

# priv_validator_key.json file for what we replace to
#	{
#		"address": "8B78E478777427CC3906B8234CB72BCEA2C78E83",
#		"pub_key": {
#			"type": "tendermint/PubKeyEd25519",
#			"value": "2OpBuqaXvXQ+lSxAoT1S7Jfyr56KiakTzvuFiuJK+X4="
#		},
#		"priv_key": {
#			"type": "tendermint/PrivKeyEd25519",
#			"value": "3OLLoEfdT+ZrLqpRCvytpXrhgKfeEBBoeaoXe1p3/mjY6kG6ppe9dD6VLEChPVLsl/KvnoqJqRPO+4WK4kr5fg=="
#		}
#	}

# replace validator cons pubkey
sed -i '' 's%b77zCh/VsRgVvfGXuW4dB+Dhg4PrMWWBC5G2K/qFgiU=%2OpBuqaXvXQ+lSxAoT1S7Jfyr56KiakTzvuFiuJK+X4=%g' $EXPORTED_GENESIS
# This is a PITA to get:
# take pubkey, do osmosisd debug pubkey {base64}
# take address out of it, feed into osmosisd debug addr {addr}
# pass that into osmosisd debug bech32-convert -p osmovalcons
sed -i '' 's%osmovalcons1z6skn9g6s7py0klztr7acutr3anqd52k9x5p70%osmovalcons13duwg7rhwsnucwgxhq35edete63v0r5rqp90es%g' $EXPORTED_GENESIS
# replace validator hex addr
sed -i '' 's%16A169951A878247DBE258FDDC71638F6606D156%8B78E478777427CC3906B8234CB72BCEA2C78E83%g' $EXPORTED_GENESIS
# replace operator address
# mnemonic: kitchen comic flower drip sick prize account cheese truth income weekend nominee segment punch call satisfy captain earth ethics wasp clump tunnel orchard advance
sed -i '' 's%osmovaloper1cyw4vw20el8e7ez8080md0r8psg25n0cq98a9n%osmovaloper1qye772qje88p7ggtzrvl9nxvty6dkuusvpqhac%g' $EXPORTED_GENESIS
# replace the actual account
sed -i '' 's%osmo1cyw4vw20el8e7ez8080md0r8psg25n0c6j07j5%osmo1qye772qje88p7ggtzrvl9nxvty6dkuuskkg52l%g' $EXPORTED_GENESIS
# replace that accounts pubkey, obtained via auth. New pubkey obtained via new debug command
sed -i '' 's%AqlNb1FM8veQrT4/apR5B3hww8VApc0LTtZnXhq7FqG0%A9zC0Sa0VCK/lVLi1Kv0C1c0MQp47d+yjFqb6dAUza0a%g' $EXPORTED_GENESIS

# now time to replace the delegation share amounts for the self-bond
# before:      5000000000.000000000000000000 (5000 osmo)
# after: 1000005000000000.000000000000000000 (1 BN + 5k osmo = 1,000,005,000)
# There are two spots for this, one is the delegator shares of the pool
# in app_state[staking][delegations]{
#    entry with delegator_address = osmo1qye772qje88p7ggtzrvl9nxvty6dkuuskkg52lm, and validator addr = osmovaloper1qye772qje88p7ggtzrvl9nxvty6dkuusvpqhac}
DELEGATOR_INDEX=$(jq '.app_state["staking"]["delegations"] | map (.delegator_address=="osmo1qye772qje88p7ggtzrvl9nxvty6dkuuskkg52l") | index(true)' $EXPORTED_GENESIS)
cat $EXPORTED_GENESIS | jq '.app_state["staking"]["delegations"]['"$DELEGATOR_INDEX"']["shares"]="1000005000000000.000000000000000000"' > tmp_genesis.json && mv tmp_genesis.json $EXPORTED_GENESIS
# and in app_state.distribution.delegator_starting_infos
DISTRIBUTION_START_INFO_INDEX=$(jq '.app_state.distribution.delegator_starting_infos | map (.delegator_address=="osmo1qye772qje88p7ggtzrvl9nxvty6dkuuskkg52l") | index(true)' $EXPORTED_GENESIS)
cat $EXPORTED_GENESIS | jq '.app_state.distribution.delegator_starting_infos['"$DISTRIBUTION_START_INFO_INDEX"'].starting_info.stake="1000005000000000.000000000000000000"' > tmp_genesis.json && mv tmp_genesis.json $EXPORTED_GENESIS

# Then correspondingly up the total tokens bonded to the validator
#           "tokens": "5743672759222", (add 1BN to this)
sed -i '' 's%5743672759222%1005743672759222%g' $EXPORTED_GENESIS
sed -i '' 's%"power": "5743672"%"power": "1005743672"%g' $EXPORTED_GENESIS
# Update last_total_power (which is last total bonded across validators)
sed -i '' 's%65600898%1065600898%g' $EXPORTED_GENESIS

# edit operator address, old: 2125267 (2.1 osmo), new: 1000000002125267 (1BN + 2.1)
sed -i '' 's%"2125267"%"1000000002125267"%g' $EXPORTED_GENESIS

# Update total osmo supply, old 429793936956313, new 2Billion + 429793936956313
sed -i '' 's%429793936956313%2429793936956313%g' $EXPORTED_GENESIS

# Fix bonded_tokens_pool balance, old 65600951578831
sed -i '' 's%65600951578831%1065600951578831%g' $EXPORTED_GENESIS

### Fix gov params

# deposit
sed -i '' 's%"voting_period": "259200s"%"voting_period": "40s"%g' $EXPORTED_GENESIS

# epoch length
    #   "epochs": [
    #     {
    #       "current_epoch": "77",
    #       "current_epoch_ended": false,
    #       "current_epoch_start_time": "2021-12-03T17:02:07.229632445Z",
    #       "duration": "86400s",
    #       "epoch_counting_started": true,
    #       "identifier": "day",
    #       "start_time": "2021-06-18T17:00:00Z"
    #     },
# replace that duration with jq
cat $EXPORTED_GENESIS | jq '.app_state["epochs"]["epochs"][0]["duration"]="3600s"' > tmp_genesis.json && mv tmp_genesis.json $EXPORTED_GENESIS
cat $EXPORTED_GENESIS | jq '.app_state["epochs"]["epochs"][0]["current_epoch_start_time"]="2021-12-08T17:02:07.229632445Z"' > tmp_genesis.json && mv tmp_genesis.json $EXPORTED_GENESIS
