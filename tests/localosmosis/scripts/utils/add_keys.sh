#!/bin/bash

# Define a function to add a new key to the test keyring backend
function add_key() {
    local key_name=$1
    local mnemonic=$2
    osmosisd keys add "$key_name" --recover --keyring-backend test <<< "$mnemonic"
}

validator_mnemonic="bottom loan skill merry east cradle onion journey palm apology verb edit desert impose absurd oil bubble sweet glove shallow size build burst effort"

# Define an array of 10 recovery phrases
account_mnemonics=(
    "notice oak worry limit wrap speak medal online prefer cluster roof addict wrist behave treat actual wasp year salad speed social layer crew genius"
    "quality vacuum heart guard buzz spike sight swarm shove special gym robust assume sudden deposit grid alcohol choice devote leader tilt noodle tide penalty"
    "symbol force gallery make bulk round subway violin worry mixture penalty kingdom boring survey tool fringe patrol sausage hard admit remember broken alien absorb"
    "bounce success option birth apple portion aunt rural episode solution hockey pencil lend session cause hedgehog slender journey system canvas decorate razor catch empty"
    "second render cat sing soup reward cluster island bench diet lumber grocery repeat balcony perfect diesel stumble piano distance caught occur example ozone loyal"
    "spatial forest elevator battle also spoon fun skirt flight initial nasty transfer glory palm drama gossip remove fan joke shove label dune debate quick"
    "noble width taxi input there patrol clown public spell aunt wish punch moment will misery eight excess arena pen turtle minimum grain vague inmate"
    "cream sport mango believe inhale text fish rely elegant below earth april wall rug ritual blossom cherry detail length blind digital proof identify ride"
    "index light average senior silent limit usual local involve delay update rack cause inmate wall render magnet common feature laundry exact casual resource hundred"
)

# Add validator key
add_key val $validator_mnemonic 

# Add test accounts
for i in {1..9}; do
    key_name="lo-test$i"
    mnemonic="${account_mnemonics[$i-1]}"
    add_key "$key_name" "$mnemonic"
done
