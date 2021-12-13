#Pool creation and upgrade for testing v4 -> v5
#TODO should be replaced with a parameterized script to just upgrade to a specific new branch / upgrade handler

yes | osmosisd tx gov submit-proposal software-upgrade v5 --upgrade-height=25 --from=validator --keyring-backend=test --title="v5 upgrade" --description="v5 upgrade" --fees=2000uosmo
yes | osmosisd tx gov deposit 1 10000000uosmo --from=faucet --keyring-backend=test --fees=2000uosmo
sleep 7
yes | osmosisd tx gov vote 1 yes --from=validator --keyring-backend=test --fees=2000uosmo
sleep 7
yes | osmosisd tx gamm create-pool --pool-file="pool1.json" --from=faucet --keyring-backend=test --gas=400000 --fees=4000uosmo

cd osmosis
git checkout v5.x
make install
