#### Modify a mainnet export based genesis file to have a single local >66% validator, and a faucet

###FIXME incomplete / WIP, needs to be finished based on original testnetify script which uses sed

#Parameters
export VALIDATOR_NAME="Validating Chaos" #Moniker of validtor to be replaced
export VALIDATOR_ADDRESS=""

#Constants
export GENESIS=".osmosisd/config/genesis.json"
export TMP_GENESIS=".osmosisd/config/genesis.json.tmp"
export PRIV_VAL=".osmosisd/config/priv_validator_key.json"



#collect indices and keys for existing validator to replace
VALIDATOR_INDEX=$(jq --arg VALIDATOR_NAME "$VALIDATOR_NAME" '.validators | map (.name==$VALIDATOR_NAME) | index(true)' $GENESIS)
echo $VALIDATOR_INDEX
NEW_VALIDATOR_POWER=$(jq --arg VALIDATOR_INDEX "$VALIDATOR_INDEX" '.validators[$VALIDATOR_INDEX | tonumber].power | tonumber + 2000000000' $GENESIS)
#read as int, add 2billion to it
echo $NEW_VALIDATOR_POWER

jq --arg VALIDATOR_INDEX "$VALIDATOR_INDEX" '.validators[$VALIDATOR_INDEX | tonumber]' $GENESIS

STAKING_VALIDATOR_INDEX=$(jq --arg VALIDATOR_NAME "$VALIDATOR_NAME" '.app_state.staking.validators | map (.description.moniker==$VALIDATOR_NAME) | index(true)' $GENESIS)
echo $STAKING_VALIDATOR_INDEX
jq --arg STAKING_VALIDATOR_INDEX "$STAKING_VALIDATOR_INDEX" '.app_state.staking.validators[$STAKING_VALIDATOR_INDEX | tonumber]' $GENESIS

VALOPER_ADDRESS=$(jq -r --arg STAKING_VALIDATOR_INDEX "$STAKING_VALIDATOR_INDEX" '.app_state.staking.validators[$STAKING_VALIDATOR_INDEX | tonumber].operator_address' $GENESIS)
echo $VALOPER_ADDRESS

VALIDATOR_ADDRESS=$(osmosisd debug bech32-convert $VALOPER_ADDRESS --prefix="osmo" 2>&1 >/dev/null)
echo $VALIDATOR_ADDRESS

DELEGATOR_INDEX=$(jq --arg VALIDATOR_ADDRESS "$VALIDATOR_ADDRESS" '.app_state.staking.delegations | map (.delegator_address==$VALIDATOR_ADDRESS) | index(true)' $GENESIS)
echo $DELEGATOR_INDEX
jq --arg DELEGATOR_INDEX "$DELEGATOR_INDEX" '.app_state.staking.delegations[$DELEGATOR_INDEX | tonumber]' $GENESIS

STARTING_INFO_INDEX=$(jq --arg VALIDATOR_ADDRESS "$VALIDATOR_ADDRESS" '.app_state.distribution.delegator_starting_infos | map (.delegator_address==$VALIDATOR_ADDRESS) | index(true)' $GENESIS)
echo $STARTING_INFO_INDEX
jq --arg STARTING_INFO_INDEX "$STARTING_INFO_INDEX" '.app_state.distribution.delegator_starting_infos[$STARTING_INFO_INDEX | tonumber]' $GENESIS

#collect newly generated validator addresses/keys
NEW_VALIDATOR_ADDRESS=$(jq '.address' $PRIV_VAL)
echo $NEW_VALIDATOR_ADDRESS
NEW_VALIDATOR_PUBKEY=$(jq '.pub_key.value' $PRIV_VAL)
echo $NEW_VALIDATOR_PUBKEY

NEW_VALIDATOR_PUBOPER=$(osmosisd debug addr $NEW_VALIDATOR_ADDRESS 2>&1 >/dev/null | tail -c 51)
echo $NEW_VALIDATOR_PUBOPER
NEW_VALIDATOR_CONS=$(osmosisd debug bech32-convert $NEW_VALIDATOR_PUBOPER -p osmovalcons)
echo $NEW_VALIDATOR_CONS


#modify validator values
jq\
    --arg VALIDATOR_INDEX "$VALIDATOR_INDEX"\
    --arg NEW_VALIDATOR_ADDRESS "$NEW_VALIDATOR_ADDRESS"\
    --arg VALIDATOR_NAME "$VALIDATOR_NAME"\
    --arg NEW_VALIDATOR_POWER "$NEW_VALIDATOR_POWER"\
    --arg NEW_VALIDATOR_PUBKEY "$NEW_VALIDATOR_CONS"\
    '.validators[$VALIDATOR_INDEX | tonumber] = {"address": NEW_VALIDATOR_ADDRESS, "name": $VALIDATOR_NAME,"power": $NEW_VALIDATOR_POWER,"pub_key": {"type": "tendermint/PubKeyEd25519","value": $NEW_VALIDATOR_PUBKEY}}'\
    $GENESIS > $TMP_GENESIS

