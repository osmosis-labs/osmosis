# LocalOsmosis

LocalOsmosis is a complete Osmosis testnet containerized with Docker and orchestrated with a simple docker-compose file. LocalOsmosis comes preconfigured with opinionated, sensible defaults for a standard testing environment.

LocalOsmosis comes in two flavors:

1. No initial state: brand new testnet with no initial state. 
2. With mainnet state: creates a testnet from a mainnet state export

## Prerequisites

Ensure you have docker and docker-compose installed:

```sh
# Docker
sudo apt-get remove docker docker-engine docker.io
sudo apt-get update
sudo apt install docker.io -y

# Docker compose
sudo apt install docker-compose -y
```

## 1. LocalOsmosis - No Initial State

The following commands must be executed from the root folder of the Osmosis repository.

1. Make any change to the osmosis code that you want to test

2. Initialize LocalOsmosis:

```bash
make localnet-init
```

The command:

- Builds a local docker image with the latest changes
- Cleans the `$HOME/.osmosisd-local` folder

3. Start LocalOsmosis:

```bash
make localnet-start
```

> Note
>
> You can also start LocalOsmosis in detach mode with:
>
> `make localnet-startd`

### Accounts

Localosmosis will spin up a single validator localnet with the following accounts pre-configured

| Account                | Address                                                                                                | Mnemonic                                                                                                                                                         |
|------------------------|--------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| localosmosis-validator | `osmo12smx2wdlyttvyzvzg54y2vnqwq2qjateuf7thj`<br/>`osmovaloper1phaxpevm5wecex2jyaqty2a4v02qj7qm9v24r6` | `bottom loan skill merry east cradle onion journey palm apology verb edit desert impose absurd oil bubble sweet glove shallow size build burst effort`           |
| localosmosis-pools     | `osmo1jllfytsz4dryxhz5tl7u73v29exsf80vz52ucc`                                                          | `traffic cool olive pottery elegant innocent aisle dial genuine install shy uncle ride federal soon shift flight program cave famous provide cute pole struggle` |
| localosmosis-faucet    | `osmo14hm982r3xzkmhjfzjh74xs7uqfkzpu5axvja3w`                                                          | `only item always south dry begin barely seed wire praise chapter bomb remind abandon erase safe point vehicle tuition release half denial receive water`        |

### Faucet

There is faucet available at `http://localhost:8000/` that can be used to get tokens for testing.

> Note: currently it sends only `OSMO` tokens

### Pools

By default, `localosmosis` will create any pools that it's located in the `pools` folder. The pools are created with the `localosmosis-pools` account.

Currently, there are the following pools are created:

```json
{
    "weights": "5uosmo,5stake",
    "initial-deposit": "100000000000uosmo,100000000000stake",
    "swap-fee": "0.01",
    "exit-fee": "0.01",
    "future-governor": ""
}

{
    "weights": "5uosmo,5uion",
    "initial-deposit": "100000000000uosmo,100000000000uion",
    "swap-fee": "0.01",
    "exit-fee": "0.01",
    "future-governor": ""
}

{
    "weights": "5stake,5uion,5uosmo",
    "initial-deposit": "100000000000stake,10000000000uion,200000000000uosmo",
    "swap-fee": "0.01",
    "exit-fee": "0.01",
    "future-governor": ""
}

{
	"initial-deposit": "1000000stake,1000uosmo",
	"swap-fee": "0.005",
	"exit-fee": "0",
	"future-governor": "",
	"scaling-factors": "1000,1"
}
```

### Teardown

1. You can stop chain, keeping the state with

```bash
make localnet-stop
```

2. When you are done you can clean up the environment with:

```bash
make localnet-clean
```

## 2. LocalOsmosis - With Mainnet State

LocalOsmosis can also be started with a mainnet state export. This is useful for testing in a more realistic environment.
The setup takes more time than the no initial state setup, as it requires downloading the state export from a remote server.

We also recommend using a machine with at least 64GB of RAM or sufficient swap.

1. Build the `local:osmosis` docker image:

```bash
make localnet-state-export-init
```

The command:

- Builds a local docker image with the latest changes
- Cleans the `$HOME/.osmosisd-local` folder

2. Start LocalOsmosis:

```bash
make localnet-state-export-start
```

> Note
>
> You can also start LocalOsmosis in detach mode with:
>
> `make localnet-state-export-startd`

When running this command for the first time, `local:osmosis` will:

- Download the state export from a remote server
- Modify the `state_export.json` to create a new state suitable for a testnet
- Start the localnet

You will then go through the genesis initialization process. This will take ~15 minutes.
You will then hit the first block (not block 1, but the block number after your snapshot was taken), and then you will just see a bunch of p2p error logs with some KV store logs.

**This will happen for about ~20 minutes**, and then you will finally hit blocks at a normal pace.

### Accounts

