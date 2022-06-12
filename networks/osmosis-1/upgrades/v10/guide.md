# v9 to v10 Upgrade Guide

All validator nodes should upgrade to v10 prior to the network restarting. At 4:00PM UTC on June 12th, 2022, we will have a coordinated re-start of the network. The sequence of events will look like to following:

* All validator nodes upgrade to v10 now, but keep their nodes offline. Even if your node is further behind (i.e. you stopped your node first when we corrdinated the shutdown and still have blocks ahead of you before reaching the halt height), you still must upgrade to v10 now.
* At exactly 4:00PM UTC on June 12th, 2022, all validators start their nodes at the same time
* Once 66% or more of the voting power gets online, block 4713065 will be reached, along with the upgrade at this height. Prior to 66 percent of validator power getting online, you will only see p2p logs. This is also an epoch block, so it will take some time to process
* After block 4713065, three more epochs will happen back to back, one per block 
* If the June 12th epoch time has not occured yet, blocks will be produced until the epoch time. If the epoch time has occured, the June 12th epoch will occur in conjunction with the four other epochs above.

The coordination of restart will happen over Discord. In the event Discord is down, we will reach out with a Telegram link over Twitter to further coordinate the network restart.


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

As always, we recommend having 64GB of memory. If having 64GB of physical memory is not possible, the next best thing is to set up swap.

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

Now, create the required folder, make the build, and copy the daemon over to that folder. **NOTE**, you must put the v10 binary in the v9 folder as shown below since this is a fork.

```{.sh}
mkdir -p ~/.osmosisd/cosmovisor/upgrades/v9/bin
cd $HOME/osmosis
git pull
git checkout v10.0.0
make build
cp build/osmosisd ~/.osmosisd/cosmovisor/upgrades/v9/bin
```

## Completely Manual Option

```{.sh}
cd $HOME/osmosis
git pull
git checkout v10.0.0
make install
```

## Further Help

If you need more help, please go to <https://docs.osmosis.zone> or join our discord at <https://discord.gg/pAxjcFnAFH>.
