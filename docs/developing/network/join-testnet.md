# Joining Testnet

Make sure you have [installed the Osmosis Binary (CLI).](../cli/install)

Use osmosisd to initialize your node (replace the NODE_NAME with a name of your choosing):

```bash
osmosisd init NODE_NAME
```

Download and place the genesis file in the osmosis config folder:

```
wget -O ~/.osmosisd/config/genesis.json https://github.com/osmosis-labs/networks/raw/main/osmosis-1/genesis.json
```

We will now set up cosmovisor to ensure any future upgrades happen flawlessly. To install Cosmovisor:

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

We must now download the latest chain data from a snapshot provider. In this example, I will use <a>https://quicksync.io/networks/osmosis.html</a> and I will use the pruned chain data. You may choose the default or archived based off your needs. 

Download liblz4-tool to handle the compressed file:

```bash
sudo apt-get install wget liblz4-tool aria2 -y
```

Before we download the chain data, we must first initialize the file name from quicksync. Hover over the download button from page linked above and you will see the file name (specifically the date and time). Replace the below with the timestamp listed for you:

EXAMPLE: If the download link is <a>https://get.quicksync.io/osmosis-1-pruned.20211119.0910.tar.lz4</a>, then

FILENAME=osmosis-1-pruned.20211119.0910.tar.lz4

```bash
FILENAME=osmosis-1-TYPE.DATE.TIME.tar.lz4
```

Download the chain data its corresponding checksum:

```bash
cd $HOME/.osmosisd/
aria2c -x5 https://get.quicksync.io/$FILENAME
OR (single thread but no double space needed on harddisk)
wget -O - https://get.quicksync.io/$FILENAME | lz4 -d | tar -xvf -
wget https://raw.githubusercontent.com/chainlayer/quicksync-playbooks/master/roles/quicksync/files/checksum.sh
wget https://get.quicksync.io/$FILENAME.checksum
```

Compare the checksum with the onchain version:

```bash
curl -s https://api-osmosis.cosmostation.io/v1/tx/hash/`curl -s https://get.quicksync.io/$FILENAME.hash`|jq -r '.data.tx.body.memo'|sha512sum -c
```

The output should state "checksum: OK"

You are now ready to start the Osmosis Daemon through cosmovisor. Lets set up a service to allow cosmovisor to run in the background as well as restart automatically if it runs into any problems:

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
Guide as of 17 November 2021.

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


Clone the osmosis repo, checkout and install v3.2.0_with_cache:

```bash
cd $HOME
git clone https://github.com/osmosis-labs/osmosis
cd osmosis
git checkout v3.2.0_with_cache
make install
```


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


Reset private validator file to genesis state:

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

The process should initialize and get to block 1122200, where it will automatically upgrade to v4.0.0-rc1


At around block height 1128853, when reading the logs you will see many notifications of "slashing and jailing validator". This is due to fact that many validators did not participate in the testnet and therefore get jailed at the same time (approx 30,000 blocks after the upgrade). In my experience, this may cause your node to reset due to a memory error. As long as you set up the service above, it will automatically reset and eventually get passed this difficult block. 

Guide as of 17 November 2021. 

## Relayers


