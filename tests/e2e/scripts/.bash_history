osmosisd keys list --keyring-backend test
osmosisd keys add wallet
exit
ls
osmosisd keys list --keyring-backend test
osmosisd q bank balances osmo16ssxe0l899487ye2rjss9yf0y7lgz9djucyadn
osmosisd tx bank send val osmo133uej2dmeukrsx9du6kvgmltcg2xg7dsyssf7q 100000000000000000000gamm/pool/1 --chain-id osmo-test-a --from val -b block --keyring-backend test
exit
