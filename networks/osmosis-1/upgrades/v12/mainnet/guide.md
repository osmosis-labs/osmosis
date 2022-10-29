# v11 to v12 Mainnet Upgrade Guide

Osmosis v12 Gov Prop: <https://www.mintscan.io/osmosis/proposals/335>

Countdown: <https://www.mintscan.io/osmosis/blocks/6246000>

Height: 6246000

## Memory Requirements

This upgrade will **not** be resource intensive. With that being said, we still recommend having 64GB of memory. If having 64GB of physical memory is not possible, the next best thing is to set up swap.

Short version swap setup instructions:

``` {.sh}
sudo swapoff -a
sudo fallocate -l 32G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
```

To persist swap after restart:

``` {.sh}
sudo cp /etc/fstab /etc/fstab.bak
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
```

In depth swap setup instructions:
<https://www.digitalocean.com/community/tutorials/how-to-add-swap-space-on-ubuntu-20-04>

## First Time Cosmovisor Setup

If you have never setup Cosmovisor before, follow the following instructions.

If you have already setup Cosmovisor, skip to the next section.

We highly recommend validators use cosmovisor to run their nodes. This
will make low-downtime upgrades smoother, as validators don't have to
manually upgrade binaries during the upgrade, and instead can
pre-install new binaries, and cosmovisor will automatically update them
based on on-chain SoftwareUpgrade proposals.

You should review the docs for cosmovisor located here:
<https://docs.cosmos.network/master/run-node/cosmovisor.html>

If you choose to use cosmovisor, please continue with these
instructions:

To install Cosmovisor:

``` {.sh}
go install github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@v1.0.0
```

After this, you must make the necessary folders for cosmosvisor in your
daemon home directory (\~/.osmosisd).

``` {.sh}
mkdir -p ~/.osmosisd
mkdir -p ~/.osmosisd/cosmovisor
mkdir -p ~/.osmosisd/cosmovisor/genesis
mkdir -p ~/.osmosisd/cosmovisor/genesis/bin
mkdir -p ~/.osmosisd/cosmovisor/upgrades
```

Copy the current v11 osmosisd binary into the
cosmovisor/genesis folder and v11 folder.

```{.sh}
cp $GOPATH/bin/osmosisd ~/.osmosisd/cosmovisor/genesis/bin
mkdir -p ~/.osmosisd/cosmovisor/upgrades/v11/bin
cp $GOPATH/bin/osmosisd ~/.osmosisd/cosmovisor/upgrades/v11/bin
```

Cosmovisor is now ready to be set up for v12.

Set these environment variables:

```{.sh}
echo "# Setup Cosmovisor" >> ~/.profile
echo "export DAEMON_NAME=osmosisd" >> ~/.profile
echo "export DAEMON_HOME=$HOME/.osmosisd" >> ~/.profile
echo "export DAEMON_ALLOW_DOWNLOAD_BINARIES=false" >> ~/.profile
echo "export DAEMON_LOG_BUFFER_SIZE=512" >> ~/.profile
echo "export DAEMON_RESTART_AFTER_UPGRADE=true" >> ~/.profile
echo "export UNSAFE_SKIP_BACKUP=true" >> ~/.profile
source ~/.profile
```

## Cosmovisor Upgrade

Create the v12 folder, make the build, and copy the daemon over to that folder

```{.sh}
mkdir -p ~/.osmosisd/cosmovisor/upgrades/v12/bin
cd $HOME/osmosis
git pull
git checkout v12.0.0
make build
cp build/osmosisd ~/.osmosisd/cosmovisor/upgrades/v12/bin
```

Now, at the upgrade height, Cosmovisor will upgrade to the v12 binary

## Manual Option

1. Wait for Osmosis to reach the upgrade height (FILL ME IN WITH UPGRADE HEIGHT)

2. Look for a panic message, followed by endless peer logs. Stop the daemon

3. Run the following commands:

```{.sh}
cd $HOME/osmosis
git pull
git checkout v12.0.0
make install
```

4. Start the osmosis daemon again, watch the upgrade happen, and then continue to hit blocks

## Further Help

If you need more help, please go to <https://docs.osmosis.zone> or join
our discord at <https://discord.gg/pAxjcFnAFH>.
