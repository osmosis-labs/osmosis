# run sed

export EXPORTED_GENESIS=testnet_genesis.json

# Replace Sentinel addrs/pubkeys

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
sed -i '' 's%osmovaloper1cyw4vw20el8e7ez8080md0r8psg25n0cq98a9n%osmovaloper1qye772qje88p7ggtzrvl9nxvty6dkuusvpqhac%g' $EXPORTED_GENESIS
# replace the actual account
sed -i '' 's%osmo1cyw4vw20el8e7ez8080md0r8psg25n0c6j07j5%osmo1qye772qje88p7ggtzrvl9nxvty6dkuuskkg52l%g' $EXPORTED_GENESIS
# replace that accounts pubkey, obtained via auth. New pubkey obtained via new debug command
sed -i '' 's%AqlNb1FM8veQrT4/apR5B3hww8VApc0LTtZnXhq7FqG0%A9zC0Sa0VCK/lVLi1Kv0C1c0MQp47d+yjFqb6dAUza0a%g' $EXPORTED_GENESIS

# now time to replace the amounts
# manually increase share amounts for 
#          "delegator_address": "osmo1cyw4vw20el8e7ez8080md0r8psg25n0c6j07j5",
# before: 5000000000.000000000000000000 (5000 osmo)
# after: 100005000000000.000000000000000000 (100M + 5k osmo = 100,005,000)
# there are two such locations

# Then correspondingly up the total tokens bonded to the validator
#           "tokens": "5041917053916", (add 100M to this)
sed -i '' 's%5041917053916%105041917053916%g' $EXPORTED_GENESIS
sed -i '' 's%"power": "5041917"%"power": "105041917"%g' $EXPORTED_GENESIS
# Update last_total_power (which is last total bonded across validators)
sed -i '' 's%37961808%137961808%g' $EXPORTED_GENESIS

# edit operator address, old: 2125267 (2.1 osmo), new: 100000002125267 (100M + 2.1)
sed -i '' 's%2125267%100000002125267%g' $EXPORTED_GENESIS

# Update total osmo supply, old 374314745901464, new 571M
sed -i '' 's%374314745901464%574314745901464%g' $EXPORTED_GENESIS

# Fix bonded tokens pool balance, old 37961850148775
sed -i '' 's%37961850148775%137961850148775%g' $EXPORTED_GENESIS

### Fix gov params

# deposit
sed -i '' 's%"voting_period": "259200s"%"voting_period": "120s"%g' $EXPORTED_GENESIS

# epoch length
    #   "epochs": [
    #     {
    #       "current_epoch": "77",
    #       "current_epoch_ended": false,
    #       "current_epoch_start_time": "2021-09-03T17:12:52.752325457Z",
    #       "duration": "86400s",
    #       "epoch_counting_started": true,
    #       "identifier": "day",
    #       "start_time": "2021-06-18T17:00:00Z"
    #     },
# replace that duration with jq
cat $EXPORTED_GENESIS | jq '.app_state["epochs"]["epochs"][0]["duration"]="1800s"' > tmp_genesis.json && mv tmp_genesis.json $EXPORTED_GENESIS
