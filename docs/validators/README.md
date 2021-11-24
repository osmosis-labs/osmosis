# Validating On Mainnet

## Synced Node

Before creating a validator, ensure you have first followed the instructions on how to [join the mainnet](../developing/network/join-mainnet)

## Initialize Wallet Keyring

If you decide you want to turn your node into a validator, you will first need to add a wallet to your keyring.

While you can add an existing wallet through your seed phrase, we will create a new wallet in this example (replace KEY_NAME with a name of your choosing):

```bash
osmosisd keys add KEY_NAME
```
Ensure you write down the mnemonic as you can not recover the wallet without it. To ensure your wallet was saved to your keyring, the WALLET_NAME is in your keys list:

```bash
osmosisd keys list
```

## Validator Public Key

The last thing needed before initializing the validator is to obtain your validator public key which was created when you first initialized your node. To obtain your validator pubkey:

```bash
osmosisd tendermint show-validator
```

## Create Validator Command

Here is the empty command:

```bash
osmosisd tx staking create-validator \
--from=[KEY_NAME] \
--amount=[staking_amount_uosmo] \
--pubkey=[osmovalconspub...]  \
--moniker="[moniker_id_of_your_node]" \
--security-contact="[security contact email/contact method]" \
--chain-id="[chain-id]" \
--commission-rate="[commission_rate]" \
--commission-max-rate="[maximum_commission_rate]" \
--commission-max-change-rate="[maximum_rate_of_change_of_commission]" \
--min-self-delegation="[min_self_delegation_amount]" \
--gas="auto" \
--gas-prices="[gas_price]" \
```

Here is the same command but with example values:

```bash
osmosisd tx staking create-validator \
--from=wallet1 \
--amount=500000000uosmo \
--pubkey=osmovalconspub1zcjduepqrevtrgcntyz04w9yzwvpy2ddf2h5pyu2tczgf9dssmywty0tzqzs0gwu0r  \
--moniker="Wosmongton" \
--security-contact="wosmongton@osmosis.labs" \
--chain-id="osmosis-1" \
--commission-rate="0.1" \
--commission-max-rate="0.2" \
--commission-max-change-rate="0.05" \
--min-self-delegation="500000000" \
--gas="auto" \
--gas-prices="0.0025uosmo" \
```

If you need further explanation for each of these command flags:
- the from flag is the KEY_NAME you created when initializing the key on your keyring
- the amount flag is the amount you will place in your own validator in uosmo (in the example, 500000000uosmo is 500osmo)
- the pubkey is the validator public key found earlier
- the moniker is a human readable name you choose for your validator 
- the security contact is an email your delegates are able to contact you at
- the chain-id is whatever chain-id you are working with (in the osmosis mainnet case it is osmosis-1)
- the commission rate is the rate you will charge your delegates (in the example above, 10 percent)
- the commission max rate is the most you are allowed to charge your delegates (in the example above, 20 percent)
- the max change rate is how much you can increase your commission rate in a 24 hour period (in the example above, 5 percent per day until reaching the max rate)
- the min self delegation is the lowest amount of personal funds the validator is required to have in their own validator to stay bonded (in the example above, 500osmo)
- the gas price is the amount of gas used to send this create-validator transaction


