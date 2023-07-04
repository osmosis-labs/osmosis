#!/bin/bash

# install go 1.19
sudo rm -rvf /usr/local/go/
wget https://golang.org/dl/go1.19.3.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.19.3.linux-amd64.tar.gz
rm go1.19.3.linux-amd64.tar.gz
echo "Installed go 1.19"

# go setting
echo 'export GOROOT=/usr/local/go' >> ~/.profile
echo 'export GOPATH=$HOME/go' >> ~/.profile
echo 'export GO111MODULE=on' >> ~/.profile
echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.profile
source ~/.profile

# install cosmovisor
go install github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@v1.0.0
echo "Installed cosmovisor"

sudo apt-get update
sudo apt-get install -y build-essential

# install v15.2.0 binary
rm -rf ~/osmosis
rm -rf ~/.osmosisd
git clone https://github.com/osmosis-labs/osmosis osmosis
cd osmosis
git checkout v15.2.0
make install
echo "Installed v15.2.0 binary"

# download genesis
osmosisd init snapshotsync --chain-id osmosis-1
wget -O genesis.json https://snapshots.polkachu.com/genesis/osmosis/genesis.json --inet4-only
mv genesis.json ~/.osmosisd/config/genesis.json
echo "Downleaded genesis"

# change seed nodes
sed -i 's/seeds = ""/seeds = "ade4d8bc8cbe014af6ebdf3cb7b1e9ad36f412c0@seeds.polkachu.com:12556"/' ~/.osmosisd/config/config.toml
echo "Changed seed nodes"

# Create Cosmovisor Folders
mkdir -p ~/.osmosisd/cosmovisor/genesis/bin
mkdir -p ~/.osmosisd/cosmovisor/upgrades
echo "Created Cosmovisor Folders"

# Load Node Binary into Cosmovisor Folder
cp ~/go/bin/osmosisd ~/.osmosisd/cosmovisor/genesis/bin
echo "Lode Node Binary into cosmovisor folder"

# Get the current username
username=$(whoami)

# Create the osmosis.service file in a temporary location
cat > osmosis.service <<EOL
[Unit]
Description="osmosis node"
After=network-online.target

[Service]
User=$username
ExecStart=/home/$username/go/bin/cosmovisor start
Restart=always
RestartSec=3
LimitNOFILE=4096
Environment="DAEMON_NAME=osmosisd"
Environment="DAEMON_HOME=/home/$username/.osmosisd"
Environment="DAEMON_ALLOW_DOWNLOAD_BINARIES=false"
Environment="DAEMON_RESTART_AFTER_UPGRADE=true"
Environment="UNSAFE_SKIP_BACKUP=true"

[Install]
WantedBy=multi-user.target
EOL

# Move the osmosis.service file to the /etc/systemd/system folder with sudo privileges
sudo mv osmosis.service /etc/systemd/system/

# Reload systemd configuration with sudo privileges
sudo systemctl daemon-reload

echo "osmosis.service file created successfully in /etc/systemd/system."

# Enable service
sudo systemctl enable osmosis.service

# Start service
sudo service osmosis start

# Check logs
sudo journalctl -fu osmosis