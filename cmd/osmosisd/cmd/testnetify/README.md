# Testnetify.py

This script takes a state export JSON file (under the name
testnet\_genesis.json) and edits multiple values to give the operator
address 1M spendable OSMO, makes your validator file possess over 1M
OSMO, makes the epoch duration 21600s, and sets the voting period to
180s.

## Instructions

Start at home directory

```sh
cd $HOME
```

Create state export from a running node (ensure to name it
testnet\_genesis.json for script to work)

```sh
osmosisd export 2> testnet_genesis.json
```

Make a copy of the genesis file, just in case you mess up

```sh
cp $HOME/testnet_genesis.json $HOME/testnet_genesis_bk.json
```

NOTE: There are three values in the python script you can change.

1. The operator address (op\_address) and its corresponding public key
    (op\_pubkey). I provided the mnemonic for the provided address
    (osmo12smx2wdlyttvyzvzg54y2vnqwq2qjateuf7thj), but you may change
    the address and corresponding pubkey if desired
2. The chain-id (new\_chain\_id)
3. The epoch duration (new\_duration)
4. The voting period (new\_voting\_period)
5. The distribution module account currently must be subtracted by 3
    due to a bug. If the account gets offset further, you must change
    dist\_offset\_amt

Mnemonic for provided address: **bottom loan skill merry east cradle
onion journey palm apology verb edit desert impose absurd oil bubble
sweet glove shallow size build burst effort**

From the same directory as the testnet\_genesis.json, run testnetify.py

```sh
python3 testnetify.py
```

After it is complete, overwrite the current genesis file with new
testnet genesis

```sh
cp testnet_genesis.json .osmosisd/config/genesis.json
```

Unsafe reset all

```sh
osmosisd unsafe-reset-all
```

Start osmosis daemon (with extra flag first time around)

```sh
osmosisd start --x-crisis-skip-assert-invariants
```

After initializing, you will get a "couldn't connect to any seeds"
error. Leave the node as it is.

Set up a second node with the same genesis.json file that was created
using the first node.

On the first node, retrieve its public IP and also run this command to
get your node ID

```sh
osmosisd tendermint show-node-id
```

On the second node, open the config.toml

```sh
nano $HOME/.osmosisd/config/config.toml
```

Under persistent\_peers and seeds, add the first nodes information like
so: node-id\@IP:26656

Example: 665ebb897edc41d691c70b15916086a9c7761dc4\@199.43.113.117:26656

On the second node, ensure the genesis file is replaced with the genesis
created on the first node and start the osmosis daemon

```sh
osmosisd start --x-crisis-skip-assert-invariants
```

Once the second peer initializes, the chain will no longer be halted
(this is necessary due to a tendermint bug). The second peer can then be
shut off if desired. If the first peer ever shuts down, the second peer
must be started in order to kickstart the chain again.

As a last note, sometimes getting testnet nodes spun up for the first
time can be finicky. If you are stuck getting the second node to connect
to the first node, sometimes doing a `unsafe-reset-all` on both nodes
fixes the issue. Also, try adding the second node as a persistent peer
to the first node. These two methods have fixed nodes that do not want
to cooperate. Lastly, if your node keeps killing the daemon, please
ensure you have a swap file set up.

Enjoy your state exported testnet!
