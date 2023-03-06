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

4. (optional) Add your validator wallet and 9 other preloaded wallets automatically:

```bash
make localnet-keys
```

- These keys are added to your `--keyring-backend test`
- If the keys are already on your keyring, you will get an `"Error: aborted"`
- Ensure you use the name of the account as listed in the table below, as well as ensure you append the `--keyring-backend test` to your txs
- Example: `osmosisd tx bank send lo-test2 osmo1cyyzpxplxdzkeea7kwsydadg87357qnahakaks --keyring-backend test --chain-id LocalOsmosis`

5. You can stop chain, keeping the state with

```bash
make localnet-stop
```

6. When you are done you can clean up the environment with:

```bash
make localnet-clean
```

## 2. LocalOsmosis - With Mainnet State

Running LocalOsmosis with mainnet state is resource intensive and can take a bit of time.
It is recommended to only use this method if you are testing a new feature that must be thoroughly tested before pushing to production.

A few things to note before getting started. The below method will only work if you are using the same version as mainnet. In other words,
if mainnet is on v8.0.0 and you try to do this on a v9.0.0 tag or on main, you will run into an error when initializing the genesis.
(yes, it is possible to create a state exported testnet on a upcoming release, but that is out of the scope of this tutorial)

Additionally, this process requires 64GB of RAM. If you do not have 64GB of RAM, you will get an OOM error.

### Create a mainnet state export

