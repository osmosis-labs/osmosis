# v8 to v9 Upgrade Guide

Osmosis v9 Gov Prop: <https://www.mintscan.io/osmosis/proposals/252>

Countdown: <https://www.mintscan.io/osmosis/blocks/4707300>

Height: 4707300

## Go Requirement

You will need to be running go1.18 for this version of Osmosis. You can check if you are running go1.18 with the following command:

```{.sh}
go version
```

If this does not say go1.18, you need to upgrade/downgrade. One of the many ways to upgrade/downgrade to/from go 1.18 on linux is as follows:

```{.sh}
wget -q -O - https://git.io/vQhTU | bash -s -- --remove
wget -q -O - https://git.io/vQhTU | bash -s -- --version 1.18
```

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

## Install and setup Cosmovisor

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

Copy the current osmosisd binary into the
cosmovisor/genesis folder.

```{.sh}
cp $GOPATH/bin/osmosisd ~/.osmosisd/cosmovisor/genesis/bin
```

Cosmovisor is now ready to be started. We will now set up Cosmovisor for the upgrade

### Cosmovisor: Automatic Method

You can set up your cosmovisor to automatically pull the upgrade at the upgrade height on its own and seamlessly upgrade. To do so:

```{.sh}
echo "# Setup Cosmovisor" >> ~/.profile
echo "export DAEMON_NAME=osmosisd" >> ~/.profile
echo "export DAEMON_HOME=$HOME/.osmosisd" >> ~/.profile
echo "export DAEMON_ALLOW_DOWNLOAD_BINARIES=true" >> ~/.profile
echo "export DAEMON_LOG_BUFFER_SIZE=512" >> ~/.profile
echo "export DAEMON_RESTART_AFTER_UPGRADE=true" >> ~/.profile
echo "export UNSAFE_SKIP_BACKUP=true" >> ~/.profile
source ~/.profile
```

Now, at the upgrade height, Cosmovisor will do the rest


### Cosmovisor: Manual Method

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

Now, create the required folder, make the build, and copy the daemon over to that folder

```{.sh}
mkdir -p ~/.osmosisd/cosmovisor/upgrades/v10/bin
cd $HOME/osmosis
git pull
git checkout v9.0.0
make build
cp build/osmosisd ~/.osmosisd/cosmovisor/upgrades/v10/bin
```

Now, at the upgrade height, Cosmovisor will upgrade to the v9 binary

## Completely Manual Option

For those of you that like to do things completely manually:

1. Wait for Osmosis to reach the upgrade height (4707300)

2. Look for a panic message, followed by endless peer logs. Stop the daemon

3. Run the following commands:
```{.sh}
cd $HOME/osmosis
git pull
git checkout v9.0.0
make install
```

4. Start the osmosis daemon again, watch the upgrade happen, and then continue to hit blocks

## Further Help

If you need more help, please go to <https://docs.osmosis.zone> or join
our discord at <https://discord.gg/pAxjcFnAFH>.