| Account                | Address                                                                                                | Mnemonic                                                                                                                                                  |
|------------------------|--------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|
| localosmosis-validator | `osmo12smx2wdlyttvyzvzg54y2vnqwq2qjateuf7thj`<br/>`osmovaloper1phaxpevm5wecex2jyaqty2a4v02qj7qm9v24r6` | `bottom loan skill merry east cradle onion journey palm apology verb edit desert impose absurd oil bubble sweet glove shallow size build burst effort`    |

In this setup, there is only one validator account that serves also as the faucet account.

### Faucet

There is faucet available at `http://localhost:8000/` that can be used to get tokens for testing.

> Note: currently it sends only `OSMO` tokens

### Teardown 

1. You can stop chain, keeping the state with

```bash
make localnet-state-export-stop
```

2. When you are done you can clean up the environment with:

```bash
make localnet-state-export-clean
```

Note: At some point, all the validators (except yours) will get jailed at the same block due to them being offline.

When this happens, it may take a little bit of time to process. Once all validators are jailed, you will continue to hit blocks as you did before.
If you are only running the validator for a short time (< 24 hours) you will not experience this.

## Interacting with LocalOsmosis

You can run `osmosisd` commands against the local chain with the following command:

```bash
osmosisd --home ~/.osmosisd-local --chain-id localosmosis q bank total
```

or by querying the node directly:

```bash
osmosisd localosmosis q bank total --node http://localhost:26657
```

If you need to sign transactions you can use the following keys available in the `test` keyring-backend located in `~/.osmosisd-local`:

```bash
osmosisd keys list --keyring-backend test --home ~/.osmosisd-local 
```

```json
- name: localosmosis-faucet
  type: local
  address: osmo14hm982r3xzkmhjfzjh74xs7uqfkzpu5axvja3w
  pubkey: '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AkLNaw9Xz5J+O+FbOLGXO8Pz5S19+bqRve1hAVI2cJ+F"}'
  mnemonic: ""
- name: localosmosis-pools
  type: local
  address: osmo1jllfytsz4dryxhz5tl7u73v29exsf80vz52ucc
  pubkey: '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A6xsM4oR8iJRVSZKXr3Xa36vpCDUjhbNXiWy6Q1xJAHk"}'
  mnemonic: ""
- name: localosmosis-validator
  type: local
  address: osmo12smx2wdlyttvyzvzg54y2vnqwq2qjateuf7thj
  pubkey: '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A2MR6q+pOpLtdxh0tHHe2JrEY2KOcvRogtLxHDHzJvOh"}'
  mnemonic: ""
```

Examples:

```bash
# Create a pool from `localosmosis-pools` accounts
osmosisd tx gamm create-pool --pool-file pools/nativeDenomPoolB.json \
    --chain-id=localosmosis \
    --home ~/.osmosisd-local \
    --keyring-backend=test\
    -b block \
    --from localosmosis-pools \
    --fees 10000uosmo


# Send 100000uosmo from validator to faucet address
osmosisd tx bank send localosmosis-validator osmo14hm982r3xzkmhjfzjh74xs7uqfkzpu5axvja3w 100000uosmo \
    --chain-id=localosmosis \
    --home ~/.osmosisd-local \
    --keyring-backend=test\
    -b block \
    --from localosmosis-validator \
    --fees 10000uosmo
```

### Software-upgrade test

To test a software upgrade, you can use the `submit_upgrade_proposal.sh` script located in the `utils/` folder. This script automatically creates a proposal to upgrade the software to the specified version and votes "yes" on the proposal. Once the proposal passes and the upgrade height is reached, you can update your `localosmosis` instance to use the new version.

To use the script:

1. make sure you have a running LocalOsmosis instance

2. run the following command:

```bash
./utils/submit_upgrade_proposal.sh <upgrade version>
```

Replace `<upgrade version>` with the version of the software you want to upgrade to, for example. If no version is specified, the script will default to `v15` version.

The script does the following:

- Creates an upgrade proposal with the specified version and description.
- Votes "yes" on the proposal.

Once the upgrade height is reached, you need to update your `localosmosis` instance to use the new software. 

There are two ways to do this:

1. Change the image in the `docker-compose.yml` file to use the new version, and then restart LocalOsmosis using `make localnet-start`. For example:

```yaml
services:
  localosmosis:
    image: <NEW_IMAGE_I_WANT_TO_USE>
    # All this needs to be commented to don't build the image with local changes
    # 
    # build:
    #     context: ../../
    #     dockerfile: Dockerfile
    #     args:
    #     RUNNER_IMAGE: alpine:3.16
    #     GO_VERSION: 1.19
```

2. Checkout the Osmosis repository to a different `ref` that includes the new version, and then rebuild and restart LocalOsmosis using `make localnet-start`. Make sure to don't delete your `~/.osmosisd-local` folder.
