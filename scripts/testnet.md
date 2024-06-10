# Testnet script

## Instructions
0) Run with `./testnet.sh <NODE1_API> <NODE2_API> <NODE3_API> <NODE4_API>`


1) Start separately all 4 validators on different nodes
`symphonyd start --home=$HOME/.symphonyd`


2) Send `note` to other validators from 1-st validator (node1) 

Note: to get validator address can use `symphonyd keys show` command

Example: `$(symphonyd keys show validator2 -a --keyring-backend=test --home=$HOME/.symphonyd/validator2)`


- send `note` from 1-st validator to 2-nd validator

`symphonyd tx bank send validator1 <VALIDATOR2_ADDRESS> 500000000note --keyring-backend=test --home=$HOME/.symphonyd --chain-id=symphony-testnet-1 --broadcast-mode sync --yes --fees 1000000note
`
- send `note` from 1-st validator to 3-rd validator

`symphonyd tx bank send validator1 <VALIDATOR3_ADDRESS> 500000000note --keyring-backend=test --home=$HOME/.symphonyd --chain-id=symphony-testnet-1 --broadcast-mode sync --yes --fees 1000000note
`
- send `note` from 1-st validator to 4-rd validator

`symphonyd tx bank send validator1 <VALIDATOR4_ADDRESS> 500000000note --keyring-backend=test --home=$HOME/.symphonyd --chain-id=symphony-testnet-1 --broadcast-mode sync --yes --fees 1000000note
`

3) Add 2, 3, 4 validators 
- add validator2

`symphonyd tx staking create-validator --amount=500000000note --from=validator2 --pubkey=$(symphonyd tendermint show-validator --home=$HOME/.symphonyd) --moniker="validator2" --chain-id=symphony-testnet-1 --commission-rate="0.1" --commission-max-rate="0.2" --commission-max-change-rate="0.05" --min-self-delegation="500000000" --keyring-backend=test --home=$HOME/.symphonyd --broadcast-mode sync  --yes --fees 1000000note
`
- add validator3

`symphonyd tx staking create-validator --amount=500000000note --from=validator3 --pubkey=$(symphonyd tendermint show-validator --home=$HOME/.symphonyd) --moniker="validator3" --chain-id=symphony-testnet-1 --commission-rate="0.1" --commission-max-rate="0.2" --commission-max-change-rate="0.05" --min-self-delegation="500000000" --keyring-backend=test --home=$HOME/.symphonyd --broadcast-mode sync  --yes --fees 1000000note
`
- add validator4

`symphonyd tx staking create-validator --amount=500000000note --from=validator4 --pubkey=$(symphonyd tendermint show-validator --home=$HOME/.symphonyd) --moniker="validator4" --chain-id=symphony-testnet-1 --commission-rate="0.1" --commission-max-rate="0.2" --commission-max-change-rate="0.05" --min-self-delegation="500000000" --keyring-backend=test --home=$HOME/.symphonyd --broadcast-mode sync  --yes --fees 1000000note`