# v16 to v17 Testnet Upgrade Guide

|                 |                                                              |
|-----------------|--------------------------------------------------------------|
| Chain-id        | `osmo-test-5`                                                |
| Upgrade Version | `v17.0.0-rc0`                                                |
| Upgrade Height  | 2200500                                                      |
| Countdown       | <https://www.mintscan.io/osmosis-testnet/block/2200500> |

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
<https://docs.cosmos.network/main/tooling/cosmovisor>

If you choose to use cosmovisor, please continue with these
instructions:

To install Cosmovisor:

``` {.sh}
go install github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@v1.0.0
```

After this, you must make the necessary folders for cosmovisor in your
daemon home directory (\~/.osmosisd).

``` {.sh}
mkdir -p ~/.osmosisd
mkdir -p ~/.osmosisd/cosmovisor
mkdir -p ~/.osmosisd/cosmovisor/genesis
mkdir -p ~/.osmosisd/cosmovisor/genesis/bin
mkdir -p ~/.osmosisd/cosmovisor/upgrades
```

Copy the current v16 osmosisd binary into the
cosmovisor/genesis folder and v16 folder.

```{.sh}
cp $GOPATH/bin/osmosisd ~/.osmosisd/cosmovisor/genesis/bin
mkdir -p ~/.osmosisd/cosmovisor/upgrades/v16/bin
cp $GOPATH/bin/osmosisd ~/.osmosisd/cosmovisor/upgrades/v16/bin
```

Cosmovisor is now ready to be set up for v17.

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

Create the v17 folder, make the build, and copy the daemon over to that folder

```{.sh}
mkdir -p ~/.osmosisd/cosmovisor/upgrades/v17/bin
cd $HOME/osmosis
git pull
git checkout v17.0.0-rc0
make build
cp build/osmosisd ~/.osmosisd/cosmovisor/upgrades/v17/bin
```

Now, at the upgrade height, Cosmovisor will upgrade to the v17 binary

## Manual Option

1. Wait for Osmosis to reach the upgrade height (2200500)

2. Look for a panic message, followed by endless peer logs. Stop the daemon

3. Run the following commands:

```{.sh}
cd $HOME/osmosis
git pull
git checkout v17.0.0-rc0
make install
```

4. Start the osmosis daemon again, watch the upgrade happen, and then continue to hit blocks

## Further Help

If you need more help, please:
- go to <https://docs.osmosis.zone>
- join our discord at <https://discord.gg/pAxjcFnAFH>.
