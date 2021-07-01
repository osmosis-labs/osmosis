# run old binary on terminal1
git checkout 6574912d71c41d591859239964162a2d3ee3a57e
go install ./cmd/osmosisd/
Modify startnode.sh script to below`
    #!/bin/bash

    rm -rf $HOME/.osmosisd/

    cd $HOME

    osmosisd init --chain-id=testing testing --home=$HOME/.osmosisd
    osmosisd keys add validator --keyring-backend=test --home=$HOME/.osmosisd
    osmosisd add-genesis-account $(osmosisd keys show validator -a --keyring-backend=test --home=$HOME/.osmosisd) 1000000000stake,1000000000valtoken --home=$HOME/.osmosisd
    osmosisd gentx validator 500000000stake --keyring-backend=test --home=$HOME/.osmosisd --chain-id=testing
    osmosisd collect-gentxs --home=$HOME/.osmosisd

    cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["gov"]["voting_params"]["voting_period"]="10s"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json

    osmosisd start --home=$HOME/.osmosisd
`
sh startnode.sh

# operations on terminal2
osmosisd tx lockup lock-tokens 100stake --duration="5s" --from=validator --chain-id=testing --keyring-backend=test --yes
osmosisd tx gov submit-proposal software-upgrade upgrade-lockup-module-store-management --title="lockup module upgrade" --description="lockup module upgrade for gas efficiency"  --from=validator --upgrade-height=10 --deposit=10000000stake --chain-id=testing --keyring-backend=test -y
osmosisd tx gov vote 1 yes --from=validator --keyring-backend=test --chain-id=testing --yes
osmosisd query gov proposal 1
osmosisd query upgrade plan

# on terminal1
Wait until consensus failure happen and stop binary using Ctrl + C
git checkout lockup_module_genesis_export
go install ./cmd/osmosisd/
osmosisd start --home=$HOME/.osmosisd

# check on terminal2
osmosisd query lockup account-locked-longer-duration $(osmosisd keys show -a --keyring-backend=test validator) 1s