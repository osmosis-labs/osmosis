# Mainnet Upgrade Guide: From Version v18 to v19

## Overview

- **v19 Proposal**: [Proposal Page](https://www.mintscan.io/osmosis/proposals/606)
- **v19 Upgrade Block Height**: `11317300`
- **v19 Upgrade Countdown**: [Block Countdown](https://www.mintscan.io/osmosis/blocks/11317300)

## Hardware Requirements

### Memory Specifications

Although this upgrade is not expected to be resource-intensive, a minimum of 64GB of RAM is advised. If you cannot meet this requirement, setting up a swap space is recommended.

#### Configuring Swap Space

*Execute these commands to set up a 32GB swap space*:

```sh
sudo swapoff -a
sudo fallocate -l 32G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
```

*To ensure the swap space persists after reboot*:

```sh
sudo cp /etc/fstab /etc/fstab.bak
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
```

For an in-depth guide on swap configuration, please refer to [this tutorial](https://www.digitalocean.com/community/tutorials/how-to-add-swap-space-on-ubuntu-20-04).

---

## Cosmovisor Configuration

### Initial Setup (For First-Time Users)

If you have not previously configured Cosmovisor, follow this section; otherwise, proceed to the next section.

Cosmovisor is strongly recommended for validators to minimize downtime during upgrades. It automates the binary replacement process according to on-chain `SoftwareUpgrade` proposals.

Documentation for Cosmovisor can be found [here](https://docs.cosmos.network/main/tooling/cosmovisor).

#### Installation Steps

*Run these commands to install and configure Cosmovisor*:

```sh
go install github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@v1.0.0
mkdir -p ~/.osmosisd
mkdir -p ~/.osmosisd/cosmovisor
mkdir -p ~/.osmosisd/cosmovisor/genesis
mkdir -p ~/.osmosisd/cosmovisor/genesis/bin
mkdir -p ~/.osmosisd/cosmovisor/upgrades
cp $GOPATH/bin/osmosisd ~/.osmosisd/cosmovisor/genesis/bin
mkdir -p ~/.osmosisd/cosmovisor/upgrades/v18/bin
cp $GOPATH/bin/osmosisd ~/.osmosisd/cosmovisor/upgrades/v18/bin
```

*Add these lines to your profile to set up environment variables*:

```sh
echo "# Cosmovisor Setup" >> ~/.profile
echo "export DAEMON_NAME=osmosisd" >> ~/.profile
echo "export DAEMON_HOME=$HOME/.osmosisd" >> ~/.profile
echo "export DAEMON_ALLOW_DOWNLOAD_BINARIES=false" >> ~/.profile
echo "export DAEMON_LOG_BUFFER_SIZE=512" >> ~/.profile
echo "export DAEMON_RESTART_AFTER_UPGRADE=true" >> ~/.profile
echo "export UNSAFE_SKIP_BACKUP=true" >> ~/.profile
source ~/.profile
```

### Upgrading to v19

*To prepare for the upgrade, execute these commands*:

```sh
mkdir -p ~/.osmosisd/cosmovisor/upgrades/v19/bin
cd $HOME/osmosis
git pull
git checkout v19.2.0
make build
cp build/osmosisd ~/.osmosisd/cosmovisor/upgrades/v19/bin
```

At the designated block height, Cosmovisor will automatically upgrade to version v19.

---

## Manual Upgrade Procedure

Follow these steps if you opt for a manual upgrade:

1. Monitor Osmosis until it reaches the specified upgrade block height: 11317300.
2. Observe for a panic message followed by continuous peer logs, then halt the daemon.
3. Perform these steps:

```sh
cd $HOME/osmosis
git pull
git checkout v19.2.0
make install
```

4. Restart the Osmosis daemon and observe the upgrade.

---

## Additional Resources

- Osmosis Documentation: [Website](https://docs.osmosis.zone)
- Community Support: [Discord](https://discord.gg/pAxjcFnAFH)

