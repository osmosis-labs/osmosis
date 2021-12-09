# Joining Testnet

## Install Osmosis Binary

Make sure you have [installed the Osmosis Binary (CLI).](../cli/install)

## Initialize Osmosis Node

You have now installed the Osmosis Daemon (osmosisd). Use osmosisd to initialize your node (replace the NODE_NAME with a name of your choosing):

```bash
osmosisd init NODE_NAME --chain-id=osmosis-testnet-0
```


We now need to open the config.toml to edit the seed list:

```bash
cd $HOME/.osmosisd/config
nano config.toml
```

Use page down or arrow keys to get to the line that says seeds = "" and replace it with the following:

```bash
seeds = "4eaed17781cd948149098d55f80a28232a365236@testmosis.blockpane.com:26656"
```

Then pres ```Ctrl+O```, then enter to save, then ```Ctrl+X``` to exit

## Set Up Cosmovisor

We will now set up cosmovisor to ensure the upgrade happens flawlessly. To install Cosmovisor:

```bash
cd $HOME
git clone https://github.com/cosmos/cosmos-sdk
cd cosmos-sdk
git checkout v0.42.9
make cosmovisor
cp cosmovisor/cosmovisor $GOPATH/bin/cosmovisor
cd $HOME
```


Create the required directories:

```bash
mkdir -p ~/.osmosisd/cosmovisor
mkdir -p ~/.osmosisd/cosmovisor/genesis
mkdir -p ~/.osmosisd/cosmovisor/genesis/bin
mkdir -p ~/.osmosisd/cosmovisor/upgrades
```


Set the environment variables:

```bash
echo "# Setup Cosmovisor" >> ~/.profile
echo "export DAEMON_NAME=osmosisd" >> ~/.profile
echo "export DAEMON_HOME=$HOME/.osmosisd" >> ~/.profile
echo "export DAEMON_ALLOW_DOWNLOAD_BINARIES=false" >> ~/.profile
echo "export DAEMON_LOG_BUFFER_SIZE=512" >> ~/.profile
echo "export DAEMON_RESTART_AFTER_UPGRADE=true" >> ~/.profile
source ~/.profile
```

Download and replace the genesis file:

```bash
cd $HOME/.osmosisd/config
wget https://github.com/osmosis-labs/networks/raw/unity/v4/osmosis-1/upgrades/v4/testnet/genesis.tar.bz2
tar -xjf genesis.tar.bz2
```


Copy the current osmosisd binary into the cosmovisor/genesis folder:

```bash
cp $GOPATH/bin/osmosisd ~/.osmosisd/cosmovisor/genesis/bin
```

To check your work, ensure the version of cosmovisor and osmosisd are the same:

```bash
cosmovisor version
osmosisd version
```

These two command should both output 4.2.0

Reset private validator file to genesis state:

```bash
osmosisd unsafe-reset-all
```

## Download Chain Data

We must now download the latest chain data from a snapshot provider. In this example, I will use [the validator MP-20's latest testnet snapshot](https://mp20.net/snapshots/osmosis-testnet/) and I will use the pruned chain data.

Download liblz4-tool to handle the compressed file:

```bash
sudo apt-get install wget liblz4-tool aria2 -y
```

Download, decompress, and replace the chain data:

```bash
cd $HOME/.osmosisd/
wget -O - https://mp20.net/snapshots/osmosis-testnet/osmosis-testnet-mp20-latest.tar.xz | xz -d -v | tar xf - |
```

## Set Up Osmosis Service

While we could start cosmovisor now with "cosmovisor start", lets set up a service to allow cosmovisor to run in the background as well as restart automatically if it runs into any problems:

```bash
echo "[Unit]
Description=Cosmovisor daemon
After=network-online.target
[Service]
Environment="DAEMON_NAME=osmosisd"
Environment="DAEMON_HOME=${HOME}/.osmosisd"
Environment="DAEMON_RESTART_AFTER_UPGRADE=true"
Environment="DAEMON_ALLOW_DOWNLOAD_BINARIES=false"
Environment="DAEMON_LOG_BUFFER_SIZE=512"
User=$USER
ExecStart=${HOME}/go/bin/cosmovisor start
Restart=always
RestartSec=3
LimitNOFILE=4096
[Install]
WantedBy=multi-user.target
" >cosmovisor.service
```

Move this new file to the systemd directory:

```bash
sudo mv cosmovisor.service /lib/systemd/system/cosmovisor.service
```

## Start Osmosis Service

Reload and start the service:

```bash
sudo systemctl daemon-reload
sudo systemctl start cosmovisor
```

Check the status of your service:

```bash
sudo systemctl status cosmovisor
```

To see live logs of your service:

```bash
journalctl -u cosmovisor -f
``` 