jq --arg VALIDATOR_INDEX "$VALIDATOR_INDEX" '.validators[$VALIDATOR_INDEX | tonumber' $TMP_GENESIS

#TODO should separate out standard validator replacement stuff, from state specific changes / setup

#modify bank values

#modify params (voting period, epochs, claim decay)



## Copy of testnetify script for quick reverence, remove before actual finalization for usage

# export EXPORTED_GENESIS=testnet_genesis.json

# # Replace Sentinel addrs/pubkeys

# # replace validator cons pubkey
# sed -i '' 's%b77zCh/VsRgVvfGXuW4dB+Dhg4PrMWWBC5G2K/qFgiU=%2OpBuqaXvXQ+lSxAoT1S7Jfyr56KiakTzvuFiuJK+X4=%g' $EXPORTED_GENESIS
# # This is a PITA to get:
# # take pubkey, do osmosisd debug pubkey {base64}
# # take address out of it, feed into osmosisd debug addr {addr}
# # pass that into osmosisd debug bech32-convert -p osmovalcons
# sed -i '' 's%osmovalcons1z6skn9g6s7py0klztr7acutr3anqd52k9x5p70%osmovalcons13duwg7rhwsnucwgxhq35edete63v0r5rqp90es%g' $EXPORTED_GENESIS
# # replace validator hex addr
# sed -i '' 's%16A169951A878247DBE258FDDC71638F6606D156%8B78E478777427CC3906B8234CB72BCEA2C78E83%g' $EXPORTED_GENESIS
# # replace operator address
# sed -i '' 's%osmovaloper1cyw4vw20el8e7ez8080md0r8psg25n0cq98a9n%osmovaloper1qye772qje88p7ggtzrvl9nxvty6dkuusvpqhac%g' $EXPORTED_GENESIS
# # replace the actual account
# sed -i '' 's%osmo1cyw4vw20el8e7ez8080md0r8psg25n0c6j07j5%osmo1qye772qje88p7ggtzrvl9nxvty6dkuuskkg52l%g' $EXPORTED_GENESIS
# # replace that accounts pubkey, obtained via auth. New pubkey obtained via new debug command
# sed -i '' 's%AqlNb1FM8veQrT4/apR5B3hww8VApc0LTtZnXhq7FqG0%A9zC0Sa0VCK/lVLi1Kv0C1c0MQp47d+yjFqb6dAUza0a%g' $EXPORTED_GENESIS

# # now time to replace the amounts
# # manually increase share amounts for 
# #          "delegator_address": "osmo1cyw4vw20el8e7ez8080md0r8psg25n0c6j07j5",
# # before: 5000000000.000000000000000000 (5000 osmo)
# # after: 100005000000000.000000000000000000 (100M + 5k osmo = 100,005,000)
# # there are two such locations

# # Then correspondingly up the total tokens bonded to the validator
# #           "tokens": "5979280136171", (add 100M to this)
# sed -i '' 's%5979280136171%105979280136171%g' $EXPORTED_GENESIS
# sed -i '' 's%"power": "5979280"%"power": "105979280"%g' $EXPORTED_GENESIS
# # Update last_total_power (which is last total bonded across validators)
# sed -i '' 's%57368851%157368851%g' $EXPORTED_GENESIS

# # edit operator address, old: 2125267 (2.1 osmo), new: 100000002125267 (100M + 2.1)
# sed -i '' 's%2125267%100000002125267%g' $EXPORTED_GENESIS

# # Update total osmo supply, old 413150362339859, new 613M
# sed -i '' 's%413150362339859%613150362339859%g' $EXPORTED_GENESIS

# # Fix bonded_tokens_pool balance, old 57368900009013
# sed -i '' 's%57368900009013%157368900009013%g' $EXPORTED_GENESIS

# ### Fix gov params

# # deposit
# sed -i '' 's%"voting_period": "259200s"%"voting_period": "120s"%g' $EXPORTED_GENESIS

# # epoch length
#     #   "epochs": [
#     #     {
#     #       "current_epoch": "77",
#     #       "current_epoch_ended": false,
#     #       "current_epoch_start_time": "2021-09-03T17:12:52.752325457Z",
#     #       "duration": "86400s",
#     #       "epoch_counting_started": true,
#     #       "identifier": "day",
#     #       "start_time": "2021-06-18T17:00:00Z"
#     #     },
# # replace that duration with jq
# cat $EXPORTED_GENESIS | jq '.app_state["epochs"]["epochs"][0]["duration"]="1800s"' > tmp_genesis.json && mv tmp_genesis.json $EXPORTED_GENESIS