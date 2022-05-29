# LocalOsmosis

You can now quickly test your changes to Osmosis with just a few commands:

1. Make any change to the osmosis code that you want to test

2. From the Osmosis home folder, run `make localnet-build`
    - This compiles all your changes to docker image called local:osmosis (~60 seconds)

3. Once complete, run `make localnet-start`
    - You will now be running a local network with your changes!

4. To add your validator wallet and 9 other preloaded wallets automatically, run `make localnet-keys`
    - These keys are added to your --keyring-backend test
    - If the keys are already on your keyring, you will get an "Error: aborted"
    - Ensure you use the name of the account as listed in the table below, as well as ensure you append the `--keyring-backend test` to your txs
        - Example: `osmosisd tx bank send lo-test2 osmo1cyyzpxplxdzkeea7kwsydadg87357qnahakaks --keyring-backend test --chain-id localosmosis`

5. To remove all block history and start from scratch, run `make localnet-remove`


# LocalOsmosis with Mainnet State

Running LocalOsmosis with mainnet state is resource intensive and can take a bit of time. It is recommmended to only use this method if you are testing a new feature that must be throughly tested before pushing to production.

1. Set up a node on mainnet (easiest to use the https://get.osmosis.zone tool)
```
curl -sL https://get.osmosis.zone/install > i.py && python3 i.py
```

2. Ensure your node is hitting blocks. Once it does, stop your Osmosis daemon
```
systemctl stop osmosisd.service
```

3. Take a state export snapshot with the following command:
```
cd $HOME
osmosisd export 2> testnet_genesis.json
```
After a while, this will create a file called `testnet_genesis.json` which is a snapshot of the current mainnet state.


4. Move `testnet_genesis.json` to the localosmosis folder within the osmosis repo
```
mv $HOME/testnet_genesis.json $HOME/osmosis/tests/localosmosis
```

5. Ensure you have docker and docker compose installed/running:
Docker
```
sudo apt-get remove docker docker-engine docker.io
sudo apt-get update
sudo apt install docker.io -y
```

Docker Compose
```
sudo apt install docker-compose -y
```

6. Compile to local:osmosis/stateExport docker image
```
make localnet-build-state-export
```

7. Start the docker image
```
localnet-start-state-export
```


## Accounts

LocalOsmosis is pre-configured with one validator and 9 accounts with ION and OSMO balances.


| Account   | Address                                                                                                  | Mnemonic                                                                                                                                                                   |
| --------- | -------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| lo-val | `osmo1phaxpevm5wecex2jyaqty2a4v02qj7qmlmzk5a`<br/>`osmovaloper1phaxpevm5wecex2jyaqty2a4v02qj7qm9v24r6` | `satisfy adjust timber high purchase tuition stool faith fine install that you unaware feed domain license impose boss human eager hat rent enjoy dawn`                    |
| lo-test1     | `osmo1cyyzpxplxdzkeea7kwsydadg87357qnahakaks`                                                           | `notice oak worry limit wrap speak medal online prefer cluster roof addict wrist behave treat actual wasp year salad speed social layer crew genius`                       |
| lo-test2     | `osmo18s5lynnmx37hq4wlrw9gdn68sg2uxp5rgk26vv`                                                           | `quality vacuum heart guard buzz spike sight swarm shove special gym robust assume sudden deposit grid alcohol choice devote leader tilt noodle tide penalty`              |
| lo-test3     | `osmo1qwexv7c6sm95lwhzn9027vyu2ccneaqad4w8ka`                                                           | `symbol force gallery make bulk round subway violin worry mixture penalty kingdom boring survey tool fringe patrol sausage hard admit remember broken alien absorb`        |
| lo-test4     | `osmo14hcxlnwlqtq75ttaxf674vk6mafspg8xwgnn53`                                                           | `bounce success option birth apple portion aunt rural episode solution hockey pencil lend session cause hedgehog slender journey system canvas decorate razor catch empty` |
| lo-test5     | `osmo12rr534cer5c0vj53eq4y32lcwguyy7nndt0u2t`                                                           | `second render cat sing soup reward cluster island bench diet lumber grocery repeat balcony perfect diesel stumble piano distance caught occur example ozone loyal`        |
| lo-test6     | `osmo1nt33cjd5auzh36syym6azgc8tve0jlvklnq7jq`                                                           | `spatial forest elevator battle also spoon fun skirt flight initial nasty transfer glory palm drama gossip remove fan joke shove label dune debate quick`                  |
| lo-test7     | `osmo10qfrpash5g2vk3hppvu45x0g860czur8ff5yx0`                                                           | `noble width taxi input there patrol clown public spell aunt wish punch moment will misery eight excess arena pen turtle minimum grain vague inmate`                       |
| lo-test8     | `osmo1f4tvsdukfwh6s9swrc24gkuz23tp8pd3e9r5fa`                                                           | `cream sport mango believe inhale text fish rely elegant below earth april wall rug ritual blossom cherry detail length blind digital proof identify ride`                 |
| lo-test9     | `osmo1myv43sqgnj5sm4zl98ftl45af9cfzk7nhjxjqh`                                                           | `index light average senior silent limit usual local involve delay update rack cause inmate wall render magnet common feature laundry exact casual resource hundred`       |
| lo-test10    | `osmo14gs9zqh8m49yy9kscjqu9h72exyf295afg6kgk`                                                           | `prefer forget visit mistake mixture feel eyebrow autumn shop pair address airport diesel street pass vague innocent poem method awful require hurry unhappy shoulder`     |
