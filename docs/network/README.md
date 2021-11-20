# Network
## Quickstart

## Joining Mainnet

## Joining Testnet

Start by ensuring your system is up to date:

```bash
sudo apt update
sudo apt upgrade
```

Install tools if not already installed (git, gcc, make, etc.):

```bash
sudo apt install git build-essential ufw curl jq snapd --yes
```

Install go:

```bash
wget -q -O - https://git.io/vQhTU | bash -s -- --version 1.17.2
```

After installed, open new terminal to properly load go

Run the following to clone the osmosis repo:

```bash
cd $HOME
git clone https://github.com/osmosis-labs/osmosis
cd osmosis
git checkout v3.2.0_with_cache
make install
```

You have now installed the Osmosis Daemon (osmosisd). Use osmosisd to initialize your node with the following (replace the NODE_NAME with a name of your choosing):

```bash
osmosisd init NODE_NAME --chain-id=osmosis-testnet-0
```

We now need to open the config.toml to edit the seed list with the following:

```bash
cd $HOME/.osmosisd/config
nano config.toml
```

Use page down or arrow keys to get to the line that says seeds = "" and replace it with the following:

```bash
seeds = "4eaed17781cd948149098d55f80a28232a365236@testmosis.blockpane.com:26656"
```

Then pres Ctrl+O, then enter to save, then Ctrl+X to exit

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

After this, create the required directories with the following commands:

```bash
mkdir -p ~/.osmosisd/cosmovisor
mkdir -p ~/.osmosisd/cosmovisor/genesis
mkdir -p ~/.osmosisd/cosmovisor/genesis/bin
mkdir -p ~/.osmosisd/cosmovisor/upgrades
```

Next, run the following commands to set the environment variables:

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

Next, copy the current osmosisd binary into the cosmovisor/genesis folder:

```bash
cp $GOPATH/bin/osmosisd ~/.osmosisd/cosmovisor/genesis/bin
```

To check your work, ensure the version of cosmovisor and osmosisd are the same:

```bash
cosmovisor version
osmosisd version
```

These two command should both output 3.1.0-23-g517562d

Before we start cosmovisor, lets prep the upgrade to v4.0.0-rc1:

```bash
mkdir -p ~/.osmosisd/cosmovisor/upgrades/v4/bin
cd $HOME/osmosis
git checkout v4.0.0-rc1
make build
cp build/osmosisd ~/.osmosisd/cosmovisor/upgrades/v4/bin
cd $HOME
```

Use the following command to ensure the validator file is in the gensesis state:

```bash
osmosisd unsafe-reset-all
```

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

Finally, reload and start the service:

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

The process should initialize and get to block 1122200, where it will automatically upgrade to v4.0.0-rc1


At around block height 1128853, when reading the logs you will see many notifications of "slashing and jailing validator". This is due to fact that many validators did not participate in the testnet and therefore get jailed at the same time (approx 30,000 blocks after the upgrade). In my experience, this may cause your node to reset due to a memory error. As long as you set up the service above, it will automatically reset and eventually get passed this difficult block. 

Guide as of 17 November 2021. 

## Relayers


