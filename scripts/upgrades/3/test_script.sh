# Create a genesis.json for testing. The node that you this on will be your "validator"
# It should be on version v3.0.0-rc0
symphonyd init --chain-id=testing testing --home=$HOME/.symphonyd
symphonyd keys add validator --keyring-backend=test --home=$HOME/.symphonyd
symphonyd add-genesis-account $(symphonyd keys show validator -a --keyring-backend=test --home=$HOME/.symphonyd) 1000000000note,1000000000valtoken --home=$HOME/.symphonyd
sed -i -e "s/stake/note/g" $HOME/.symphonyd/config/genesis.json
symphonyd gentx validator 500000000note --commission-rate="0.0" --keyring-backend=test --home=$HOME/.symphonyd --chain-id=testing
symphonyd collect-gentxs --home=$HOME/.symphonyd

cat $HOME/.symphonyd/config/genesis.json | jq '.initial_height="711800"' > $HOME/.symphonyd/config/tmp_genesis.json && mv $HOME/.symphonyd/config/tmp_genesis.json $HOME/.symphonyd/config/genesis.json
cat $HOME/.symphonyd/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"]["denom"]="valtoken"' > $HOME/.symphonyd/config/tmp_genesis.json && mv $HOME/.symphonyd/config/tmp_genesis.json $HOME/.symphonyd/config/genesis.json
cat $HOME/.symphonyd/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"]["amount"]="100"' > $HOME/.symphonyd/config/tmp_genesis.json && mv $HOME/.symphonyd/config/tmp_genesis.json $HOME/.symphonyd/config/genesis.json
cat $HOME/.symphonyd/config/genesis.json | jq '.app_state["gov"]["voting_params"]["voting_period"]="120s"' > $HOME/.symphonyd/config/tmp_genesis.json && mv $HOME/.symphonyd/config/tmp_genesis.json $HOME/.symphonyd/config/genesis.json
cat $HOME/.symphonyd/config/genesis.json | jq '.app_state["staking"]["params"]["min_commission_rate"]="0.050000000000000000"' > $HOME/.symphonyd/config/tmp_genesis.json && mv $HOME/.symphonyd/config/tmp_genesis.json $HOME/.symphonyd/config/genesis.json

# Now setup a second full node, and peer it with this v3.0.0-rc0 node.

# start the chain on both machines
symphonyd start
# Create proposals

symphonyd tx gov submit-proposal --title="existing passing prop" --description="passing prop"  --from=validator --deposit=1000valtoken --chain-id=testing --keyring-backend=test --broadcast-mode=block  --type="Text"
symphonyd tx gov vote 1 yes --from=validator --keyring-backend=test --chain-id=testing --yes
symphonyd tx gov submit-proposal --title="prop with enough melody deposit" --description="prop w/ enough deposit"  --from=validator --deposit=500000000note --chain-id=testing --keyring-backend=test --broadcast-mode=block  --type="Text"
# Check that we have proposal 1 passed, and proposal 2 in deposit period
symphonyd q gov proposals
# CHeck that validator commission is under min_commission_rate
symphonyd q staking validators
# Wait for upgrade block.
# Upgrade happened
# your full node should have crashed with consensus failure

# Now we test post-upgrade behavior is as intended

# Everything in deposit stayed in deposit
symphonyd q gov proposals
# Check that commissions was bumped to min_commission_rate
symphonyd q staking validators
# pushes 2 into voting period
symphonyd tx gov deposit 2 1valtoken --from=validator --keyring-backend=test --chain-id=testing --yes