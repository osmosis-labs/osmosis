# Joining Mainnet

## Install Osmosis Binary

Make sure you have [installed the Osmosis Binary (CLI).](../cli/install)

## Initialize Osmosis Node

Use osmosisd to initialize your node (replace the NODE_NAME with a name of your choosing):

```bash
osmosisd init NODE_NAME
```

Download and place the genesis file in the osmosis config folder:

```
wget -O ~/.osmosisd/config/genesis.json https://github.com/osmosis-labs/networks/raw/main/osmosis-1/genesis.json
```

## Set Up Cosmovisor

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

## Download Chain Data

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
wget -O - https://get.quicksync.io/$FILENAME | lz4 -d | tar -xvf -
wget https://raw.githubusercontent.com/chainlayer/quicksync-playbooks/master/roles/quicksync/files/checksum.sh
wget https://get.quicksync.io/$FILENAME.checksum
```

Compare the checksum with the onchain version:

```bash
curl -s https://api-osmosis.cosmostation.io/v1/tx/hash/`curl -s https://dl2.quicksync.io/$FILENAME.hash`|jq -r '.data.tx.body.memo'|sha512sum -c
```

The output should state "checksum: OK"

:::: tabs cache-lifetime="10" :options="{ useUrlFragment: false }"

::: tab Default id="first-tab"
``` bash
FILENAME=`curl https://quicksync.io/osmosis.json | jq -r --arg MODE "default" '.[] | select(.network=="default")|select (.mirror=="Netherlands")|.filename'`
cd $HOME/.osmosisd/
wget -O - https://dl2.quicksync.io/$FILENAME | lz4 -d | tar -xvf -
wget https://raw.githubusercontent.com/chainlayer/quicksync-playbooks/master/roles/quicksync/files/checksum.sh
wget https://dl2.quicksync.io/$FILENAME.checksum
```
:::


::: tab Pruned id="second-tab"
``` bash
FILENAME=`curl https://quicksync.io/osmosis.json | jq -r --arg MODE "pruned" '.[] | select(.network=="pruned")|select (.mirror=="Netherlands")|.filename'`
cd $HOME/.osmosisd/
wget -O - https://dl2.quicksync.io/$FILENAME | lz4 -d | tar -xvf -
wget https://raw.githubusercontent.com/chainlayer/quicksync-playbooks/master/roles/quicksync/files/checksum.sh
wget https://dl2.quicksync.io/$FILENAME.checksum
```
:::

::: tab Archive id="third-tab"
``` bash
FILENAME=`curl https://quicksync.io/osmosis.json | jq -r --arg MODE "archive" '.[] | select(.network=="archive")|select (.mirror=="Netherlands")|.filename'`
cd $HOME/.osmosisd/
wget -O - https://dl2.quicksync.io/$FILENAME | lz4 -d | tar -xvf -
wget https://raw.githubusercontent.com/chainlayer/quicksync-playbooks/master/roles/quicksync/files/checksum.sh
wget https://dl2.quicksync.io/$FILENAME.checksum
```
:::

::::




## Set Up Osmosis Service

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