1. Set up a node on mainnet (easiest to use the [get.osmosis.zone](https://get.osmosis.zone) tool). This will be the node you use to run the state exported testnet, so ensure it has at least 64GB of RAM.

```sh
curl -sL https://get.osmosis.zone/install > i.py && python3 i.py
```

2. Once the installer is done, ensure your node is hitting blocks.

```sh
source ~/.profile
journalctl -u osmosisd.service -f
```

3. Stop your Osmosis daemon

```sh
systemctl stop osmosisd.service
```

4. Take a state export snapshot with the following command:

```sh
cd $HOME
osmosisd export 2> state_export.json
```

After a while (~15 minutes), this will create a file called `state_export.json` which is a snapshot of the current mainnet state.

### Use the state export in LocalOsmosis

1. Copy the `state_export.json` to the `LocalOsmosis/state_export` folder within the osmosis repo

```sh
cp $HOME/state_export.json $HOME/osmosis/tests/LocalOsmosis/state_export/
```

6. Ensure you have docker and docker-compose installed:

```sh
# Docker
sudo apt-get remove docker docker-engine docker.io
sudo apt-get update
sudo apt install docker.io -y

# Docker compose
sudo apt install docker-compose -y
```

7. Build the `local:osmosis` docker image:

```bash
make localnet-state-export-init
```

The command:

- Builds a local docker image with the latest changes
- Cleans the `$HOME/.osmosisd` folder

3. Start LocalOsmosis:

```bash
make localnet-state-export-start
```

> Note
>
> You can also start LocalOsmosis in detach mode with:
>
> `make localnet-state-export-startd`

When running this command for the first time, `local:osmosis` will:

- Modify the provided `state_export.json` to create a new state suitable for a testnet
- Start the chain

You will then go through the genesis initialization process. This will take ~15 minutes.
You will then hit the first block (not block 1, but the block number after your snapshot was taken), and then you will just see a bunch of p2p error logs with some KV store logs.
**This will happen for about 1 hour**, and then you will finally hit blocks at a normal pace.

9. On your host machine, add this specific wallet which holds a large amount of osmo funds

```sh
MNEMONIC="bottom loan skill merry east cradle onion journey palm apology verb edit desert impose absurd oil bubble sweet glove shallow size build burst effort"
echo $MNEMONIC | osmosisd keys add wallet --recover --keyring-backend test
```

You now are running a validator with a majority of the voting power with the same state as mainnet state (at the time you took the snapshot)

10. On your host machine, you can now query the state-exported testnet:

```sh
osmosisd status
```

11. Here is an example command to ensure complete understanding:

```sh
osmosisd tx bank send wallet osmo1nyphwl8p5yx6fxzevjwqunsfqpcxukmtk8t60m 10000000uosmo --chain-id testing1 --keyring-backend test
```

12. You can stop chain, keeping the state with

```bash
make localnet-state-export-stop
```

13. When you are done you can clean up the environment with:

```bash
make localnet-state-export-clean
```

Note: At some point, all the validators (except yours) will get jailed at the same block due to them being offline.

When this happens, it may take a little bit of time to process. Once all validators are jailed, you will continue to hit blocks as you did before.
If you are only running the validator for a short time (< 24 hours) you will not experience this.

## LocalOsmosis Accounts

LocalOsmosis is pre-configured with one validator and 9 accounts with ION and OSMO balances.

| Account   | Address                                                                                                | Mnemonic                                                                                                                                                                   |
|-----------|--------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| lo-val    | `osmo1phaxpevm5wecex2jyaqty2a4v02qj7qmlmzk5a`<br/>`osmovaloper1phaxpevm5wecex2jyaqty2a4v02qj7qm9v24r6` | `satisfy adjust timber high purchase tuition stool faith fine install that you unaware feed domain license impose boss human eager hat rent enjoy dawn`                    |
| lo-test1  | `osmo1cyyzpxplxdzkeea7kwsydadg87357qnahakaks`                                                          | `notice oak worry limit wrap speak medal online prefer cluster roof addict wrist behave treat actual wasp year salad speed social layer crew genius`                       |
| lo-test2  | `osmo18s5lynnmx37hq4wlrw9gdn68sg2uxp5rgk26vv`                                                          | `quality vacuum heart guard buzz spike sight swarm shove special gym robust assume sudden deposit grid alcohol choice devote leader tilt noodle tide penalty`              |
| lo-test3  | `osmo1qwexv7c6sm95lwhzn9027vyu2ccneaqad4w8ka`                                                          | `symbol force gallery make bulk round subway violin worry mixture penalty kingdom boring survey tool fringe patrol sausage hard admit remember broken alien absorb`        |
| lo-test4  | `osmo14hcxlnwlqtq75ttaxf674vk6mafspg8xwgnn53`                                                          | `bounce success option birth apple portion aunt rural episode solution hockey pencil lend session cause hedgehog slender journey system canvas decorate razor catch empty` |
| lo-test5  | `osmo12rr534cer5c0vj53eq4y32lcwguyy7nndt0u2t`                                                          | `second render cat sing soup reward cluster island bench diet lumber grocery repeat balcony perfect diesel stumble piano distance caught occur example ozone loyal`        |
| lo-test6  | `osmo1nt33cjd5auzh36syym6azgc8tve0jlvklnq7jq`                                                          | `spatial forest elevator battle also spoon fun skirt flight initial nasty transfer glory palm drama gossip remove fan joke shove label dune debate quick`                  |
| lo-test7  | `osmo10qfrpash5g2vk3hppvu45x0g860czur8ff5yx0`                                                          | `noble width taxi input there patrol clown public spell aunt wish punch moment will misery eight excess arena pen turtle minimum grain vague inmate`                       |
| lo-test8  | `osmo1f4tvsdukfwh6s9swrc24gkuz23tp8pd3e9r5fa`                                                          | `cream sport mango believe inhale text fish rely elegant below earth april wall rug ritual blossom cherry detail length blind digital proof identify ride`                 |
| lo-test9  | `osmo1myv43sqgnj5sm4zl98ftl45af9cfzk7nhjxjqh`                                                          | `index light average senior silent limit usual local involve delay update rack cause inmate wall render magnet common feature laundry exact casual resource hundred`       |
| lo-test10 | `osmo14gs9zqh8m49yy9kscjqu9h72exyf295afg6kgk`                                                          | `prefer forget visit mistake mixture feel eyebrow autumn shop pair address airport diesel street pass vague innocent poem method awful require hurry unhappy shoulder`     |

## Tests

### Software-upgrade test

To test a software upgrade, you can use the `submit_upgrade_proposal.sh` script located in the `scripts/` folder. This script automatically creates a proposal to upgrade the software to the specified version and votes "yes" on the proposal. Once the proposal passes and the upgrade height is reached, you can update your localosmosis instance to use the new version.

#### Usage 

To use the script:

1. make sure you have a running LocalOsmosis instance

2. run the following command:

```bash
./scripts/submit_upgrade_proposal.sh <upgrade version>
```

Replace `<upgrade version>` with the version of the software you want to upgrade to, for example. If no version is specified, the script will default to `v15` version.

The script does the following:

- Creates an upgrade proposal with the specified version and description.
- Votes "yes" on the proposal.

#### Upgrade

Once the upgrade height is reached, you need to update your `localosmosis` instance to use the new software. 

There are two ways to do this:

1. Change the image in the `docker-compose.yml` file to use the new version, and then restart LocalOsmosis using `make localnet-start`. For example:

```yaml
services:
  osmosisd:
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
